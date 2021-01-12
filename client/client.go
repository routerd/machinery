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

package client

import (
	"context"
	"io"
	"reflect"

	grpc "google.golang.org/grpc"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/anypb"

	"routerd.net/machinery/api"
	machineryv1 "routerd.net/machinery/api/v1"
)

var _ api.Client = (*GRPCClient)(nil)

type GRPCClient struct {
	get          func(ctx context.Context, nn api.NamespacedName, obj api.Object) error
	list         func(ctx context.Context, listObj api.ListObject, opts api.ListOptions) error
	delete       func(ctx context.Context, obj api.Object, opts api.DeleteOptions) error
	deleteAllOf  func(ctx context.Context, opts api.DeleteAllOfOptions) error
	create       func(ctx context.Context, obj api.Object, opts api.CreateOptions) error
	update       func(ctx context.Context, obj api.Object, opts api.UpdateOptions) error
	updateStatus func(ctx context.Context, obj api.Object, opts api.UpdateOptions) error
	watch        func(ctx context.Context, opts api.WatchOptions) (grpc.ClientStream, error)
}

func NewGRPCClient(client interface{}) (*GRPCClient, error) {
	c := &GRPCClient{}
	rv := reflect.ValueOf(client)

	if getter := rv.MethodByName("Get"); getter.IsValid() {
		// TODO: check signature
		c.get = func(ctx context.Context, nn api.NamespacedName, obj api.Object) error {

			rargs := getter.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(&machineryv1.GetRequest{
					Name:      nn.Name,
					Namespace: nn.Namespace,
				}),
			})
			if !rargs[0].IsNil() {
				// write into given object
				proto.Merge(obj, rargs[0].Interface().(api.Object))
			}
			err, _ := rargs[1].Interface().(error)
			return err
		}
	}

	if lister := rv.MethodByName("List"); lister.IsValid() {
		// TODO: check signature
		c.list = func(ctx context.Context, listObj api.ListObject, opts api.ListOptions) error {

			req := &machineryv1.ListRequest{
				Namespace: opts.Namespace,
			}
			if opts.LabelSelector != nil {
				req.LabelSelector = opts.LabelSelector.String()
			}

			rargs := lister.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(req),
			})
			if !rargs[0].IsNil() {
				// write into given object
				proto.Merge(listObj, rargs[0].Interface().(api.ListObject))
			}
			err, _ := rargs[1].Interface().(error)
			return err
		}
	}

	if delete := rv.MethodByName("Delete"); delete.IsValid() {
		deleteType := delete.Type()
		// TODO: check signature
		c.delete = func(ctx context.Context, obj api.Object, opts api.DeleteOptions) error {
			req := reflect.New(deleteType.In(1).Elem())

			req.Elem().FieldByName("Object").
				Set(reflect.ValueOf(obj))

			rargs := delete.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				req,
			})
			if !rargs[0].IsNil() {
				// write into given object
				proto.Merge(obj, rargs[0].Interface().(api.Object))
			}
			err, _ := rargs[1].Interface().(error)
			return err
		}
	}

	if create := rv.MethodByName("Create"); create.IsValid() {
		createType := create.Type()
		// TODO: check signature
		c.create = func(ctx context.Context, obj api.Object, opts api.CreateOptions) error {
			req := reflect.New(createType.In(1).Elem())

			req.Elem().FieldByName("Object").
				Set(reflect.ValueOf(obj))

			rargs := create.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				req,
			})
			if !rargs[0].IsNil() {
				// write into given object
				proto.Merge(obj, rargs[0].Interface().(api.Object))
			}
			err, _ := rargs[1].Interface().(error)
			return err
		}
	}

	if delete := rv.MethodByName("DeleteAllOf"); delete.IsValid() {
		// TODO: check signature
		c.deleteAllOf = func(ctx context.Context, opts api.DeleteAllOfOptions) error {

			req := &machineryv1.DeleteAllOfRequest{
				Namespace: opts.Namespace,
			}
			if opts.LabelSelector != nil {
				req.LabelSelector = opts.LabelSelector.String()
			}

			rargs := delete.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(req),
			})
			err, _ := rargs[1].Interface().(error)
			return err
		}
	}

	if update := rv.MethodByName("Update"); update.IsValid() {
		updateType := update.Type()
		// TODO: check signature
		c.update = func(ctx context.Context, obj api.Object, opts api.UpdateOptions) error {
			req := reflect.New(updateType.In(1).Elem())

			req.Elem().FieldByName("Object").
				Set(reflect.ValueOf(obj))

			rargs := update.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				req,
			})
			if !rargs[0].IsNil() {
				// write into given object
				proto.Merge(obj, rargs[0].Interface().(api.Object))
			}
			err, _ := rargs[1].Interface().(error)
			return err
		}
	}

	if update := rv.MethodByName("UpdateStatus"); update.IsValid() {
		updateType := update.Type()
		// TODO: check signature
		c.updateStatus = func(ctx context.Context, obj api.Object, opts api.UpdateOptions) error {
			req := reflect.New(updateType.In(1).Elem())

			req.Elem().FieldByName("Object").
				Set(reflect.ValueOf(obj))

			rargs := update.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				req,
			})
			if !rargs[0].IsNil() {
				// write into given object
				proto.Merge(obj, rargs[0].Interface().(api.Object))
			}
			err, _ := rargs[1].Interface().(error)
			return err
		}
	}

	if watch := rv.MethodByName("Watch"); watch.IsValid() {
		// TODO: check signature
		c.watch = func(ctx context.Context, opts api.WatchOptions) (grpc.ClientStream, error) {

			req := &machineryv1.WatchRequest{
				Namespace: opts.Namespace,
			}
			if opts.LabelSelector != nil {
				req.LabelSelector = opts.LabelSelector.String()
			}

			rargs := watch.Call([]reflect.Value{
				reflect.ValueOf(ctx),
				reflect.ValueOf(req),
			})

			stream, _ := rargs[0].Interface().(grpc.ClientStream)
			err, _ := rargs[1].Interface().(error)
			return stream, err
		}
	}

	return c, nil
}

