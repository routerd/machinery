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
	"encoding/json"
	"fmt"
	"reflect"
	"strconv"
	"strings"
	"sync"

	"github.com/google/uuid"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"routerd.net/machinery/errors"
	"routerd.net/machinery/validate"
)

type InMemoryStorage struct {
	hub                *eventHub
	objectFullName     protoreflect.FullName
	listObjectFullName protoreflect.FullName
	newObject          func() Object
	// newListObject func() ListObject
	mux      sync.RWMutex
	data     map[string][]byte
	sequence uint64
}

func NewInMemoryStorage(objType Object) *InMemoryStorage {
	objName := objType.ProtoReflect().Descriptor().FullName()
	listObjName := protoreflect.FullName(objName + "List")

	s := &InMemoryStorage{
		objectFullName:     objName,
		listObjectFullName: listObjName,
		newObject: func() Object {
			return objType.ProtoReflect().
				New().Interface().(Object)
		},
		data: map[string][]byte{},
	}
	s.hub = NewHub(s.list)
	return s
}

func (s *InMemoryStorage) Run(stopCh <-chan struct{}) {
	s.hub.Run(stopCh)
}

func (s *InMemoryStorage) Get(
	ctx context.Context,
	name, namespace string,
	obj Object,
) error {
	if err := s.checkObject(obj); err != nil {
		return err
	}
	if err := validate.ValidateNamespacedName(name, namespace); err != nil {
		return err
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	return s.load(name, namespace, obj)
}

func (s *InMemoryStorage) List(ctx context.Context, listObj ListObject, opts ...ListOption) error {
	if err := s.checkListObject(listObj); err != nil {
		return err
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	objects, err := s.list(opts...)
	if err != nil {
		return err
	}
	protoMessages := make([]proto.Message, len(objects))
	for i := 0; i < len(objects); i++ {
		protoMessages[i] = objects[i]
	}
	listObj.SetItems(protoMessages)
	return nil
}

func (s *InMemoryStorage) Watch(ctx context.Context, obj Object, opts ...ListOption) (WatchClient, error) {
	if err := s.checkObject(obj); err != nil {
		return nil, err
	}

	return s.hub.Register(
		obj.ObjectMetaAccessor().GetResourceVersion(), opts...)
}

func (s *InMemoryStorage) Create(ctx context.Context, obj Object, opts ...CreateOption) error {
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

	meta := obj.ObjectMetaAccessor()
	key := s.key(meta.GetName(), meta.GetNamespace())
	if _, ok := s.data[key]; ok { // Already Exists?
		return errors.ErrAlreadyExists{
			Name: meta.GetName(), Namespace: meta.GetNamespace(),
			TypeFullName: string(obj.ProtoReflect().Descriptor().FullName()),
		}
	}

	// Ensure correct metadata
	s.sequence++
	meta.SetGeneration(1)
	meta.SetResourceVersion(strconv.FormatUint(s.sequence, 10))
	meta.SetUID(uuid.New().String())

	// Store
	if err := s.store(obj); err != nil {
		return err
	}
	s.hub.Broadcast(Added, obj)
	return nil
}

func (s *InMemoryStorage) Delete(ctx context.Context, obj Object, opts ...DeleteOption) error {
	// Input validation
	if err := s.checkObject(obj); err != nil {
		return err
	}

	// Defaulting
	// because we don't know what the Validation next is expecting
	if err := Default(obj); err != nil {
		return err
	}
	// Validation
	if err := ValidateDelete(obj); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	// Load Existing
	meta := obj.ObjectMetaAccessor()
	key := s.key(meta.GetName(), meta.GetNamespace())
	if err := s.load(meta.GetName(), meta.GetNamespace(), obj); err != nil {
		return err
	}

	// Delete
	delete(s.data, key)
	s.hub.Broadcast(Deleted, obj)
	return nil
}

func (s *InMemoryStorage) Update(ctx context.Context, obj Object, opts ...UpdateOption) error {
	// Input validation
	if err := s.checkObject(obj); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	// Load Existing
	meta := obj.ObjectMetaAccessor()
	existingObj := s.newObject()
	if err := s.load(meta.GetName(), meta.GetNamespace(), existingObj); err != nil {
		return err
	}
	existingMeta := existingObj.ObjectMetaAccessor()

	// Defaulting
	if err := Default(obj); err != nil {
		return err
	}
	// Validation
	if err := ValidateUpdate(existingObj, obj); err != nil {
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
	if existingMeta.GetResourceVersion() != meta.GetResourceVersion() {
		return errors.ErrConflict{
			Name: meta.GetName(), Namespace: meta.GetNamespace(),
			TypeFullName: string(obj.ProtoReflect().Descriptor().FullName()),
		}
	}

	// Ensure correct metadata
	s.sequence++
	meta.SetGeneration(meta.GetGeneration() + 1)
	meta.SetResourceVersion(strconv.FormatUint(s.sequence, 10))
	meta.SetUID(existingMeta.GetUID())

	// Store
	if err := s.store(obj); err != nil {
		return err
	}
	s.hub.Broadcast(Modified, obj)
	return nil
}

func (s *InMemoryStorage) UpdateStatus(ctx context.Context, obj Object, opts ...UpdateOption) error {
	// Input validation
	if err := s.checkObject(obj); err != nil {
		return err
	}

	s.mux.Lock()
	defer s.mux.Unlock()

	// Load Existing
	meta := obj.ObjectMetaAccessor()
	existingObj := s.newObject()
	if err := s.load(meta.GetName(), meta.GetNamespace(), existingObj); err != nil {
		return err
	}
	existingMeta := existingObj.ObjectMetaAccessor()

	// Defaulting
	if err := Default(obj); err != nil {
		return err
	}
	// Validation
	if err := ValidateUpdate(existingObj, obj); err != nil {
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
	meta.SetResourceVersion(strconv.FormatUint(s.sequence, 10))
	meta.SetUID(existingMeta.GetUID())

	// Store
	if err := s.store(obj); err != nil {
		return err
	}
	s.hub.Broadcast(Modified, obj)
	return nil
}

func (s *InMemoryStorage) list(opts ...ListOption) ([]Object, error) {
	var options ListOptions
	for _, opt := range opts {
		opt.ApplyToList(&options)
	}

	s.mux.RLock()
	defer s.mux.RUnlock()

	var out []Object
	for key, entryData := range s.data {
		if options.Namespace != "" {
			if !strings.HasPrefix(key, options.Namespace+"/") {
				continue
			}
		}

		obj := s.newObject()
		if err := json.Unmarshal(entryData, obj); err != nil {
			return nil, err
		}
		out = append(out, obj)
	}
	return out, nil
}

func (s *InMemoryStorage) key(name, namespace string) string {
	return namespace + "/" + name
}

func (s *InMemoryStorage) load(name, namespace string, obj Object) error {
	data, ok := s.data[s.key(name, namespace)]
	if !ok {
		return errors.ErrNotFound{
			Name: name, Namespace: namespace,
			TypeFullName: string(s.objectFullName),
		}
	}
	if err := proto.Unmarshal(data, obj); err != nil {
		return err
	}
	return nil
}

func (s *InMemoryStorage) store(obj Object) error {
	data, err := proto.Marshal(obj)
	if err != nil {
		return err
	}
	key := s.key(
		obj.ObjectMetaAccessor().GetName(),
		obj.ObjectMetaAccessor().GetNamespace())
	s.data[key] = data
	return nil
}

func (s *InMemoryStorage) checkObject(obj Object) error {
	objFullName := obj.ProtoReflect().Descriptor().FullName()
	if s.objectFullName != objFullName {
		return fmt.Errorf(
			"wrong type for storage. Is: %s, want: %s",
			objFullName, s.objectFullName)
	}
	return nil
}

func (s *InMemoryStorage) checkListObject(obj ListObject) error {
	objFullName := obj.ProtoReflect().Descriptor().FullName()
	if s.listObjectFullName != objFullName {
		return fmt.Errorf(
			"wrong list type for storage. Is: %s, want: %s",
			objFullName, s.listObjectFullName)
	}
	return nil
}
