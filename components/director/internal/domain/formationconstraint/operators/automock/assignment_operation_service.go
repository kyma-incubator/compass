// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// AssignmentOperationService is an autogenerated mock type for the assignmentOperationService type
type AssignmentOperationService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, in
func (_m *AssignmentOperationService) Create(ctx context.Context, in *model.AssignmentOperationInput) (string, error) {
	ret := _m.Called(ctx, in)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.AssignmentOperationInput) (string, error)); ok {
		return rf(ctx, in)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *model.AssignmentOperationInput) string); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *model.AssignmentOperationInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLatestOperation provides a mock function with given fields: ctx, assignmentID, formationID
func (_m *AssignmentOperationService) GetLatestOperation(ctx context.Context, assignmentID string, formationID string) (*model.AssignmentOperation, error) {
	ret := _m.Called(ctx, assignmentID, formationID)

	var r0 *model.AssignmentOperation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.AssignmentOperation, error)); ok {
		return rf(ctx, assignmentID, formationID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.AssignmentOperation); ok {
		r0 = rf(ctx, assignmentID, formationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AssignmentOperation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, assignmentID, formationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewAssignmentOperationService creates a new instance of AssignmentOperationService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAssignmentOperationService(t interface {
	mock.TestingT
	Cleanup(func())
}) *AssignmentOperationService {
	mock := &AssignmentOperationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}