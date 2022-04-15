// Code generated by mockery v2.10.6. DO NOT EDIT.

package automock

import (
	context "context"

	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	director "github.com/kyma-incubator/compass/components/hydrator/internal/director"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/pkg/model"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// GetRuntimeByTokenIssuer provides a mock function with given fields: ctx, issuer
func (_m *Client) GetRuntimeByTokenIssuer(ctx context.Context, issuer string) (*graphql.Runtime, error) {
	ret := _m.Called(ctx, issuer)

	var r0 *graphql.Runtime
	if rf, ok := ret.Get(0).(func(context.Context, string) *graphql.Runtime); ok {
		r0 = rf(ctx, issuer)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.Runtime)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, issuer)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSystemAuthByID provides a mock function with given fields: ctx, authID
func (_m *Client) GetSystemAuthByID(ctx context.Context, authID string) (*model.SystemAuth, error) {
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

// GetSystemAuthByToken provides a mock function with given fields: ctx, token
func (_m *Client) GetSystemAuthByToken(ctx context.Context, token string) (*model.SystemAuth, error) {
	ret := _m.Called(ctx, token)

	var r0 *model.SystemAuth
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.SystemAuth); ok {
		r0 = rf(ctx, token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.SystemAuth)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTenantByExternalID provides a mock function with given fields: ctx, tenantID
func (_m *Client) GetTenantByExternalID(ctx context.Context, tenantID string) (*graphql.Tenant, error) {
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

// GetTenantByInternalID provides a mock function with given fields: ctx, tenantID
func (_m *Client) GetTenantByInternalID(ctx context.Context, tenantID string) (*graphql.Tenant, error) {
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

// GetTenantByLowestOwnerForResource provides a mock function with given fields: ctx, resourceID, resourceType
func (_m *Client) GetTenantByLowestOwnerForResource(ctx context.Context, resourceID string, resourceType string) (string, error) {
	ret := _m.Called(ctx, resourceID, resourceType)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string) string); ok {
		r0 = rf(ctx, resourceID, resourceType)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, resourceID, resourceType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InvalidateSystemAuthOneTimeToken provides a mock function with given fields: ctx, authID
func (_m *Client) InvalidateSystemAuthOneTimeToken(ctx context.Context, authID string) error {
	ret := _m.Called(ctx, authID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, authID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateSystemAuth provides a mock function with given fields: ctx, sysAuth
func (_m *Client) UpdateSystemAuth(ctx context.Context, sysAuth *model.SystemAuth) (director.UpdateAuthResult, error) {
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
