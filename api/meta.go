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

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"
)

type Object interface {
	proto.Message
	ObjectMeta() ObjectMeta
}

type ObjectMeta interface {
	// GetName returns a unique, user defined ID for the object.
	GetName() string
	SetName(string)
	// GetNamespace returns the "keyspace" or "folder" of this object.
	GetNamespace() string
	SetNamespace(string)

	// GetLabels returns a set of labels that can be used to filter and group objects.
	GetLabels() map[string]string
	SetLabels(map[string]string)

	GetAnnotations() map[string]string
	SetAnnotations(map[string]string)

	// GetUid returns the UUID of the object.
	// UUIDs are added by the storage implementation and ensures that
	// objects with the same name and namespace, but different livecycle
	// can be differentiated.
	GetUid() string
	SetUid(string)

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

	GetGenerateName() string
	SetGenerateName(string)

	GetOwnerReferences() []OwnerReference
	SetOwnerReferences([]OwnerReference)
}

type OwnerReference interface {
	SetTypeURL(string)
	GetTypeURL() string

	GetNamespace() string
	SetNamespace(string)

	GetName() string
	SetName(string)

	GetUid() string
	SetUid(string)
}

type ListObject interface {
	proto.Message
	ListMeta() ListMeta
}

type ListMeta interface {
	GetResourceVersion() string
	SetResourceVersion(string)
}
