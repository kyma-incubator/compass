// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// AspectService is an autogenerated mock type for the AspectService type
type AspectService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, resourceType, resourceID, integrationDependencyID, in
func (_m *AspectService) Create(ctx context.Context, resourceType resource.Type, resourceID string, integrationDependencyID string, in model.AspectInput) (string, error) {
	ret := _m.Called(ctx, resourceType, resourceID, integrationDependencyID, in)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, string, model.AspectInput) (string, error)); ok {
		return rf(ctx, resourceType, resourceID, integrationDependencyID, in)
	}
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, string, model.AspectInput) string); ok {
		r0 = rf(ctx, resourceType, resourceID, integrationDependencyID, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string, string, model.AspectInput) error); ok {
		r1 = rf(ctx, resourceType, resourceID, integrationDependencyID, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteByIntegrationDependencyID provides a mock function with given fields: ctx, integrationDependencyID
func (_m *AspectService) DeleteByIntegrationDependencyID(ctx context.Context, integrationDependencyID string) error {
	ret := _m.Called(ctx, integrationDependencyID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, integrationDependencyID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewAspectService creates a new instance of AspectService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAspectService(t interface {
	mock.TestingT
	Cleanup(func())
}) *AspectService {
	mock := &AspectService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
