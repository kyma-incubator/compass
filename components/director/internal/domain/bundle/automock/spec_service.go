// Code generated by mockery v2.5.1. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// SpecService is an autogenerated mock type for the SpecService type
type SpecService struct {
	mock.Mock
}

// CreateByReferenceObjectID provides a mock function with given fields: ctx, in, objectType, objectID
func (_m *SpecService) CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) (string, error) {
	ret := _m.Called(ctx, in, objectType, objectID)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, model.SpecInput, model.SpecReferenceObjectType, string) string); ok {
		r0 = rf(ctx, in, objectType, objectID)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.SpecInput, model.SpecReferenceObjectType, string) error); ok {
		r1 = rf(ctx, in, objectType, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByReferenceObjectID provides a mock function with given fields: ctx, objectType, objectID
func (_m *SpecService) GetByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) (*model.Spec, error) {
	ret := _m.Called(ctx, objectType, objectID)

	var r0 *model.Spec
	if rf, ok := ret.Get(0).(func(context.Context, model.SpecReferenceObjectType, string) *model.Spec); ok {
		r0 = rf(ctx, objectType, objectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Spec)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.SpecReferenceObjectType, string) error); ok {
		r1 = rf(ctx, objectType, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RefetchSpec provides a mock function with given fields: ctx, id
func (_m *SpecService) RefetchSpec(ctx context.Context, id string) (*model.Spec, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Spec
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Spec); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Spec)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateByReferenceObjectID provides a mock function with given fields: ctx, id, in, objectType, objectID
func (_m *SpecService) UpdateByReferenceObjectID(ctx context.Context, id string, in model.SpecInput, objectType model.SpecReferenceObjectType, objectID string) error {
	ret := _m.Called(ctx, id, in, objectType, objectID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.SpecInput, model.SpecReferenceObjectType, string) error); ok {
		r0 = rf(ctx, id, in, objectType, objectID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
