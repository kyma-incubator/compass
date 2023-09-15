// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
	mock "github.com/stretchr/testify/mock"
)

// SubscriptionService is an autogenerated mock type for the SubscriptionService type
type SubscriptionService struct {
	mock.Mock
}

// DetermineSubscriptionFlow provides a mock function with given fields: ctx, providerID, region
func (_m *SubscriptionService) DetermineSubscriptionFlow(ctx context.Context, providerID string, region string) (resource.Type, error) {
	ret := _m.Called(ctx, providerID, region)

	var r0 resource.Type
	if rf, ok := ret.Get(0).(func(context.Context, string, string) resource.Type); ok {
		r0 = rf(ctx, providerID, region)
	} else {
		r0 = ret.Get(0).(resource.Type)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, providerID, region)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SubscribeTenantToApplication provides a mock function with given fields: ctx, providerID, subaccountTenantID, consumerTenantID, providerSubaccountID, region, subscribedAppName, subscriptionID, subscriptionPayload
func (_m *SubscriptionService) SubscribeTenantToApplication(ctx context.Context, providerID string, subaccountTenantID string, consumerTenantID string, providerSubaccountID string, region string, subscribedAppName string, subscriptionID string, subscriptionPayload string) (bool, string, string, error) {
	ret := _m.Called(ctx, providerID, subaccountTenantID, consumerTenantID, providerSubaccountID, region, subscribedAppName, subscriptionID, subscriptionPayload)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string, string, string, string, string) bool); ok {
		r0 = rf(ctx, providerID, subaccountTenantID, consumerTenantID, providerSubaccountID, region, subscribedAppName, subscriptionID, subscriptionPayload)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 string
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, string, string, string, string, string) string); ok {
		r1 = rf(ctx, providerID, subaccountTenantID, consumerTenantID, providerSubaccountID, region, subscribedAppName, subscriptionID, subscriptionPayload)
	} else {
		r1 = ret.Get(1).(string)
	}

	var r2 string
	if rf, ok := ret.Get(2).(func(context.Context, string, string, string, string, string, string, string, string) string); ok {
		r2 = rf(ctx, providerID, subaccountTenantID, consumerTenantID, providerSubaccountID, region, subscribedAppName, subscriptionID, subscriptionPayload)
	} else {
		r2 = ret.Get(2).(string)
	}

	var r3 error
	if rf, ok := ret.Get(3).(func(context.Context, string, string, string, string, string, string, string, string) error); ok {
		r3 = rf(ctx, providerID, subaccountTenantID, consumerTenantID, providerSubaccountID, region, subscribedAppName, subscriptionID, subscriptionPayload)
	} else {
		r3 = ret.Error(3)
	}

	return r0, r1, r2, r3
}

// SubscribeTenantToRuntime provides a mock function with given fields: ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionAppName, subscriptionID
func (_m *SubscriptionService) SubscribeTenantToRuntime(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, consumerTenantID string, region string, subscriptionAppName string, subscriptionID string) (bool, error) {
	ret := _m.Called(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionAppName, subscriptionID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string, string, string, string) bool); ok {
		r0 = rf(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionAppName, subscriptionID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, string, string, string, string) error); ok {
		r1 = rf(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionAppName, subscriptionID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnsubscribeTenantFromApplication provides a mock function with given fields: ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID
func (_m *SubscriptionService) UnsubscribeTenantFromApplication(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, consumerTenantID string, region string, subscriptionID string) (bool, error) {
	ret := _m.Called(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string, string, string) bool); ok {
		r0 = rf(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, string, string, string) error); ok {
		r1 = rf(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnsubscribeTenantFromRuntime provides a mock function with given fields: ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID
func (_m *SubscriptionService) UnsubscribeTenantFromRuntime(ctx context.Context, providerID string, subaccountTenantID string, providerSubaccountID string, consumerTenantID string, region string, subscriptionID string) (bool, error) {
	ret := _m.Called(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string, string, string) bool); ok {
		r0 = rf(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, string, string, string) error); ok {
		r1 = rf(ctx, providerID, subaccountTenantID, providerSubaccountID, consumerTenantID, region, subscriptionID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewSubscriptionService interface {
	mock.TestingT
	Cleanup(func())
}

// NewSubscriptionService creates a new instance of SubscriptionService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSubscriptionService(t mockConstructorTestingTNewSubscriptionService) *SubscriptionService {
	mock := &SubscriptionService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
