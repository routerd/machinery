/*
routerd
Copyright (C) 2020  The routerd Authors

This program is free software: you can redistribute it and/or modify
it under the terms of the GNU Affero General Public License as published by
the Free Software Foundation, either version 3 of the License, or
(at your option) any later version.

This program is distributed in the hope that it will be useful,
but WITHOUT ANY WARRANTY; without even the implied warranty of
MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
GNU Affero General Public License for more details.

You should have received a copy of the GNU Affero General Public License
along with this program.  If not, see <http://www.gnu.org/licenses/>.
*/

package boltdb

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"

	"github.com/google/uuid"
	bolt "go.etcd.io/bbolt"

	"routerd.net/machinery/errors"
	"routerd.net/machinery/runtime"
	storagev1 "routerd.net/machinery/storage/api/v1"
	"routerd.net/machinery/storage/event"
)

// defaulted may be implement
type defaulted interface {
	Default() error
}

type ErrBucketNotFound string

func (e ErrBucketNotFound) Error() string {
	return fmt.Sprintf("bucket %s not found", string(e))
}

// GroupVersionKind key
// this special key remembers the GVK of all objects in the bucket.
const gvkKey = "__GVK__"

var (
	_ storagev1.Client = (*BoltDBStorage)(nil)
)

type BoltDBStorage struct {
	db     *bolt.DB
	scheme *runtime.Scheme

	hub             *event.Hub
	objGVK, listGVK runtime.GroupVersionKind
	bucket          string
}

func NewBoltDBStorage(
	scheme *runtime.Scheme, obj runtime.Object, db *bolt.DB) (*BoltDBStorage, error) {
	objGVK, err := scheme.GroupVersionKind(obj)
	if err != nil {
		return nil, err
	}

	listGVK, err := scheme.ListGroupVersionKind(obj)
	if err != nil {
		return nil, err
	}

	s := &BoltDBStorage{
		db:     db,
		scheme: scheme,

		objGVK:  objGVK,
		listGVK: listGVK,
		bucket:  objGVK.GroupKind().String(),
	}
	s.hub = event.NewHub(100, s.list)

	return s, s.ensureStorageVersion()
}

func (s *BoltDBStorage) Run(stopCh <-chan struct{}) {
	s.hub.Run(stopCh)
}

func (s *BoltDBStorage) Watch(ctx context.Context, obj storagev1.Object, opts ...storagev1.ListOption) (storagev1.WatchClient, error) {
	if err := s.ensureObjectGKV(obj); err != nil {
		return nil, err
	}
	return s.hub.Register(obj.GetResourceVersion(), opts...), nil
}

// Get a single object from the underlying storage.
// Returns ErrNotFound if no item with the given key exists.
func (s *BoltDBStorage) Get(ctx context.Context, key storagev1.NamespacedName, obj storagev1.Object) error {
	if err := s.ensureObjectGKV(obj); err != nil {
		return err
	}

	var b []byte
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.bucket))
		if bucket == nil {
			return ErrBucketNotFound(string(s.bucket))
		}

		b = bucket.Get([]byte(key.String()))
		if b == nil {
			return errors.ErrNotFound{
				Key: key.String(),
				GVK: s.objGVK,
			}
		}

		return nil
	})
	if err != nil {
		return err
	}

	return json.Unmarshal(b, obj)
}

func (s *BoltDBStorage) list(opts ...storagev1.ListOption) ([]runtime.Object, error) {
	var options storagev1.ListOptions
	for _, opt := range opts {
		opt.ApplyToList(&options)
	}

	var out []runtime.Object
	err := s.db.View(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.bucket))
		if bucket == nil {
			return ErrBucketNotFound(s.bucket)
		}

		return bucket.ForEach(func(k, v []byte) error {
			// skip key containing the type information
			if bytes.Equal(k, []byte(gvkKey)) {
				return nil
			}

			if options.Namespace != "" {
				if !bytes.HasSuffix(k, []byte("."+options.Namespace)) {
					return nil
				}
			}

			obj, err := s.scheme.New(s.objGVK)
			if err != nil {
				return err
			}
			if err := json.Unmarshal(v, obj.(storagev1.Object)); err != nil {
				return err
			}
			out = append(out, obj)
			return nil
		})
	})
	if err != nil {
		return nil, err
	}
	return out, nil
}

// List returns all entries, matching the given ListOptions.
func (s *BoltDBStorage) List(ctx context.Context, listObj storagev1.ListObject, opts ...storagev1.ListOption) error {
	if err := s.ensureListObjectGKV(listObj); err != nil {
		return err
	}

	var options storagev1.ListOptions
	for _, opt := range opts {
		opt.ApplyToList(&options)
	}

	rv := reflect.ValueOf(listObj).Elem()

	objects, err := s.list(opts...)
	if err != nil {
		return err
	}
	for _, obj := range objects {
		rv.FieldByName("Items").Set(
			reflect.Append(rv.FieldByName("Items"), reflect.ValueOf(obj).Elem()),
		)
	}
	return nil
}

