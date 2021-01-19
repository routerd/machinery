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

package admission

import (
	"context"
	"fmt"

	protoV1 "github.com/golang/protobuf/proto"
	"github.com/google/uuid"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"

	"routerd.net/machinery/api"
	v1 "routerd.net/machinery/api/v1"
	"routerd.net/machinery/runtime"
	"routerd.net/machinery/validate"
)

type GenericValidation struct {
	Scheme *runtime.Scheme
	Getter api.Getter
}

func (v *GenericValidation) validateMetadata(obj api.Object) (
	violations []proto.Message) {
	meta := obj.ObjectMeta()
	if meta == nil {
		violations = append(violations, &v1.BadRequest_FieldViolation{
			Field:       ".meta",
			Description: api.NotEmptyDescription,
		})
		return
	}

	if len(meta.GetName()) == 0 {
		// .meta.name MUST be set.
		violations = append(violations, &v1.BadRequest_FieldViolation{
			Field:       ".meta.name",
			Description: api.NotEmptyDescription,
		})
	} else {
		if err := validate.ValidateName(meta.GetName()); err != nil {
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       ".meta.name",
				Description: err.Error(),
			})
		}
	}

	if v.Scheme.IsNamespaced(obj) {
		if len(meta.GetNamespace()) == 0 {
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       ".meta.namespace",
				Description: api.NotEmptyDescription,
			})
		} else if err := validate.ValidateNamespace(meta.GetNamespace()); err != nil {
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       ".meta.namespace",
				Description: err.Error(),
			})
		}
	}

	var i int
	for k, v := range meta.GetLabels() {
		if err := validate.ValidateQualifiedKey(k); err != nil {
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       fmt.Sprintf(".meta.labels[%d]", i),
				Description: fmt.Sprintf("invalid key: %s", err.Error()),
			})
		}

		if err := validate.ValidateKey(v); err != nil {
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       fmt.Sprintf(".meta.labels[%d]", i),
				Description: fmt.Sprintf("invalid value: %s", err.Error()),
			})
		}
		i++
	}

	i = 0
	for k, v := range meta.GetAnnotations() {
		if err := validate.ValidateQualifiedKey(k); err != nil {
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       fmt.Sprintf(".meta.annotations[%d]", i),
				Description: fmt.Sprintf("invalid key: %s", err.Error()),
			})
		}

		if len(v) > 1024 {
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       fmt.Sprintf(".meta.annotations[%d]", i),
				Description: "invalid value: Must be 1024 characters or less",
			})
		}
		i++
	}

	finalizers := map[string]struct{}{}
	i = 0
	for _, finalizer := range meta.GetFinalizers() {
		if _, ok := finalizers[finalizer]; ok {
			// no duplicated finalizers
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       fmt.Sprintf(".meta.finalizers[%d]", i),
				Description: "finalizers can not contain the same key more than once",
			})
		} else {
			finalizers[finalizer] = struct{}{}
		}

		if err := validate.ValidateQualifiedKey(finalizer); err != nil {
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       fmt.Sprintf(".meta.finalizers[%d]", i),
				Description: err.Error(),
			})
		}
		i++
	}

	if len(meta.GetUid()) > 0 {
		if _, err := uuid.Parse(meta.GetUid()); err != nil {
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       ".meta.uid",
				Description: err.Error(),
			})
		}
	}

	return violations
}

type validateCreate interface {
	ValidateCreate(ctx context.Context) error
}

func (v *GenericValidation) OnCreate(ctx context.Context, obj api.Object) error {
	violations := v.validateMetadata(obj)

	if vobj, ok := obj.(validateCreate); ok {
		err := vobj.ValidateCreate(ctx)
		s, ok := status.FromError(err)
		if ok && s.Code() == codes.InvalidArgument {
			// add the details from the status
			violations = append(violations, asProtoMessages(s.Details())...)
		} else if err != nil {
			// unknown error/other status code
			return err
		}
	}
	return checkViolations(violations)
}

type validateUpdate interface {
	ValidateUpdate(ctx context.Context, old api.Object) error
}

