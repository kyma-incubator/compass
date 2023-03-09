// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	webhook "github.com/kyma-incubator/compass/components/director/pkg/webhook"

	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
)

// NotificationsService is an autogenerated mock type for the NotificationsService type
type NotificationsService struct {
	mock.Mock
}

// GenerateFormationAssignmentNotifications provides a mock function with given fields: ctx, tenant, objectID, _a3, operation, objectType
func (_m *NotificationsService) GenerateFormationAssignmentNotifications(ctx context.Context, tenant string, objectID string, _a3 *model.Formation, operation model.FormationOperation, objectType graphql.FormationObjectType) ([]*webhookclient.FormationAssignmentNotificationRequest, error) {
	ret := _m.Called(ctx, tenant, objectID, _a3, operation, objectType)

	var r0 []*webhookclient.FormationAssignmentNotificationRequest
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *model.Formation, model.FormationOperation, graphql.FormationObjectType) []*webhookclient.FormationAssignmentNotificationRequest); ok {
		r0 = rf(ctx, tenant, objectID, _a3, operation, objectType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*webhookclient.FormationAssignmentNotificationRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, *model.Formation, model.FormationOperation, graphql.FormationObjectType) error); ok {
		r1 = rf(ctx, tenant, objectID, _a3, operation, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GenerateFormationNotifications provides a mock function with given fields: ctx, formationTemplateWebhooks, tenantID, _a3, formationTemplateName, formationTemplateID, formationOperation
func (_m *NotificationsService) GenerateFormationNotifications(ctx context.Context, formationTemplateWebhooks []*model.Webhook, tenantID string, _a3 *model.Formation, formationTemplateName string, formationTemplateID string, formationOperation model.FormationOperation) ([]*webhookclient.FormationNotificationRequest, error) {
	ret := _m.Called(ctx, formationTemplateWebhooks, tenantID, _a3, formationTemplateName, formationTemplateID, formationOperation)

	var r0 []*webhookclient.FormationNotificationRequest
	if rf, ok := ret.Get(0).(func(context.Context, []*model.Webhook, string, *model.Formation, string, string, model.FormationOperation) []*webhookclient.FormationNotificationRequest); ok {
		r0 = rf(ctx, formationTemplateWebhooks, tenantID, _a3, formationTemplateName, formationTemplateID, formationOperation)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*webhookclient.FormationNotificationRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []*model.Webhook, string, *model.Formation, string, string, model.FormationOperation) error); ok {
		r1 = rf(ctx, formationTemplateWebhooks, tenantID, _a3, formationTemplateName, formationTemplateID, formationOperation)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SendNotification provides a mock function with given fields: ctx, webhookNotificationReq
func (_m *NotificationsService) SendNotification(ctx context.Context, webhookNotificationReq webhookclient.WebhookRequest) (*webhook.Response, error) {
	ret := _m.Called(ctx, webhookNotificationReq)

	var r0 *webhook.Response
	if rf, ok := ret.Get(0).(func(context.Context, webhookclient.WebhookRequest) *webhook.Response); ok {
		r0 = rf(ctx, webhookNotificationReq)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*webhook.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, webhookclient.WebhookRequest) error); ok {
		r1 = rf(ctx, webhookNotificationReq)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewNotificationsServiceT interface {
	mock.TestingT
	Cleanup(func())
}

// NewNotificationsService creates a new instance of NotificationsService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewNotificationsService(t NewNotificationsServiceT) *NotificationsService {
	mock := &NotificationsService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
