// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	director "github.com/kyma-incubator/compass/components/hydrator/internal/director"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/pkg/model"

	testing "testing"
)

// DirectorClient is an autogenerated mock type for the DirectorClient type
type DirectorClient struct {
	mock.Mock
}

// GetSystemAuthByID provides a mock function with given fields: ctx, authID
func (_m *DirectorClient) GetSystemAuthByID(ctx context.Context, authID string) (*model.SystemAuth, error) {
	ret := _m.Called(ctx, authID)

	var r0 *model.SystemAuth
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.SystemAuth); ok {
		r0 = rf(ctx, authID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.SystemAuth)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, authID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTenantByExternalID provides a mock function with given fields: ctx, tenantID
func (_m *DirectorClient) GetTenantByExternalID(ctx context.Context, tenantID string) (*graphql.Tenant, error) {
	ret := _m.Called(ctx, tenantID)

	var r0 *graphql.Tenant
	if rf, ok := ret.Get(0).(func(context.Context, string) *graphql.Tenant); ok {
		r0 = rf(ctx, tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.Tenant)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateSystemAuth provides a mock function with given fields: ctx, sysAuth
func (_m *DirectorClient) UpdateSystemAuth(ctx context.Context, sysAuth *model.SystemAuth) (director.UpdateAuthResult, error) {
	ret := _m.Called(ctx, sysAuth)

	var r0 director.UpdateAuthResult
	if rf, ok := ret.Get(0).(func(context.Context, *model.SystemAuth) director.UpdateAuthResult); ok {
		r0 = rf(ctx, sysAuth)
	} else {
		r0 = ret.Get(0).(director.UpdateAuthResult)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *model.SystemAuth) error); ok {
		r1 = rf(ctx, sysAuth)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WriteTenants provides a mock function with given fields: ctx, tenants
func (_m *DirectorClient) WriteTenants(ctx context.Context, tenants []graphql.BusinessTenantMappingInput) error {
	ret := _m.Called(ctx, tenants)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []graphql.BusinessTenantMappingInput) error); ok {
		r0 = rf(ctx, tenants)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewDirectorClient creates a new instance of DirectorClient. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewDirectorClient(t testing.TB) *DirectorClient {
	mock := &DirectorClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
