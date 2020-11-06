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

	"sigs.k8s.io/controller-runtime/pkg/client"

	"routerd.net/machinery/runtime"
)

type WatchClient interface {
	Close() error
	EventChan() <-chan Event
}

type Reader interface {
	Get(ctx context.Context, name string, obj runtime.Object) error
	List(ctx context.Context, obj runtime.Object, opts ...ListOption) error
}

type Watcher interface {
	Watch(ctx context.Context,
		obj runtime.Object, opts ...ListOption) (WatchClient, error)
}

type Writer interface {
	Create(ctx context.Context, obj runtime.Object, opts ...client.CreateOption) error
	Delete(ctx context.Context, obj runtime.Object, opts ...client.DeleteOption) error
	Update(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error
	UpdateStatus(ctx context.Context, obj runtime.Object, opts ...client.UpdateOption) error
}
