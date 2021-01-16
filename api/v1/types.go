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
	"google.golang.org/protobuf/types/known/timestamppb"

	"routerd.net/machinery/api"
)

// ObjectMeta setters

func (m *ObjectMeta) SetUid(uid string) {
	m.Uid = uid
}

func (m *ObjectMeta) SetNamespace(namespace string) {
	m.Namespace = namespace
}

func (m *ObjectMeta) SetName(name string) {
	m.Name = name
}

func (m *ObjectMeta) SetLabels(labels map[string]string) {
	m.Labels = labels
}

func (m *ObjectMeta) SetAnnotations(annotations map[string]string) {
	m.Annotations = annotations
}

func (m *ObjectMeta) SetResourceVersion(resourceVersion string) {
	m.ResourceVersion = resourceVersion
}

func (m *ObjectMeta) SetFinalizers(finalizers []string) {
	m.Finalizers = finalizers
}

func (m *ObjectMeta) SetDeletedTimestamp(timestamp *timestamppb.Timestamp) {
	m.DeletedTimestamp = timestamp
}

func (m *ObjectMeta) SetCreatedTimestamp(timestamp *timestamppb.Timestamp) {
	m.CreatedTimestamp = timestamp
}

func (m *ObjectMeta) SetGeneration(generation int64) {
	m.Generation = generation
}

func (m *ObjectMeta) SetGenerateName(generateName string) {
	m.GenerateName = generateName
}

// ListMeta setters

func (m *ListMeta) SetResourceVersion(resourceVersion string) {
	m.ResourceVersion = resourceVersion
}

// OwnerReference setters

func (m *OwnerReference) SetTypeURL(typeURL string) {
	m.TypeURL = typeURL
}

func (m *OwnerReference) SetUid(uid string) {
	m.Uid = uid
}

func (m *OwnerReference) SetNamespace(namespace string) {
	m.Namespace = namespace
}

func (m *OwnerReference) SetName(name string) {
	m.Name = name
}

type ObjectMetaAdapter struct {
	*ObjectMeta
}

func (m *ObjectMetaAdapter) GetOwnerReferences() []api.OwnerReference {
	apiOwnerReferences := make(
		[]api.OwnerReference, len(m.OwnerReferences))
	for _, or := range m.OwnerReferences {
		apiOwnerReferences = append(apiOwnerReferences, or)
	}
	return apiOwnerReferences
}

func (m *ObjectMetaAdapter) SetOwnerReferences(
	ownerReferences []api.OwnerReference) {
	v1OwnerReferences := make(
		[]*OwnerReference, len(ownerReferences))
	for _, or := range ownerReferences {
		v1OwnerReferences = append(v1OwnerReferences, or.(*OwnerReference))
	}
	m.OwnerReferences = v1OwnerReferences
}
