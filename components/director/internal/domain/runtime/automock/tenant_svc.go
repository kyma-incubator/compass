// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// TenantSvc is an autogenerated mock type for the tenantSvc type
type TenantSvc struct {
	mock.Mock
}

// GetTenantByID provides a mock function with given fields: ctx, id
func (_m *TenantSvc) GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.BusinessTenantMapping
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.BusinessTenantMapping); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BusinessTenantMapping)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewTenantSvc interface {
	mock.TestingT
	Cleanup(func())
}

// NewTenantSvc creates a new instance of TenantSvc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTenantSvc(t mockConstructorTestingTNewTenantSvc) *TenantSvc {
	mock := &TenantSvc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
