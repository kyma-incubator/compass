// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	formationconstraint "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	webhook "github.com/kyma-incubator/compass/components/director/pkg/webhook"

	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
)

// NotificationBuilder is an autogenerated mock type for the notificationBuilder type
type NotificationBuilder struct {
	mock.Mock
}

// BuildFormationAssignmentNotificationRequest provides a mock function with given fields: ctx, formationTemplateID, joinPointDetails, _a3
func (_m *NotificationBuilder) BuildFormationAssignmentNotificationRequest(ctx context.Context, formationTemplateID string, joinPointDetails *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails, _a3 *model.Webhook) (*webhookclient.FormationAssignmentNotificationRequest, error) {
	ret := _m.Called(ctx, formationTemplateID, joinPointDetails, _a3)

	var r0 *webhookclient.FormationAssignmentNotificationRequest
	if rf, ok := ret.Get(0).(func(context.Context, string, *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails, *model.Webhook) *webhookclient.FormationAssignmentNotificationRequest); ok {
		r0 = rf(ctx, formationTemplateID, joinPointDetails, _a3)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*webhookclient.FormationAssignmentNotificationRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails, *model.Webhook) error); ok {
		r1 = rf(ctx, formationTemplateID, joinPointDetails, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// BuildFormationNotificationRequests provides a mock function with given fields: ctx, joinPointDetails, _a2, formationTemplateWebhooks
func (_m *NotificationBuilder) BuildFormationNotificationRequests(ctx context.Context, joinPointDetails *formationconstraint.GenerateFormationNotificationOperationDetails, _a2 *model.Formation, formationTemplateWebhooks []*model.Webhook) ([]*webhookclient.FormationNotificationRequest, error) {
	ret := _m.Called(ctx, joinPointDetails, _a2, formationTemplateWebhooks)

	var r0 []*webhookclient.FormationNotificationRequest
	if rf, ok := ret.Get(0).(func(context.Context, *formationconstraint.GenerateFormationNotificationOperationDetails, *model.Formation, []*model.Webhook) []*webhookclient.FormationNotificationRequest); ok {
		r0 = rf(ctx, joinPointDetails, _a2, formationTemplateWebhooks)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*webhookclient.FormationNotificationRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *formationconstraint.GenerateFormationNotificationOperationDetails, *model.Formation, []*model.Webhook) error); ok {
		r1 = rf(ctx, joinPointDetails, _a2, formationTemplateWebhooks)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PrepareDetailsForApplicationTenantMappingNotificationGeneration provides a mock function with given fields: operation, formationID, formationTemplateID, sourceApplicationTemplate, sourceApplication, targetApplicationTemplate, targetApplication, assignment, reverseAssignment, tenantContext, tenantID
func (_m *NotificationBuilder) PrepareDetailsForApplicationTenantMappingNotificationGeneration(operation model.FormationOperation, formationID string, formationTemplateID string, sourceApplicationTemplate *webhook.ApplicationTemplateWithLabels, sourceApplication *webhook.ApplicationWithLabels, targetApplicationTemplate *webhook.ApplicationTemplateWithLabels, targetApplication *webhook.ApplicationWithLabels, assignment *webhook.FormationAssignment, reverseAssignment *webhook.FormationAssignment, tenantContext *webhook.CustomerTenantContext, tenantID string) (*formationconstraint.GenerateFormationAssignmentNotificationOperationDetails, error) {
	ret := _m.Called(operation, formationID, formationTemplateID, sourceApplicationTemplate, sourceApplication, targetApplicationTemplate, targetApplication, assignment, reverseAssignment, tenantContext, tenantID)

	var r0 *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails
	if rf, ok := ret.Get(0).(func(model.FormationOperation, string, string, *webhook.ApplicationTemplateWithLabels, *webhook.ApplicationWithLabels, *webhook.ApplicationTemplateWithLabels, *webhook.ApplicationWithLabels, *webhook.FormationAssignment, *webhook.FormationAssignment, *webhook.CustomerTenantContext, string) *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails); ok {
		r0 = rf(operation, formationID, formationTemplateID, sourceApplicationTemplate, sourceApplication, targetApplicationTemplate, targetApplication, assignment, reverseAssignment, tenantContext, tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*formationconstraint.GenerateFormationAssignmentNotificationOperationDetails)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.FormationOperation, string, string, *webhook.ApplicationTemplateWithLabels, *webhook.ApplicationWithLabels, *webhook.ApplicationTemplateWithLabels, *webhook.ApplicationWithLabels, *webhook.FormationAssignment, *webhook.FormationAssignment, *webhook.CustomerTenantContext, string) error); ok {
		r1 = rf(operation, formationID, formationTemplateID, sourceApplicationTemplate, sourceApplication, targetApplicationTemplate, targetApplication, assignment, reverseAssignment, tenantContext, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PrepareDetailsForConfigurationChangeNotificationGeneration provides a mock function with given fields: operation, formationID, formationTemplateID, applicationTemplate, application, runtime, runtimeContext, assignment, reverseAssignment, targetType, tenantContext, tenantID
func (_m *NotificationBuilder) PrepareDetailsForConfigurationChangeNotificationGeneration(operation model.FormationOperation, formationID string, formationTemplateID string, applicationTemplate *webhook.ApplicationTemplateWithLabels, application *webhook.ApplicationWithLabels, runtime *webhook.RuntimeWithLabels, runtimeContext *webhook.RuntimeContextWithLabels, assignment *webhook.FormationAssignment, reverseAssignment *webhook.FormationAssignment, targetType model.ResourceType, tenantContext *webhook.CustomerTenantContext, tenantID string) (*formationconstraint.GenerateFormationAssignmentNotificationOperationDetails, error) {
	ret := _m.Called(operation, formationID, formationTemplateID, applicationTemplate, application, runtime, runtimeContext, assignment, reverseAssignment, targetType, tenantContext, tenantID)

	var r0 *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails
	if rf, ok := ret.Get(0).(func(model.FormationOperation, string, string, *webhook.ApplicationTemplateWithLabels, *webhook.ApplicationWithLabels, *webhook.RuntimeWithLabels, *webhook.RuntimeContextWithLabels, *webhook.FormationAssignment, *webhook.FormationAssignment, model.ResourceType, *webhook.CustomerTenantContext, string) *formationconstraint.GenerateFormationAssignmentNotificationOperationDetails); ok {
		r0 = rf(operation, formationID, formationTemplateID, applicationTemplate, application, runtime, runtimeContext, assignment, reverseAssignment, targetType, tenantContext, tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*formationconstraint.GenerateFormationAssignmentNotificationOperationDetails)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.FormationOperation, string, string, *webhook.ApplicationTemplateWithLabels, *webhook.ApplicationWithLabels, *webhook.RuntimeWithLabels, *webhook.RuntimeContextWithLabels, *webhook.FormationAssignment, *webhook.FormationAssignment, model.ResourceType, *webhook.CustomerTenantContext, string) error); ok {
		r1 = rf(operation, formationID, formationTemplateID, applicationTemplate, application, runtime, runtimeContext, assignment, reverseAssignment, targetType, tenantContext, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewNotificationBuilder interface {
	mock.TestingT
	Cleanup(func())
}

// NewNotificationBuilder creates a new instance of NotificationBuilder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewNotificationBuilder(t mockConstructorTestingTNewNotificationBuilder) *NotificationBuilder {
	mock := &NotificationBuilder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
