// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	operation "github.com/kyma-incubator/compass/components/director/pkg/operation"
	mock "github.com/stretchr/testify/mock"
)

// Scheduler is an autogenerated mock type for the Scheduler type
type Scheduler struct {
	mock.Mock
}

// Schedule provides a mock function with given fields: ctx, op
func (_m *Scheduler) Schedule(ctx context.Context, op *operation.Operation) (string, error) {
	ret := _m.Called(ctx, op)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, *operation.Operation) string); ok {
		r0 = rf(ctx, op)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *operation.Operation) error); ok {
		r1 = rf(ctx, op)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewScheduler interface {
	mock.TestingT
	Cleanup(func())
}

// NewScheduler creates a new instance of Scheduler. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewScheduler(t mockConstructorTestingTNewScheduler) *Scheduler {
	mock := &Scheduler{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
