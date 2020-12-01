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
	"google.golang.org/protobuf/proto"

	"routerd.net/machinery/api"
	machineryv1 "routerd.net/machinery/api/v1"
	testdatav1 "routerd.net/machinery/testdata/v1"
)

func TestInMemoryStorage(t *testing.T) {
	s := NewInMemoryStorage(&testdatav1.TestObject{})

	// Run
	stopCh := make(chan struct{})
	go s.Run(stopCh)
	defer close(stopCh)

	StorageTestSuite(t, s)
}

func StorageTestSuite(t *testing.T, s api.Client) {
	t.Helper()
	ctx := context.Background()

	initObjectNamespaced := &testdatav1.TestObject{
		Meta: &machineryv1.ObjectMeta{
			Name:      "test-namespaced",
			Namespace: "test",
		},
	}

	// Create
	// ------
	createObjects := []*testdatav1.TestObject{
		initObjectNamespaced,
		{
			Meta: &machineryv1.ObjectMeta{
				Name:       "finalizer-obj",
				Namespace:  "test",
				Finalizers: []string{"wait-for-me"},
			},
		},
		{
			Meta: &machineryv1.ObjectMeta{
				Name:      "test-obj-1",
				Namespace: "test",
				Labels: map[string]string{
					"test": "1",
				},
			},
		},
		{
			Meta: &machineryv1.ObjectMeta{
				Name:      "test-obj-2",
				Namespace: "test",
			},
		},
	}
	for _, obj := range createObjects {
		require.NoError(t, s.Create(ctx, obj))

		// ensure metadata is set
		assert.NotEmpty(t, obj.Meta.CreatedTimestamp)
		assert.Equal(t, int64(1), obj.Meta.Generation)
		assert.NotEmpty(t, obj.Meta.ResourceVersion)
		assert.NotEmpty(t, obj.Meta.Uid)
	}

	// Start Watch
	// -----------
	watcher, err := s.Watch(ctx, &testdatav1.TestObject{})
	require.NoError(t, err)
	defer watcher.Close()

	var (
		eventsWG sync.WaitGroup
		events   []api.ResourceEvent
	)
	eventsWG.Add(13) // we expect 13 events from start to finish
	go func() {
		for event := range watcher.Events() {
			events = append(events, event)
			eventsWG.Done()
		}
	}()

	// Create (generateName)
	// ------
	generateNameObj := &testdatav1.TestObject{
		Meta: &machineryv1.ObjectMeta{
			GenerateName: "test-",
			Namespace:    "test",
		},
	}
	require.NoError(t, s.Create(ctx, generateNameObj))
	assert.NotEmpty(t, generateNameObj.Meta.Name)

	// List
	// ----

	// list everything
	listAll := &testdatav1.TestObjectList{}
	require.NoError(t, s.List(ctx, listAll))

	assert.Len(t, listAll.Items, len(createObjects)+1)

	// list in namespace "test"
	listNamespaced := &testdatav1.TestObjectList{}
	require.NoError(t, s.List(ctx, listNamespaced, api.InNamespace("test")))

	assert.Len(t, listNamespaced.Items, 5)

	// list by label match
	listWithLabel := &testdatav1.TestObjectList{}
	require.NoError(t, s.List(ctx, listWithLabel, api.MatchLabels{"test": "1"}))

	assert.Len(t, listWithLabel.Items, 1)

	// Get
	// ---

	// namespaced
	namespacedObj := &testdatav1.TestObject{}
	require.NoError(t, s.Get(ctx, api.NamespacedName{
		Name:      "test-namespaced",
		Namespace: "test",
	}, namespacedObj))
	if !proto.Equal(initObjectNamespaced, namespacedObj) {
		t.Error("wrong object when getting namespaced object")
	}

	// Update
	// ------
	namespacedObj.Field1 = "setting-field-1"
	namespacedObj.Field2 = "setting-field-2"
	namespacedObj.Status = &testdatav1.TestObjectStatus{
		Status: "ok",
	}

	require.NoError(t, s.Update(ctx, namespacedObj))
	require.NoError(t, s.Get(ctx, api.NamespacedName{
		Name:      "test-namespaced",
		Namespace: "test",
	}, namespacedObj))
	assert.Equal(t, "setting-field-1", namespacedObj.Field1)
	assert.Equal(t, "setting-field-2", namespacedObj.Field2)
	assert.Empty(t, namespacedObj.Status) // Status is not set!

	// UpdateStatus
	// ------------
	namespacedObj.Field1 = "updating-field-1"
	namespacedObj.Field2 = "updating-field-2"
	namespacedObj.Status = &testdatav1.TestObjectStatus{
		Status: "ok",
	}

	require.NoError(t, s.UpdateStatus(ctx, namespacedObj))
	require.NoError(t, s.Get(ctx, api.NamespacedName{
		Name:      "test-namespaced",
		Namespace: "test",
	}, namespacedObj))
	assert.Equal(t, "setting-field-1", namespacedObj.Field1) // not updated
	assert.Equal(t, "setting-field-2", namespacedObj.Field2) // not updated
	assert.NotEmpty(t, namespacedObj.Status)

	// Delete
	// ------
	require.NoError(t, s.Delete(ctx, namespacedObj))
	assert.EqualError(t, s.Get(ctx, api.NamespacedName{
		Name:      "test-namespaced",
		Namespace: "test",
	}, namespacedObj), "machinery.testdata.v1.TestObject test/test-namespaced: not found")

	// delete with finalizer
	finalizerObj := &testdatav1.TestObject{}
	require.NoError(t, s.Get(ctx, api.NamespacedName{
		Name:      "finalizer-obj",
		Namespace: "test",
	}, finalizerObj))
	require.NoError(t, s.Delete(ctx, finalizerObj))
	assert.NotEmpty(t, finalizerObj.Meta.DeletedTimestamp)

	finalizerObj.Meta.Finalizers = nil
	require.NoError(t, s.Update(ctx, finalizerObj))
	assert.EqualError(t, s.Get(ctx, api.NamespacedName{
		Name:      "finalizer-obj",
		Namespace: "test",
	}, finalizerObj), "machinery.testdata.v1.TestObject test/finalizer-obj: not found")

	// DeleteAllOf
	// -----------

	// delete all namespaced objects
	require.NoError(t, s.DeleteAllOf(ctx, &testdatav1.TestObject{}, api.InNamespace("test")))

	namespacedListAfterDeletion := &testdatav1.TestObjectList{}
	require.NoError(t, s.List(ctx, namespacedListAfterDeletion, api.InNamespace("test")))
	assert.Len(t, namespacedListAfterDeletion.Items, 0)

	// Check Events
	// ------------

	// wait for all events to be received
	eventsWG.Wait()
	watcher.Close()

	require.Len(t, events, 13)
	// first 5 events should be generated added events from storage
	for i := 0; i < 5; i++ {
		assert.Equal(t, api.Added, events[i].Type)
	}

	// Update test
	assertEventTypeAndNamespaceName(t, api.NamespacedName{
		Name: "test-namespaced", Namespace: "test",
	}, api.Modified, events[5])

	// StatusUpdate test
	assertEventTypeAndNamespaceName(t, api.NamespacedName{
		Name: "test-namespaced", Namespace: "test",
	}, api.Modified, events[6])

	// Delete test
	assertEventTypeAndNamespaceName(t, api.NamespacedName{
		Name: "test-namespaced", Namespace: "test",
	}, api.Deleted, events[7])

	// Delete finalizer-obj
	assertEventTypeAndNamespaceName(t, api.NamespacedName{
		Name: "finalizer-obj", Namespace: "test",
	}, api.Modified, events[8]) // "deleted" but waiting for finalizer
	assertEventTypeAndNamespaceName(t, api.NamespacedName{
		Name: "finalizer-obj", Namespace: "test",
	}, api.Deleted, events[9]) // deleted, -> finalizer removed

	// Rest of Namespaced object (DeleteAllOf)
	for i := 0; i < 3; i++ {
		assert.Equal(t, api.Deleted, events[i+10].Type)
	}
}

func assertEventTypeAndNamespaceName(
	t *testing.T, nn api.NamespacedName, et api.ResourceEventType, e api.ResourceEvent) {
	if e.Type != et {
		t.Errorf("event type should be %q, is %q, event: %v", et, e.Type, e)
	}
	name := e.Object.ObjectMeta().GetName()
	if name != nn.Name {
		t.Errorf("object name should be %q, is %q, event: %v", nn.Name, name, e)
	}
	namespace := e.Object.ObjectMeta().GetNamespace()
	if namespace != nn.Namespace {
		t.Errorf("object namespace should be %q, is %q, event: %v", nn.Namespace, namespace, e)
	}
}
