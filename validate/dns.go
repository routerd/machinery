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
)

var rfc1035SubdomainRegEx = regexp.
	MustCompile(`^[[:lower:]]([[:lower:]]|-|\.)*[[:lower:]]$`)

func ValidateRFC1035Subdomain(subdomain string) error {
	if len(subdomain) > 253 {
		return fmt.Errorf(
			"rfc1035 DNS subdomain MUST not exceed 253 characters")
	}

	if !rfc1035SubdomainRegEx.MatchString(subdomain) {
		return fmt.Errorf("rfc1035 DNS labels MUST be lowercase, start and end with an alphanumeric character and MUST only contain alphanumeric characters, - or .")
	}

	return nil
}

var rfc1035LabelRegEx = regexp.
	MustCompile(`^[a-z]([-a-z0-9]*[a-z0-9])?$`)

func ValidateRFC1035Label(label string) error {
	if len(label) > 63 {
		return fmt.Errorf("rfc1035 DNS labels MUST not exceed 63 characters")
	}

	if !rfc1035LabelRegEx.MatchString(label) {
		return fmt.Errorf("rfc1035 DNS labels MUST be lowercase, start and end with an alphanumeric character and MUST only contain alphanumeric characters or -")
	}

	return nil
}
