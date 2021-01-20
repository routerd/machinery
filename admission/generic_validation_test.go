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
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/timestamppb"

	"routerd.net/machinery/api"
	v1 "routerd.net/machinery/api/v1"
	"routerd.net/machinery/clientmock"
	"routerd.net/machinery/runtime"
	testdatav1 "routerd.net/machinery/testdata/v1"
)

type validatedTestObject struct {
	api.Object
	mock.Mock
}

func (o *validatedTestObject) ValidateCreate(ctx context.Context) error {
	args := o.Called(ctx)
	err, _ := args.Error(0).(error)
	return err
}

func (o *validatedTestObject) ValidateDelete(ctx context.Context) error {
	args := o.Called(ctx)
	err, _ := args.Error(0).(error)
	return err
}

func (o *validatedTestObject) ValidateUpdate(ctx context.Context, old api.Object) error {
	args := o.Called(ctx)
	err, _ := args.Error(0).(error)
	return err
}

func TestGenericValidation(t *testing.T) {
	t.Run("validateMetadata", func(t *testing.T) {
		tests := []struct {
			name       string
			obj        api.Object
			violations []proto.Message
		}{
			{
				name: "missing metadata",
				obj:  &testdatav1.TestObject{},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta",
						Description: api.NotEmptyDescription,
					},
				},
			},

			{
				name: "name and namespace  missing",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.name",
						Description: api.NotEmptyDescription,
					},
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.namespace",
						Description: api.NotEmptyDescription,
					},
				},
			},

			{
				name: "valid",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
				violations: []proto.Message{},
			},

			{
				name: "invalid name",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "Test",
						Namespace: "test",
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.name",
						Description: "invalid name \"Test\": rfc1035 DNS labels MUST be lowercase, start and end with an alphanumeric character and MUST only contain alphanumeric characters or -",
					},
				},
			},

			{
				name: "invalid namespace",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "testT",
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.namespace",
						Description: "invalid namespace \"testT\": rfc1035 DNS subdomains MUST be lowercase, start and end with an alphanumeric character and MUST only contain alphanumeric characters, - or .",
					},
				},
			},

			{
				name: "invalid label",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Labels: map[string]string{
							"?123": "?",
						},
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.labels[0]",
						Description: "invalid key: must start and end with an alphanumeric character with [a-zA-Z0-9], dashes (-), underscores (_) or dots (.) are allowed between",
					},
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.labels[0]",
						Description: "invalid value: must start and end with an alphanumeric character with [a-zA-Z0-9], dashes (-), underscores (_) or dots (.) are allowed between",
					},
				},
			},

			{
				name: "invalid annotation",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Annotations: map[string]string{
							"?123": strings.Repeat("??", 514),
						},
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.annotations[0]",
						Description: "invalid key: must start and end with an alphanumeric character with [a-zA-Z0-9], dashes (-), underscores (_) or dots (.) are allowed between",
					},
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.annotations[0]",
						Description: "invalid value: Must be 1024 characters or less",
					},
				},
			},

			{
				name: "duplicated finalizer",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Finalizers: []string{
							"fin",
							"fin",
						},
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.finalizers[1]",
						Description: "finalizers can not contain the same key more than once",
					},
				},
			},

			{
				name: "invalid finalizer",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Finalizers: []string{
							"?",
						},
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.finalizers[0]",
						Description: "must start and end with an alphanumeric character with [a-zA-Z0-9], dashes (-), underscores (_) or dots (.) are allowed between",
					},
				},
			},

			{
				name: "invalid uid",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Uid:       "???",
						Name:      "test",
						Namespace: "test",
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.uid",
						Description: "invalid UUID length: 3",
					},
				},
			},
		}

		scheme := runtime.NewScheme()
		scheme.RegisterNamespaced(&testdatav1.TestObject{})

		v := &GenericValidation{
			Scheme: scheme,
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				violations := v.validateMetadata(test.obj)
				if assert.Len(t, violations, len(test.violations)) {
					for i, violation := range test.violations {
						assert.True(t,
							proto.Equal(violation, violations[i]),
							fmt.Sprintf("%v, want %v", violation, violations[i]))
					}
				}
			})
		}
	})

	t.Run("OnCreate", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.RegisterNamespaced(&testdatav1.TestObject{})

		v := &GenericValidation{
			Scheme: scheme,
		}
		ctx := context.Background()

		t.Run("success", func(t *testing.T) {
			obj := &validatedTestObject{
				Object: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
			}

			obj.On("ValidateCreate", mock.Anything).Return(nil)

			err := v.OnCreate(ctx, obj)
			require.NoError(t, err)

			obj.AssertExpectations(t)
		})

		t.Run("error", func(t *testing.T) {
			obj := &validatedTestObject{
				Object: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
			}

			merr := fmt.Errorf("explosion")
			obj.On("ValidateCreate", mock.Anything).Return(merr)

			err := v.OnCreate(ctx, obj)
			assert.EqualError(t, err, merr.Error())

			obj.AssertExpectations(t)
		})
	})

	t.Run("OnUpdate", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.RegisterNamespaced(&testdatav1.TestObject{})

		getter := &clientmock.Getter{}
		v := &GenericValidation{
			Scheme: scheme,
			Getter: getter,
		}
		ctx := context.Background()

		getter.
			On("Get", mock.Anything, api.NamespacedName{
				Name: "test", Namespace: "test",
			}, mock.AnythingOfType("*v1.TestObject")).
			Return(nil)

		t.Run("success", func(t *testing.T) {
			obj := &validatedTestObject{
				Object: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
			}

			obj.On("ValidateUpdate", mock.Anything).Return(nil)

			err := v.OnUpdate(ctx, obj)
			require.NoError(t, err)

			obj.AssertExpectations(t)
		})

		t.Run("error", func(t *testing.T) {
			obj := &validatedTestObject{
				Object: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
			}

			merr := fmt.Errorf("explosion")
			obj.On("ValidateUpdate", mock.Anything, mock.Anything).Return(merr)

			err := v.OnUpdate(ctx, obj)
			assert.EqualError(t, err, merr.Error())

			obj.AssertExpectations(t)
		})
	})

	t.Run("OnDelete", func(t *testing.T) {
		scheme := runtime.NewScheme()
		scheme.RegisterNamespaced(&testdatav1.TestObject{})

		v := &GenericValidation{
			Scheme: scheme,
		}
		ctx := context.Background()

		t.Run("success", func(t *testing.T) {
			obj := &validatedTestObject{
				Object: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
			}

			obj.On("ValidateDelete", mock.Anything).Return(nil)

			err := v.OnDelete(ctx, obj)
			require.NoError(t, err)

			obj.AssertExpectations(t)
		})

		t.Run("error", func(t *testing.T) {
			obj := &validatedTestObject{
				Object: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
			}

			merr := fmt.Errorf("explosion")
			obj.On("ValidateDelete", mock.Anything).Return(merr)

			err := v.OnDelete(ctx, obj)
			assert.EqualError(t, err, merr.Error())

			obj.AssertExpectations(t)
		})
	})

	t.Run("validateMetadata", func(t *testing.T) {
		tests := []struct {
			name       string
			obj        api.Object
			violations []proto.Message
		}{
			{
				name: "missing metadata",
				obj:  &testdatav1.TestObject{},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta",
						Description: api.NotEmptyDescription,
					},
				},
			},

			{
				name: "name and namespace  missing",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.name",
						Description: api.NotEmptyDescription,
					},
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.namespace",
						Description: api.NotEmptyDescription,
					},
				},
			},

			{
				name: "valid",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
				violations: []proto.Message{},
			},

			{
				name: "invalid name",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "Test",
						Namespace: "test",
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.name",
						Description: "invalid name \"Test\": rfc1035 DNS labels MUST be lowercase, start and end with an alphanumeric character and MUST only contain alphanumeric characters or -",
					},
				},
			},

			{
				name: "invalid namespace",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "testT",
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.namespace",
						Description: "invalid namespace \"testT\": rfc1035 DNS subdomains MUST be lowercase, start and end with an alphanumeric character and MUST only contain alphanumeric characters, - or .",
					},
				},
			},

			{
				name: "invalid label",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Labels: map[string]string{
							"?123": "?",
						},
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.labels[0]",
						Description: "invalid key: must start and end with an alphanumeric character with [a-zA-Z0-9], dashes (-), underscores (_) or dots (.) are allowed between",
					},
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.labels[0]",
						Description: "invalid value: must start and end with an alphanumeric character with [a-zA-Z0-9], dashes (-), underscores (_) or dots (.) are allowed between",
					},
				},
			},

			{
				name: "invalid annotation",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Annotations: map[string]string{
							"?123": strings.Repeat("??", 514),
						},
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.annotations[0]",
						Description: "invalid key: must start and end with an alphanumeric character with [a-zA-Z0-9], dashes (-), underscores (_) or dots (.) are allowed between",
					},
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.annotations[0]",
						Description: "invalid value: Must be 1024 characters or less",
					},
				},
			},

			{
				name: "duplicated finalizer",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Finalizers: []string{
							"fin",
							"fin",
						},
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.finalizers[1]",
						Description: "finalizers can not contain the same key more than once",
					},
				},
			},

			{
				name: "invalid finalizer",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Finalizers: []string{
							"?",
						},
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.finalizers[0]",
						Description: "must start and end with an alphanumeric character with [a-zA-Z0-9], dashes (-), underscores (_) or dots (.) are allowed between",
					},
				},
			},

			{
				name: "invalid uid",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Uid:       "???",
						Name:      "test",
						Namespace: "test",
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.uid",
						Description: "invalid UUID length: 3",
					},
				},
			},
		}

		scheme := runtime.NewScheme()
		scheme.RegisterNamespaced(&testdatav1.TestObject{})

		v := &GenericValidation{
			Scheme: scheme,
		}
		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				violations := v.validateMetadata(test.obj)
				if assert.Len(t, violations, len(test.violations)) {
					for i, violation := range test.violations {
						assert.True(t,
							proto.Equal(violation, violations[i]),
							fmt.Sprintf("%v, want %v", violation, violations[i]))
					}
				}
			})
		}
	})

	t.Run("validateUpdate", func(t *testing.T) {
		nowpb := timestamppb.Now()
		tests := []struct {
			name       string
			obj, old   api.Object
			violations []proto.Message
		}{
			{
				name: "success",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
				old: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
					},
				},
			},

			{
				name: "generateName is immutable",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:         "test",
						Namespace:    "test",
						GenerateName: "test-1234",
					},
				},
				old: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:         "test",
						Namespace:    "test",
						GenerateName: "test-123",
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.generateName",
						Description: "is immutable",
					},
				},
			},

			{
				name: "uid is immutable",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Uid:       "b78fcec2-b5ca-4361-a958-290bc231692d",
					},
				},
				old: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:      "test",
						Namespace: "test",
						Uid:       "79703ad0-cdf6-4799-96bf-1d89c8970636",
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.uid",
						Description: "is immutable",
					},
				},
			},

			{
				name: "createdTimestamp is immutable",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:             "test",
						Namespace:        "test",
						CreatedTimestamp: timestamppb.Now(),
					},
				},
				old: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:             "test",
						Namespace:        "test",
						CreatedTimestamp: nowpb,
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.createdTimestamp",
						Description: "is immutable",
					},
				},
			},

			// After Deletion

			{
				name: "deletedTimestamp is immutable after set",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:             "test",
						Namespace:        "test",
						DeletedTimestamp: nil,
					},
				},
				old: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:             "test",
						Namespace:        "test",
						DeletedTimestamp: nowpb,
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.deletedTimestamp",
						Description: "immutable after being set",
					},
				},
			},

			{
				name: "can't add to finalizers",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:             "test",
						Namespace:        "test",
						Finalizers:       []string{"fin1", "fin2"},
						DeletedTimestamp: nowpb,
					},
				},
				old: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:             "test",
						Namespace:        "test",
						Finalizers:       []string{"fin1"},
						DeletedTimestamp: nowpb,
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.finalizers[1]",
						Description: "can't add new finalizers after object deletion",
					},
				},
			},

			{
				name: "can't update something else",
				obj: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:       "test",
						Namespace:  "test",
						Finalizers: []string{"fin1"},
						Annotations: map[string]string{
							"key": "new_value",
						},
						DeletedTimestamp: nowpb,
					},
				},
				old: &testdatav1.TestObject{
					Meta: &v1.ObjectMeta{
						Name:       "test",
						Namespace:  "test",
						Finalizers: []string{"fin1", "fin2"},
						Annotations: map[string]string{
							"key": "value",
						},
						DeletedTimestamp: nowpb,
					},
				},
				violations: []proto.Message{
					&v1.BadRequest_FieldViolation{
						Field:       ".meta.finalizers",
						Description: "object deleted, only finalizers can be updated",
					},
				},
			},
		}

		scheme := runtime.NewScheme()
		scheme.RegisterNamespaced(&testdatav1.TestObject{})

		for _, test := range tests {
			t.Run(test.name, func(t *testing.T) {
				v := &GenericValidation{
					Scheme: scheme,
				}
				ctx := context.Background()

				violations := v.validateUpdate(ctx, test.obj, test.old)
				if assert.Len(t, violations, len(test.violations)) {
					for i, violation := range test.violations {
						assert.True(t,
							proto.Equal(violation, violations[i]),
							fmt.Sprintf("got %v, want %v", violations[i], violation))
					}
				}
			})
		}
	})
}
