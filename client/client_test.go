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
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"

	"routerd.net/machinery/api"
	machineryv1 "routerd.net/machinery/api/v1"
	testdatav1 "routerd.net/machinery/testdata/v1"
)

func TestGRPCClient(t *testing.T) {
	apiObject := &testdatav1.TestObject{
		Meta: &machineryv1.ObjectMeta{
			Name:      "test",
			Namespace: "test",
		},
		Field1: "lorem",
		Field2: "ipsum",
	}

	apiObjectList := &testdatav1.TestObjectList{
		Items: []*testdatav1.TestObject{
			{
				Meta: &machineryv1.ObjectMeta{
					Name:      "test",
					Namespace: "test",
				},
				Field1: "lorem",
				Field2: "ipsum",
			},
		},
	}

	t.Run("Get", func(t *testing.T) {
		// Init
		m := &testdatav1.TestObjectServiceClientMock{}
		c, err := NewGRPCClient(m)
		require.NoError(t, err)

		// Mock Setup
		m.
			On("Get", mock.Anything, mock.Anything, mock.Anything).
			Return(apiObject, nil)

		// Test
		ctx := context.Background()
		obj := &testdatav1.TestObject{}
		err = c.Get(ctx, api.NamespacedName{
			Name: "test", Namespace: "test-ns",
		}, obj)
		require.NoError(t, err)

		assert.True(t, proto.Equal(apiObject, obj), "obj and apiObject should be equal")
		m.AssertCalled(t, "Get", ctx, &machineryv1.GetRequest{
			Name: "test", Namespace: "test-ns",
		}, mock.Anything)
	})

	t.Run("List", func(t *testing.T) {
		// Init
		m := &testdatav1.TestObjectServiceClientMock{}
		c, err := NewGRPCClient(m)
		require.NoError(t, err)

		// Mock Setup
		m.
			On("List", mock.Anything, mock.Anything, mock.Anything).
			Return(apiObjectList, nil)

		// Test
		ctx := context.Background()
		obj := &testdatav1.TestObjectList{}
		err = c.List(ctx, obj,
			api.InNamespace("test"), api.MatchLabels{"test": "test"})
		require.NoError(t, err)

		assert.True(t, proto.Equal(apiObjectList, obj), "obj and apiObject should be equal")
		m.AssertCalled(t, "List", ctx, &machineryv1.ListRequest{
			Namespace:     "test",
			LabelSelector: "test=test",
		}, mock.Anything)
	})

	t.Run("Delete", func(t *testing.T) {
		// Init
		m := &testdatav1.TestObjectServiceClientMock{}
		c, err := NewGRPCClient(m)
		require.NoError(t, err)

		// Mock Setup
		m.
			On("Delete", mock.Anything, mock.Anything, mock.Anything).
			Return(apiObject, nil)

		// Test
		ctx := context.Background()
		obj := &testdatav1.TestObject{}
		err = c.Delete(ctx, obj)
		require.NoError(t, err)

		assert.True(t, proto.Equal(apiObject, obj), "obj and apiObject should be equal")
		m.AssertCalled(t, "Delete", ctx, &testdatav1.TestObjectDeleteRequest{
			Object: obj,
		}, mock.Anything)
	})

	t.Run("Update", func(t *testing.T) {
		// Init
		m := &testdatav1.TestObjectServiceClientMock{}
		c, err := NewGRPCClient(m)
		require.NoError(t, err)

		// Mock Setup
		m.
			On("Update", mock.Anything, mock.Anything, mock.Anything).
			Return(apiObject, nil)

		// Test
		ctx := context.Background()
		obj := &testdatav1.TestObject{}
		err = c.Update(ctx, obj)
		require.NoError(t, err)

		assert.True(t, proto.Equal(apiObject, obj), "obj and apiObject should be equal")
		m.AssertCalled(t, "Update", ctx, &testdatav1.TestObjectUpdateRequest{
			Object: obj,
		}, mock.Anything)
	})

	t.Run("UpdateStatus", func(t *testing.T) {
		// Init
		m := &testdatav1.TestObjectServiceClientMock{}
		c, err := NewGRPCClient(m)
		require.NoError(t, err)

		// Mock Setup
		m.
			On("UpdateStatus", mock.Anything, mock.Anything, mock.Anything).
			Return(apiObject, nil)

		// Test
		ctx := context.Background()
		obj := &testdatav1.TestObject{}
		err = c.UpdateStatus(ctx, obj)
		require.NoError(t, err)

		assert.True(t, proto.Equal(apiObject, obj), "obj and apiObject should be equal")
		m.AssertCalled(t, "UpdateStatus", ctx, &testdatav1.TestObjectUpdateRequest{
			Object: obj,
		}, mock.Anything)
	})

	t.Run("DeleteAllOf", func(t *testing.T) {
		// Init
		m := &testdatav1.TestObjectServiceClientMock{}
		c, err := NewGRPCClient(m)
		require.NoError(t, err)

		// Mock Setup
		m.
			On("DeleteAllOf", mock.Anything, mock.Anything, mock.Anything).
			Return((*machineryv1.Status)(nil), nil)

		// Test
		ctx := context.Background()
		err = c.DeleteAllOf(ctx, &testdatav1.TestObject{},
			api.InNamespace("test"), api.MatchLabels{"test": "test"})
		require.NoError(t, err)

		m.AssertCalled(t, "DeleteAllOf", ctx, &machineryv1.DeleteAllOfRequest{
			Namespace:     "test",
			LabelSelector: "test=test",
		}, mock.Anything)
	})
}
