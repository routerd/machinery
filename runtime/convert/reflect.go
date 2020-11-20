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
	"reflect"
)

// reflectConvert converts src to dest via runtime reflection.
func (c *Converter) reflectConvert(src, dest reflect.Value, scope *scope) error {
	srcType, destType := src.Type(), dest.Type()
	if !dest.CanSet() {
		return scope.errorf("Cannot set dest.")
	}

	switch srcType.Kind() {
	case reflect.Map, reflect.Ptr, reflect.Slice, reflect.Interface, reflect.Struct:
		// Don't copy via assignment/conversion

	default:
		if srcType.AssignableTo(destType) {
			dest.Set(src)
			return nil
		}
		if srcType.ConvertibleTo(destType) {
			dest.Set(src.Convert(destType))
			return nil
		}
	}

	scope.push(scopeStackEntry{value: src}, scopeStackEntry{value: dest})
	defer scope.pop()

	switch dest.Kind() {
	case reflect.Slice:
		if src.IsNil() {
			dest.Set(reflect.Zero(destType))
			return nil
		}

		dest.Set(reflect.MakeSlice(destType, src.Len(), src.Cap()))
		for i := 0; i < src.Len(); i++ {
			scope.setIndices(i, i)
			if err := c.convert(src.Index(i), dest.Index(i), scope); err != nil {
				return err
			}
		}
		return nil

	case reflect.Ptr:
		if src.IsNil() {
			dest.Set(reflect.Zero(destType))
			return nil
		}

		dest.Set(reflect.New(destType.Elem()))
		switch srcType.Kind() {
		case reflect.Ptr, reflect.Interface:
			return c.convert(src.Elem(), dest.Elem(), scope)
		default:
			return c.convert(src, dest.Elem(), scope)
		}

	case reflect.Struct:
		dest.Set(reflect.New(destType).Elem())
		for i := 0; i < src.NumField(); i++ {
			srcFieldValue := src.Field(i)
			srcFieldType := srcType.Field(i)

			destFieldValue := dest.FieldByName(srcFieldType.Name)
			if !destFieldValue.IsValid() {
				continue
			}
			scope.setKeys(srcFieldType.Name, srcFieldType.Name)
			if err := c.convert(srcFieldValue, destFieldValue, scope); err != nil {
				return err
			}
		}
		return nil

	case reflect.Map:
		if src.IsNil() {
			dest.Set(reflect.Zero(destType))
			return nil
		}

		dest.Set(reflect.MakeMap(destType))
		for _, srcKey := range src.MapKeys() {
			destKey := reflect.New(destType.Key()).Elem()
			if err := c.convert(srcKey, destKey, scope); err != nil {
				return err
			}
			destValue := reflect.New(destType.Elem()).Elem()
			scope.setKeys(srcKey.Interface(), destKey.Interface())
			if err := c.convert(src.MapIndex(srcKey), destValue, scope); err != nil {
				return err
			}
			dest.SetMapIndex(destKey, destValue)
		}
		return nil

	default:
		return scope.errorf("couldn't copy '%v' into '%v'; didn't understand types", srcType, destType)
	}
}
