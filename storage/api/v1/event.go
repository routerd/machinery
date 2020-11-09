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

import "routerd.net/machinery/runtime"

// Event is emitted for every change of an object in storage.
type Event struct {
	Type   EventType
	Object runtime.Object
}

type EventType string

const (
	Added    EventType = "Added"
	Modified EventType = "Modified"
	Deleted  EventType = "Deleted"
	Error    EventType = "Error"
)
