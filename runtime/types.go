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

package runtime

type GroupVersionKind struct {
	Group   string `json:"group"`
	Version string `json:"version"`
	Kind    string `json:"kind"`
}

func (gvk GroupVersionKind) Empty() bool {
	return len(gvk.Version) == 0 && len(gvk.Kind) == 0
}

func (gvk GroupVersionKind) String() string {
	return gvk.Group + "." + gvk.Kind + "/" + gvk.Version
}

func (gvk GroupVersionKind) GroupVersion() GroupVersion {
	return GroupVersion{Group: gvk.Group, Version: gvk.Version}
}

type GroupVersion struct {
	Group   string `json:"group"`
	Version string `json:"version"`
}

func (gv GroupVersion) String() string {
	return gv.Group + "/" + gv.Version
}

type ObjectKind interface {
	SetGroupVersionKind(vk GroupVersionKind)
	GetGroupVersionKind() GroupVersionKind
}

type Object interface {
	GetObjectKind() ObjectKind
	DeepCopyObject() Object
}
