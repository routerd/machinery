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

package storage

import (
	"context"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/types/known/timestamppb"
	"k8s.io/apimachinery/pkg/labels"

	"routerd.net/machinery/api"
	"routerd.net/machinery/errors"
	"routerd.net/machinery/validate"
)

type InMemoryStorage struct {
	hub                *eventHub
	objectFullName     protoreflect.FullName
	listObjectFullName protoreflect.FullName
	newObject          func() api.Object
	mux                sync.RWMutex
	data               map[string][]byte
	sequence           uint64
}

var _ Storage = (*InMemoryStorage)(nil)

func NewInMemoryStorage(objType api.Object) *InMemoryStorage {
	objName := objType.ProtoReflect().Descriptor().FullName()
	listObjName := protoreflect.FullName(objName + "List")

	s := &InMemoryStorage{
		objectFullName:     objName,
		listObjectFullName: listObjName,
		newObject: func() api.Object {
			return objType.ProtoReflect().
				New().Interface().(api.Object)
		},
		data: map[string][]byte{},
	}
	s.hub = NewHub(s.list)
	return s
}

func (s *InMemoryStorage) Run(stopCh <-chan struct{}) {
	s.hub.Run(stopCh)
}

func (s *InMemoryStorage) Get(ctx context.Context, nn api.NamespacedName, obj api.Object) error {
	if err := s.checkObject(obj); err != nil {
		return err
	}
	if err := validate.ValidateNamespacedName(nn); err != nil {
		return err
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.load(nn, obj)
}

func (s *InMemoryStorage) List(ctx context.Context, listObj api.ListObject, opts ...ListOption) error {
	if err := s.checkListObject(listObj); err != nil {
		return err
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	objects, err := s.list(opts...)
	if err != nil {
		return err
	}

	rv := reflect.ValueOf(listObj).Elem()
	for _, obj := range objects {
		rv.FieldByName("Items").Set(
			reflect.Append(rv.FieldByName("Items"), reflect.ValueOf(obj)),
		)
	}
	return nil
}

func (s *InMemoryStorage) Watch(ctx context.Context, obj api.Object, opts ...ListOption) (WatchClient, error) {
	if err := s.checkObject(obj); err != nil {
		return nil, err
	}

	return s.hub.Register(
		obj.ObjectMeta().GetResourceVersion(), opts...)
}

func (s *InMemoryStorage) Create(ctx context.Context, obj api.Object, opts ...CreateOption) error {
	// Input validation
	if err := s.checkObject(obj); err != nil {
		return err
	}

	// Defaulting
	if err := Default(obj); err != nil {
		return err
	}
	// Validation
	if err := ValidateCreate(obj); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	meta := obj.ObjectMeta()
	key := api.NamespacedName{Name: meta.GetName(), Namespace: meta.GetNamespace()}
	if _, ok := s.data[key.String()]; ok { // Already Exists?
		return errors.ErrAlreadyExists{
			NamespacedName: key,
			TypeFullName:   string(obj.ProtoReflect().Descriptor().FullName()),
		}
	}

	// Ensure correct metadata
	s.sequence++
	meta.SetGeneration(1)
	meta.SetResourceVersion(strconv.FormatUint(s.sequence, 10))
	meta.SetUid(uuid.New().String())
	meta.SetCreatedTimestamp(timestamppb.Now())

	// Store
	if err := s.store(obj); err != nil {
		return err
	}
	s.hub.Broadcast(api.Added, obj)
	return nil
}

func (s *InMemoryStorage) Delete(ctx context.Context, obj api.Object, opts ...DeleteOption) error {
	// Input validation
	if err := s.checkObject(obj); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	return s.delete(ctx, obj, opts...)
}

func (s *InMemoryStorage) delete(ctx context.Context, obj api.Object, opts ...DeleteOption) error {
	// Defaulting
	// because we don't know what the Validation next is expecting
	if err := Default(obj); err != nil {
		return err
	}
	// Validation
	if err := ValidateDelete(obj); err != nil {
		return err
	}

	// Finalizer Handling
	if len(obj.ObjectMeta().GetFinalizers()) != 0 {
		obj.ObjectMeta().SetDeletedTimestamp(timestamppb.Now())
		return s.update(ctx, obj)
	}

	// Load Existing
	meta := obj.ObjectMeta()
	key := api.NamespacedName{Name: meta.GetName(), Namespace: meta.GetNamespace()}
	if err := s.load(key, obj); err != nil {
		return err
	}

	// Delete
	delete(s.data, key.String())
	s.hub.Broadcast(api.Deleted, obj)
	return nil
}

func (s *InMemoryStorage) Update(ctx context.Context, obj api.Object, opts ...UpdateOption) error {
	// Input validation
	if err := s.checkObject(obj); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()
	return s.update(ctx, obj, opts...)
}

func (s *InMemoryStorage) update(ctx context.Context, obj api.Object, opts ...UpdateOption) error {
	// Load Existing
	meta := obj.ObjectMeta()
	existingObj := s.newObject()
	key := api.NamespacedName{Name: meta.GetName(), Namespace: meta.GetNamespace()}
	if err := s.load(key, existingObj); err != nil {
		return err
	}
	existingMeta := existingObj.ObjectMeta()

	// Defaulting
	if err := Default(obj); err != nil {
		return err
	}
	// Validation
	if err := ValidateUpdate(existingObj, obj); err != nil {
		return err
	}

	// Finalizer Handling
	if obj.ObjectMeta().GetDeletedTimestamp() != nil &&
		len(obj.ObjectMeta().GetFinalizers()) == 0 {
		return s.delete(ctx, obj)
	}

	// Ensure Status is not updated, if the field exists
	statusField := reflect.ValueOf(obj).Elem().FieldByName("Status")
	if statusField.IsValid() {
		statusField.Set(
			reflect.ValueOf(existingObj).Elem().FieldByName("Status"),
		)
	}

	// Check if there is a change
	if reflect.DeepEqual(existingObj, obj) {
		return nil
	}

	// Check ResourceVersion
	if existingMeta.GetResourceVersion() != meta.GetResourceVersion() {
		return errors.ErrConflict{
			NamespacedName: key,
			TypeFullName:   string(obj.ProtoReflect().Descriptor().FullName()),
		}
	}

	// Ensure correct metadata
	s.sequence++
	meta.SetGeneration(meta.GetGeneration() + 1)
	meta.SetResourceVersion(strconv.FormatUint(s.sequence, 10))
	meta.SetUid(existingMeta.GetUid())

	// Store
	if err := s.store(obj); err != nil {
		return err
	}
	s.hub.Broadcast(api.Modified, obj)
	return nil
}

func (s *InMemoryStorage) UpdateStatus(ctx context.Context, obj api.Object, opts ...UpdateOption) error {
	// Input validation
	if err := s.checkObject(obj); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	// Load Existing
	meta := obj.ObjectMeta()
	existingObj := s.newObject()
	key := api.NamespacedName{Name: meta.GetName(), Namespace: meta.GetNamespace()}
	if err := s.load(key, existingObj); err != nil {
		return err
	}

	// Defaulting
	if err := Default(obj); err != nil {
		return err
	}
	// Validation
	if err := ValidateUpdate(existingObj, obj); err != nil {
		return err
	}

	// Ensure _only_ Status is updated
	updatedObj := proto.Clone(existingObj).(api.Object)
	statusField := reflect.ValueOf(updatedObj).Elem().FieldByName("Status")
	if statusField.IsValid() {
		statusField.Set(
			reflect.ValueOf(obj).Elem().FieldByName("Status"),
		)
	}

	// Check if there is a change
	if reflect.DeepEqual(existingObj, updatedObj) {
		return nil
	}

	// Ensure correct metadata
	s.sequence++
	updatedObj.ObjectMeta().
		SetResourceVersion(strconv.FormatUint(s.sequence, 10))

	// Store
	if err := s.store(updatedObj); err != nil {
		return err
	}
	s.hub.Broadcast(api.Modified, updatedObj)
	return nil
}

func (s *InMemoryStorage) list(opts ...ListOption) ([]api.Object, error) {
	var options ListOptions
	for _, opt := range opts {
		opt.ApplyToList(&options)
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	var out []api.Object
	for key, entryData := range s.data {
		if options.Namespace != nil {
			if !strings.HasPrefix(key, *options.Namespace+"/") {
				continue
			}
		}

		obj := s.newObject()
		if err := proto.Unmarshal(entryData, obj); err != nil {
			return nil, err
		}

		if options.LabelSelector != nil {
			if !options.LabelSelector.Matches(
				labels.Set(obj.ObjectMeta().GetLabels())) {
				continue
			}
		}
		out = append(out, obj)
	}
	return out, nil
}

func (s *InMemoryStorage) load(nn api.NamespacedName, obj api.Object) error {
	data, ok := s.data[nn.String()]
	if !ok {
		return errors.ErrNotFound{
			NamespacedName: nn,
			TypeFullName:   string(s.objectFullName),
		}
	}
	if err := proto.Unmarshal(data, obj); err != nil {
		return fmt.Errorf("loading from storage: %w", err)
	}
	return nil
}

func (s *InMemoryStorage) store(obj api.Object) error {
	data, err := proto.Marshal(obj)
	if err != nil {
		return err
	}
	key := api.NamespacedName{
		Name:      obj.ObjectMeta().GetName(),
		Namespace: obj.ObjectMeta().GetNamespace(),
	}
	s.data[key.String()] = data
	return nil
}

func (s *InMemoryStorage) checkObject(obj api.Object) error {
	objFullName := obj.ProtoReflect().Descriptor().FullName()
	if s.objectFullName != objFullName {
		return fmt.Errorf(
			"wrong type for storage. Is: %s, want: %s",
			objFullName, s.objectFullName)
	}
	return nil
}

func (s *InMemoryStorage) checkListObject(obj api.ListObject) error {
	objFullName := obj.ProtoReflect().Descriptor().FullName()
	if s.listObjectFullName != objFullName {
		return fmt.Errorf(
			"wrong list type for storage. Is: %s, want: %s",
			objFullName, s.listObjectFullName)
	}
	return nil
}
