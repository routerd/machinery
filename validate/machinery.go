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

package validate

import (
	"fmt"

	"routerd.net/machinery/api"
)

func ValidateNamespacedName(nn api.NamespacedName) error {
	nameErr := ValidateName(nn.Name)
	namespaceErr := ValidateNamespace(nn.Namespace)

	if nameErr != nil || namespaceErr != nil {
		return InvalidNamespaceNameErr{
			NameErr:      nameErr,
			NamespaceErr: namespaceErr,
		}
	}
	return nil
}

type InvalidNamespaceNameErr struct {
	NameErr, NamespaceErr error
}

func (e InvalidNamespaceNameErr) Error() string {
	if e.NamespaceErr == nil {
		return fmt.Sprintf("invalid name: %v", e.NameErr)
	}
	if e.NameErr == nil {
		return fmt.Sprintf("invalid namespace: %v", e.NamespaceErr)
	}
	return fmt.Sprintf("invalid name and namespace: %v, %v", e.NameErr, e.NamespaceErr)
}

func ValidateName(name string) error {
	if err := ValidateRFC1035Label(name); err != nil {
		return fmt.Errorf("invalid name %s: %w", name, err)
	}
	return nil
}

func ValidateNamespace(namespace string) error {
	if err := ValidateRFC1035Subdomain(namespace); err != nil {
		return fmt.Errorf("invalid namespace %s: %w", namespace, err)
	}
	return nil
}
