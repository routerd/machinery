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

package utils

import (
	"fmt"

	"routerd.net/machinery/runtime"
)

type GVKGate struct {
	scheme          *runtime.Scheme
	ObjGVK, ListGVK runtime.GroupVersionKind
}

func NewGVKGate(scheme *runtime.Scheme, obj runtime.Object) (*GVKGate, error) {
	var err error
	g := &GVKGate{
		scheme: scheme,
	}

	g.ObjGVK, err = scheme.GroupVersionKind(obj)
	if err != nil {
		return nil, err
	}

	g.ListGVK, err = scheme.ListGroupVersionKind(obj)
	if err != nil {
		return nil, err
	}
	return g, nil
}

func (g *GVKGate) CheckObject(obj runtime.Object) error {
	objGVK, err := g.scheme.GroupVersionKind(obj)
	if err != nil {
		return err
	}

	if objGVK != g.ObjGVK {
		return fmt.Errorf("invalid GVK for Object, expected: %s", g.ObjGVK)
	}
	return nil
}

func (g *GVKGate) CheckList(obj runtime.Object) error {
	listGVK, err := g.scheme.ListGroupVersionKind(obj)
	if err != nil {
		return err
	}

	if listGVK != g.ListGVK {
		return fmt.Errorf("invalid GVK for List, expected: %s", g.ListGVK)
	}
	return nil
}
