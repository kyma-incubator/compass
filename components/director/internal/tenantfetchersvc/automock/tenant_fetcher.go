// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// TenantFetcher is an autogenerated mock type for the TenantFetcher type
type TenantFetcher struct {
	mock.Mock
}

// SynchronizeTenant provides a mock function with given fields: ctx, parentTenantID, tenantID
func (_m *TenantFetcher) SynchronizeTenant(ctx context.Context, parentTenantID string, tenantID string) error {
	ret := _m.Called(ctx, parentTenantID, tenantID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, parentTenantID, tenantID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewTenantFetcher creates a new instance of TenantFetcher. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTenantFetcher(t interface {
	mock.TestingT
	Cleanup(func())
}) *TenantFetcher {
	mock := &TenantFetcher{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
