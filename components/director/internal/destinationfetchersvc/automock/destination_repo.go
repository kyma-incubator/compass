// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// DestinationRepo is an autogenerated mock type for the DestinationRepo type
type DestinationRepo struct {
	mock.Mock
}

// DeleteOld provides a mock function with given fields: ctx, latestRevision, tenantID
func (_m *DestinationRepo) DeleteOld(ctx context.Context, latestRevision string, tenantID string) error {
	ret := _m.Called(ctx, latestRevision, tenantID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, latestRevision, tenantID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetDestinationByNameAndTenant provides a mock function with given fields: ctx, destinationName, tenantID
func (_m *DestinationRepo) GetDestinationByNameAndTenant(ctx context.Context, destinationName string, tenantID string) (*model.Destination, error) {
	ret := _m.Called(ctx, destinationName, tenantID)

	var r0 *model.Destination
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.Destination, error)); ok {
		return rf(ctx, destinationName, tenantID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Destination); ok {
		r0 = rf(ctx, destinationName, tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Destination)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, destinationName, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Upsert provides a mock function with given fields: ctx, in, id, tenantID, bundleID, revision
func (_m *DestinationRepo) Upsert(ctx context.Context, in model.DestinationInput, id string, tenantID string, bundleID string, revision string) error {
	ret := _m.Called(ctx, in, id, tenantID, bundleID, revision)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.DestinationInput, string, string, string, string) error); ok {
		r0 = rf(ctx, in, id, tenantID, bundleID, revision)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewDestinationRepo creates a new instance of DestinationRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDestinationRepo(t interface {
	mock.TestingT
	Cleanup(func())
}) *DestinationRepo {
	mock := &DestinationRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
