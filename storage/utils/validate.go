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

package utils

import (
	"fmt"

	storagev1 "routerd.net/machinery/storage/api/v1"
)

type validateCreate interface {
	ValidateCreate() error
}

func ValidateCreate(obj storagev1.Object) error {
	if d, ok := obj.(validateCreate); ok {
		if err := d.ValidateCreate(); err != nil {
			return fmt.Errorf("calling obj.ValidateCreate(): %w", err)
		}
	}
	return nil
}

type validateUpdate interface {
	ValidateUpdate() error
}

func ValidateUpdate(obj storagev1.Object) error {
	if d, ok := obj.(validateUpdate); ok {
		if err := d.ValidateUpdate(); err != nil {
			return fmt.Errorf("calling obj.ValidateUpdate(): %w", err)
		}
	}
	return nil
}

type validateDelete interface {
	ValidateDelete() error
}

func ValidateDelete(obj storagev1.Object) error {
	if d, ok := obj.(validateDelete); ok {
		if err := d.ValidateDelete(); err != nil {
			return fmt.Errorf("calling obj.ValidateDelete(): %w", err)
		}
	}
	return nil
}
