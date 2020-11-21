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
	"fmt"
	"strings"

	"routerd.net/machinery/api"
)

type ErrNotFound struct {
	api.NamespacedName
	TypeFullName string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s %s: not found", e.TypeFullName, e.String())
}

type ErrExpired struct {
	Message string
}

func (e *ErrExpired) Error() string {
	return e.Message
}

type ErrAlreadyExists struct {
	api.NamespacedName
	TypeFullName string
}

func (e ErrAlreadyExists) Error() string {
	return fmt.Sprintf("%s %s: already exists", e.TypeFullName, e.String())
}

type ErrConflict struct {
	api.NamespacedName
	TypeFullName string
}

func (e ErrConflict) Error() string {
	return fmt.Sprintf("%s %s: conflicting resource version", e.TypeFullName, e.String())
}

type ErrBadRequest struct {
	api.NamespacedName
	TypeFullName    string
	FieldViolations []FieldViolation
}

func (e ErrBadRequest) Error() string {
	var msg []string
	for _, fv := range e.FieldViolations {
		msg = append(msg, fv.Message)
	}
	return fmt.Sprintf("%s %s: is invalid: %s", e.TypeFullName, e.String(), strings.Join(msg, ", "))
}

type FieldViolation struct {
	Field, Message string
}
