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
	"fmt"

	"routerd.net/machinery/runtime"
)

type NamespacedName struct {
	Name      string `json:"name"`
	Namespace string `json:"namespace"`
}

func (nn NamespacedName) String() string {
	return nn.Name + "." + nn.Namespace
}

func (nn NamespacedName) Validate() error {
	nameErr := ValidateName(nn.Name)
	namespaceErr := ValidateNamespace(nn.Namespace)

	if nameErr != nil || namespaceErr != nil {
		return InvalidNamespaceNameErr{
			NamespacedName: nn,
			NameErr:        nameErr,
			NamespaceErr:   namespaceErr,
		}
	}
	return nil
}

type InvalidNamespaceNameErr struct {
	NamespacedName
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

func Key(obj Object) NamespacedName {
	return NamespacedName{
		Name:      obj.GetName(),
		Namespace: obj.GetNamespace(),
	}
}

// Object can be persistent in storage.
type Object interface {
	runtime.Object
	GetName() string
	GetNamespace() string
	GetUID() string
	SetUID(string)
	GetResourceVersion() string
	SetResourceVersion(string)
	GetGeneration() int64
	SetGeneration(int64)
}

// ListObject
type ListObject interface {
	runtime.Object
}
