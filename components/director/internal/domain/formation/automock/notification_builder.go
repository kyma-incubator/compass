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

// BuildNotificationRequest provides a mock function with given fields: ctx, formationTemplateID, joinPointDetails, _a3
func (_m *NotificationBuilder) BuildNotificationRequest(ctx context.Context, formationTemplateID string, joinPointDetails *formationconstraint.GenerateNotificationOperationDetails, _a3 *model.Webhook) (*webhookclient.NotificationRequest, error) {
	ret := _m.Called(ctx, formationTemplateID, joinPointDetails, _a3)

	var r0 *webhookclient.NotificationRequest
	if rf, ok := ret.Get(0).(func(context.Context, string, *formationconstraint.GenerateNotificationOperationDetails, *model.Webhook) *webhookclient.NotificationRequest); ok {
		r0 = rf(ctx, formationTemplateID, joinPointDetails, _a3)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*webhookclient.NotificationRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *formationconstraint.GenerateNotificationOperationDetails, *model.Webhook) error); ok {
		r1 = rf(ctx, formationTemplateID, joinPointDetails, _a3)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PrepareDetailsForApplicationTenantMappingNotificationGeneration provides a mock function with given fields: operation, formationID, sourceApplicationTemplate, sourceApplication, targetApplicationTemplate, targetApplication, assignment, reverseAssignment
func (_m *NotificationBuilder) PrepareDetailsForApplicationTenantMappingNotificationGeneration(operation model.FormationOperation, formationID string, sourceApplicationTemplate *webhook.ApplicationTemplateWithLabels, sourceApplication *webhook.ApplicationWithLabels, targetApplicationTemplate *webhook.ApplicationTemplateWithLabels, targetApplication *webhook.ApplicationWithLabels, assignment *webhook.FormationAssignment, reverseAssignment *webhook.FormationAssignment) (*formationconstraint.GenerateNotificationOperationDetails, error) {
	ret := _m.Called(operation, formationID, sourceApplicationTemplate, sourceApplication, targetApplicationTemplate, targetApplication, assignment, reverseAssignment)

	var r0 *formationconstraint.GenerateNotificationOperationDetails
	if rf, ok := ret.Get(0).(func(model.FormationOperation, string, *webhook.ApplicationTemplateWithLabels, *webhook.ApplicationWithLabels, *webhook.ApplicationTemplateWithLabels, *webhook.ApplicationWithLabels, *webhook.FormationAssignment, *webhook.FormationAssignment) *formationconstraint.GenerateNotificationOperationDetails); ok {
		r0 = rf(operation, formationID, sourceApplicationTemplate, sourceApplication, targetApplicationTemplate, targetApplication, assignment, reverseAssignment)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*formationconstraint.GenerateNotificationOperationDetails)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.FormationOperation, string, *webhook.ApplicationTemplateWithLabels, *webhook.ApplicationWithLabels, *webhook.ApplicationTemplateWithLabels, *webhook.ApplicationWithLabels, *webhook.FormationAssignment, *webhook.FormationAssignment) error); ok {
		r1 = rf(operation, formationID, sourceApplicationTemplate, sourceApplication, targetApplicationTemplate, targetApplication, assignment, reverseAssignment)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PrepareDetailsForConfigurationChangeNotificationGeneration provides a mock function with given fields: operation, formationID, applicationTemplate, application, runtime, runtimeContext, assignment, reverseAssignment, targetType
func (_m *NotificationBuilder) PrepareDetailsForConfigurationChangeNotificationGeneration(operation model.FormationOperation, formationID string, applicationTemplate *webhook.ApplicationTemplateWithLabels, application *webhook.ApplicationWithLabels, runtime *webhook.RuntimeWithLabels, runtimeContext *webhook.RuntimeContextWithLabels, assignment *webhook.FormationAssignment, reverseAssignment *webhook.FormationAssignment, targetType model.ResourceType) (*formationconstraint.GenerateNotificationOperationDetails, error) {
	ret := _m.Called(operation, formationID, applicationTemplate, application, runtime, runtimeContext, assignment, reverseAssignment, targetType)

	var r0 *formationconstraint.GenerateNotificationOperationDetails
	if rf, ok := ret.Get(0).(func(model.FormationOperation, string, *webhook.ApplicationTemplateWithLabels, *webhook.ApplicationWithLabels, *webhook.RuntimeWithLabels, *webhook.RuntimeContextWithLabels, *webhook.FormationAssignment, *webhook.FormationAssignment, model.ResourceType) *formationconstraint.GenerateNotificationOperationDetails); ok {
		r0 = rf(operation, formationID, applicationTemplate, application, runtime, runtimeContext, assignment, reverseAssignment, targetType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*formationconstraint.GenerateNotificationOperationDetails)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.FormationOperation, string, *webhook.ApplicationTemplateWithLabels, *webhook.ApplicationWithLabels, *webhook.RuntimeWithLabels, *webhook.RuntimeContextWithLabels, *webhook.FormationAssignment, *webhook.FormationAssignment, model.ResourceType) error); ok {
		r1 = rf(operation, formationID, applicationTemplate, application, runtime, runtimeContext, assignment, reverseAssignment, targetType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewNotificationBuilderT interface {
	mock.TestingT
	Cleanup(func())
}

// NewNotificationBuilder creates a new instance of NotificationBuilder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewNotificationBuilder(t NewNotificationBuilderT) *NotificationBuilder {
	mock := &NotificationBuilder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
