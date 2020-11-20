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

import "fmt"

type ErrNotFound struct {
	Name, Namespace string
	TypeFullName    string
}

func (e ErrNotFound) Error() string {
	return fmt.Sprintf("%s: %s/%s", e.TypeFullName, e.Namespace, e.Name)
}

type ErrExpired struct {
	Message string
}

func (e *ErrExpired) Error() string {
	return e.Message
}

type ErrAlreadyExists struct {
	Name, Namespace string
	TypeFullName    string
}

func (e ErrAlreadyExists) Error() string {
	return fmt.Sprintf("%s already exists: %s/%s", e.TypeFullName, e.Namespace, e.Name)
}

type ErrConflict struct {
	Name, Namespace string
	TypeFullName    string
}

func (e ErrConflict) Error() string {
	return fmt.Sprintf("%s conflicting resource version: %s/%s", e.TypeFullName, e.Namespace, e.Name)
}
