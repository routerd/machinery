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
	"routerd.net/machinery/runtime"
	"routerd.net/machinery/util/ringbuffer"
)

// Hub emits Events to registered clients.
type Hub struct {
	buffer  *ringbuffer.Buffer
	listAll listAllFn

	broadcast chan Event
	clients   map[EventClient]struct{}
	register  chan EventClient
}

type listAllFn func(opts ...ListOption) ([]runtime.Object, error)

func NewHub(eventBufferSize int, listAll listAllFn) *Hub {
	return &Hub{
		buffer:  ringbuffer.NewBuffer(eventBufferSize),
		listAll: listAll,

		broadcast: make(chan Event),
		clients:   map[EventClient]struct{}{},
		register:  make(chan EventClient),
	}
}

// Broadcast emits an event to all connected clients.
func (h *Hub) Broadcast(eventType EventType, obj runtime.Object) {
	e := Event{
		Type:   eventType,
		Object: obj,
	}

	h.broadcast <- e
}

// Register returns a new client connected to the event hub.
func (h *Hub) Register(resourceVersion string, opts ...ListOption) EventClient {
	c := newEventClient(100, resourceVersion, opts)
	h.register <- c
	return c
}

// Run executes the Hub loops until the given channel is closed.
func (h *Hub) Run(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			// close all clients
			for c := range h.clients {
				c.Close()
				delete(h.clients, c)
			}
			return

		case c := <-h.register:
			h.clients[c] = struct{}{}
			h.seed(c)

		case event := <-h.broadcast:
			h.buffer.Add(event)

			for c := range h.clients {
				select {
				case c.events() <- event:
				default:
					// can't send -> Close()
					c.Close()
					delete(h.clients, c)
				}
			}
		}
	}
}

type versioned interface {
	GetResourceVersion() string
}

func (h *Hub) seed(c EventClient) {
	if c.resourceVersion() == "" {
		// add all items from storage
		objects, err := h.listAll(c.options()...)
		if err != nil {
			c.events() <- Event{
				Type: Error,
				Object: &Status{
					Status:  StatusFailure,
					Reason:  "InternalError",
					Message: "An internal error occured.",
				},
			}
			c.Close()
			return
		}
		for _, obj := range objects {
			c.events() <- Event{
				Type:   Added,
				Object: obj,
			}
		}
		return
	}

	// Start at specific resource version
	var start bool
	for _, obj := range h.buffer.List() {
		e := obj.(Event)
		o := e.Object.(versioned)
		if !start {
			start = o.GetResourceVersion() == c.resourceVersion()
		}
		if start {
			c.events() <- e
		}
	}
}

// EventClient receives events until Closed.
// On error the Event channel is closed and an Error event is emitted.
type EventClient interface {
	Close() error
	EventChan() <-chan Event
	events() chan<- Event
	resourceVersion() string
	options() []ListOption
}

type eventClient struct {
	rv   string
	recv chan Event
	opts []ListOption
}

func newEventClient(bufferSize int, resourceVersion string, opts []ListOption) *eventClient {
	return &eventClient{
		rv:   resourceVersion,
		recv: make(chan Event, bufferSize),
		opts: opts,
	}
}

func (c *eventClient) Close() error {
	close(c.recv)
	return nil
}

func (c *eventClient) EventChan() <-chan Event {
	return c.recv
}

func (c *eventClient) resourceVersion() string {
	return c.rv
}

func (c *eventClient) options() []ListOption {
	return c.opts
}

func (c *eventClient) events() chan<- Event {
	return c.recv
}
