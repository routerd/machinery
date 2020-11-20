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

// RingBuffer stores a preset amount of objects.
// If the buffer is full the oldest item will be discarded.
type RingBuffer struct {
	values  []*entry
	indexes map[string]int
	ptr     int
}

type entry struct {
	object interface{}
	clean  func()
}

func NewRingBuffer(size int) *RingBuffer {
	return &RingBuffer{
		values:  make([]*entry, size),
		indexes: map[string]int{},
	}
}

// Size returns the size of the buffer.
func (b *RingBuffer) Size() int {
	return len(b.values)
}

// List returns a slice of all entries in the buffer.
func (b *RingBuffer) List() []interface{} {
	return b.seek(0, len(b.values))
}

func (b *RingBuffer) FromIndex(index string) ([]interface{}, bool) {
	if _, ok := b.indexes[index]; !ok {
		return nil, false
	}
	return b.seek(b.index(b.indexes[index]), len(b.values)), true
}

func (b *RingBuffer) Append(index string, value interface{}) {
	b.append(index, value)
}

func (b *RingBuffer) Clear() {
	for i := range b.values {
		b.values[i] = nil
	}
}

func (b *RingBuffer) seek(start, end int) []interface{} {
	var out []interface{}
	for i := start; i < end; i++ {
		v := b.values[b.pointer(i)]
		if v == nil {
			continue
		}
		out = append(out, v.object)
	}
	return out
}

func (b *RingBuffer) append(index string, value interface{}) {
	b.ptr = b.nextPtr()
	if b.values[b.ptr] != nil {
		b.values[b.ptr].clean()
	}

	b.indexes[index] = b.ptr
	b.values[b.ptr] = &entry{
		object: value,
		clean:  func() { delete(b.indexes, index) },
	}
}

func (b *RingBuffer) nextPtr() int {
	ptr := b.ptr + 1
	if ptr >= len(b.values) {
		return ptr - len(b.values)
	}
	return ptr
}

// takes a pointer and normalizes it into an index.
func (b *RingBuffer) index(ptr int) int {
	i := len(b.values) + (ptr - b.ptr) - 1
	if i >= len(b.values) {
		i = i - len(b.values)
	}
	return i
}

// takes an index and shifts it by the current ptr of the buffer.
func (b *RingBuffer) pointer(index int) int {
	// b.ptr is always at the last index ([len(b.values)]-1)
	if index == len(b.values)-1 {
		return b.ptr
	}

	i := b.ptr + index + 1
	if i >= len(b.values) {
		return i - len(b.values)
	}
	return i
}
