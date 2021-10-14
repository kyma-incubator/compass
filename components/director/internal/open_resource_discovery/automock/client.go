// Code generated by mockery (devel). DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// FetchOpenResourceDiscoveryDocuments provides a mock function with given fields: ctx, app, webhook
func (_m *Client) FetchOpenResourceDiscoveryDocuments(ctx context.Context, app *model.Application, webhook *model.Webhook) (ord.Documents, error) {
	ret := _m.Called(ctx, app, webhook)

	var r0 ord.Documents
	if rf, ok := ret.Get(0).(func(context.Context, *model.Application, *model.Webhook) ord.Documents); ok {
		r0 = rf(ctx, app, webhook)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(ord.Documents)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *model.Application, *model.Webhook) error); ok {
		r1 = rf(ctx, app, webhook)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
