// Code generated by mockery 2.9.0. DO NOT EDIT.

package automock

import (
	context "context"

	oathkeeper "github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	tenantmapping "github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	mock "github.com/stretchr/testify/mock"
)

// ObjectContextProvider is an autogenerated mock type for the ObjectContextProvider type
type ObjectContextProvider struct {
	mock.Mock
}

// GetObjectContext provides a mock function with given fields: ctx, reqData, authDetails, extraTenantKeys
func (_m *ObjectContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails, extraTenantKeys tenantmapping.KeysExtra) (tenantmapping.ObjectContext, error) {
	ret := _m.Called(ctx, reqData, authDetails, extraTenantKeys)

	var r0 tenantmapping.ObjectContext
	if rf, ok := ret.Get(0).(func(context.Context, oathkeeper.ReqData, oathkeeper.AuthDetails, tenantmapping.KeysExtra) tenantmapping.ObjectContext); ok {
		r0 = rf(ctx, reqData, authDetails, extraTenantKeys)
	} else {
		r0 = ret.Get(0).(tenantmapping.ObjectContext)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, oathkeeper.ReqData, oathkeeper.AuthDetails, tenantmapping.KeysExtra) error); ok {
		r1 = rf(ctx, reqData, authDetails, extraTenantKeys)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Match provides a mock function with given fields: ctx, data
func (_m *ObjectContextProvider) Match(ctx context.Context, data oathkeeper.ReqData) (bool, *oathkeeper.AuthDetails, error) {
	ret := _m.Called(ctx, data)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, oathkeeper.ReqData) bool); ok {
		r0 = rf(ctx, data)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 *oathkeeper.AuthDetails
	if rf, ok := ret.Get(1).(func(context.Context, oathkeeper.ReqData) *oathkeeper.AuthDetails); ok {
		r1 = rf(ctx, data)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*oathkeeper.AuthDetails)
		}
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(context.Context, oathkeeper.ReqData) error); ok {
		r2 = rf(ctx, data)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
