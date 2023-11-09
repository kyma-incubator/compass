// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// IntegrationDependencyService is an autogenerated mock type for the IntegrationDependencyService type
type IntegrationDependencyService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, resourceType, resourceID, packageID, in, integrationDependencyHash
func (_m *IntegrationDependencyService) Create(ctx context.Context, resourceType resource.Type, resourceID string, packageID *string, in model.IntegrationDependencyInput, integrationDependencyHash uint64) (string, error) {
	ret := _m.Called(ctx, resourceType, resourceID, packageID, in, integrationDependencyHash)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, *string, model.IntegrationDependencyInput, uint64) (string, error)); ok {
		return rf(ctx, resourceType, resourceID, packageID, in, integrationDependencyHash)
	}
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, *string, model.IntegrationDependencyInput, uint64) string); ok {
		r0 = rf(ctx, resourceType, resourceID, packageID, in, integrationDependencyHash)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string, *string, model.IntegrationDependencyInput, uint64) error); ok {
		r1 = rf(ctx, resourceType, resourceID, packageID, in, integrationDependencyHash)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByApplicationID provides a mock function with given fields: ctx, appID
func (_m *IntegrationDependencyService) ListByApplicationID(ctx context.Context, appID string) ([]*model.IntegrationDependency, error) {
	ret := _m.Called(ctx, appID)

	var r0 []*model.IntegrationDependency
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*model.IntegrationDependency, error)); ok {
		return rf(ctx, appID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.IntegrationDependency); ok {
		r0 = rf(ctx, appID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.IntegrationDependency)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, appID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByApplicationTemplateVersionID provides a mock function with given fields: ctx, appTemplateVersionID
func (_m *IntegrationDependencyService) ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.IntegrationDependency, error) {
	ret := _m.Called(ctx, appTemplateVersionID)

	var r0 []*model.IntegrationDependency
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*model.IntegrationDependency, error)); ok {
		return rf(ctx, appTemplateVersionID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.IntegrationDependency); ok {
		r0 = rf(ctx, appTemplateVersionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.IntegrationDependency)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, appTemplateVersionID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, resourceType, resourceID, id, in, integrationDependencyHash
func (_m *IntegrationDependencyService) Update(ctx context.Context, resourceType resource.Type, resourceID string, id string, in model.IntegrationDependencyInput, integrationDependencyHash uint64) error {
	ret := _m.Called(ctx, resourceType, resourceID, id, in, integrationDependencyHash)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, string, model.IntegrationDependencyInput, uint64) error); ok {
		r0 = rf(ctx, resourceType, resourceID, id, in, integrationDependencyHash)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewIntegrationDependencyService creates a new instance of IntegrationDependencyService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIntegrationDependencyService(t interface {
	mock.TestingT
	Cleanup(func())
}) *IntegrationDependencyService {
	mock := &IntegrationDependencyService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}