// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	statusreport "github.com/kyma-incubator/compass/components/director/internal/domain/statusreport"
)

// StatusService is an autogenerated mock type for the statusService type
type StatusService struct {
	mock.Mock
}

// DeleteWithConstraints provides a mock function with given fields: ctx, id, notificationStatusReport
func (_m *StatusService) DeleteWithConstraints(ctx context.Context, id string, notificationStatusReport *statusreport.NotificationStatusReport) error {
	ret := _m.Called(ctx, id, notificationStatusReport)

	if len(ret) == 0 {
		panic("no return value specified for DeleteWithConstraints")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *statusreport.NotificationStatusReport) error); ok {
		r0 = rf(ctx, id, notificationStatusReport)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateWithConstraints provides a mock function with given fields: ctx, notificationStatusReport, fa, operation
func (_m *StatusService) UpdateWithConstraints(ctx context.Context, notificationStatusReport *statusreport.NotificationStatusReport, fa *model.FormationAssignment, operation model.FormationOperation) error {
	ret := _m.Called(ctx, notificationStatusReport, fa, operation)

	if len(ret) == 0 {
		panic("no return value specified for UpdateWithConstraints")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *statusreport.NotificationStatusReport, *model.FormationAssignment, model.FormationOperation) error); ok {
		r0 = rf(ctx, notificationStatusReport, fa, operation)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewStatusService creates a new instance of StatusService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewStatusService(t interface {
	mock.TestingT
	Cleanup(func())
}) *StatusService {
	mock := &StatusService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
