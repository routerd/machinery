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

// Buffer stores a preset amount of objects.
// If the buffer is full the oldest item will be discarded.
type Buffer struct {
	values []interface{}
	ptr    int
}

func NewBuffer(size int) *Buffer {
	return &Buffer{values: make([]interface{}, size)}
}

// Size returns the size of the buffer.
func (b *Buffer) Size() int {
	return len(b.values)
}

// List returns a slice of all entries in the buffer.
func (b *Buffer) List() []interface{} {
	i := b.index(b.ptr + 1)
	var out []interface{}
	for _, v := range append(b.values[i:], b.values[:i]...) {
		if v == nil {
			break
		}
		out = append(out, v)
	}
	return out
}

func (b *Buffer) Add(values ...interface{}) {
	for _, v := range values {
		b.add(v)
	}
}

func (b *Buffer) Clear() {
	for i := range b.values {
		b.values[i] = nil
	}
}

func (b *Buffer) add(value interface{}) {
	b.ptr = b.index(b.ptr + 1)
	b.values[b.ptr] = value
}

func (b *Buffer) index(i int) int {
	if i == len(b.values) {
		return 0
	}
	return i
}
