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
	"errors"
	"fmt"
	"reflect"
)

// ErrIsNil is returned if one of the inputs is nil
var ErrIsNil = errors.New("expected pointer, but got nil")

// dereferencePointer checks if the given obj is a pointer and dereferences it.
func dereferencePointer(obj interface{}) (reflect.Value, error) {
	v := reflect.ValueOf(obj)
	if v.Kind() == reflect.Invalid {
		return reflect.Value{}, fmt.Errorf("expected pointer, but got invalid kind")
	}

	if v.Kind() != reflect.Ptr {
		return reflect.Value{}, fmt.Errorf("expected pointer, but got %v type", v.Type())
	}

	if v.IsNil() {
		return reflect.Value{}, ErrIsNil
	}
	return v.Elem(), nil
}
