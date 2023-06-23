// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// APIService is an autogenerated mock type for the APIService type
type APIService struct {
	mock.Mock
}

// CreateInBundle provides a mock function with given fields: ctx, resourceType, resourceID, bundleID, in, spec
func (_m *APIService) CreateInBundle(ctx context.Context, resourceType resource.Type, resourceID string, bundleID string, in model.APIDefinitionInput, spec *model.SpecInput) (string, error) {
	ret := _m.Called(ctx, resourceType, resourceID, bundleID, in, spec)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, string, model.APIDefinitionInput, *model.SpecInput) string); ok {
		r0 = rf(ctx, resourceType, resourceID, bundleID, in, spec)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string, string, model.APIDefinitionInput, *model.SpecInput) error); ok {
		r1 = rf(ctx, resourceType, resourceID, bundleID, in, spec)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteAllByBundleID provides a mock function with given fields: ctx, bundleID
func (_m *APIService) DeleteAllByBundleID(ctx context.Context, bundleID string) error {
	ret := _m.Called(ctx, bundleID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, bundleID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetForBundle provides a mock function with given fields: ctx, id, bundleID
func (_m *APIService) GetForBundle(ctx context.Context, id string, bundleID string) (*model.APIDefinition, error) {
	ret := _m.Called(ctx, id, bundleID)

	var r0 *model.APIDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.APIDefinition); ok {
		r0 = rf(ctx, id, bundleID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.APIDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, bundleID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByBundleIDs provides a mock function with given fields: ctx, bundleIDs, pageSize, cursor
func (_m *APIService) ListByBundleIDs(ctx context.Context, bundleIDs []string, pageSize int, cursor string) ([]*model.APIDefinitionPage, error) {
	ret := _m.Called(ctx, bundleIDs, pageSize, cursor)

	var r0 []*model.APIDefinitionPage
	if rf, ok := ret.Get(0).(func(context.Context, []string, int, string) []*model.APIDefinitionPage); ok {
		r0 = rf(ctx, bundleIDs, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.APIDefinitionPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string, int, string) error); ok {
		r1 = rf(ctx, bundleIDs, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewAPIService interface {
	mock.TestingT
	Cleanup(func())
}

// NewAPIService creates a new instance of APIService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAPIService(t mockConstructorTestingTNewAPIService) *APIService {
	mock := &APIService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
