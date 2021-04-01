// Code generated by mockery v2.5.1. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// WebhookService is an autogenerated mock type for the WebhookService type
type WebhookService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, resourceID, in, converterFunc
func (_m *WebhookService) Create(ctx context.Context, resourceID string, in model.WebhookInput, converterFunc model.WebhookConverterFunc) (string, error) {
	ret := _m.Called(ctx, resourceID, in, converterFunc)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, model.WebhookInput, model.WebhookConverterFunc) string); ok {
		r0 = rf(ctx, resourceID, in, converterFunc)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, model.WebhookInput, model.WebhookConverterFunc) error); ok {
		r1 = rf(ctx, resourceID, in, converterFunc)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *WebhookService) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, id
func (_m *WebhookService) Get(ctx context.Context, id string) (*model.Webhook, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Webhook
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Webhook); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Webhook)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllApplicationWebhooks provides a mock function with given fields: ctx, applicationID
func (_m *WebhookService) ListAllApplicationWebhooks(ctx context.Context, applicationID string) ([]*model.Webhook, error) {
	ret := _m.Called(ctx, applicationID)

	var r0 []*model.Webhook
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.Webhook); ok {
		r0 = rf(ctx, applicationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Webhook)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, applicationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, in
func (_m *WebhookService) Update(ctx context.Context, id string, in model.WebhookInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.WebhookInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
