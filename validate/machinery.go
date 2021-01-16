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
	"regexp"
	"strings"

	"routerd.net/machinery/api"
)

func ValidateNamespacedName(nn api.NamespacedName) error {
	nameErr := ValidateName(nn.Name)

	var namespaceErr error
	if nn.Namespace != "" {
		namespaceErr = ValidateNamespace(nn.Namespace)
	}

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
		return fmt.Errorf("invalid name %q: %w", name, err)
	}
	return nil
}

func ValidateNamespace(namespace string) error {
	if err := ValidateRFC1035Subdomain(namespace); err != nil {
		return fmt.Errorf("invalid namespace %q: %w", namespace, err)
	}
	return nil
}

var qualifiedKeyName = regexp.
	MustCompile(`^[a-zA-Z0-9]([-_\.a-zA-Z0-9]*[a-zA-Z0-9])?$`)

func ValidateKey(key string) error {
	if len(key) > 63 {
		return fmt.Errorf("must be 63 characters or less")
	}
	if !qualifiedKeyName.MatchString(key) {
		return fmt.Errorf("must start and end with an alphanumeric character with [a-zA-Z0-9], dashes (-), underscores (_) or dots (.) are allowed between")
	}
	return nil
}

func ValidateQualifiedKey(key string) error {
	if len(key) == 0 {
		return nil
	}

	var name, prefix string
	if i := strings.LastIndex(key, "/"); i == -1 {
		name = key
	} else {
		name = key[:i]
		prefix = key[i+1:]
	}
	if err := ValidateKey(name); err != nil {
		return err
	}

	if len(prefix) > 0 {
		return ValidateRFC1035Subdomain(prefix)
	}
	return nil
}