func (c *GRPCClient) Get(ctx context.Context, nn api.NamespacedName, obj api.Object) error {
	return c.get(ctx, nn, obj)
}

func (c *GRPCClient) List(ctx context.Context, listObj api.ListObject, opts ...api.ListOption) error {
	var listOpts api.ListOptions
	for _, opt := range opts {
		opt.ApplyToList(&listOpts)
	}
	return c.list(ctx, listObj, listOpts)
}

func (c *GRPCClient) Create(ctx context.Context, obj api.Object, opts ...api.CreateOption) error {
	var createOpts api.CreateOptions
	for _, opt := range opts {
		opt.ApplyToCreate(&createOpts)
	}
	return c.create(ctx, obj, createOpts)
}

func (c *GRPCClient) Delete(ctx context.Context, obj api.Object, opts ...api.DeleteOption) error {
	var deleteOpts api.DeleteOptions
	for _, opt := range opts {
		opt.ApplyToDelete(&deleteOpts)
	}
	return c.delete(ctx, obj, deleteOpts)
}

func (c *GRPCClient) DeleteAllOf(ctx context.Context, obj api.Object, opts ...api.DeleteAllOfOption) error {
	var deleteOpts api.DeleteAllOfOptions
	for _, opt := range opts {
		opt.ApplyToDeleteAllOf(&deleteOpts)
	}
	return c.deleteAllOf(ctx, deleteOpts)
}

func (c *GRPCClient) Update(ctx context.Context, obj api.Object, opts ...api.UpdateOption) error {
	var updateOpts api.UpdateOptions
	for _, opt := range opts {
		opt.ApplyToUpdate(&updateOpts)
	}
	return c.update(ctx, obj, updateOpts)
}

func (c *GRPCClient) UpdateStatus(ctx context.Context, obj api.Object, opts ...api.UpdateOption) error {
	var updateOpts api.UpdateOptions
	for _, opt := range opts {
		opt.ApplyToUpdate(&updateOpts)
	}
	return c.updateStatus(ctx, obj, updateOpts)
}

func (c *GRPCClient) Watch(ctx context.Context,
	obj api.Object, opts ...api.WatchOption) (api.WatchClient, error) {
	var watchOpts api.WatchOptions
	for _, opt := range opts {
		opt.ApplyToWatch(&watchOpts)
	}

	w := &grpcWatcher{}
	ctx, w.cancel = context.WithCancel(ctx)

	var err error
	w.stream, err = c.watch(ctx, watchOpts)
	if err != nil {
		return nil, err
	}
	return w, nil
}

type grpcWatcher struct {
	stream grpc.ClientStream
	cancel func()
}

func (w *grpcWatcher) Events() <-chan api.ResourceEvent {
	e := make(chan api.ResourceEvent, 10)
	go func(events chan api.ResourceEvent) {
		defer close(events)
		for {
			event := &machineryv1.ResourceEvent{}
			err := w.stream.RecvMsg(event)
			if err == io.EOF {
				return
			}
			if err != nil {
				events <- api.ResourceEvent{
					Type: api.Error,
					// TODO: add error details
				}
				return
			}

			obj, err := anypb.UnmarshalNew(
				event.Object, proto.UnmarshalOptions{})
			if err != nil {
				events <- api.ResourceEvent{
					Type: api.Error,
				}
				return
			}

			events <- api.ResourceEvent{
				Type:   api.ResourceEventType(event.Type),
				Object: obj.(api.Object),
			}
		}
	}(e)
	return e
}

func (w *grpcWatcher) Close() error {
	w.cancel()
	return nil
}
