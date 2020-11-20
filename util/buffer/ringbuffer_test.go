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

package buffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRingBuffer(t *testing.T) {
	t.Run("overflow", func(t *testing.T) {
		b := NewRingBuffer(3)

		b.Append("1", 1)
		b.Append("2", 2)
		b.Append("3", 3)
		b.Append("4", 4)
		b.Append("5", 5)

		assert.Equal(t, 3, b.Size())
		assert.Equal(t, []interface{}{3, 4, 5}, b.List())

		b.Append("6", 6)
		b.Append("7", 7)
		assert.Equal(t, []interface{}{5, 6, 7}, b.List())
	})

	t.Run("empty", func(t *testing.T) {
		b := NewRingBuffer(10)
		b.Append("1", 1)
		b.Append("2", 2)
		b.Append("3", 3)

		assert.Equal(t, 10, b.Size())
		assert.Equal(t, []interface{}{1, 2, 3}, b.List())
	})

	t.Run("clear", func(t *testing.T) {
		b := NewRingBuffer(3)

		b.Append("1", 1)
		b.Append("2", 2)
		b.Append("3", 3)

		b.Clear()
		assert.Equal(t, []interface{}(nil), b.List())
	})

	t.Run("index empty", func(t *testing.T) {
		b := NewRingBuffer(5)

		b.Append("1", 1)
		b.Append("2", 2)
		b.Append("3", 3)

		out, ok := b.FromIndex("2")
		assert.True(t, ok)
		assert.Equal(t, []interface{}{2, 3}, out)
	})

	t.Run("index full", func(t *testing.T) {
		b := NewRingBuffer(5)

		b.Append("1", 1)
		b.Append("2", 2)
		b.Append("3", 3)
		b.Append("4", 4)
		b.Append("5", 5)
		b.Append("6", 6)

		out, ok := b.FromIndex("2")
		assert.True(t, ok)
		assert.Equal(t, []interface{}{2, 3, 4, 5, 6}, out)

		b.Append("7", 7)
		out, ok = b.FromIndex("2")
		assert.False(t, ok)
		assert.Equal(t, []interface{}(nil), out)

		out, ok = b.FromIndex("5")
		assert.True(t, ok)
		assert.Equal(t, []interface{}{5, 6, 7}, out)
	})
}
