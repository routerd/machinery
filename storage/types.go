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

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// Client implements all storage interfaces.
type Client interface {
	Reader
	Watcher
	Writer
}

// Reader provides read methods for storage access.
type Reader interface {
	Get(ctx context.Context, name, namespace string, obj Object) error
	List(ctx context.Context, obj ListObject, opts ...ListOption) error
}

type WatchClient interface {
	Close() error
	Events() <-chan Event
}

// Watcher can be used to watch for to the specified object type.
type Watcher interface {
	Watch(ctx context.Context,
		obj Object, opts ...ListOption) (WatchClient, error)
}

// Writer provides write methods for storage access.
type Writer interface {
	Create(ctx context.Context, obj Object, opts ...CreateOption) error
	Delete(ctx context.Context, obj Object, opts ...DeleteOption) error
	Update(ctx context.Context, obj Object, opts ...UpdateOption) error
	UpdateStatus(ctx context.Context, obj Object, opts ...UpdateOption) error
}

// Objects can be persistet in a storage.
type Object interface {
	proto.Message
	ObjectMetaAccessor() ObjectMetaAccessor
}

type ObjectMetaAccessor interface {
	// GetName returns a unique, user defined ID for the object.
	GetName() string
	// GetNamespace returns the "keyspace" or "folder" of this object.
	GetNamespace() string

	// GetLabels returns a set of labels that can be used to filter and group objects.
	GetLabels() map[string]string
	SetLabels(map[string]string)

	// GetUID returns the UUID of the object.
	// UUIDs are added by the storage implementation and ensures that
	// objects with the same name and namespace, but different livecycle
	// can be differentiated.
	GetUID() string
	SetUID(string)

	// ResourceVersion keeps track of the change revision of an object.
	GetResourceVersion() string
	SetResourceVersion(string)

	// Finalizers delay the actual removal of an object from storage after it's deletion.
	// When an object has outstanding finalizers, the storage will just mark the object as deleted by setting the DeletedTimestamp.
	// The object will be automatically deleted from storage when the last finalizer has been removed.
	GetFinalizers() []string
	SetFinalizers([]string)

	// DeletedTimestamp marks ab object as deleted. Only used in conjunction with finalizers.
	SetDeletedTimestamp(timestamp *timestamppb.Timestamp)
	GetDeletedTimestamp() *timestamppb.Timestamp

	// CreatedTimestamp tracks the time when this object was submitted to storage.
	SetCreatedTimestamp(timestamp *timestamppb.Timestamp)
	GetCreatedTimestamp() *timestamppb.Timestamp

	// Generation is incremented for every non-status update to this object.
	GetGeneration() int64
	SetGeneration(int64)
}

type ListObject interface {
	proto.Message
	SetItems([]proto.Message)
	GetItems() proto.Message
}

// Event is emitted for every change of an object in storage.
type Event struct {
	Type   EventType
	Object Object
}

type EventType string

const (
	Added    EventType = "Added"
	Modified EventType = "Modified"
	Deleted  EventType = "Deleted"
	Error    EventType = "Error"
)
