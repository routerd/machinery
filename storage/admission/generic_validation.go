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

	if v.Scheme.IsNamespaced(obj) &&
		len(meta.GetNamespace()) == 0 {
		violations = append(violations, &v1.BadRequest_FieldViolation{
			Field:       ".meta.namespace",
			Description: api.NotEmptyDescription,
		})
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
				Field:       fmt.Sprintf(".meta.annoations[%d]", i),
				Description: fmt.Sprintf("invalid key: %s", err.Error()),
			})
		}

		if len(v) > 1024 {
			violations = append(violations, &v1.BadRequest_FieldViolation{
				Field:       fmt.Sprintf(".meta.annoations[%d]", i),
				Description: "invalid value: Must be 1024 characters or less.",
			})
		}
		i++
	}

	i = 0
	for _, finalizer := range meta.GetFinalizers() {
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

func (v *GenericValidation) OnCreate(obj api.Object) error {
	violations := v.validateMetadata(obj)

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

func (v *GenericValidation) OnUpdate(obj api.Object) error {
	return nil
}

func (v *GenericValidation) OnDelete(obj api.Object) error {
	return nil
}
