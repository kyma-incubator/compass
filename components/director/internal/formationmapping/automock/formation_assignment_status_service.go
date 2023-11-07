// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	formationassignment "github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	notificationresponse "github.com/kyma-incubator/compass/components/director/internal/domain/statusresponse"
)

// FormationAssignmentStatusService is an autogenerated mock type for the formationAssignmentStatusService type
type FormationAssignmentStatusService struct {
	mock.Mock
}

// DeleteWithConstraints provides a mock function with given fields: ctx, id, notificationResponse
func (_m *FormationAssignmentStatusService) DeleteWithConstraints(ctx context.Context, id string, notificationResponse *notificationresponse.NotificationResponse) error {
	ret := _m.Called(ctx, id, notificationResponse)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *notificationresponse.NotificationResponse) error); ok {
		r0 = rf(ctx, id, notificationResponse)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetAssignmentToErrorStateWithConstraints provides a mock function with given fields: ctx, notificationResponse, assignment, errorMessage, errorCode, state, operation
func (_m *FormationAssignmentStatusService) SetAssignmentToErrorStateWithConstraints(ctx context.Context, notificationResponse *notificationresponse.NotificationResponse, assignment *model.FormationAssignment, errorMessage string, errorCode formationassignment.AssignmentErrorCode, state model.FormationAssignmentState, operation model.FormationOperation) error {
	ret := _m.Called(ctx, notificationResponse, assignment, errorMessage, errorCode, state, operation)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *notificationresponse.NotificationResponse, *model.FormationAssignment, string, formationassignment.AssignmentErrorCode, model.FormationAssignmentState, model.FormationOperation) error); ok {
		r0 = rf(ctx, notificationResponse, assignment, errorMessage, errorCode, state, operation)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateWithConstraints provides a mock function with given fields: ctx, notificationResponse, fa, operation
func (_m *FormationAssignmentStatusService) UpdateWithConstraints(ctx context.Context, notificationResponse *notificationresponse.NotificationResponse, fa *model.FormationAssignment, operation model.FormationOperation) error {
	ret := _m.Called(ctx, notificationResponse, fa, operation)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *notificationresponse.NotificationResponse, *model.FormationAssignment, model.FormationOperation) error); ok {
		r0 = rf(ctx, notificationResponse, fa, operation)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewFormationAssignmentStatusService creates a new instance of FormationAssignmentStatusService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFormationAssignmentStatusService(t interface {
	mock.TestingT
	Cleanup(func())
}) *FormationAssignmentStatusService {
	mock := &FormationAssignmentStatusService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
