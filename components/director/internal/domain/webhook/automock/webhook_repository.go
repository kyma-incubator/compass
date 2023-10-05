// Code generated by mockery. DO NOT EDIT.

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
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.WebhookReferenceObjectType) (*model.Webhook, error)); ok {
		return rf(ctx, tenant, id, objectType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.WebhookReferenceObjectType) *model.Webhook); ok {
		r0 = rf(ctx, tenant, id, objectType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Webhook)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, model.WebhookReferenceObjectType) error); ok {
		r1 = rf(ctx, tenant, id, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByIDAndWebhookTypeGlobal provides a mock function with given fields: ctx, objectID, objectType, webhookType
func (_m *WebhookRepository) GetByIDAndWebhookTypeGlobal(ctx context.Context, objectID string, objectType model.WebhookReferenceObjectType, webhookType model.WebhookType) (*model.Webhook, error) {
	ret := _m.Called(ctx, objectID, objectType, webhookType)

	var r0 *model.Webhook
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.WebhookReferenceObjectType, model.WebhookType) (*model.Webhook, error)); ok {
		return rf(ctx, objectID, objectType, webhookType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, model.WebhookReferenceObjectType, model.WebhookType) *model.Webhook); ok {
		r0 = rf(ctx, objectID, objectType, webhookType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Webhook)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, model.WebhookReferenceObjectType, model.WebhookType) error); ok {
		r1 = rf(ctx, objectID, objectType, webhookType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByIDGlobal provides a mock function with given fields: ctx, id
func (_m *WebhookRepository) GetByIDGlobal(ctx context.Context, id string) (*model.Webhook, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Webhook
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Webhook, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Webhook); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Webhook)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByApplicationTemplateID provides a mock function with given fields: ctx, applicationTemplateID
func (_m *WebhookRepository) ListByApplicationTemplateID(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error) {
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

// ListByReferenceObjectID provides a mock function with given fields: ctx, tenant, objID, objType
func (_m *WebhookRepository) ListByReferenceObjectID(ctx context.Context, tenant string, objID string, objType model.WebhookReferenceObjectType) ([]*model.Webhook, error) {
	ret := _m.Called(ctx, tenant, objID, objType)

	var r0 []*model.Webhook
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.WebhookReferenceObjectType) ([]*model.Webhook, error)); ok {
		return rf(ctx, tenant, objID, objType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.WebhookReferenceObjectType) []*model.Webhook); ok {
		r0 = rf(ctx, tenant, objID, objType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Webhook)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, model.WebhookReferenceObjectType) error); ok {
		r1 = rf(ctx, tenant, objID, objType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByReferenceObjectIDGlobal provides a mock function with given fields: ctx, objID, objType
func (_m *WebhookRepository) ListByReferenceObjectIDGlobal(ctx context.Context, objID string, objType model.WebhookReferenceObjectType) ([]*model.Webhook, error) {
	ret := _m.Called(ctx, objID, objType)

	var r0 []*model.Webhook
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.WebhookReferenceObjectType) ([]*model.Webhook, error)); ok {
		return rf(ctx, objID, objType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, model.WebhookReferenceObjectType) []*model.Webhook); ok {
		r0 = rf(ctx, objID, objType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Webhook)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, model.WebhookReferenceObjectType) error); ok {
		r1 = rf(ctx, objID, objType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByWebhookType provides a mock function with given fields: ctx, webhookType
func (_m *WebhookRepository) ListByWebhookType(ctx context.Context, webhookType model.WebhookType) ([]*model.Webhook, error) {
	ret := _m.Called(ctx, webhookType)

	var r0 []*model.Webhook
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.WebhookType) ([]*model.Webhook, error)); ok {
		return rf(ctx, webhookType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.WebhookType) []*model.Webhook); ok {
		r0 = rf(ctx, webhookType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Webhook)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.WebhookType) error); ok {
		r1 = rf(ctx, webhookType)
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

// NewWebhookRepository creates a new instance of WebhookRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewWebhookRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *WebhookRepository {
	mock := &WebhookRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
