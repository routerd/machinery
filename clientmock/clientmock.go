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

package clientmock

import (
	"context"

	"github.com/stretchr/testify/mock"

	"routerd.net/machinery/api"
)

type Reader struct {
	mock.Mock
}

func (m *Reader) Get(ctx context.Context, nn api.NamespacedName, obj api.Object) error {
	args := m.Called(ctx, nn, obj)
	err, _ := args.Error(0).(error)
	return err
}

func (m *Reader) List(ctx context.Context, listObj api.ListObject, opts ...api.ListOption) error {
	args := m.Called(ctx, listObj, opts)
	err, _ := args.Error(0).(error)
	return err
}

type Getter struct {
	mock.Mock
}

func (m *Getter) Get(ctx context.Context, nn api.NamespacedName, obj api.Object) error {
	args := m.Called(ctx, nn, obj)
	err, _ := args.Error(0).(error)
	return err
}

type Lister struct {
	mock.Mock
}

func (m *Lister) List(ctx context.Context, listObj api.ListObject, opts ...api.ListOption) error {
	args := m.Called(ctx, listObj, opts)
	err, _ := args.Error(0).(error)
	return err
}
