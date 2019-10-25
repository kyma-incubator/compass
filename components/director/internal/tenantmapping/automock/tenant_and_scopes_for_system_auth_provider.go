// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import context "context"
import mock "github.com/stretchr/testify/mock"
import tenantmapping "github.com/kyma-incubator/compass/components/director/internal/tenantmapping"

// TenantAndScopesForSystemAuthProvider is an autogenerated mock type for the TenantAndScopesForSystemAuthProvider type
type TenantAndScopesForSystemAuthProvider struct {
	mock.Mock
}

// GetTenantAndScopes provides a mock function with given fields: ctx, reqData, authID, authFlow
func (_m *TenantAndScopesForSystemAuthProvider) GetTenantAndScopes(ctx context.Context, reqData tenantmapping.ReqData, authID string, authFlow tenantmapping.AuthFlow) (string, string, error) {
	ret := _m.Called(ctx, reqData, authID, authFlow)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, tenantmapping.ReqData, string, tenantmapping.AuthFlow) string); ok {
		r0 = rf(ctx, reqData, authID, authFlow)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 string
	if rf, ok := ret.Get(1).(func(context.Context, tenantmapping.ReqData, string, tenantmapping.AuthFlow) string); ok {
		r1 = rf(ctx, reqData, authID, authFlow)
	} else {
		r1 = ret.Get(1).(string)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, tenantmapping.ReqData, string, tenantmapping.AuthFlow) error); ok {
		r2 = rf(ctx, reqData, authID, authFlow)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
