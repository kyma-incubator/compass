// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
)

// FormationAssignmentNotificationService is an autogenerated mock type for the FormationAssignmentNotificationService type
type FormationAssignmentNotificationService struct {
	mock.Mock
}

// GenerateNotification provides a mock function with given fields: ctx, formationAssignment
func (_m *FormationAssignmentNotificationService) GenerateNotification(ctx context.Context, formationAssignment *model.FormationAssignment) (*webhookclient.NotificationRequest, error) {
	ret := _m.Called(ctx, formationAssignment)

	var r0 *webhookclient.NotificationRequest
	if rf, ok := ret.Get(0).(func(context.Context, *model.FormationAssignment) *webhookclient.NotificationRequest); ok {
		r0 = rf(ctx, formationAssignment)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*webhookclient.NotificationRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *model.FormationAssignment) error); ok {
		r1 = rf(ctx, formationAssignment)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewFormationAssignmentNotificationService interface {
	mock.TestingT
	Cleanup(func())
}

// NewFormationAssignmentNotificationService creates a new instance of FormationAssignmentNotificationService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationAssignmentNotificationService(t mockConstructorTestingTNewFormationAssignmentNotificationService) *FormationAssignmentNotificationService {
	mock := &FormationAssignmentNotificationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