// Create persists the given Object to
func (s *BoltDBStorage) Create(ctx context.Context, obj storagev1.Object, opts ...storagev1.CreateOption) error {
	if err := s.ensureObjectGKV(obj); err != nil {
		return err
	}
	if err := s.validateNameNamespace(obj); err != nil {
		return err
	}

	// Defaulting
	if d, ok := obj.(defaulted); ok {
		if err := d.Default(); err != nil {
			return fmt.Errorf("calling Default(): %w", err)
		}
	}

	key := storagev1.Key(obj).String()

	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.bucket))
		if bucket == nil {
			return ErrBucketNotFound(string(s.bucket))
		}

		if v := bucket.Get([]byte(key)); v != nil {
			return errors.ErrAlreadyExists{
				Key: key,
				GVK: s.objGVK,
			}
		}

		// Set Generation/ResourceVersion
		rv, err := bucket.NextSequence()
		if err != nil {
			return err
		}
		obj.GetObjectKind().SetGroupVersionKind(s.objGVK)
		obj.SetGeneration(1)
		obj.SetResourceVersion(strconv.FormatUint(rv, 10))
		obj.SetUID(uuid.New().String())

		// Put into Storage
		b, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		if err := bucket.Put([]byte(key), b); err != nil {
			return err
		}
		s.hub.Broadcast(storagev1.Added, obj)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// Update updates metadata and the Objects spec, while ignoring changes to the Objects Status.
func (s *BoltDBStorage) Update(ctx context.Context, obj storagev1.Object, opts ...storagev1.UpdateOption) error {
	if err := s.ensureObjectGKV(obj); err != nil {
		return err
	}
	if err := s.validateNameNamespace(obj); err != nil {
		return err
	}

	// Defaulting
	if d, ok := obj.(defaulted); ok {
		if err := d.Default(); err != nil {
			return fmt.Errorf("calling Default(): %w", err)
		}
	}

	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.bucket))
		if bucket == nil {
			return ErrBucketNotFound(string(s.bucket))
		}

		// Get existing
		key := storagev1.Key(obj).String()
		existing, err := s.loadNew(bucket, key)
		if err != nil {
			return err
		}

		// Ensure Status is not updated, if the field exists
		statusField := reflect.ValueOf(obj).Elem().FieldByName("Status")
		if statusField.IsValid() {
			statusField.Set(
				reflect.ValueOf(existing).Elem().FieldByName("Status"),
			)
		}

		// Check if there is a change
		if reflect.DeepEqual(existing, obj) {
			return nil
		}

		// Check ResourceVersion
		if existing.GetResourceVersion() != obj.GetResourceVersion() {
			return errors.ErrConflict{Key: key, GVK: s.objGVK}
		}

		// Set Generation/ResourceVersion
		rv, err := bucket.NextSequence()
		if err != nil {
			return err
		}
		obj.SetGeneration(obj.GetGeneration() + 1)
		obj.SetResourceVersion(strconv.FormatUint(rv, 10))

		// Put into Storage
		b, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		if err := bucket.Put([]byte(key), b); err != nil {
			return err
		}
		s.hub.Broadcast(storagev1.Modified, obj)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

// UpdateStatus ignores metadata and the Objects spec and only persists status changes.
func (s *BoltDBStorage) UpdateStatus(ctx context.Context, obj storagev1.Object, opts ...storagev1.UpdateOption) error {
	if err := s.ensureObjectGKV(obj); err != nil {
		return err
	}
	if err := s.validateNameNamespace(obj); err != nil {
		return err
	}

	// Defaulting
	if d, ok := obj.(defaulted); ok {
		if err := d.Default(); err != nil {
			return fmt.Errorf("calling Default(): %w", err)
		}
	}

	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.bucket))
		if bucket == nil {
			return ErrBucketNotFound(string(s.bucket))
		}

		// Get existing
		key := storagev1.Key(obj).String()
		existing, err := s.loadNew(bucket, key)
		if err != nil {
			return err
		}

		// Ensure ObjectMeta and Spec is not updated
		reflect.ValueOf(obj).Elem().FieldByName("ObjectMeta").Set(
			reflect.ValueOf(existing).Elem().FieldByName("ObjectMeta"),
		)
		specField := reflect.ValueOf(obj).Elem().FieldByName("Spec")
		if specField.IsValid() {
			specField.Set(
				reflect.ValueOf(existing).Elem().FieldByName("Spec"),
			)
		}

		// Check if there is a change
		if reflect.DeepEqual(existing, obj) {
			return nil
		}

		// Check ResourceVersion
		if existing.GetResourceVersion() != obj.GetResourceVersion() {
			return errors.ErrConflict{Key: key, GVK: s.objGVK}
		}

		// Set Generation/ResourceVersion
		rv, err := bucket.NextSequence()
		if err != nil {
			return err
		}
		obj.SetResourceVersion(strconv.FormatUint(rv, 10))

		// Put into Storage
		b, err := json.Marshal(obj)
		if err != nil {
			return err
		}
		if err := bucket.Put([]byte(key), b); err != nil {
			return err
		}
		s.hub.Broadcast(storagev1.Modified, obj)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *BoltDBStorage) Delete(ctx context.Context, obj storagev1.Object, opts ...storagev1.DeleteOption) error {
	if err := s.ensureObjectGKV(obj); err != nil {
		return err
	}
	if err := s.validateNameNamespace(obj); err != nil {
		return err
	}

	err := s.db.Update(func(tx *bolt.Tx) error {
		bucket := tx.Bucket([]byte(s.bucket))
		if bucket == nil {
			return ErrBucketNotFound(string(s.bucket))
		}

		// Get existing
		key := storagev1.Key(obj).String()
		existing, err := s.loadNew(bucket, key)
		if err != nil {
			return err
		}

		// Delete
		if err := bucket.Delete([]byte(key)); err != nil {
			return err
		}
		s.hub.Broadcast(storagev1.Deleted, existing)
		return nil
	})
	if err != nil {
		return err
	}
	return nil
}

