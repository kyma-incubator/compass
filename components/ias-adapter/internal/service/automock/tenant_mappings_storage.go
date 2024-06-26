// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	types "github.com/kyma-incubator/compass/components/ias-adapter/internal/types"
)

// TenantMappingsStorage is an autogenerated mock type for the TenantMappingsStorage type
type TenantMappingsStorage struct {
	mock.Mock
}

// DeleteTenantMapping provides a mock function with given fields: ctx, formationID, applicationID
func (_m *TenantMappingsStorage) DeleteTenantMapping(ctx context.Context, formationID string, applicationID string) error {
	ret := _m.Called(ctx, formationID, applicationID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, formationID, applicationID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListTenantMappings provides a mock function with given fields: ctx, formationID
func (_m *TenantMappingsStorage) ListTenantMappings(ctx context.Context, formationID string) (map[string]types.TenantMapping, error) {
	ret := _m.Called(ctx, formationID)

	var r0 map[string]types.TenantMapping
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (map[string]types.TenantMapping, error)); ok {
		return rf(ctx, formationID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) map[string]types.TenantMapping); ok {
		r0 = rf(ctx, formationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]types.TenantMapping)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, formationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpsertTenantMapping provides a mock function with given fields: ctx, tenantMapping
func (_m *TenantMappingsStorage) UpsertTenantMapping(ctx context.Context, tenantMapping types.TenantMapping) error {
	ret := _m.Called(ctx, tenantMapping)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, types.TenantMapping) error); ok {
		r0 = rf(ctx, tenantMapping)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewTenantMappingsStorage interface {
	mock.TestingT
	Cleanup(func())
}

// NewTenantMappingsStorage creates a new instance of TenantMappingsStorage. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTenantMappingsStorage(t mockConstructorTestingTNewTenantMappingsStorage) *TenantMappingsStorage {
	mock := &TenantMappingsStorage{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
