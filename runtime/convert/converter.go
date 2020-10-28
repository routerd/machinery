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

// Converter holds conversion functions and can use them to convert one type to another.
type Converter struct {
	conversionFuncs conversionFuncs
}

func NewConverter() *Converter {
	c := &Converter{
		conversionFuncs: newConversionFunctions(),
	}
	_ = c.conversionFuncs.Add(DefaultConversions...)
	return c
}

func (c *Converter) Convert(src, dest interface{}) error {
	return c.doConvert(src, dest, c.convert)
}

func (c *Converter) RegisterConversionFunc(conversionFunc interface{}) error {
	return c.conversionFuncs.Add(conversionFunc)
}

type conversionFunc func(sv, dv reflect.Value, scope *scope) error

func (c *Converter) doConvert(src, dest interface{}, f conversionFunc) error {
	destValue, err := dereferencePointer(dest)
	if err != nil {
		return fmt.Errorf("%v: for (src: %v) (dest: %v)", err, src, dest)
	}
	if !destValue.CanAddr() || !destValue.CanSet() {
		return fmt.Errorf("can't write to dest")
	}

	srcValue, err := dereferencePointer(src)
	if err != nil {
		if err == ErrIsNil {
			// nothing to do
			return nil
		}
		return err
	}

	s := &scope{
		converter: c,
	}
	s.init()

	return f(srcValue, destValue, s)
}

func (c *Converter) convert(srcValue, destValue reflect.Value, scope *scope) error {
	// Convert sv to dv.
	srcType, destType := srcValue.Type(), destValue.Type()
	pair := conversionPair{srcType, destType}

	// check registered conversion functions
	if fv, ok := c.conversionFuncs.fns[pair]; ok {
		return c.callConversionFunc(srcValue, destValue, fv, scope)
	}

	// dynamic conversion
	return c.reflectConvert(srcValue, destValue, scope)
}

type conversionPair struct {
	src, dest reflect.Type
}

// conversionFuncs holds conversion functions of a type to another type.
type conversionFuncs struct {
	fns map[conversionPair]reflect.Value
}

func newConversionFunctions() conversionFuncs {
	return conversionFuncs{
		fns: map[conversionPair]reflect.Value{},
	}
}

// registers conversion functions with this signature:
// `func(type1, type2, Scope) error`
func (c conversionFuncs) Add(fns ...interface{}) error {
	for _, fn := range fns {
		fv := reflect.ValueOf(fn)
		ft := fv.Type()

		if err := verifyConversionFunctionSignature(ft); err != nil {
			return err
		}
		c.fns[conversionPair{ft.In(0).Elem(), ft.In(1).Elem()}] = fv
	}
	return nil
}
