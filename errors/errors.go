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

package errors

import (
	"errors"
	"fmt"

	"routerd.net/machinery/runtime"
)

type ErrConflict struct {
	Key string
	GVK runtime.GroupVersionKind
}

func (e ErrConflict) Error() string {
	return fmt.Sprintf("%s: %s conflicting resource version", e.GVK, e.Key)
}

func IsConflict(err error) bool {
	_, ok := errors.Unwrap(err).(ErrConflict)
	return ok
}

type ErrAlreadyExists struct {
	Key string
	GVK runtime.GroupVersionKind
}

func (e ErrAlreadyExists) Error() string {
	return fmt.Sprintf("%s: %s already exists", e.GVK, e.Key)
}

func IsAlreadyExists(err error) bool {
	_, ok := errors.Unwrap(err).(ErrAlreadyExists)
	return ok
}

type ErrNotFound struct {
	Key string
	GVK runtime.GroupVersionKind
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s: %s not found", e.GVK, e.Key)
}

func IsNotFound(err error) bool {
	_, ok := errors.Unwrap(err).(ErrNotFound)
	return ok
}

type ErrInternal struct {
	Err error
}

func (e ErrInternal) Error() string {
	return "internal error"
}