func (v *GenericValidation) OnUpdate(ctx context.Context, obj api.Object) error {
	violations := v.validateMetadata(obj)

	old := proto.Clone(obj).(api.Object)
	err := v.Getter.Get(ctx, api.NamespacedName{
		Name:      obj.ObjectMeta().GetName(),
		Namespace: obj.ObjectMeta().GetNamespace(),
	}, old)
	if err != nil {
		return err
	}

	violations = append(violations, v.validateUpdate(ctx, obj, old)...)

	if vobj, ok := obj.(validateUpdate); ok {
		err := vobj.ValidateUpdate(ctx, old)
		s, ok := status.FromError(err)
		if ok && s.Code() == codes.InvalidArgument {
			// add the details from the status
			violations = append(violations, asProtoMessages(s.Details())...)
		} else if err != nil {
			// unknown error/other status code
			return err
		}
	}
	return checkViolations(violations)
}

func (v *GenericValidation) validateUpdate(ctx context.Context, obj, old api.Object) []proto.Message {
	var violations []proto.Message

	// Finalizer Handling
	if obj.ObjectMeta().GetDeletedTimestamp() != nil ||
		old.ObjectMeta().GetDeletedTimestamp() != nil {
		if old.ObjectMeta().GetDeletedTimestamp() != nil &&
			!obj.ObjectMeta().GetDeletedTimestamp().AsTime().Equal(
				old.ObjectMeta().GetDeletedTimestamp().AsTime()) {
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       ".meta.deletedTimestamp",
				Description: "immutable after being set",
			})
		}

		// don't allow adding new finalizers
		oldFinalizers := map[string]struct{}{}
		for _, finalizer := range old.ObjectMeta().GetFinalizers() {
			oldFinalizers[finalizer] = struct{}{}
		}
		for i, finalizer := range obj.ObjectMeta().GetFinalizers() {
			if _, ok := oldFinalizers[finalizer]; !ok {
				violations = append(violations, &v1.BadRequest_FieldViolation{
					Field:       fmt.Sprintf(".meta.finalizers[%d]", i),
					Description: "can't add new finalizers after object deletion",
				})
			}
		}

		// only allow updates to finalizers
		oldWithNewFinalizers := proto.Clone(old).(api.Object)
		oldWithNewFinalizers.ObjectMeta().SetFinalizers(
			obj.ObjectMeta().GetFinalizers())
		oldWithNewFinalizers.ObjectMeta().SetDeletedTimestamp(
			obj.ObjectMeta().GetDeletedTimestamp())

		if !proto.Equal(obj, oldWithNewFinalizers) {
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       ".meta.finalizers",
				Description: "object deleted, only finalizers can be updated",
			})
		}
	}

	// immutable metadata fields
	if obj.ObjectMeta().GetGenerateName() !=
		old.ObjectMeta().GetGenerateName() {
		violations = append(violations, &v1.BadRequest_FieldViolation{
			Field:       ".meta.generateName",
			Description: api.Immutable,
		})
	}
	if obj.ObjectMeta().GetUid() !=
		old.ObjectMeta().GetUid() {
		violations = append(violations, &v1.BadRequest_FieldViolation{
			Field:       ".meta.uid",
			Description: api.Immutable,
		})
	}
	if !obj.ObjectMeta().GetCreatedTimestamp().AsTime().Equal(
		old.ObjectMeta().GetCreatedTimestamp().AsTime()) {
		violations = append(violations, &v1.BadRequest_FieldViolation{
			Field:       ".meta.createdTimestamp",
			Description: api.Immutable,
		})
	}

	return violations
}

type validateDelete interface {
	ValidateDelete(ctx context.Context) error
}

func (v *GenericValidation) OnDelete(ctx context.Context, obj api.Object) error {
	violations := v.validateMetadata(obj)

	if vobj, ok := obj.(validateDelete); ok {
		err := vobj.ValidateDelete(ctx)
		s, ok := status.FromError(err)
		if ok && s.Code() == codes.InvalidArgument {
			// add the details from the status
			violations = append(violations, asProtoMessages(s.Details())...)
		} else if err != nil {
			// unknown error/other status code
			return err
		}
	}
	return checkViolations(violations)
}

func asProtoMessages(list []interface{}) []proto.Message {
	var messages []proto.Message
	for _, i := range list {
		msg, ok := i.(proto.Message)
		if ok {
			messages = append(messages, msg)
		}
	}
	return messages
}

func checkViolations(violations []proto.Message) error {
	if len(violations) > 0 {
		violationsV1 := make([]protoV1.Message, len(violations))
		for i, v := range violations {
			violationsV1[i] = protoV1.MessageV1(v)
		}

		s, err := status.
			New(codes.InvalidArgument, "Validation failed.").
			WithDetails(violationsV1...)
		if err != nil {
			return err
		}
		return s.Err()
	}
	return nil
}
