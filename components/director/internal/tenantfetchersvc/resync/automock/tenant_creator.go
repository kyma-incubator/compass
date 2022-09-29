// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// TenantCreator is an autogenerated mock type for the TenantCreator type
type TenantCreator struct {
	mock.Mock
}

// CreateTenants provides a mock function with given fields: ctx, eventsTenants
func (_m *TenantCreator) CreateTenants(ctx context.Context, eventsTenants []model.BusinessTenantMappingInput) error {
	ret := _m.Called(ctx, eventsTenants)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []model.BusinessTenantMappingInput) error); ok {
		r0 = rf(ctx, eventsTenants)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// FetchTenant provides a mock function with given fields: ctx, externalTenantID
func (_m *TenantCreator) FetchTenant(ctx context.Context, externalTenantID string) (*model.BusinessTenantMappingInput, error) {
	ret := _m.Called(ctx, externalTenantID)

	var r0 *model.BusinessTenantMappingInput
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.BusinessTenantMappingInput); ok {
		r0 = rf(ctx, externalTenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BusinessTenantMappingInput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, externalTenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TenantsToCreate provides a mock function with given fields: ctx, region, fromTimestamp
func (_m *TenantCreator) TenantsToCreate(ctx context.Context, region string, fromTimestamp string) ([]model.BusinessTenantMappingInput, error) {
	ret := _m.Called(ctx, region, fromTimestamp)

	var r0 []model.BusinessTenantMappingInput
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []model.BusinessTenantMappingInput); ok {
		r0 = rf(ctx, region, fromTimestamp)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.BusinessTenantMappingInput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, region, fromTimestamp)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewTenantCreator creates a new instance of TenantCreator. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewTenantCreator(t testing.TB) *TenantCreator {
	mock := &TenantCreator{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
