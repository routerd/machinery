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

type ListOptions struct {
	Namespace string
}

type ListOption interface {
	ApplyToList(opt *ListOptions)
}

type WatchOptions struct {
	Namespace string
}

type WatchOption interface {
	ApplyToWatch(opt *WatchOptions)
}

type DeleteOptions struct{}

type DeleteOption interface {
	ApplyToDelete(opt *DeleteOptions)
}

type CreateOptions struct{}

type CreateOption interface {
	ApplyToCreate(opt *CreateOptions)
}

type UpdateOptions struct{}

type UpdateOption interface {
	ApplyToUpdate(opt *UpdateOptions)
}

type InNamespace string

func (n InNamespace) ApplyToList(opt *ListOptions) {
	opt.Namespace = string(n)
}

func (n InNamespace) ApplyToWatch(opt *WatchOptions) {
	opt.Namespace = string(n)
}
