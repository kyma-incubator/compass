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

// Create provides a mock function with given fields: ctx, resourceType, resourceID, bundleID, packageID, in, spec, targetURLsPerBundle, apiHash, defaultBundleID
func (_m *APIService) Create(ctx context.Context, resourceType resource.Type, resourceID string, bundleID *string, packageID *string, in model.APIDefinitionInput, spec []*model.SpecInput, targetURLsPerBundle map[string]string, apiHash uint64, defaultBundleID string) (string, error) {
	ret := _m.Called(ctx, resourceType, resourceID, bundleID, packageID, in, spec, targetURLsPerBundle, apiHash, defaultBundleID)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, *string, *string, model.APIDefinitionInput, []*model.SpecInput, map[string]string, uint64, string) string); ok {
		r0 = rf(ctx, resourceType, resourceID, bundleID, packageID, in, spec, targetURLsPerBundle, apiHash, defaultBundleID)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string, *string, *string, model.APIDefinitionInput, []*model.SpecInput, map[string]string, uint64, string) error); ok {
		r1 = rf(ctx, resourceType, resourceID, bundleID, packageID, in, spec, targetURLsPerBundle, apiHash, defaultBundleID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *APIService) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListByApplicationID provides a mock function with given fields: ctx, appID
func (_m *APIService) ListByApplicationID(ctx context.Context, appID string) ([]*model.APIDefinition, error) {
	ret := _m.Called(ctx, appID)

	var r0 []*model.APIDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.APIDefinition); ok {
		r0 = rf(ctx, appID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.APIDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, appID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByApplicationTemplateVersionID provides a mock function with given fields: ctx, appTemplateVersionID
func (_m *APIService) ListByApplicationTemplateVersionID(ctx context.Context, appTemplateVersionID string) ([]*model.APIDefinition, error) {
	ret := _m.Called(ctx, appTemplateVersionID)

	var r0 []*model.APIDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.APIDefinition); ok {
		r0 = rf(ctx, appTemplateVersionID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.APIDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, appTemplateVersionID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateInManyBundles provides a mock function with given fields: ctx, resourceType, id, in, specIn, defaultTargetURLPerBundle, defaultTargetURLPerBundleToBeCreated, bundleIDsToBeDeleted, apiHash, defaultBundleID
func (_m *APIService) UpdateInManyBundles(ctx context.Context, resourceType resource.Type, id string, in model.APIDefinitionInput, specIn *model.SpecInput, defaultTargetURLPerBundle map[string]string, defaultTargetURLPerBundleToBeCreated map[string]string, bundleIDsToBeDeleted []string, apiHash uint64, defaultBundleID string) error {
	ret := _m.Called(ctx, resourceType, id, in, specIn, defaultTargetURLPerBundle, defaultTargetURLPerBundleToBeCreated, bundleIDsToBeDeleted, apiHash, defaultBundleID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, model.APIDefinitionInput, *model.SpecInput, map[string]string, map[string]string, []string, uint64, string) error); ok {
		r0 = rf(ctx, resourceType, id, in, specIn, defaultTargetURLPerBundle, defaultTargetURLPerBundleToBeCreated, bundleIDsToBeDeleted, apiHash, defaultBundleID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
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
