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

package v1

import (
	"routerd.net/machinery/meta"
	"routerd.net/machinery/runtime"
)

var _ runtime.Object = (*Status)(nil)

// Status may be returned for calls that don't return an object.
type Status struct {
	meta.TypeMeta `json:",inline"`
	Status        StatusType `json:"status"`
	Reason        string     `json:"reason"`
	Message       string     `json:"message"`
}

type StatusType string

const (
	StatusSuccess StatusType = "Success"
	StatusFailure StatusType = "Failure"
)

func (s *Status) DeepCopyObject() runtime.Object {
	statusClone := &Status{}
	runtime.DeepCopy(s, statusClone)
	return statusClone
}
