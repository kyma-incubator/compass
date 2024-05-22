// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// SystemFieldDiscoveryService is an autogenerated mock type for the SystemFieldDiscoveryService type
type SystemFieldDiscoveryService struct {
	mock.Mock
}

// ProcessSaasRegistryApplication provides a mock function with given fields: ctx, appID, tenantID
func (_m *SystemFieldDiscoveryService) ProcessSaasRegistryApplication(ctx context.Context, appID string, tenantID string) error {
	ret := _m.Called(ctx, appID, tenantID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, appID, tenantID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewSystemFieldDiscoveryService creates a new instance of SystemFieldDiscoveryService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSystemFieldDiscoveryService(t interface {
	mock.TestingT
	Cleanup(func())
}) *SystemFieldDiscoveryService {
	mock := &SystemFieldDiscoveryService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}