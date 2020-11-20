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

package storage

import (
	"k8s.io/apimachinery/pkg/labels"
	"k8s.io/apimachinery/pkg/selection"
)

type ListOptions struct {
	Namespace     string
	LabelSelector labels.Selector
}

type ListOption interface {
	ApplyToList(opt *ListOptions)
}

type WatchOptions struct {
	ListOptions
}

type WatchOption interface {
	ApplyToWatch(opt *WatchOptions)
}

type DeleteOptions struct{}

type DeleteOption interface {
	ApplyToDelete(opt *DeleteOptions)
}

type CreateOptions struct{}

type CreateOption interface {
	ApplyToCreate(opt *CreateOptions)
}

type UpdateOptions struct{}

type UpdateOption interface {
	ApplyToUpdate(opt *UpdateOptions)
}

type InNamespace string

func (n InNamespace) ApplyToList(opt *ListOptions) {
	opt.Namespace = string(n)
}

func (n InNamespace) ApplyToWatch(opt *WatchOptions) {
	opt.Namespace = string(n)
}

type MatchLabels map[string]string

func (m MatchLabels) ApplyToList(opts *ListOptions) {
	sel := labels.SelectorFromValidatedSet(map[string]string(m))
	opts.LabelSelector = sel
}

func (m MatchLabels) ApplyToWatch(opts *WatchOptions) {
	m.ApplyToList(&opts.ListOptions)
}

type HasLabels []string

func (m HasLabels) ApplyToList(opts *ListOptions) {
	sel := labels.NewSelector()
	for _, label := range m {
		r, err := labels.NewRequirement(label, selection.Exists, nil)
		if err == nil {
			sel = sel.Add(*r)
		}
	}
	opts.LabelSelector = sel
}

func (m HasLabels) ApplyToWatch(opts *WatchOptions) {
	m.ApplyToList(&opts.ListOptions)
}
