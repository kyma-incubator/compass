// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// SystemFetcherService is an autogenerated mock type for the SystemFetcherService type
type SystemFetcherService struct {
	mock.Mock
}

// ProcessTenant provides a mock function with given fields: ctx, tenantID
func (_m *SystemFetcherService) ProcessTenant(ctx context.Context, tenantID string) error {
	ret := _m.Called(ctx, tenantID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, tenantID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewSystemFetcherService creates a new instance of SystemFetcherService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSystemFetcherService(t interface {
	mock.TestingT
	Cleanup(func())
}) *SystemFetcherService {
	mock := &SystemFetcherService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}