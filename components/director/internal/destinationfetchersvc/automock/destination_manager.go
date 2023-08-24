// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// DestinationManager is an autogenerated mock type for the DestinationManager type
type DestinationManager struct {
	mock.Mock
}

// FetchDestinationsSensitiveData provides a mock function with given fields: ctx, tenantID, destinationNames
func (_m *DestinationManager) FetchDestinationsSensitiveData(ctx context.Context, tenantID string, destinationNames []string) ([]byte, error) {
	ret := _m.Called(ctx, tenantID, destinationNames)

	var r0 []byte
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) []byte); ok {
		r0 = rf(ctx, tenantID, destinationNames)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string) error); ok {
		r1 = rf(ctx, tenantID, destinationNames)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSubscribedTenantIDs provides a mock function with given fields: ctx
func (_m *DestinationManager) GetSubscribedTenantIDs(ctx context.Context) ([]string, error) {
	ret := _m.Called(ctx)

	var r0 []string
	if rf, ok := ret.Get(0).(func(context.Context) []string); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SyncTenantDestinations provides a mock function with given fields: ctx, tenantID
func (_m *DestinationManager) SyncTenantDestinations(ctx context.Context, tenantID string) error {
	ret := _m.Called(ctx, tenantID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, tenantID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewDestinationManager interface {
	mock.TestingT
	Cleanup(func())
}

// NewDestinationManager creates a new instance of DestinationManager. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDestinationManager(t mockConstructorTestingTNewDestinationManager) *DestinationManager {
	mock := &DestinationManager{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
