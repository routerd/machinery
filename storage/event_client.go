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
	"sync"

	"routerd.net/machinery/api"
)

// EventClient represents a client connection to the event hub.
type EventClient interface {
	Events() <-chan api.ResourceEvent
	Close() error

	eventSink() chan<- api.ResourceEvent
	options() ListOptions
	resourceVersion() string
	error(error)
	initDone()
	wait()
}

var _ EventClient = (*eventClient)(nil)

type eventClient struct {
	eventCh    chan api.ResourceEvent
	rv         string
	opts       ListOptions
	err        error
	initCh     chan struct{}
	deregister chan<- EventClient
	closeOnce  sync.Once
}

func newEventClient(
	bufferSize int,
	resourceVersion string,
	opts ListOptions,
	deregister chan<- EventClient,
) *eventClient {
	return &eventClient{
		eventCh:    make(chan api.ResourceEvent, bufferSize),
		initCh:     make(chan struct{}),
		rv:         resourceVersion,
		opts:       opts,
		deregister: deregister,
	}
}

func (c *eventClient) Close() error {
	c.closeOnce.Do(func() {
		c.deregister <- c
	})
	return nil
}

func (c *eventClient) Events() <-chan api.ResourceEvent {
	return c.eventCh
}

func (c *eventClient) error(err error) {
	c.err = err
}

func (c *eventClient) eventSink() chan<- api.ResourceEvent {
	return c.eventCh
}

func (c *eventClient) options() ListOptions {
	return c.opts
}

func (c *eventClient) resourceVersion() string {
	return c.rv
}

func (c *eventClient) initDone() {
	close(c.initCh)
}

func (c *eventClient) wait() {
	<-c.initCh
}
