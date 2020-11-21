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
	"routerd.net/machinery/api"
	"routerd.net/machinery/errors"
	"routerd.net/machinery/util/buffer"
)

// ListerFn is used to get all existing objects to seed new clients with "Add" events.
type ListerFn func(opts ListOptions) ([]api.Object, error)

type indexedBuffer interface {
	Append(index string, value interface{})
	FromIndex(index string) ([]interface{}, bool)
}

type event struct {
	Type            api.EventType
	Object          api.Object
	ResourceVersion string
}

// eventHub emits Events to registered clients.
// The Hub seeds new clients with all objects in storage via add events
// and maintains a buffer of sent events to replay to clients.
type eventHub struct {
	list ListerFn

	buffer     indexedBuffer
	broadcast  chan event
	clients    map[EventClient]struct{}
	register   chan EventClient
	deregister chan EventClient
}

func NewHub(list ListerFn) *eventHub {
	return &eventHub{
		list: list,

		buffer:     buffer.NewRingBuffer(100),
		broadcast:  make(chan event),
		clients:    map[EventClient]struct{}{},
		register:   make(chan EventClient),
		deregister: make(chan EventClient),
	}
}

func (h *eventHub) Broadcast(eventType api.EventType, obj api.Object) {
	e := event{
		Type:            eventType,
		Object:          obj,
		ResourceVersion: obj.ObjectMeta().GetResourceVersion(),
	}
	h.broadcast <- e
}

func (h *eventHub) Register(
	resourceVersion string, opts ListOptions,
) (EventClient, error) {
	c := newEventClient(50, resourceVersion, opts, h.deregister)
	h.register <- c
	c.wait()
	return c, c.err
}

func (h *eventHub) Run(stopCh <-chan struct{}) {
	for {
		select {
		case <-stopCh:
			// close all clients and stop
			for c := range h.clients {
				h.closeEventClient(c)
			}
			return

		case c := <-h.register:
			h.clients[c] = struct{}{}
			h.seed(c)

		case c := <-h.deregister:
			h.closeEventClient(c)

		case event := <-h.broadcast:
			e := api.Event{
				Type:   event.Type,
				Object: event.Object,
			}
			h.buffer.Append(event.ResourceVersion, &e)

			for c := range h.clients {
				select {
				case c.eventSink() <- e:
				default:
					// can't send -> Close()
					h.closeEventClient(c)
				}
			}
		}
	}
}

func (h *eventHub) seed(c EventClient) {
	defer c.initDone()

	if len(c.resourceVersion()) > 0 {
		// client requests to continue at a specific resource version.
		// lets see if we have that in the buffer.
		events, ok := h.buffer.FromIndex(c.resourceVersion())
		if !ok {
			c.error(&errors.ErrExpired{
				Message: "Requested ResourceVersion no longer available.",
			})
			h.closeEventClient(c)
			return
		}

		for i, obj := range events {
			if i == 0 {
				// skip first index as the client already has it.
				continue
			}

			e := obj.(*api.Event)
			c.eventSink() <- *e
		}
		return
	}

	// client is not requesting a specific revision,
	// so we seed the client by sending all known objects.
	objects, err := h.list(c.options())
	if err != nil {
		c.error(err)
		h.closeEventClient(c)
		return
	}
	for _, obj := range objects {
		c.eventSink() <- api.Event{
			Type:   api.Added,
			Object: obj,
		}
	}
}

func (h *eventHub) closeEventClient(c EventClient) {
	c.Close()
	delete(h.clients, c)
}