func (s *BoltDBStorage) loadNew(bucket *bolt.Bucket, key string) (storagev1.Object, error) {
	v := bucket.Get([]byte(key))
	if v == nil {
		return nil, errors.ErrNotFound{Key: key, GVK: s.objGVK}
	}

	runtimeObj, err := s.scheme.New(s.objGVK)
	if err != nil {
		return nil, err
	}
	obj := runtimeObj.(storagev1.Object)
	if err := json.Unmarshal(v, obj); err != nil {
		return nil, err
	}
	return obj, nil
}

// ensures all objects in storage are in the desired storage version.
// performs object conversion via schema if needed.
func (s *BoltDBStorage) ensureStorageVersion() error {
	return s.db.Update(func(tx *bolt.Tx) error {
		bucket, err := tx.CreateBucketIfNotExists([]byte(s.bucket))
		if err != nil {
			return err
		}

		gvkBytes := bucket.Get([]byte(gvkKey))
		if gvkBytes == nil {
			// no GVK value set, bucket was just created
			gvkBytes, err = json.Marshal(s.objGVK)
			if err != nil {
				return err
			}
			return bucket.Put([]byte(gvkKey), gvkBytes)
		}

		var storedGVK runtime.GroupVersionKind
		if err := json.Unmarshal(gvkBytes, &storedGVK); err != nil {
			return fmt.Errorf("unmarshal stored GVK: %w", err)
		}
		if storedGVK == s.objGVK {
			return nil
		}

		c := bucket.Cursor()
		for k, _ := c.First(); k != nil; k, _ = c.Next() {
			if bytes.Equal(k, []byte(gvkKey)) {
				continue
			}

			storageObj, err := s.loadNew(bucket, string(k))
			if err != nil {
				return err
			}
			if storageObj.GetObjectKind().GetGroupVersionKind() == s.objGVK {
				// object is of right version
				// -> nothing todo
				continue
			}

			// object needs to be converted
			convertedObj, err := s.scheme.New(s.objGVK)
			if err != nil {
				return err
			}
			if err := s.scheme.Convert(storageObj, convertedObj); err != nil {
				return err
			}
		}

		// update GVK
		gvkBytes, err = json.Marshal(s.objGVK)
		if err != nil {
			return err
		}
		return bucket.Put([]byte(gvkKey), gvkBytes)
	})
}

// Ensure that the object type is of the GKV this repository is persisting.
func (s *BoltDBStorage) ensureObjectGKV(obj storagev1.Object) error {
	objGVK, err := s.scheme.GroupVersionKind(obj)
	if err != nil {
		return err
	}

	if objGVK != s.objGVK {
		return fmt.Errorf("invalid GVK, expected: %s", s.objGVK)
	}
	return nil
}

// Ensure that the list object type is of the GKV this repository is persisting.
func (s *BoltDBStorage) ensureListObjectGKV(obj storagev1.ListObject) error {
	listGVK, err := s.scheme.ListGroupVersionKind(obj)
	if err != nil {
		return err
	}

	if listGVK != s.listGVK {
		return fmt.Errorf("invalid GVK, expected: %s", s.listGVK)
	}
	return nil
}

func (s *BoltDBStorage) validateNameNamespace(obj storagev1.Object) error {
	if err := storagev1.ValidateName(obj.GetName()); err != nil {
		return err
	}
	if err := storagev1.ValidateNamespace(obj.GetNamespace()); err != nil {
		return err
	}
	return nil
}
