// Code generated by mockery v2.10.5. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// TenantService is an autogenerated mock type for the tenantService type
type TenantService struct {
	mock.Mock
}

// CreateManyIfNotExists provides a mock function with given fields: ctx, tenantInputs
func (_m *TenantService) CreateManyIfNotExists(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) error {
	_va := make([]interface{}, len(tenantInputs))
	for _i := range tenantInputs {
		_va[_i] = tenantInputs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, ...model.BusinessTenantMappingInput) error); ok {
		r0 = rf(ctx, tenantInputs...)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetExternalTenant provides a mock function with given fields: ctx, id
func (_m *TenantService) GetExternalTenant(ctx context.Context, id string) (string, error) {
	ret := _m.Called(ctx, id)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetInternalTenant provides a mock function with given fields: ctx, externalTenant
func (_m *TenantService) GetInternalTenant(ctx context.Context, externalTenant string) (string, error) {
	ret := _m.Called(ctx, externalTenant)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, externalTenant)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, externalTenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
