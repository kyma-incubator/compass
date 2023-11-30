// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// WebhookService is an autogenerated mock type for the webhookService type
type WebhookService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, owningResourceID, in, objectType
func (_m *WebhookService) Create(ctx context.Context, owningResourceID string, in model.WebhookInput, objectType model.WebhookReferenceObjectType) (string, error) {
	ret := _m.Called(ctx, owningResourceID, in, objectType)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.WebhookInput, model.WebhookReferenceObjectType) (string, error)); ok {
		return rf(ctx, owningResourceID, in, objectType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, model.WebhookInput, model.WebhookReferenceObjectType) string); ok {
		r0 = rf(ctx, owningResourceID, in, objectType)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, model.WebhookInput, model.WebhookReferenceObjectType) error); ok {
		r1 = rf(ctx, owningResourceID, in, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id, objectType
func (_m *WebhookService) Delete(ctx context.Context, id string, objectType model.WebhookReferenceObjectType) error {
	ret := _m.Called(ctx, id, objectType)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.WebhookReferenceObjectType) error); ok {
		r0 = rf(ctx, id, objectType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListForApplicationTemplate provides a mock function with given fields: ctx, applicationTemplateID
func (_m *WebhookService) ListForApplicationTemplate(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error) {
	ret := _m.Called(ctx, applicationTemplateID)

	var r0 []*model.Webhook
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*model.Webhook, error)); ok {
		return rf(ctx, applicationTemplateID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.Webhook); ok {
		r0 = rf(ctx, applicationTemplateID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Webhook)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, applicationTemplateID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, in, objectType
func (_m *WebhookService) Update(ctx context.Context, id string, in model.WebhookInput, objectType model.WebhookReferenceObjectType) error {
	ret := _m.Called(ctx, id, in, objectType)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.WebhookInput, model.WebhookReferenceObjectType) error); ok {
		r0 = rf(ctx, id, in, objectType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewWebhookService interface {
	mock.TestingT
	Cleanup(func())
}

// NewWebhookService creates a new instance of WebhookService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewWebhookService(t mockConstructorTestingTNewWebhookService) *WebhookService {
	mock := &WebhookService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}