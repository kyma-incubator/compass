// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	oathkeeper "github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	mock "github.com/stretchr/testify/mock"

	tenantmapping "github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping"

	testing "testing"
)

// ObjectContextProvider is an autogenerated mock type for the ObjectContextProvider type
type ObjectContextProvider struct {
	mock.Mock
}

// GetObjectContext provides a mock function with given fields: ctx, reqData, authDetails
func (_m *ObjectContextProvider) GetObjectContext(ctx context.Context, reqData oathkeeper.ReqData, authDetails oathkeeper.AuthDetails) (tenantmapping.ObjectContext, error) {
	ret := _m.Called(ctx, reqData, authDetails)

	var r0 tenantmapping.ObjectContext
	if rf, ok := ret.Get(0).(func(context.Context, oathkeeper.ReqData, oathkeeper.AuthDetails) tenantmapping.ObjectContext); ok {
		r0 = rf(ctx, reqData, authDetails)
	} else {
		r0 = ret.Get(0).(tenantmapping.ObjectContext)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, oathkeeper.ReqData, oathkeeper.AuthDetails) error); ok {
		r1 = rf(ctx, reqData, authDetails)
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

// NewObjectContextProvider creates a new instance of ObjectContextProvider. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewObjectContextProvider(t testing.TB) *ObjectContextProvider {
	mock := &ObjectContextProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
