// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// SystemsSyncService is an autogenerated mock type for the SystemsSyncService type
type SystemsSyncService struct {
	mock.Mock
}

// ListByTenant provides a mock function with given fields: ctx, tenant
func (_m *SystemsSyncService) ListByTenant(ctx context.Context, tenant string) ([]*model.SystemSynchronizationTimestamp, error) {
	ret := _m.Called(ctx, tenant)

	var r0 []*model.SystemSynchronizationTimestamp
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*model.SystemSynchronizationTimestamp, error)); ok {
		return rf(ctx, tenant)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.SystemSynchronizationTimestamp); ok {
		r0 = rf(ctx, tenant)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.SystemSynchronizationTimestamp)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Upsert provides a mock function with given fields: ctx, in
func (_m *SystemsSyncService) Upsert(ctx context.Context, in *model.SystemSynchronizationTimestamp) error {
	ret := _m.Called(ctx, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.SystemSynchronizationTimestamp) error); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewSystemsSyncService creates a new instance of SystemsSyncService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSystemsSyncService(t interface {
	mock.TestingT
	Cleanup(func())
}) *SystemsSyncService {
	mock := &SystemsSyncService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
