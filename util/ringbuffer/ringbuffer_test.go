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

package ringbuffer

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestRingbuffer(t *testing.T) {
	b := NewBuffer(3)

	b.Add(1, 2, 3, 4, 5)

	assert.Equal(t, 3, b.Size())
	assert.Equal(t, []interface{}{3, 4, 5}, b.List())
	b.Add(6, 7)
	assert.Equal(t, []interface{}{5, 6, 7}, b.List())
	b.Clear()
	assert.Len(t, b.List(), 0)

	t.Fail()
}
