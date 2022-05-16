// Code generated by mockery 2.9.0. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// WebhookRepository is an autogenerated mock type for the WebhookRepository type
type WebhookRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, tenant, item
func (_m *WebhookRepository) Create(ctx context.Context, tenant string, item *model.Webhook) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Webhook) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, id
func (_m *WebhookRepository) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByID provides a mock function with given fields: ctx, tenant, id, objectType
func (_m *WebhookRepository) GetByID(ctx context.Context, tenant string, id string, objectType model.WebhookReferenceObjectType) (*model.Webhook, error) {
	ret := _m.Called(ctx, tenant, id, objectType)

	var r0 *model.Webhook
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.WebhookReferenceObjectType) *model.Webhook); ok {
		r0 = rf(ctx, tenant, id, objectType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Webhook)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, model.WebhookReferenceObjectType) error); ok {
		r1 = rf(ctx, tenant, id, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByIDGlobal provides a mock function with given fields: ctx, id
func (_m *WebhookRepository) GetByIDGlobal(ctx context.Context, id string) (*model.Webhook, error) {
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

// ListByApplicationIDWithSelectForUpdate provides a mock function with given fields: ctx, tenant, applicationID
func (_m *WebhookRepository) ListByApplicationIDWithSelectForUpdate(ctx context.Context, tenant string, applicationID string) ([]*model.Webhook, error) {
	ret := _m.Called(ctx, tenant, applicationID)

	var r0 []*model.Webhook
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []*model.Webhook); ok {
		r0 = rf(ctx, tenant, applicationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Webhook)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, applicationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByApplicationTemplateID provides a mock function with given fields: ctx, applicationTemplateID
func (_m *WebhookRepository) ListByApplicationTemplateID(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error) {
	ret := _m.Called(ctx, applicationTemplateID)

	var r0 []*model.Webhook
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.Webhook); ok {
		r0 = rf(ctx, applicationTemplateID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Webhook)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, applicationTemplateID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByReferenceObjectID provides a mock function with given fields: ctx, tenant, objId, objType
func (_m *WebhookRepository) ListByReferenceObjectID(ctx context.Context, tenant string, objId string, objType model.WebhookReferenceObjectType) ([]*model.Webhook, error) {
	ret := _m.Called(ctx, tenant, objId, objType)

	var r0 []*model.Webhook
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.WebhookReferenceObjectType) []*model.Webhook); ok {
		r0 = rf(ctx, tenant, objId, objType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Webhook)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, model.WebhookReferenceObjectType) error); ok {
		r1 = rf(ctx, tenant, objId, objType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, tenant, item
func (_m *WebhookRepository) Update(ctx context.Context, tenant string, item *model.Webhook) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Webhook) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
