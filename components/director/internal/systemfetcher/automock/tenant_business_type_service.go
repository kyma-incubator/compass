// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// TenantBusinessTypeService is an autogenerated mock type for the tenantBusinessTypeService type
type TenantBusinessTypeService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, in
func (_m *TenantBusinessTypeService) Create(ctx context.Context, in *model.TenantBusinessTypeInput) (string, error) {
	ret := _m.Called(ctx, in)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, *model.TenantBusinessTypeInput) string); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *model.TenantBusinessTypeInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: ctx, id
func (_m *TenantBusinessTypeService) GetByID(ctx context.Context, id string) (*model.TenantBusinessType, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.TenantBusinessType
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.TenantBusinessType); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.TenantBusinessType)
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

// ListAll provides a mock function with given fields: ctx
func (_m *TenantBusinessTypeService) ListAll(ctx context.Context) ([]*model.TenantBusinessType, error) {
	ret := _m.Called(ctx)

	var r0 []*model.TenantBusinessType
	if rf, ok := ret.Get(0).(func(context.Context) []*model.TenantBusinessType); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.TenantBusinessType)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewTenantBusinessTypeService interface {
	mock.TestingT
	Cleanup(func())
}

// NewTenantBusinessTypeService creates a new instance of TenantBusinessTypeService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewTenantBusinessTypeService(t mockConstructorTestingTNewTenantBusinessTypeService) *TenantBusinessTypeService {
	mock := &TenantBusinessTypeService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
