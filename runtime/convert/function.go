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
	"fmt"
	"reflect"
)

func (c *Converter) callConversionFunc(sv, dv, custom reflect.Value, scope *scope) error {
	if !sv.CanAddr() {
		sv2 := reflect.New(sv.Type())
		sv2.Elem().Set(sv)
		sv = sv2
	} else {
		sv = sv.Addr()
	}
	if !dv.CanAddr() {
		if !dv.CanSet() {
			return scope.errorf("can't addr or set dest")
		}
		dvOrig := dv
		dvv := reflect.New(dvOrig.Type())
		defer func() { dvOrig.Set(dvv) }()
	} else {
		dv = dv.Addr()
	}
	args := []reflect.Value{sv, dv, reflect.ValueOf(scope)}
	ret := custom.Call(args)[0].Interface()
	// This convolution is necessary because nil interfaces won't convert
	// to errors.
	if ret == nil {
		return nil
	}
	return ret.(error)
}

func verifyConversionFunctionSignature(ft reflect.Type) error {
	if ft.Kind() != reflect.Func {
		return fmt.Errorf("expected func, got: %v", ft)
	}
	if ft.NumIn() != 3 {
		return fmt.Errorf("expected three 'in' params, got: %v", ft)
	}
	if ft.NumOut() != 1 {
		return fmt.Errorf("expected one 'out' param, got: %v", ft)
	}
	if ft.In(0).Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer arg for 'in' param 0, got: %v", ft)
	}
	if ft.In(1).Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer arg for 'in' param 1, got: %v", ft)
	}
	scopeType := Scope(nil)
	if e, a := reflect.TypeOf(&scopeType).Elem(), ft.In(2); e != a {
		return fmt.Errorf("expected '%v' arg for 'in' param 2, got '%v' (%v)", e, a, ft)
	}

	var forErrorType error
	// This convolution is necessary, otherwise TypeOf picks up on the fact
	// that forErrorType is nil.
	errorType := reflect.TypeOf(&forErrorType).Elem()
	if ft.Out(0) != errorType {
		return fmt.Errorf("expected error return, got: %v", ft)
	}
	return nil
}
