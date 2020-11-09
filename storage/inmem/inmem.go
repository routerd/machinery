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

package inmem

import (
	"context"
	"encoding/json"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"

	"routerd.net/machinery/errors"
	"routerd.net/machinery/runtime"
	storagev1 "routerd.net/machinery/storage/api/v1"
	"routerd.net/machinery/storage/event"
	"routerd.net/machinery/storage/utils"
)

var (
	_ storagev1.Client = (*Storage)(nil)
)

type Storage struct {
	scheme *runtime.Scheme

	gvkGate  *utils.GVKGate
	sequence uint64
	data     map[string][]byte
	hub      *event.Hub
	mux      sync.RWMutex
}

func NewStorage(scheme *runtime.Scheme, obj runtime.Object) (*Storage, error) {
	gate, err := utils.NewGVKGate(scheme, obj)
	if err != nil {
		return nil, err
	}

	s := &Storage{
		scheme: scheme,

		gvkGate: gate,
		data:    map[string][]byte{},
	}
	s.hub = event.NewHub(100, s.list)

	return s, nil
}

func (s *Storage) Run(stopCh <-chan struct{}) {
	s.hub.Run(stopCh)
}

func (s *Storage) Get(ctx context.Context, key storagev1.NamespacedName, obj storagev1.Object) error {
	if err := s.gvkGate.CheckObject(obj); err != nil {
		return err
	}
	if err := key.Validate(); err != nil {
		return err
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.load(key.String(), obj)
}

func (s *Storage) list(opts ...storagev1.ListOption) ([]runtime.Object, error) {
	var options storagev1.ListOptions
	for _, opt := range opts {
		opt.ApplyToList(&options)
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	var out []runtime.Object
	for _, entryData := range s.data {
		o, err := s.scheme.New(s.gvkGate.ObjGVK)
		if err != nil {
			return nil, err
		}
		obj := o.(storagev1.Object)

		if options.Namespace != "" {
			if !strings.HasSuffix(obj.GetNamespace(), "."+options.Namespace) {
				continue
			}
		}

		if err := json.Unmarshal(entryData, obj); err != nil {
			return nil, err
		}
		out = append(out, obj)
	}
	return out, nil
}

func (s *Storage) List(ctx context.Context, listObj storagev1.ListObject, opts ...storagev1.ListOption) error {
	if err := s.gvkGate.CheckList(listObj); err != nil {
		return err
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

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

func (s *Storage) Watch(ctx context.Context, obj storagev1.Object, opts ...storagev1.ListOption) (storagev1.WatchClient, error) {
	if err := s.gvkGate.CheckObject(obj); err != nil {
		return nil, err
	}

	return s.hub.Register(obj.GetResourceVersion(), opts...), nil
}

func (s *Storage) Create(ctx context.Context, obj storagev1.Object, opts ...storagev1.CreateOption) error {
	// Input validation
	if err := s.gvkGate.CheckObject(obj); err != nil {
		return err
	}
	key := storagev1.Key(obj)
	if err := key.Validate(); err != nil {
		return err
	}

	// Defaulting
	if err := utils.Default(obj); err != nil {
		return err
	}
	// Validation
	if err := utils.ValidateCreate(obj); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	if _, ok := s.data[key.String()]; ok { // Already Exists?
		return errors.ErrAlreadyExists{
			Key: key.String(), GVK: s.gvkGate.ObjGVK}
	}

	// Ensure correct metadata
	s.sequence++
	obj.GetObjectKind().SetGroupVersionKind(s.gvkGate.ObjGVK)
	obj.SetGeneration(1)
	obj.SetResourceVersion(strconv.FormatUint(s.sequence, 10))
	obj.SetUID(uuid.New().String())

	// Store
	if err := s.store(obj); err != nil {
		return err
	}
	s.hub.Broadcast(storagev1.Added, obj)
	return nil
}

func (s *Storage) Delete(ctx context.Context, obj storagev1.Object, opts ...storagev1.DeleteOption) error {
	// Input validation
	if err := s.gvkGate.CheckObject(obj); err != nil {
		return err
	}
	key := storagev1.Key(obj)
	if err := key.Validate(); err != nil {
		return err
	}

	// Defaulting
	// because we don't know what the Validation next is expecting
	if err := utils.Default(obj); err != nil {
		return err
	}
	// Validation
	if err := utils.ValidateDelete(obj); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	// Load Existing
	if err := s.load(key.String(), obj); err != nil {
		return err
	}

	// Delete
	delete(s.data, key.String())
	s.hub.Broadcast(storagev1.Deleted, obj)
	return nil
}

func (s *Storage) Update(ctx context.Context, obj storagev1.Object, opts ...storagev1.UpdateOption) error {
	// Input validation
	if err := s.gvkGate.CheckObject(obj); err != nil {
		return err
	}
	key := storagev1.Key(obj)
	if err := key.Validate(); err != nil {
		return err
	}

	// Defaulting
	if err := utils.Default(obj); err != nil {
		return err
	}
	// Validation
	if err := utils.ValidateUpdate(obj); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	// Load Existing
	e, err := s.scheme.New(s.gvkGate.ObjGVK)
	if err != nil {
		return err
	}
	existingObj := e.(storagev1.Object)
	if err := s.load(key.String(), existingObj); err != nil {
		return err
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
	if existingObj.GetResourceVersion() != obj.GetResourceVersion() {
		return errors.ErrConflict{Key: key.String(), GVK: s.gvkGate.ObjGVK}
	}

	// Ensure correct metadata
	s.sequence++
	obj.SetGeneration(obj.GetGeneration() + 1)
	obj.SetResourceVersion(strconv.FormatUint(s.sequence, 10))
	obj.GetObjectKind().SetGroupVersionKind(s.gvkGate.ObjGVK)
	obj.SetUID(existingObj.GetUID())

	// Store
	if err := s.store(obj); err != nil {
		return err
	}
	s.hub.Broadcast(storagev1.Modified, obj)
	return nil
}

func (s *Storage) UpdateStatus(ctx context.Context, obj storagev1.Object, opts ...storagev1.UpdateOption) error {
	// Input validation
	if err := s.gvkGate.CheckObject(obj); err != nil {
		return err
	}
	key := storagev1.Key(obj)
	if err := key.Validate(); err != nil {
		return err
	}

	// Defaulting
	if err := utils.Default(obj); err != nil {
		return err
	}
	// Validation
	if err := utils.ValidateUpdate(obj); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	// Load Existing
	e, err := s.scheme.New(s.gvkGate.ObjGVK)
	if err != nil {
		return err
	}
	existingObj := e.(storagev1.Object)
	if err := s.load(key.String(), existingObj); err != nil {
		return err
	}

	// Ensure ObjectMeta and Spec is not updated
	reflect.ValueOf(obj).Elem().FieldByName("ObjectMeta").Set(
		reflect.ValueOf(existingObj).Elem().FieldByName("ObjectMeta"),
	)
	specField := reflect.ValueOf(obj).Elem().FieldByName("Spec")
	if specField.IsValid() {
		specField.Set(
			reflect.ValueOf(existingObj).Elem().FieldByName("Spec"),
		)
	}

	// Check if there is a change
	if reflect.DeepEqual(existingObj, obj) {
		return nil
	}

	// Ensure correct metadata
	s.sequence++
	obj.SetResourceVersion(strconv.FormatUint(s.sequence, 10))
	obj.GetObjectKind().SetGroupVersionKind(s.gvkGate.ObjGVK)
	obj.SetUID(existingObj.GetUID())

	// Store
	if err := s.store(obj); err != nil {
		return err
	}
	s.hub.Broadcast(storagev1.Modified, obj)
	return nil
}

func (s *Storage) load(key string, obj storagev1.Object) error {
	data, ok := s.data[key]
	if !ok {
		return errors.ErrNotFound{Key: key, GVK: s.gvkGate.ObjGVK}
	}
	if err := json.Unmarshal(data, obj); err != nil {
		return err
	}
	return nil
}

func (s *Storage) store(obj storagev1.Object) error {
	obj.GetObjectKind().SetGroupVersionKind(s.gvkGate.ObjGVK)

	key := obj.GetName()
	data, err := json.Marshal(obj)
	if err != nil {
		return err
	}
	s.data[key] = data
	return nil
}
