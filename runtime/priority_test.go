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
	"sort"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAPIVersionByPriority(t *testing.T) {
	v10 := APIVersion{Major: 10}
	v2 := APIVersion{Major: 2}
	v1 := APIVersion{Major: 1}

	v11beta2 := APIVersion{Major: 11, Phase: "beta", Revision: 2}
	v10beta3 := APIVersion{Major: 10, Phase: "beta", Revision: 3}
	v3beta1 := APIVersion{Major: 3, Phase: "beta", Revision: 1}
	v12alpha1 := APIVersion{Major: 12, Phase: "alpha", Revision: 1}
	v11alpha1 := APIVersion{Major: 11, Phase: "alpha", Revision: 2}

	l := APIVersionByPriority{v3beta1, v1, v12alpha1, v10beta3, v11alpha1, v2, v11beta2, v10}
	sort.Sort(l)

	var versions []string
	for _, v := range l {
		versions = append(versions, v.String())
	}
	assert.Equal(t, []string{"v10", "v2", "v1", "v11beta2", "v10beta3", "v3beta1", "v12alpha1", "v11alpha2"}, versions)
}

func TestParseAPIVersion(t *testing.T) {
	tests := []struct {
		name       string
		version    string
		apiVersion APIVersion
	}{
		{
			name:    "major",
			version: "v3",
			apiVersion: APIVersion{
				Major: 3,
			},
		},
		{
			name:    "phased",
			version: "v1alpha",
			apiVersion: APIVersion{
				Major: 1,
				Phase: "alpha",
			},
		},
		{
			name:    "phased with revision",
			version: "v2beta4",
			apiVersion: APIVersion{
				Major:    2,
				Phase:    "beta",
				Revision: 4,
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			apiVersion, err := ParseAPIVersion(test.version)
			require.NoError(t, err)
			assert.Equal(t, test.apiVersion, apiVersion)
		})
	}
}
