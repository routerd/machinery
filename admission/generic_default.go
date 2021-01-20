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

package admission

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"routerd.net/machinery/api"
)

// defaulted may be implemented by objects to set field default values.
type defaulted interface {
	Default(ctx context.Context) error
}

// The GenericDefault AdmissionController calls the .Default() method on the given object, if this method is implemented.
type GenericDefault struct{}

var _ AdmissionController = (*GenericDefault)(nil)

func (d *GenericDefault) callDefault(ctx context.Context, obj api.Object) error {
	if d, ok := obj.(defaulted); ok {
		if err := d.Default(ctx); err != nil {
			return fmt.Errorf(
				"defaulting object of type %s: %w",
				obj.ProtoReflect().Descriptor().FullName(), err)
		}
	}
	return nil
}

func (d *GenericDefault) OnCreate(ctx context.Context, obj api.Object) error {
	if obj.ObjectMeta() != nil &&
		len(obj.ObjectMeta().GetName()) == 0 &&
		len(obj.ObjectMeta().GetGenerateName()) > 0 {
		obj.ObjectMeta().SetName(obj.ObjectMeta().GetGenerateName() + generateNameSuffix())
	}

	return d.callDefault(ctx, obj)
}

func (d *GenericDefault) OnUpdate(ctx context.Context, obj api.Object) error {
	return d.callDefault(ctx, obj)
}

func (d *GenericDefault) OnDelete(ctx context.Context, obj api.Object) error {
	return d.callDefault(ctx, obj)
}

var r *rand.Rand

const (
	suffixLength = 4
	charset      = "abcdefghijklmnopqrstuvwxyz" +
		"ABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
)

func generateNameSuffix() string {
	c := make([]byte, suffixLength)
	for i := range c {
		c[i] = charset[r.Intn(len(c))]
	}
	return string(c)
}

func init() {
	r = rand.New(rand.NewSource(time.Now().UnixNano()))
}
