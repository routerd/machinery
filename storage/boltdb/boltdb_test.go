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
	"context"
	"os"
	"sync"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	bolt "go.etcd.io/bbolt"

	"routerd.net/machinery/meta"
	"routerd.net/machinery/runtime"
	storagev1 "routerd.net/machinery/storage/api/v1"
)

func TestStorage(t *testing.T) {
	// make sure the test database file is always gone
	os.Remove("test.db")
	defer os.Remove("test.db")

	db, err := bolt.Open("test.db", os.ModePerm, &bolt.Options{})
	require.NoError(t, err)

	s, err := NewBoltDBStorage(testScheme, &testObject{}, db)
	require.NoError(t, err)
	stopCh := make(chan struct{})
	defer close(stopCh)
	go s.Run(stopCh)

	obj := &testObject{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test",
			Namespace: "stuff",
		},
	}
	ctx := context.Background()

	// Create a new Object
	require.NoError(t, s.Create(ctx, obj))
	assert.Equal(t, s.objGVK, obj.GetGroupVersionKind())
	assert.NotEmpty(t, obj.UID)
	assert.Equal(t, int64(1), obj.GetGeneration())
	assert.Equal(t, "1", obj.GetResourceVersion())

	// Get it again
	getObj := &testObject{}
	require.NoError(t, s.Get(ctx, storagev1.NamespacedName{
		Name: obj.Name, Namespace: obj.Namespace,
	}, getObj))
	assert.Equal(t, s.objGVK, getObj.GetGroupVersionKind())
	assert.NotEmpty(t, getObj.UID)
	assert.Equal(t, int64(1), getObj.GetGeneration())
	assert.Equal(t, "1", getObj.GetResourceVersion())

	// List all objects
	list := &testObjectList{}
	require.NoError(t, s.List(ctx, list))
	if assert.Len(t, list.Items, 1) {
		assert.Equal(t, *getObj, list.Items[0])
	}

	// Update
	getObj.Annotations = map[string]string{"test": "test"}
	getObj.Status.Prop = "something"
	require.NoError(t, s.Update(ctx, getObj))
	assert.Equal(t, int64(2), getObj.GetGeneration())
	assert.Equal(t, "2", getObj.GetResourceVersion())
	// meta updated
	assert.Equal(t, map[string]string{"test": "test"}, getObj.Annotations)
	// status NOT updated
	assert.Equal(t, "", getObj.Status.Prop)

	// Update Status
	getObj.Annotations = map[string]string{"test": "test2"}
	getObj.Status.Prop = "something"
	require.NoError(t, s.UpdateStatus(ctx, getObj))
	assert.Equal(t, int64(2), getObj.GetGeneration())
	assert.Equal(t, "3", getObj.GetResourceVersion())
	// no update to meta
	assert.Equal(t, map[string]string{"test": "test"}, getObj.Annotations)
	// status updated
	assert.Equal(t, "something", getObj.Status.Prop)

	// Delete
	require.NoError(t, s.Delete(ctx, getObj))

	// List all objects
	list = &testObjectList{}
	require.NoError(t, s.List(ctx, list))
	assert.Len(t, list.Items, 0)
}

func TestWatch(t *testing.T) {
	// make sure the test database file is always gone
	os.Remove("watch.db")
	defer os.Remove("watch.db")

	db, err := bolt.Open("watch.db", os.ModePerm, &bolt.Options{})
	require.NoError(t, err)

	s, err := NewBoltDBStorage(testScheme, &testObject{}, db)
	require.NoError(t, err)
	stopCh := make(chan struct{})
	defer close(stopCh)
	go s.Run(stopCh)

	// Watch
	ctx := context.Background()
	watcher, err := s.Watch(ctx, &testObject{})
	require.NoError(t, err)

	// already create an object
	obj1 := &testObject{ObjectMeta: meta.ObjectMeta{Name: "test1a", Namespace: "test"}}
	require.NoError(t, s.Create(ctx, obj1))

	var events []storagev1.Event
	var wg sync.WaitGroup
	wg.Add(3) // wait for 3 events
	go func() {
		for event := range watcher.EventChan() {
			events = append(events, event)
			wg.Done()
		}
	}()

	// generate an "Added" event
	obj2 := &testObject{ObjectMeta: meta.ObjectMeta{Name: "test2a", Namespace: "test"}}
	require.NoError(t, s.Create(ctx, obj2))

	// generate an Update
	obj2update := &testObject{ObjectMeta: meta.ObjectMeta{
		Name:            "test2a",
		Namespace:       "test",
		ResourceVersion: obj2.ResourceVersion,
		Annotations:     map[string]string{"test": "test"},
	}}
	require.NoError(t, s.Update(ctx, obj2update))

	// Assertions
	wg.Wait()
	if assert.Len(t, events, 3) {
		assert.Equal(t, storagev1.Event{Type: storagev1.Added, Object: obj1}, events[0])
		assert.Equal(t, storagev1.Event{Type: storagev1.Added, Object: obj2}, events[1])
		assert.Equal(t, storagev1.Event{Type: storagev1.Modified, Object: obj2update}, events[2])
	}
}

func TestStorageConversion(t *testing.T) {
	// make sure the test database file is always gone
	os.Remove("conversion.db")
	defer os.Remove("conversion.db")

	db, err := bolt.Open("conversion.db", os.ModePerm, &bolt.Options{})
	require.NoError(t, err)

	s, err := NewBoltDBStorage(testScheme, &testObject{}, db)
	require.NoError(t, err)
	stopCh := make(chan struct{})
	defer close(stopCh)
	go s.Run(stopCh)

	v1 := &testObject{
		ObjectMeta: meta.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
	}
	ctx := context.Background()
	require.NoError(t, s.Create(ctx, v1))

	// change desired storage types
	s.objGVK = v2GVK
	s.listGVK = v2ListGVK

	require.NoError(t, s.ensureStorageVersion())
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

type testObjectV2 struct {
	meta.TypeMeta   `json:",inline"`
	meta.ObjectMeta `json:"metadata"`
	Status          testObjectStatusV2 `json:"status"`
}

type testObjectStatusV2 struct {
	Prop string `json:"prop"`
}

func (o *testObjectV2) DeepCopyObject() runtime.Object {
	clone := &testObjectV2{}
	runtime.DeepCopy(o, clone)
	return clone
}

type testObjectV2List struct {
	meta.TypeMeta `json:",inline"`
	Items         []testObjectV2 `json:"items"`
}

func (o *testObjectV2List) DeepCopyObject() runtime.Object {
	clone := &testObjectV2List{}
	runtime.DeepCopy(o, clone)
	return clone
}

var (
	v2GVK = runtime.GroupVersionKind{
		Group:   "testing",
		Version: "v2",
		Kind:    "testObject",
	}
	v2ListGVK = runtime.GroupVersionKind{
		Group:   "testing",
		Version: "v2",
		Kind:    "testObjectList",
	}
)

func init() {
	testScheme.AddKnownTypes(runtime.GroupVersion{
		Group:   "testing",
		Version: "v1",
	}, &testObject{}, &testObjectList{})

	testScheme.AddKnownTypeWithKind(v2GVK, &testObjectV2{})
	testScheme.AddKnownTypeWithKind(v2ListGVK, &testObjectV2List{})
}
