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

package api

import "context"

type Client interface {
	Reader
	Watcher
	Writer
}

// Reader allows to get objects from storage or cache.
type Reader interface {
	Getter
	Lister
}

type Getter interface {
	Get(ctx context.Context, nn NamespacedName, obj Object) error
}

type Lister interface {
	List(ctx context.Context, listObj ListObject, opts ...ListOption) error
}

// Writer provides write methods for storage access.
type Writer interface {
	Create(ctx context.Context, obj Object, opts ...CreateOption) error
	Delete(ctx context.Context, obj Object, opts ...DeleteOption) error
	DeleteAllOf(ctx context.Context, obj Object, opts ...DeleteAllOfOption) error
	Update(ctx context.Context, obj Object, opts ...UpdateOption) error
	UpdateStatus(ctx context.Context, obj Object, opts ...UpdateOption) error
}

type WatchClient interface {
	Close() error
	Events() <-chan ResourceEvent
}

// Watcher can be used to watch for to the specified object type.
type Watcher interface {
	Watch(ctx context.Context,
		obj Object, opts ...WatchOption) (WatchClient, error)
}
