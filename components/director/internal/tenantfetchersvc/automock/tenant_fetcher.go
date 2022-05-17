// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// TenantFetcher is an autogenerated mock type for the TenantFetcher type
type TenantFetcher struct {
	mock.Mock
}

// FetchTenantOnDemand provides a mock function with given fields: ctx, tenantID
func (_m *TenantFetcher) FetchTenantOnDemand(ctx context.Context, tenantID string) error {
	ret := _m.Called(ctx, tenantID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, tenantID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewTenantFetcher creates a new instance of TenantFetcher. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewTenantFetcher(t testing.TB) *TenantFetcher {
	mock := &TenantFetcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
