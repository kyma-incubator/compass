// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	formationassignment "github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	mock "github.com/stretchr/testify/mock"
)

// FormationAssignmentService is an autogenerated mock type for the formationAssignmentService type
type FormationAssignmentService struct {
	mock.Mock
}

// CleanupFormationAssignment provides a mock function with given fields: ctx, mappingPair
func (_m *FormationAssignmentService) CleanupFormationAssignment(ctx context.Context, mappingPair *formationassignment.AssignmentMappingPairWithOperation) (bool, error) {
	ret := _m.Called(ctx, mappingPair)

	if len(ret) == 0 {
		panic("no return value specified for CleanupFormationAssignment")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *formationassignment.AssignmentMappingPairWithOperation) (bool, error)); ok {
		return rf(ctx, mappingPair)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *formationassignment.AssignmentMappingPairWithOperation) bool); ok {
		r0 = rf(ctx, mappingPair)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, *formationassignment.AssignmentMappingPairWithOperation) error); ok {
		r1 = rf(ctx, mappingPair)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewFormationAssignmentService creates a new instance of FormationAssignmentService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFormationAssignmentService(t interface {
	mock.TestingT
	Cleanup(func())
}) *FormationAssignmentService {
	mock := &FormationAssignmentService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
