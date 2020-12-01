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
	"strings"

	"google.golang.org/protobuf/reflect/protoreflect"

	"routerd.net/machinery/api"
)

type NamespacedStorage struct {
	namespaceType   protoreflect.FullName
	newNamespace    func() api.Object
	typeStorages    map[protoreflect.FullName]api.Client
	namespacedTypes map[protoreflect.FullName]struct{}
}

func NewNamespacedStorage(
	namespaceObj api.Object,
) *NamespacedStorage {
	return &NamespacedStorage{
		namespaceType: namespaceObj.ProtoReflect().Descriptor().FullName(),
		newNamespace: func() api.Object {
			return namespaceObj.
				ProtoReflect().New().Interface().(api.Object)
		},
		typeStorages:    map[protoreflect.FullName]api.Client{},
		namespacedTypes: map[protoreflect.FullName]struct{}{},
	}
}

func (s *NamespacedStorage) storage(objType protoreflect.FullName) (api.Client, error) {
	typeStorage, ok := s.typeStorages[objType]
	if !ok {
		return nil, fmt.Errorf("unknown type %q", objType)
	}
	return typeStorage, nil
}

func (s *NamespacedStorage) validNamespace(
	ctx context.Context, objType protoreflect.FullName, namespace string) (ns api.Object, err error) {
	if len(namespace) == 0 {
		if _, ok := s.namespacedTypes[objType]; ok {
			return nil, fmt.
				Errorf("objects of type %q need to have a namespace specified.", objType)
		}

		// object is not namespaced
		return nil, nil
	} else if objType == s.namespaceType {
		return nil, fmt.
			Errorf("namespace objects cannot be namespaced.")
	}

	namespaceStorage, err := s.storage(s.namespaceType)
	if err != nil {
		return nil, err
	}

	ns = s.newNamespace()
	if err := namespaceStorage.Get(ctx, api.NamespacedName{
		Name: namespace,
	}, ns); err != nil {
		return nil, err
	}
	return ns, nil
}

func (s *NamespacedStorage) RegisterStorage(obj api.Object, typeStorage api.Client, namespaced bool) {
	objType := obj.ProtoReflect().Descriptor().FullName()
	s.typeStorages[objType] = typeStorage
	if namespaced {
		s.namespacedTypes[objType] = struct{}{}
	}
}

func (s *NamespacedStorage) Get(ctx context.Context, nn api.NamespacedName, obj api.Object) error {
	objType := obj.ProtoReflect().Descriptor().FullName()
	if _, err := s.validNamespace(ctx, objType, nn.Namespace); err != nil {
		return err
	}
	typeStorage, err := s.storage(objType)
	if err != nil {
		return err
	}
	return typeStorage.Get(ctx, nn, obj)
}

func (s *NamespacedStorage) List(ctx context.Context, listObj api.ListObject, opts ...api.ListOption) error {
	objType := protoreflect.FullName(strings.TrimSuffix(
		string(listObj.ProtoReflect().Descriptor().FullName()), "List"))

	var listOptions api.ListOptions
	for _, opt := range opts {
		opt.ApplyToList(&listOptions)
	}
	if len(listOptions.Namespace) > 0 {
		if _, err := s.validNamespace(
			ctx, objType, listOptions.Namespace); err != nil {
			return err
		}
	}
	typeStorage, err := s.storage(objType)
	if err != nil {
		return err
	}
	return typeStorage.List(ctx, listObj, opts...)
}

func (s *NamespacedStorage) Watch(ctx context.Context, obj api.Object, opts ...api.ListOption) (api.WatchClient, error) {
	objType := obj.ProtoReflect().Descriptor().FullName()

	var listOptions api.ListOptions
	for _, opt := range opts {
		opt.ApplyToList(&listOptions)
	}
	if len(listOptions.Namespace) > 0 {
		// Namespace is a filter here, so not specifying it is ok
		if _, err := s.validNamespace(
			ctx, objType, listOptions.Namespace); err != nil {
			return nil, err
		}
	}
	typeStorage, err := s.storage(objType)
	if err != nil {
		return nil, err
	}
	return typeStorage.Watch(ctx, obj, opts...)
}

func (s *NamespacedStorage) Create(ctx context.Context, obj api.Object, opts ...api.CreateOption) error {
	objType := obj.ProtoReflect().Descriptor().FullName()
	if ns, err := s.validNamespace(ctx, objType, obj.ObjectMeta().GetNamespace()); err != nil {
		return err
	} else if ns != nil && ns.ObjectMeta().GetDeletedTimestamp() != nil {
		return fmt.Errorf("namespace is beeing terminated, creating objects is forbidden")
	}
	typeStorage, err := s.storage(objType)
	if err != nil {
		return err
	}
	return typeStorage.Create(ctx, obj, opts...)
}

func (s *NamespacedStorage) Delete(ctx context.Context, obj api.Object, opts ...api.DeleteOption) error {
	objType := obj.ProtoReflect().Descriptor().FullName()
	if _, err := s.validNamespace(ctx, objType, obj.ObjectMeta().GetNamespace()); err != nil {
		return err
	}
	typeStorage, err := s.storage(objType)
	if err != nil {
		return err
	}
	return typeStorage.Delete(ctx, obj, opts...)
}

func (s *NamespacedStorage) DeleteAllOf(ctx context.Context, obj api.Object, opts ...api.DeleteAllOfOption) error {
	objType := obj.ProtoReflect().Descriptor().FullName()

	var deleteAllOfOptions api.DeleteAllOfOptions
	for _, opt := range opts {
		opt.ApplyToDeleteAllOf(&deleteAllOfOptions)
	}
	if len(deleteAllOfOptions.Namespace) > 0 {
		// Namespace is a filter here, so not specifying it is ok
		if _, err := s.validNamespace(
			ctx, objType, deleteAllOfOptions.Namespace); err != nil {
			return err
		}
	}
	typeStorage, err := s.storage(objType)
	if err != nil {
		return err
	}
	return typeStorage.DeleteAllOf(ctx, obj, opts...)
}

func (s *NamespacedStorage) Update(ctx context.Context, obj api.Object, opts ...api.UpdateOption) error {
	objType := obj.ProtoReflect().Descriptor().FullName()
	if _, err := s.validNamespace(ctx, objType, obj.ObjectMeta().GetNamespace()); err != nil {
		return err
	}
	typeStorage, err := s.storage(objType)
	if err != nil {
		return err
	}
	return typeStorage.Update(ctx, obj, opts...)
}

func (s *NamespacedStorage) UpdateStatus(ctx context.Context, obj api.Object, opts ...api.UpdateOption) error {
	objType := obj.ProtoReflect().Descriptor().FullName()
	if _, err := s.validNamespace(ctx, objType, obj.ObjectMeta().GetNamespace()); err != nil {
		return err
	}
	typeStorage, err := s.storage(objType)
	if err != nil {
		return err
	}
	return typeStorage.UpdateStatus(ctx, obj, opts...)
}
