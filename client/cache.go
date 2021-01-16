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

package client

import (
	"context"
	"sync"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"

	"routerd.net/machinery/api"
	"routerd.net/machinery/errors"
)

type ListWatcher interface {
	api.Lister
	api.Watcher
}

type Cache struct {
	listWatcher        ListWatcher
	objectFullName     protoreflect.FullName
	listObjectFullName protoreflect.FullName

	objects map[string]api.Object
	mux     sync.RWMutex
}

func NewCache(
	objType api.Object,
	listWatcher ListWatcher,
) *Cache {
	objName := objType.ProtoReflect().Descriptor().FullName()
	listObjName := protoreflect.FullName(objName + "List")

	return &Cache{
		listWatcher:        listWatcher,
		objectFullName:     objName,
		listObjectFullName: listObjName,
		objects:            map[string]api.Object{},
	}
}

func (c *Cache) Run(stopCh <-chan struct{}) error {
	// ctx := context.Background()
	// ctx, cancel := context.WithCancel(ctx)
	// defer cancel()

	// listObjDesc, err := protoregistry.GlobalFiles.FindDescriptorByName(c.listObjectFullName)
	// if err != nil {
	// 	panic(err)
	// }
	// list := listObjDesc

	// // protoreflect.
	// list := c.listObjectFullName.ProtoReflect().
	// 	New().Interface().(api.ObjectList)
	// c.listWatcher.List(ctx)

	// for {
	// 	select {
	// 	case <-stopCh:
	// 		return nil
	// 	default:

	// 	}
	// }
	return nil
}

func (c *Cache) Get(ctx context.Context, nn api.NamespacedName, obj api.Object) error {
	c.mux.RLock()
	defer c.mux.RUnlock()

	cachedObj, ok := c.objects[nn.String()]
	if !ok {
		return errors.ErrNotFound{
			NamespacedName: nn,
			TypeFullName:   string(obj.ProtoReflect().Descriptor().FullName()),
		}
	}

	proto.Merge(obj, cachedObj)
	return nil
}

func (c *Cache) List(ctx context.Context, listObj api.ListObject, opts ...api.ListOption) error {

	var options api.ListOptions
	for _, opt := range opts {
		opt.ApplyToList(&options)
	}

	return nil
}
