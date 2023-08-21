// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// OperationRepository is an autogenerated mock type for the OperationRepository type
type OperationRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, _a1
func (_m *OperationRepository) Create(ctx context.Context, _a1 *model.Operation) error {
	ret := _m.Called(ctx, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Operation) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, id
func (_m *OperationRepository) Get(ctx context.Context, id string) (*model.Operation, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Operation
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Operation); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Operation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// LockOperation provides a mock function with given fields: ctx, operationID
func (_m *OperationRepository) LockOperation(ctx context.Context, operationID string) (bool, error) {
	ret := _m.Called(ctx, operationID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, operationID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, operationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PriorityQueueListByType provides a mock function with given fields: ctx, queueLimit, opType
func (_m *OperationRepository) PriorityQueueListByType(ctx context.Context, queueLimit int, opType model.OperationType) ([]*model.Operation, error) {
	ret := _m.Called(ctx, queueLimit, opType)

	var r0 []*model.Operation
	if rf, ok := ret.Get(0).(func(context.Context, int, model.OperationType) []*model.Operation); ok {
		r0 = rf(ctx, queueLimit, opType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Operation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int, model.OperationType) error); ok {
		r1 = rf(ctx, queueLimit, opType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RescheduleHangedOperations provides a mock function with given fields: ctx, hangPeriod
func (_m *OperationRepository) RescheduleHangedOperations(ctx context.Context, hangPeriod time.Duration) error {
	ret := _m.Called(ctx, hangPeriod)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Duration) error); ok {
		r0 = rf(ctx, hangPeriod)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ResheduleOperations provides a mock function with given fields: ctx, reschedulePeriod
func (_m *OperationRepository) ResheduleOperations(ctx context.Context, reschedulePeriod time.Duration) error {
	ret := _m.Called(ctx, reschedulePeriod)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, time.Duration) error); ok {
		r0 = rf(ctx, reschedulePeriod)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: ctx, _a1
func (_m *OperationRepository) Update(ctx context.Context, _a1 *model.Operation) error {
	ret := _m.Called(ctx, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Operation) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewOperationRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewOperationRepository creates a new instance of OperationRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewOperationRepository(t mockConstructorTestingTNewOperationRepository) *OperationRepository {
	mock := &OperationRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
