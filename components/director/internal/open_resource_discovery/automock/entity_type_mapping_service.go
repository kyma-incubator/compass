// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// EntityTypeMappingService is an autogenerated mock type for the EntityTypeMappingService type
type EntityTypeMappingService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, resourceType, resourceID, in
func (_m *EntityTypeMappingService) Create(ctx context.Context, resourceType resource.Type, resourceID string, in *model.EntityTypeMappingInput) (string, error) {
	ret := _m.Called(ctx, resourceType, resourceID, in)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, *model.EntityTypeMappingInput) (string, error)); ok {
		return rf(ctx, resourceType, resourceID, in)
	}
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, *model.EntityTypeMappingInput) string); ok {
		r0 = rf(ctx, resourceType, resourceID, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string, *model.EntityTypeMappingInput) error); ok {
		r1 = rf(ctx, resourceType, resourceID, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, resourceType, id
func (_m *EntityTypeMappingService) Delete(ctx context.Context, resourceType resource.Type, id string) error {
	ret := _m.Called(ctx, resourceType, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string) error); ok {
		r0 = rf(ctx, resourceType, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListByOwnerResourceID provides a mock function with given fields: ctx, resourceID, resourceType
func (_m *EntityTypeMappingService) ListByOwnerResourceID(ctx context.Context, resourceID string, resourceType resource.Type) ([]*model.EntityTypeMapping, error) {
	ret := _m.Called(ctx, resourceID, resourceType)

	var r0 []*model.EntityTypeMapping
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, resource.Type) ([]*model.EntityTypeMapping, error)); ok {
		return rf(ctx, resourceID, resourceType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, resource.Type) []*model.EntityTypeMapping); ok {
		r0 = rf(ctx, resourceID, resourceType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.EntityTypeMapping)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, resource.Type) error); ok {
		r1 = rf(ctx, resourceID, resourceType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewEntityTypeMappingService creates a new instance of EntityTypeMappingService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEntityTypeMappingService(t interface {
	mock.TestingT
	Cleanup(func())
}) *EntityTypeMappingService {
	mock := &EntityTypeMappingService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}