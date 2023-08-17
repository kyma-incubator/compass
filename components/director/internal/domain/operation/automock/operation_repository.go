// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
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
