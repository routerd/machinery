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

package admission

import (
	"context"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	machinerv1 "routerd.net/machinery/api/v1"
	testdatav1 "routerd.net/machinery/testdata/v1"
)

type defaultedTestObject struct {
	*testdatav1.TestObject
	mock.Mock
}

func (o *defaultedTestObject) Default(ctx context.Context) error {
	args := o.Called(ctx)
	err, _ := args.Error(0).(error)
	return err
}

func TestGenericDefault(t *testing.T) {
	t.Run("OnCreate calls .Default()", func(t *testing.T) {
		d := &GenericDefault{}
		ctx := context.Background()
		obj := &defaultedTestObject{
			TestObject: &testdatav1.TestObject{},
		}

		obj.On("Default", ctx).Return(nil)

		err := d.OnCreate(ctx, obj)
		require.NoError(t, err)

		obj.AssertExpectations(t)
	})

	t.Run("OnCreate sets .name from .generateName", func(t *testing.T) {
		d := &GenericDefault{}
		ctx := context.Background()
		obj := &testdatav1.TestObject{
			Meta: &machinerv1.ObjectMeta{
				Name:         "",
				GenerateName: "test-",
			},
		}

		err := d.OnCreate(ctx, obj)
		require.NoError(t, err)

		if assert.NotEmpty(t, obj.Meta.Name) {
			assert.True(t,
				strings.HasPrefix(obj.Meta.Name, obj.Meta.GenerateName))
		}
	})

	t.Run("OnUpdate calls .Default()", func(t *testing.T) {
		d := &GenericDefault{}
		ctx := context.Background()
		obj := &defaultedTestObject{
			TestObject: &testdatav1.TestObject{},
		}

		obj.On("Default", ctx).Return(nil)

		err := d.OnUpdate(ctx, obj)
		require.NoError(t, err)

		obj.AssertExpectations(t)
	})

	t.Run("OnDelete calls .Default()", func(t *testing.T) {
		d := &GenericDefault{}
		ctx := context.Background()
		obj := &defaultedTestObject{
			TestObject: &testdatav1.TestObject{},
		}

		obj.On("Default", ctx).Return(nil)

		err := d.OnDelete(ctx, obj)
		require.NoError(t, err)

		obj.AssertExpectations(t)
	})
}
