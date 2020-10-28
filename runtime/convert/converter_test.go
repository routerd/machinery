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

package convert

import (
	"fmt"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type testObject1 struct {
	Str       string
	Int       int
	Bool      bool
	StrMap    map[string]string
	Struct    subObject1
	StructPtr *subObject1
	List      []subObject1
}

type subObject1 struct {
	Str  string
	Int  int
	Bool bool
}

type testObject2 struct {
	Str       string
	Int       int
	Bool      bool
	StrMap    map[string]string
	Struct    subObject2
	StructPtr *subObject2
	List      []subObject2
}

type subObject2 struct {
	Str  string
	Int  int
	Bool bool
}

// a string that is always prefixed with "test3000:"
type test3000PrefixStr string

func TestConverter(t *testing.T) {
	t.Run("reflect conversion", func(t *testing.T) {
		c := NewConverter()

		src := &testObject1{
			Str: "str", Int: 42, Bool: true,
			Struct:    subObject1{Str: "str1", Int: 421, Bool: true},
			StructPtr: &subObject1{Str: "str1", Int: 421, Bool: true},
			StrMap:    map[string]string{"key": "val"},
			List: []subObject1{
				{Str: "str1", Int: 421, Bool: true},
			},
		}
		dest := &testObject2{}
		err := c.Convert(src, dest)
		require.NoError(t, err)

		fmt.Printf("%#v", *dest)
		assert.Equal(t, src.Str, dest.Str, ".Str")
		assert.Equal(t, src.Int, dest.Int, ".Int")
		assert.Equal(t, src.Bool, dest.Bool, ".Bool")

		assert.EqualValues(t, src.Struct, dest.Struct, ".Struct")
		assert.EqualValues(t, src.StrMap, dest.StrMap, ".StrMap")

		if assert.NotNil(t, src.StructPtr, ".StructPtr") {
			assert.EqualValues(t, *src.StructPtr, *dest.StructPtr, ".StructPtr")
		}

		if assert.Len(t, dest.List, 1) {
			assert.EqualValues(t, src.List[0], dest.List[0])
		}
	})

	t.Run("static conversion", func(t *testing.T) {
		c := NewConverter()

		err := c.RegisterConversionFunc(func(src *string, dest *test3000PrefixStr, scope Scope) error {
			*dest = test3000PrefixStr("test3000:" + *src)
			return nil
		})
		require.NoError(t, err)

		var (
			src  = "somestr"
			dest test3000PrefixStr
		)
		err = c.Convert(&src, &dest)
		require.NoError(t, err)

		assert.Equal(t, test3000PrefixStr("test3000:somestr"), dest)
	})
}
