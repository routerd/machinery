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
	"sync"

	"github.com/go-logr/logr"
	bolt "go.etcd.io/bbolt"

	"routerd.net/machinery/runtime"
	storagev1 "routerd.net/machinery/storage/api/v1"
	"routerd.net/machinery/storage/boltdb"
	"routerd.net/machinery/storage/inmem"
)

type Storage struct {
	log         logr.Logger
	scheme      *runtime.Scheme
	gkToStorage map[runtime.GroupKind]storage
}

var _ storagev1.Client = (*Storage)(nil)

type storage interface {
	storagev1.Client
	Run(stopCh <-chan struct{})
}

func NewStorage(
	log logr.Logger,
	scheme *runtime.Scheme,
	db *bolt.DB,
) (*Storage, error) {
	mr := &Storage{
		log:         log,
		scheme:      scheme,
		gkToStorage: map[runtime.GroupKind]storage{},
	}

	// register storages for all known objects
	gvks := scheme.KnownObjectKinds()
	for _, gvk := range gvks {
		obj, err := scheme.New(gvk)
		if err != nil {
			return nil, err
		}

		listGVK, err := scheme.ListGroupVersionKind(obj)
		if err != nil {
			return nil, err
		}

		repo, err := boltdb.NewBoltDBStorage(scheme, obj.(storagev1.Object), db)
		if err != nil {
			return nil, err
		}

		mr.gkToStorage[gvk.GroupKind()] = repo
		mr.gkToStorage[listGVK.GroupKind()] = repo
	}
	return mr, nil
}

// Init BoltDB backed storage for the given GVKs.
func (s *Storage) InitBoltDB(db *bolt.DB, gvks ...runtime.GroupVersionKind) error {
	for _, gvk := range gvks {
		obj, err := s.scheme.New(gvk)
		if err != nil {
			return err
		}

		listGVK, err := s.scheme.ListGroupVersionKind(obj)
		if err != nil {
			return err
		}

		repo, err := boltdb.NewBoltDBStorage(s.scheme, obj.(storagev1.Object), db)
		if err != nil {
			return err
		}

		s.gkToStorage[gvk.GroupKind()] = repo
		s.gkToStorage[listGVK.GroupKind()] = repo
	}
	return nil
}

// Init In-Memory backed storage for the given GVKs.
func (s *Storage) InitInMemory(db *bolt.DB, gvks ...runtime.GroupVersionKind) error {
	for _, gvk := range gvks {
		obj, err := s.scheme.New(gvk)
		if err != nil {
			return err
		}

		listGVK, err := s.scheme.ListGroupVersionKind(obj)
		if err != nil {
			return err
		}

		store, err := inmem.NewStorage(s.scheme, obj.(storagev1.Object))
		if err != nil {
			return err
		}

		s.gkToStorage[gvk.GroupKind()] = store
		s.gkToStorage[listGVK.GroupKind()] = store
	}
	return nil
}

func (s *Storage) Run(stopCh <-chan struct{}) {
	var wg sync.WaitGroup
	for gk, s := range s.gkToStorage {
		if strings.HasSuffix(gk.Kind, "List") {
			continue
		}

		wg.Add(1)
		go func(s storage, stopCh <-chan struct{}) {
			s.Run(stopCh)
			wg.Done()
		}(s, stopCh)
	}
	wg.Wait()
}

func (s *Storage) Get(ctx context.Context, key storagev1.NamespacedName, obj storagev1.Object) error {
	store, err := s.storageForObj(obj)
	if err != nil {
		return err
	}
	return store.Get(ctx, key, obj)
}

func (s *Storage) List(ctx context.Context, listObj storagev1.ListObject, opts ...storagev1.ListOption) error {
	store, err := s.storageForObj(listObj)
	if err != nil {
		return err
	}
	return store.List(ctx, listObj, opts...)
}

func (s *Storage) Watch(ctx context.Context, obj storagev1.Object, opts ...storagev1.ListOption) (storagev1.WatchClient, error) {
	store, err := s.storageForObj(obj)
	if err != nil {
		return nil, err
	}
	return store.Watch(ctx, obj, opts...)
}

func (s *Storage) Create(ctx context.Context, obj storagev1.Object, opts ...storagev1.CreateOption) error {
	store, err := s.storageForObj(obj)
	if err != nil {
		return err
	}
	return store.Create(ctx, obj, opts...)
}

func (s *Storage) Delete(ctx context.Context, obj storagev1.Object, opts ...storagev1.DeleteOption) error {
	store, err := s.storageForObj(obj)
	if err != nil {
		return err
	}
	return store.Delete(ctx, obj, opts...)
}

func (s *Storage) Update(ctx context.Context, obj storagev1.Object, opts ...storagev1.UpdateOption) error {
	store, err := s.storageForObj(obj)
	if err != nil {
		return err
	}
	return store.Update(ctx, obj, opts...)
}

func (s *Storage) UpdateStatus(ctx context.Context, obj storagev1.Object, opts ...storagev1.UpdateOption) error {
	store, err := s.storageForObj(obj)
	if err != nil {
		return err
	}
	return store.UpdateStatus(ctx, obj, opts...)
}

func (s *Storage) storageForObj(obj runtime.Object) (storage, error) {
	gvk, err := s.scheme.GroupVersionKind(obj)
	if err != nil {
		return nil, err
	}

	storage, ok := s.gkToStorage[gvk.GroupKind()]
	if !ok {
		return nil, fmt.Errorf("no storage registered for kind %s", gvk)
	}
	return storage, nil
}
