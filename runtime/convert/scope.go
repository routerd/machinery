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
	"reflect"
)

type Scope interface {
	Convert(src, dest interface{}) error
}

// scope tracks the recursion level of a conversion and
// allows custom conversion functions to continue conversion
// of nested values.
type scope struct {
	converter           *Converter
	srcStack, destStack scopeStack
}

func (s *scope) Convert(src, dest interface{}) error {
	return s.converter.Convert(src, dest)
}

func (s *scope) init() {
	s.push(scopeStackEntry{}, scopeStackEntry{})
}

func (s *scope) pop() {
	s.srcStack.pop()
	s.destStack.pop()
}

func (s *scope) push(src, dest scopeStackEntry) {
	s.srcStack.push(src)
	s.destStack.push(dest)
}

// describe prints the path to get to the current (source, dest) values.
func (s *scope) describe() (src, dest string) {
	return s.srcStack.describe(), s.destStack.describe()
}

// error makes an error that includes information about where we were in the objects
// we were asked to convert.
func (s *scope) errorf(message string, args ...interface{}) error {
	srcPath, destPath := s.describe()
	where := fmt.Sprintf("converting %v to %v: ", srcPath, destPath)
	return fmt.Errorf(where+message, args...)
}

// Formats src & dest as indices for printing.
func (s *scope) setIndices(src, dest int) {
	s.srcStack.top().key = fmt.Sprintf("[%v]", src)
	s.destStack.top().key = fmt.Sprintf("[%v]", dest)
}

// Formats src & dest as map keys for printing.
func (s *scope) setKeys(src, dest interface{}) {
	s.srcStack.top().key = fmt.Sprintf(`["%v"]`, src)
	s.destStack.top().key = fmt.Sprintf(`["%v"]`, dest)
}

type scopeStackEntry struct {
	value reflect.Value
	key   string
}

// scopeStack holds the recursion stack for a value.
type scopeStack []scopeStackEntry

func (s *scopeStack) pop() {
	n := len(*s)
	*s = (*s)[:n-1]
}

func (s *scopeStack) push(e scopeStackEntry) {
	*s = append(*s, e)
}

func (s *scopeStack) top() *scopeStackEntry {
	return &(*s)[len(*s)-1]
}

func (s scopeStack) describe() string {
	desc := ""
	if len(s) > 1 {
		desc = "(" + s[1].value.Type().String() + ")"
	}
	for i, v := range s {
		if i < 2 {
			// First layer on stack is not real; second is handled specially above.
			continue
		}
		if v.key == "" {
			desc += fmt.Sprintf(".%v", v.value.Type())
		} else {
			desc += fmt.Sprintf(".%v", v.key)
		}
	}
	return desc
}
