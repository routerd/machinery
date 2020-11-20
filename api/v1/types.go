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

import "google.golang.org/protobuf/types/known/timestamppb"

func (m *ObjectMeta) SetLabel(labels map[string]string) {
	m.Labels = labels
}

func (m *ObjectMeta) SetUID(uid string) {
	m.Uid = uid
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

type NamespacedName struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}
