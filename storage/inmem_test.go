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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"routerd.net/machinery/api"
	machineryv1 "routerd.net/machinery/api/v1"
	testdatav1 "routerd.net/machinery/storage/testdata/v1"
)

var (
	_ api.Object = (*testdatav1.TestObject)(nil)
)

func TestStorage(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		s := NewInMemoryStorage(&testdatav1.TestObject{})

		// Run
		stopCh := make(chan struct{})
		go s.Run(stopCh)
		defer close(stopCh)

		ctx := context.Background()

		// Inject some data
		require.NoError(t, s.Create(ctx, &testdatav1.TestObject{
			Meta: &machineryv1.ObjectMeta{
				Name:      "test123",
				Namespace: "test",
			},
		}))

		// Get
		obj := &testdatav1.TestObject{}
		err := s.Get(ctx, api.NamespacedName{
			Name: "test123", Namespace: "test",
		}, obj)
		require.NoError(t, err)

		assert.Equal(t, "test123", obj.Meta.Name)
	})

	t.Run("List", func(t *testing.T) {
		s := NewInMemoryStorage(&testdatav1.TestObject{})

		// Run
		stopCh := make(chan struct{})
		go s.Run(stopCh)
		defer close(stopCh)

		ctx := context.Background()

		// Inject some data
		require.NoError(t, s.Create(ctx, &testdatav1.TestObject{
			Meta: &machineryv1.ObjectMeta{
				Name:      "test123",
				Namespace: "test",
			},
		}))

		// List
		list := &testdatav1.TestObjectList{}
		require.NoError(t, s.List(ctx, list))

		if assert.Len(t, list.Items, 1) {
			assert.Equal(t, "test123", list.Items[0].Meta.Name)
		}
	})

	t.Run("Watch", func(t *testing.T) {
		s := NewInMemoryStorage(&testdatav1.TestObject{})

		// Run
		stopCh := make(chan struct{})
		go s.Run(stopCh)
		defer close(stopCh)

		// Watch
		ctx := context.Background()
		watcher, err := s.Watch(ctx, &testdatav1.TestObject{})
		require.NoError(t, err)

		var events []api.Event
		var wg sync.WaitGroup
		wg.Add(1) // wait for 1 event
		go func() {
			for event := range watcher.Events() {
				events = append(events, event)
				wg.Done()
			}
		}()

		// generate a "Added" event
		obj := &testdatav1.TestObject{
			Meta: &machineryv1.ObjectMeta{
				Name: "test3000", Namespace: "test",
			},
		}
		require.NoError(t, s.Create(ctx, obj))

		// Assertions
		wg.Wait()
		if assert.Len(t, events, 1) {
			assert.Equal(t, api.Event{
				Type: api.Added, Object: obj}, events[0])
		}
	})
}
