// Code generated by mockery v2.5.1. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// TenantService is an autogenerated mock type for the TenantService type
type TenantService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, item
func (_m *TenantService) Create(ctx context.Context, item model.TenantModel) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.TenantModel) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteByExternalID provides a mock function with given fields: ctx, tenantId
func (_m *TenantService) DeleteByExternalID(ctx context.Context, tenantId string) error {
	ret := _m.Called(ctx, tenantId)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, tenantId)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
