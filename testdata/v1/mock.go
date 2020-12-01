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

package v1

import (
	context "context"

	"github.com/stretchr/testify/mock"
	grpc "google.golang.org/grpc"

	v1 "routerd.net/machinery/api/v1"
)

var _ TestObjectServiceClient = (*TestObjectServiceClientMock)(nil)

type TestObjectServiceClientMock struct {
	mock.Mock
}

func (m *TestObjectServiceClientMock) Get(
	ctx context.Context,
	in *v1.GetRequest,
	opts ...grpc.CallOption,
) (*TestObject, error) {
	args := m.Called(ctx, in, opts)
	err, _ := args.Get(1).(error)
	return args.Get(0).(*TestObject), err
}

func (m *TestObjectServiceClientMock) List(ctx context.Context, in *v1.ListRequest, opts ...grpc.CallOption) (*TestObjectList, error) {
	args := m.Called(ctx, in, opts)
	err, _ := args.Get(1).(error)
	return args.Get(0).(*TestObjectList), err
}

func (m *TestObjectServiceClientMock) Watch(ctx context.Context, in *v1.WatchRequest, opts ...grpc.CallOption) (TestObjectService_WatchClient, error) {
	args := m.Called(ctx, in, opts)
	err, _ := args.Get(1).(error)
	return args.Get(0).(TestObjectService_WatchClient), err
}

func (m *TestObjectServiceClientMock) Create(ctx context.Context, in *TestObjectCreateRequest, opts ...grpc.CallOption) (*TestObject, error) {
	args := m.Called(ctx, in, opts)
	err, _ := args.Get(1).(error)
	return args.Get(0).(*TestObject), err
}

func (m *TestObjectServiceClientMock) Update(ctx context.Context, in *TestObjectUpdateRequest, opts ...grpc.CallOption) (*TestObject, error) {
	args := m.Called(ctx, in, opts)
	err, _ := args.Get(1).(error)
	return args.Get(0).(*TestObject), err
}

func (m *TestObjectServiceClientMock) UpdateStatus(ctx context.Context, in *TestObjectUpdateRequest, opts ...grpc.CallOption) (*TestObject, error) {
	args := m.Called(ctx, in, opts)
	err, _ := args.Get(1).(error)
	return args.Get(0).(*TestObject), err
}

func (m *TestObjectServiceClientMock) Delete(ctx context.Context, in *TestObjectDeleteRequest, opts ...grpc.CallOption) (*TestObject, error) {
	args := m.Called(ctx, in, opts)
	err, _ := args.Get(1).(error)
	return args.Get(0).(*TestObject), err
}

func (m *TestObjectServiceClientMock) DeleteAllOf(ctx context.Context, in *v1.DeleteAllOfRequest, opts ...grpc.CallOption) (*v1.Status, error) {
	args := m.Called(ctx, in, opts)
	err, _ := args.Get(1).(error)
	return args.Get(0).(*v1.Status), err
}
