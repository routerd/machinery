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

package meta

type ObjectMeta struct {
	Name            string            `json:"name"`
	Namespace       string            `json:"namespace"`
	UID             string            `json:"uid"`
	Generation      int64             `json:"generation"`
	ResourceVersion string            `json:"resourceVersion"`
	Annotations     map[string]string `json:"annotations"`
}

func (m *ObjectMeta) GetName() string     { return m.Name }
func (m *ObjectMeta) SetName(name string) { m.Name = name }

func (m *ObjectMeta) GetNamespace() string          { return m.Namespace }
func (m *ObjectMeta) SetNamespace(namespace string) { m.Namespace = namespace }

func (m *ObjectMeta) GetUID() string    { return m.UID }
func (m *ObjectMeta) SetUID(uid string) { m.UID = uid }

func (m *ObjectMeta) GetGeneration() int64           { return m.Generation }
func (m *ObjectMeta) SetGeneration(generation int64) { m.Generation = generation }

func (m *ObjectMeta) GetResourceVersion() string { return m.ResourceVersion }
func (m *ObjectMeta) SetResourceVersion(resourceVersion string) {
	m.ResourceVersion = resourceVersion
}

func (m *ObjectMeta) GetAnnotations() map[string]string {
	return m.Annotations
}
func (m *ObjectMeta) SetAnnotations(annotations map[string]string) {
	m.Annotations = annotations
}

type Object interface {
	GetName() string
	SetName(name string)
	GetNamespace() string
	SetNamespace(namespace string)
	GetGeneration() int64
	SetGeneration(generation int64)
	GetResourceVersion() string
	SetResourceVersion(resourceVersion string)
	GetAnnotations() map[string]string
	SetAnnotations(annotations map[string]string)
}
