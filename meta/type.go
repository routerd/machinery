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

import "routerd.net/machinery/runtime"

type TypeMeta struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

func (m *TypeMeta) GetObjectKind() runtime.ObjectKind { return m }
func (m *TypeMeta) GetKind() string                   { return m.Kind }
func (m *TypeMeta) SetKind(kind string)               { m.Kind = kind }
func (m *TypeMeta) GetVersion() string                { return m.Version }
func (m *TypeMeta) SetVersion(version string)         { m.Version = version }

func (m *TypeMeta) GetGroupVersionKind() runtime.GroupVersionKind {
	return runtime.GroupVersionKind{Group: m.Group, Version: m.Version, Kind: m.Kind}
}

func (m *TypeMeta) SetGroupVersionKind(gvk runtime.GroupVersionKind) {
	m.Version = gvk.Version
	m.Kind = gvk.Kind
	m.Group = gvk.Group
}
