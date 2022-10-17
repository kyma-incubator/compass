// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	webhook "github.com/kyma-incubator/compass/components/director/pkg/webhook"

	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
)

// NotificationService is an autogenerated mock type for the notificationService type
type NotificationService struct {
	mock.Mock
}

// SendNotification provides a mock function with given fields: ctx, notification
func (_m *NotificationService) SendNotification(ctx context.Context, notification *webhookclient.NotificationRequest) (*webhook.Response, error) {
	ret := _m.Called(ctx, notification)

	var r0 *webhook.Response
	if rf, ok := ret.Get(0).(func(context.Context, *webhookclient.NotificationRequest) *webhook.Response); ok {
		r0 = rf(ctx, notification)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*webhook.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *webhookclient.NotificationRequest) error); ok {
		r1 = rf(ctx, notification)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewNotificationServiceT interface {
	mock.TestingT
	Cleanup(func())
}

// NewNotificationService creates a new instance of NotificationService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewNotificationService(t NewNotificationServiceT) *NotificationService {
	mock := &NotificationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
