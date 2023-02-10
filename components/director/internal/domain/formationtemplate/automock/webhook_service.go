// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// WebhookService is an autogenerated mock type for the WebhookService type
type WebhookService struct {
	mock.Mock
}

// ListForFormationTemplate provides a mock function with given fields: ctx, tenant, formationTemplateID
func (_m *WebhookService) ListForFormationTemplate(ctx context.Context, tenant string, formationTemplateID string) ([]*model.Webhook, error) {
	ret := _m.Called(ctx, tenant, formationTemplateID)

	var r0 []*model.Webhook
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []*model.Webhook); ok {
		r0 = rf(ctx, tenant, formationTemplateID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Webhook)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, formationTemplateID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewWebhookService creates a new instance of WebhookService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewWebhookService(t testing.TB) *WebhookService {
	mock := &WebhookService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
