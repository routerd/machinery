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
	"strings"

	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/reflect/protoreflect"
	"google.golang.org/protobuf/reflect/protoregistry"

	"routerd.net/machinery/api"
)

// Scheme serves as a central type registry and interacts with known object types.
type Scheme struct {
	global     map[protoreflect.FullName]struct{}
	namespaced map[protoreflect.FullName]struct{}
}

func NewScheme() *Scheme {
	return &Scheme{
		global:     map[protoreflect.FullName]struct{}{},
		namespaced: map[protoreflect.FullName]struct{}{},
	}
}

func (s *Scheme) RegisterGlobal(objs ...api.Object) {
	for _, obj := range objs {
		s.global[s.FullName(obj)] = struct{}{}
	}
}

func (s *Scheme) RegisterNamespaced(objs ...api.Object) {
	for _, obj := range objs {
		s.namespaced[s.FullName(obj)] = struct{}{}
	}
}

func (s *Scheme) Global() []api.Object {
	var objs []api.Object
	for name := range s.global {
		obj, err := s.NewObject(name)
		if err != nil {
			panic(err)
		}
		objs = append(objs, obj)
	}
	return objs
}

func (s *Scheme) Namespaced() []api.Object {
	var objs []api.Object
	for name := range s.namespaced {
		obj, err := s.NewObject(name)
		if err != nil {
			panic(err)
		}
		objs = append(objs, obj)
	}
	return objs
}

func (s *Scheme) IsGlobal(obj api.Object) bool {
	_, found := s.global[s.FullName(obj)]
	return found
}

func (s *Scheme) IsNamespaced(obj api.Object) bool {
	_, found := s.namespaced[s.FullName(obj)]
	return found
}

func (s *Scheme) FullName(obj api.Object) protoreflect.FullName {
	return obj.ProtoReflect().Descriptor().FullName()
}

func (s *Scheme) ListFullName(obj api.Object) protoreflect.FullName {
	objName := obj.ProtoReflect().Descriptor().FullName()
	if strings.HasSuffix(string(objName), "List") {
		return objName
	}
	return protoreflect.FullName(objName + "List")
}

func (s *Scheme) New(name protoreflect.FullName) (
	proto.Message, error) {
	msg, err := protoregistry.GlobalTypes.FindMessageByName(name)
	if err != nil {
		return nil, err
	}
	return msg.New().Interface(), nil
}

func (s *Scheme) NewObject(name protoreflect.FullName) (api.Object, error) {
	msg, err := s.New(name)
	if err != nil {
		return nil, err
	}
	obj, ok := msg.(api.Object)
	if !ok {
		return nil, fmt.Errorf("Not a valid api.Object %s", name)
	}
	return obj, nil
}

func (s *Scheme) NewList(name protoreflect.FullName) (api.ListObject, error) {
	msg, err := s.New(name)
	if err != nil {
		return nil, err
	}
	obj, ok := msg.(api.ListObject)
	if !ok {
		return nil, fmt.Errorf("Not a valid api.ListObject %s", name)
	}
	return obj, nil
}
