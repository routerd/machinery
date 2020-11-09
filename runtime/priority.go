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

package runtime

import (
	"fmt"
	"regexp"
	"strconv"
)

// APIVersion is the parsed representation of a API version.
// example: v1alpha4
type APIVersion struct {
	Major    int    // example: 1
	Phase    string // example: alpha
	Revision int    // example: 4
}

func (v APIVersion) String() string {
	s := fmt.Sprintf("v%d%s", v.Major, v.Phase)
	if v.Revision != 0 {
		s += strconv.Itoa(v.Revision)
	}
	return s
}

type APIVersionByPriority []APIVersion

func (l APIVersionByPriority) Len() int { return len(l) }
func (l APIVersionByPriority) Less(i, j int) bool {
	if l[i].Phase == l[j].Phase {
		if l[i].Major == l[j].Major {
			return l[i].Revision > l[j].Revision
		}
		if l[i].Major > l[j].Major {
			return true
		}
		return false
	}

	// always prefer _stable_ APIs before choosing a phased/unstable API
	if len(l[i].Phase) == 0 && len(l[j].Phase) > 0 {
		// [i] is not a phased API, but [j] is
		// [i] must come BEFORE
		return true
	}
	// beta comes before alpha
	if l[i].Phase == "beta" && l[j].Phase == "alpha" {
		return true
	}
	return false
}

func (l APIVersionByPriority) Swap(i, j int) {
	l[i], l[j] = l[j], l[i]
}

var (
	versionRegex = regexp.MustCompile(`(?m)v([[:digit:]]+)(beta|alpha)?([[:digit:]]*)`)
)

type InvalidVersionError struct {
	version string
}

func (e InvalidVersionError) Error() string {
	return fmt.Sprintf(
		"version must start with 'v', folowed by an integer, optionally followed by 'alpha' or 'beta' and an optional a revision number. (e.g. v1alpha4, v2, v1beta), got %q", e.version)
}

func ParseAPIVersion(version string) (APIVersion, error) {
	match := versionRegex.FindStringSubmatch(version)
	if len(match) != 4 || len(match[1]) == 0 {
		return APIVersion{}, InvalidVersionError{version: version}
	}

	var (
		v   APIVersion
		err error
	)
	v.Major, err = strconv.Atoi(match[1])
	if err != nil {
		return v, err
	}
	if len(match[2]) == 0 {
		return v, nil
	}
	v.Phase = match[2]
	if len(match[3]) != 0 {
		v.Revision, err = strconv.Atoi(match[3])
		if err != nil {
			return v, err
		}
	}
	return v, nil
}
