// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	resources "github.com/kyma-incubator/compass/components/instance-creator/internal/client/resources"

	testing "testing"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// CreateResource provides a mock function with given fields: ctx, region, subaccountID, resourceReqBody, resource
func (_m *Client) CreateResource(ctx context.Context, region string, subaccountID string, resourceReqBody resources.ResourceRequestBody, resource resources.Resource) (string, error) {
	ret := _m.Called(ctx, region, subaccountID, resourceReqBody, resource)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string, resources.ResourceRequestBody, resources.Resource) string); ok {
		r0 = rf(ctx, region, subaccountID, resourceReqBody, resource)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, resources.ResourceRequestBody, resources.Resource) error); ok {
		r1 = rf(ctx, region, subaccountID, resourceReqBody, resource)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteMultipleResources provides a mock function with given fields: ctx, region, subaccountID, _a3, resourceMatchParams
func (_m *Client) DeleteMultipleResources(ctx context.Context, region string, subaccountID string, _a3 resources.Resources, resourceMatchParams resources.ResourceMatchParameters) error {
	ret := _m.Called(ctx, region, subaccountID, _a3, resourceMatchParams)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, resources.Resources, resources.ResourceMatchParameters) error); ok {
		r0 = rf(ctx, region, subaccountID, _a3, resourceMatchParams)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteResource provides a mock function with given fields: ctx, region, subaccountID, resource, resourceMatchParams
func (_m *Client) DeleteResource(ctx context.Context, region string, subaccountID string, resource resources.Resource, resourceMatchParams resources.ResourceMatchParameters) error {
	ret := _m.Called(ctx, region, subaccountID, resource, resourceMatchParams)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, resources.Resource, resources.ResourceMatchParameters) error); ok {
		r0 = rf(ctx, region, subaccountID, resource, resourceMatchParams)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// RetrieveResource provides a mock function with given fields: ctx, region, subaccountID, _a3, resourceMatchParams
func (_m *Client) RetrieveResource(ctx context.Context, region string, subaccountID string, _a3 resources.Resources, resourceMatchParams resources.ResourceMatchParameters) (string, error) {
	ret := _m.Called(ctx, region, subaccountID, _a3, resourceMatchParams)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, string, resources.Resources, resources.ResourceMatchParameters) string); ok {
		r0 = rf(ctx, region, subaccountID, _a3, resourceMatchParams)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, resources.Resources, resources.ResourceMatchParameters) error); ok {
		r1 = rf(ctx, region, subaccountID, _a3, resourceMatchParams)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RetrieveResourceByID provides a mock function with given fields: ctx, region, subaccountID, resource, resourceMatchParams
func (_m *Client) RetrieveResourceByID(ctx context.Context, region string, subaccountID string, resource resources.Resource, resourceMatchParams resources.ResourceMatchParameters) (resources.Resource, error) {
	ret := _m.Called(ctx, region, subaccountID, resource, resourceMatchParams)

	var r0 resources.Resource
	if rf, ok := ret.Get(0).(func(context.Context, string, string, resources.Resource, resources.ResourceMatchParameters) resources.Resource); ok {
		r0 = rf(ctx, region, subaccountID, resource, resourceMatchParams)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(resources.Resource)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, resources.Resource, resources.ResourceMatchParameters) error); ok {
		r1 = rf(ctx, region, subaccountID, resource, resourceMatchParams)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewClient creates a new instance of Client. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewClient(t testing.TB) *Client {
	mock := &Client{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
