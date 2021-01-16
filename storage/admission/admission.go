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

package admission

import (
	"routerd.net/machinery/api"
)

// AdmissionController checks objects from requests before they are commited to storage.
type AdmissionController interface {
	OnCreate(obj api.Object) error
	OnUpdate(obj api.Object) error
	OnDelete(obj api.Object) error
}

// AdmissionControllerFns wraps closures to match the AdmissionController interface.
type AdmissionControllerFns struct {
	OnCreateFn func(obj api.Object) error
	OnUpdateFn func(obj api.Object) error
	OnDeleteFn func(obj api.Object) error
}

func (a *AdmissionControllerFns) OnCreate(obj api.Object) error {
	if a.OnCreateFn != nil {
		return a.OnCreateFn(obj)
	}
	return nil
}

func (a *AdmissionControllerFns) OnUpdate(obj api.Object) error {
	if a.OnUpdateFn != nil {
		return a.OnUpdateFn(obj)
	}
	return nil
}

func (a *AdmissionControllerFns) OnDelete(obj api.Object) error {
	if a.OnDeleteFn != nil {
		return a.OnDeleteFn(obj)
	}
	return nil
}

// AdmissionControllerList executes a ordered list of AdmissionControllers, stopping at the first error.
type AdmissionControllerList []AdmissionController

func (a *AdmissionControllerList) OnCreate(obj api.Object) error {
	for _, subA := range *a {
		if err := subA.OnCreate(obj); err != nil {
			return err
		}
	}
	return nil
}

func (a *AdmissionControllerList) OnUpdate(obj api.Object) error {
	for _, subA := range *a {
		if err := subA.OnUpdate(obj); err != nil {
			return err
		}
	}
	return nil
}

func (a *AdmissionControllerList) OnDelete(obj api.Object) error {
	for _, subA := range *a {
		if err := subA.OnDelete(obj); err != nil {
			return err
		}
	}
	return nil
}
