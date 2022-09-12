// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	testing "testing"

	webhook "github.com/kyma-incubator/compass/components/director/pkg/webhook"

	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
)

// WebhookClient is an autogenerated mock type for the webhookClient type
type WebhookClient struct {
	mock.Mock
}

// Do provides a mock function with given fields: ctx, request
func (_m *WebhookClient) Do(ctx context.Context, request *webhookclient.Request) (*webhook.Response, error) {
	ret := _m.Called(ctx, request)

	var r0 *webhook.Response
	if rf, ok := ret.Get(0).(func(context.Context, *webhookclient.Request) *webhook.Response); ok {
		r0 = rf(ctx, request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*webhook.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *webhookclient.Request) error); ok {
		r1 = rf(ctx, request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewWebhookClient creates a new instance of WebhookClient. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewWebhookClient(t testing.TB) *WebhookClient {
	mock := &WebhookClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
