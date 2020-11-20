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

package convert

import (
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"
)

type AnyMarshaller struct {
	TypeURL string
}

func NewAnyMarshaller(typeURL string) *AnyMarshaller {
	return &AnyMarshaller{TypeURL: typeURL}
}

func (m *AnyMarshaller) Marshal(pb proto.Message) (*anypb.Any, error) {
	value, err := proto.Marshal(pb)
	if err != nil {
		return nil, err
	}
	return &anypb.Any{
		TypeUrl: m.TypeURL + "/" + string(
			pb.ProtoReflect().Descriptor().FullName()),
		Value: value,
	}, nil
}

func (m *AnyMarshaller) MustMarshal(pb proto.Message) *anypb.Any {
	a, err := m.Marshal(pb)
	if err != nil {
		panic(err)
	}
	return a
}

func (m *AnyMarshaller) Unmarshal(any *anypb.Any) (proto.Message, error) {
	return anypb.UnmarshalNew(any, proto.UnmarshalOptions{})
}
