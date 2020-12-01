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
	"testing"

	"github.com/stretchr/testify/require"

	machineryv1 "routerd.net/machinery/api/v1"
	testdatav1 "routerd.net/machinery/testdata/v1"
)

func TestNamespacedStorage(t *testing.T) {
	s := NewNamespacedStorage(&testdatav1.Namespace{})

	nsStorage := NewInMemoryStorage(&testdatav1.Namespace{})
	s.RegisterStorage(&testdatav1.Namespace{}, nsStorage, false)

	testObjStorage := NewInMemoryStorage(&testdatav1.TestObject{})
	s.RegisterStorage(&testdatav1.TestObject{}, testObjStorage, true)

	// Run
	stopCh := make(chan struct{})
	go nsStorage.Run(stopCh)
	go testObjStorage.Run(stopCh)
	defer close(stopCh)

	// Create namespace test
	ctx := context.Background()
	require.NoError(t, s.Create(ctx, &testdatav1.Namespace{
		Meta: &machineryv1.ObjectMeta{
			Name: "test",
		},
	}))
	StorageTestSuite(t, s)

	// Create an Object in a non-existing Namespace
	require.EqualError(t, s.Create(ctx, &testdatav1.TestObject{
		Meta: &machineryv1.ObjectMeta{
			Name:      "test",
			Namespace: "does-not-exist",
		},
	}), "machinery.testdata.v1.Namespace does-not-exist: not found")
}
