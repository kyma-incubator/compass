// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// SystemFieldDiscoveryEngine is an autogenerated mock type for the SystemFieldDiscoveryEngine type
type SystemFieldDiscoveryEngine struct {
	mock.Mock
}

// CreateLabelForApplicationWebhook provides a mock function with given fields: ctx, appID
func (_m *SystemFieldDiscoveryEngine) CreateLabelForApplicationWebhook(ctx context.Context, appID string) error {
	ret := _m.Called(ctx, appID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, appID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnrichApplicationWebhookIfNeeded provides a mock function with given fields: ctx, appCreateInputModel, systemFieldDiscovery, region, subacountID, appTemplateName, appName
func (_m *SystemFieldDiscoveryEngine) EnrichApplicationWebhookIfNeeded(ctx context.Context, appCreateInputModel model.ApplicationRegisterInput, systemFieldDiscovery bool, region string, subacountID string, appTemplateName string, appName string) ([]*model.WebhookInput, bool) {
	ret := _m.Called(ctx, appCreateInputModel, systemFieldDiscovery, region, subacountID, appTemplateName, appName)

	var r0 []*model.WebhookInput
	var r1 bool
	if rf, ok := ret.Get(0).(func(context.Context, model.ApplicationRegisterInput, bool, string, string, string, string) ([]*model.WebhookInput, bool)); ok {
		return rf(ctx, appCreateInputModel, systemFieldDiscovery, region, subacountID, appTemplateName, appName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.ApplicationRegisterInput, bool, string, string, string, string) []*model.WebhookInput); ok {
		r0 = rf(ctx, appCreateInputModel, systemFieldDiscovery, region, subacountID, appTemplateName, appName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.WebhookInput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.ApplicationRegisterInput, bool, string, string, string, string) bool); ok {
		r1 = rf(ctx, appCreateInputModel, systemFieldDiscovery, region, subacountID, appTemplateName, appName)
	} else {
		r1 = ret.Get(1).(bool)
	}

	return r0, r1
}

// NewSystemFieldDiscoveryEngine creates a new instance of SystemFieldDiscoveryEngine. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewSystemFieldDiscoveryEngine(t interface {
	mock.TestingT
	Cleanup(func())
}) *SystemFieldDiscoveryEngine {
	mock := &SystemFieldDiscoveryEngine{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}