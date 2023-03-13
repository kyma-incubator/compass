// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

// TenantMappingsService is an autogenerated mock type for the TenantMappingsService type
type TenantMappingsService struct {
	mock.Mock
}

// UpdateApplicationsConsumedAPIs provides a mock function with given fields: ctx, tenantMapping
func (_m *TenantMappingsService) UpdateApplicationsConsumedAPIs(ctx context.Context, tenantMapping types.TenantMapping) error {
	ret := _m.Called(ctx, tenantMapping)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.TenantMapping) error); ok {
		r0 = rf(ctx, tenantMapping)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewTenantMappingsService interface {
	mock.TestingT
	Cleanup(func())
}

// NewTenantMappingsService creates a new instance of TenantMappingsService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTenantMappingsService(t mockConstructorTestingTNewTenantMappingsService) *TenantMappingsService {
	mock := &TenantMappingsService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
