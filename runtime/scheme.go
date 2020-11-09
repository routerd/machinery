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

import (
	"fmt"
	"reflect"
	"sort"
	"strings"

	"routerd.net/machinery/runtime/convert"
)

// Scheme holds references to known types.
type Scheme struct {
	converter       *convert.Converter
	gvkToType       map[GroupVersionKind]reflect.Type
	typeToGVK       map[reflect.Type]GroupVersionKind
	versionPriority map[string][]string
}

func NewScheme() *Scheme {
	return &Scheme{
		converter:       convert.NewConverter(),
		gvkToType:       map[GroupVersionKind]reflect.Type{},
		typeToGVK:       map[reflect.Type]GroupVersionKind{},
		versionPriority: map[string][]string{},
	}
}

func (s *Scheme) Convert(src, dest interface{}) error {
	return s.converter.Convert(src, dest)
}

func (s *Scheme) RegisterConversionFunc(conversionFunc interface{}) error {
	return s.converter.RegisterConversionFunc(conversionFunc)
}

func (s *Scheme) AddKnownTypes(gv GroupVersion, types ...Object) {
	for _, obj := range types {
		rt := dereferenceType(obj)
		s.AddKnownTypeWithKind(GroupVersionKind{
			Group:   gv.Group,
			Version: gv.Version,
			Kind:    rt.Name(),
		}, obj)
	}
}

func (s *Scheme) AddKnownTypeWithKind(gvk GroupVersionKind, obj Object) {
	if len(gvk.Version) == 0 {
		panic("Version is required on all types.")
	}
	rt := dereferenceType(obj)

	s.gvkToType[gvk] = rt
	s.typeToGVK[rt] = gvk
	s.refreshVersionPriority()
}

// New creates a new instance for the given GroupVersionKind
func (s *Scheme) New(gvk GroupVersionKind) (Object, error) {
	if rt, exists := s.gvkToType[gvk]; exists {
		new := reflect.New(rt).Interface().(Object)
		new.GetObjectKind().SetGroupVersionKind(gvk)
		return new, nil
	}
	return nil, fmt.Errorf("kind %s is not registered", gvk)
}

func (s *Scheme) GroupVersionKind(obj Object) (GroupVersionKind, error) {
	rt := dereferenceType(obj)
	if gvk, ok := s.typeToGVK[rt]; ok {
		return gvk, nil
	}
	return GroupVersionKind{}, fmt.Errorf("object %T is not registered", obj)
}

func (s *Scheme) ListGroupVersionKind(obj Object) (GroupVersionKind, error) {
	gvk, err := s.GroupVersionKind(obj)
	if err != nil {
		return GroupVersionKind{}, err
	}

	listGVK := GroupVersionKind{
		Group:   gvk.Group,
		Version: gvk.Version,
		Kind:    gvk.Kind,
	}
	if !strings.HasSuffix(listGVK.Kind, "List") {
		listGVK.Kind = listGVK.Kind + "List"
	}
	if _, ok := s.gvkToType[listGVK]; !ok {
		return GroupVersionKind{},
			fmt.Errorf("no list type for %s is registered", gvk)
	}
	return listGVK, nil
}

func (s *Scheme) KnownObjectKinds() []GroupVersionKind {
	var gvks []GroupVersionKind
	for gvk := range s.gvkToType {
		if strings.HasSuffix(gvk.Kind, "List") {
			continue
		}
		if gvk.Version != s.versionPriority[gvk.Group][0] {
			// skip versions not of highest priority
			continue
		}
		gvks = append(gvks, gvk)
	}
	return gvks
}

// dereferenceType returns the Value-Type of given pointer.
func dereferenceType(obj Object) reflect.Type {
	rt := reflect.TypeOf(obj)
	if rt.Kind() != reflect.Ptr {
		panic("All types must be pointers to structs.")
	}

	rt = rt.Elem()
	if rt.Kind() != reflect.Struct {
		panic("All types must be pointers to structs.")
	}
	return rt
}

func (s *Scheme) refreshVersionPriority() {
	groupVersions := map[string]APIVersionByPriority{}
	for gvk := range s.gvkToType {
		version, err :=
			ParseAPIVersion(gvk.Version)
		if err != nil {
			panic(err)
		}
		groupVersions[gvk.Group] = append(groupVersions[gvk.Group], version)
	}

	for group, v := range groupVersions {
		sort.Sort(v)
		s.versionPriority[group] = make([]string, len(v))
		for i, apiVersion := range v {
			s.versionPriority[group][i] = apiVersion.String()
		}
	}
}
