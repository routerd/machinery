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
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"routerd.net/machinery/meta"
	"routerd.net/machinery/runtime"
	storagev1 "routerd.net/machinery/storage/api/v1"
)

func TestStorage(t *testing.T) {
	t.Run("Get", func(t *testing.T) {
		s, err := NewStorage(testScheme, &testObject{})
		require.NoError(t, err)

		// Run
		stopCh := make(chan struct{})
		go s.Run(stopCh)
		defer close(stopCh)

		// Inject some data
		s.data["test123.test"] = []byte(
			`{"metadata": {"name":"test123","namespace":"test"}}`)

		// List
		ctx := context.Background()
		obj := &testObject{}
		err = s.Get(ctx, storagev1.NamespacedName{
			Name: "test123", Namespace: "test",
		}, obj)
		require.NoError(t, err)

		assert.Equal(t, "test123", obj.Name)
	})

	t.Run("List", func(t *testing.T) {
		s, err := NewStorage(testScheme, &testObject{})
		require.NoError(t, err)

		// Run
		stopCh := make(chan struct{})
		go s.Run(stopCh)
		defer close(stopCh)

		// Inject some data
		s.data["test123.test"] = []byte(
			`{"metadata": {"name":"test123","namespace":"test"}}`)

		// List
		ctx := context.Background()
		list := &testObjectList{}
		err = s.List(ctx, list)
		require.NoError(t, err)

		if assert.Len(t, list.Items, 1) {
			assert.Equal(t, "test123", list.Items[0].Name)
		}
	})

	t.Run("Watch", func(t *testing.T) {
		s, err := NewStorage(testScheme, &testObject{})
		require.NoError(t, err)

		// Run
		stopCh := make(chan struct{})
		go s.Run(stopCh)
		defer close(stopCh)

		// Watch
		ctx := context.Background()
		watcher, err := s.Watch(ctx, &testObject{})
		require.NoError(t, err)

		var events []storagev1.Event
		var wg sync.WaitGroup
		wg.Add(1) // wait for 1 event
		go func() {
			for event := range watcher.EventChan() {
				events = append(events, event)
				wg.Done()
			}
		}()

		// generate a "Added" event
		obj := &testObject{
			ObjectMeta: meta.ObjectMeta{Name: "test3000", Namespace: "test"}}
		require.NoError(t, s.Create(ctx, obj))

		// Assertions
		wg.Wait()
		if assert.Len(t, events, 1) {
			assert.Equal(t, storagev1.Event{
				Type: storagev1.Added, Object: obj}, events[0])
		}
	})
}

var testScheme = runtime.NewScheme()

type testObject struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata"`
	Status          testObjectStatus `json:"status"`
}

type testObjectStatus struct {
	Prop string `json:"prop"`
}

func (o *testObject) DeepCopyObject() runtime.Object {
	clone := &testObject{}
	runtime.DeepCopy(o, clone)
	return clone
}

type testObjectList struct {
	meta.TypeMeta `json:",inline"`
	Items         []testObject `json:"items"`
}

func (o *testObjectList) DeepCopyObject() runtime.Object {
	clone := &testObjectList{}
	runtime.DeepCopy(o, clone)
	return clone
}

func init() {
	testScheme.AddKnownTypes(runtime.GroupVersion{
		Group:   "testing",
		Version: "v1",
	}, &testObject{}, &testObjectList{})
}
