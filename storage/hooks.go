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

package storage

import (
	"fmt"

	"routerd.net/machinery/api"
	"routerd.net/machinery/validate"
)

// defaulted may be implemented by objects to set field default values.
type defaulted interface {
	Default() error
}

func Default(obj api.Object) error {
	if d, ok := obj.(defaulted); ok {
		if err := d.Default(); err != nil {
			return fmt.Errorf("calling obj.Default(): %w", err)
		}
	}
	return nil
}

type validateCreate interface {
	ValidateCreate() error
}

func ValidateCreate(obj api.Object) error {
	meta := obj.ObjectMeta()
	if len(meta.GetUid()) != 0 {
		return fmt.Errorf("UID must be empty when creating an object.")
	}
	if len(meta.GetResourceVersion()) != 0 {
		return fmt.Errorf("ResourceVersion must be empty when creating an object.")
	}
	if meta.GetGeneration() != 0 {
		return fmt.Errorf("Generation must be empty when creating an object.")
	}

	if err := validate.ValidateName(meta.GetName()); err != nil {
		return err
	}
	if err := validate.ValidateNamespace(meta.GetNamespace()); err != nil {
		return err
	}

	if d, ok := obj.(validateCreate); ok {
		if err := d.ValidateCreate(); err != nil {
			return fmt.Errorf("calling obj.ValidateCreate(): %w", err)
		}
	}
	return nil
}

type validateUpdate interface {
	ValidateUpdate(old api.Object) error
}

func ValidateUpdate(old, new api.Object) error {
	newMeta := new.ObjectMeta()
	if err := validate.ValidateName(newMeta.GetName()); err != nil {
		return err
	}
	if err := validate.ValidateNamespace(newMeta.GetNamespace()); err != nil {
		return err
	}

	if d, ok := new.(validateUpdate); ok {
		if err := d.ValidateUpdate(old); err != nil {
			return fmt.Errorf("calling obj.ValidateUpdate(): %w", err)
		}
	}
	return nil
}

type validateDelete interface {
	ValidateDelete() error
}

func ValidateDelete(obj api.Object) error {
	meta := obj.ObjectMeta()
	if err := validate.ValidateName(meta.GetName()); err != nil {
		return err
	}
	if err := validate.ValidateNamespace(meta.GetNamespace()); err != nil {
		return err
	}

	if d, ok := obj.(validateDelete); ok {
		if err := d.ValidateDelete(); err != nil {
			return fmt.Errorf("calling obj.ValidateDelete(): %w", err)
		}
	}
	return nil
}
