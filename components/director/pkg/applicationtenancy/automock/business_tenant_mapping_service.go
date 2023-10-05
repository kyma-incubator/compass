// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	tenant "github.com/kyma-incubator/compass/components/director/pkg/tenant"
)

// BusinessTenantMappingService is an autogenerated mock type for the BusinessTenantMappingService type
type BusinessTenantMappingService struct {
	mock.Mock
}

// CreateTenantAccessForResource provides a mock function with given fields: ctx, tenantAccess
func (_m *BusinessTenantMappingService) CreateTenantAccessForResource(ctx context.Context, tenantAccess *model.TenantAccess) error {
	ret := _m.Called(ctx, tenantAccess)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.TenantAccess) error); ok {
		r0 = rf(ctx, tenantAccess)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetCustomerIDParentRecursively provides a mock function with given fields: ctx, tenantID
func (_m *BusinessTenantMappingService) GetCustomerIDParentRecursively(ctx context.Context, tenantID string) (string, error) {
	ret := _m.Called(ctx, tenantID)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, tenantID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, tenantID)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTenantByExternalID provides a mock function with given fields: ctx, id
func (_m *BusinessTenantMappingService) GetTenantByExternalID(ctx context.Context, id string) (*model.BusinessTenantMapping, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.BusinessTenantMapping
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.BusinessTenantMapping, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.BusinessTenantMapping); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BusinessTenantMapping)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTenantByID provides a mock function with given fields: ctx, id
func (_m *BusinessTenantMappingService) GetTenantByID(ctx context.Context, id string) (*model.BusinessTenantMapping, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.BusinessTenantMapping
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.BusinessTenantMapping, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.BusinessTenantMapping); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BusinessTenantMapping)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByParentAndType provides a mock function with given fields: ctx, parentID, tenantType
func (_m *BusinessTenantMappingService) ListByParentAndType(ctx context.Context, parentID string, tenantType tenant.Type) ([]*model.BusinessTenantMapping, error) {
	ret := _m.Called(ctx, parentID, tenantType)

	var r0 []*model.BusinessTenantMapping
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, tenant.Type) ([]*model.BusinessTenantMapping, error)); ok {
		return rf(ctx, parentID, tenantType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, tenant.Type) []*model.BusinessTenantMapping); ok {
		r0 = rf(ctx, parentID, tenantType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.BusinessTenantMapping)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, tenant.Type) error); ok {
		r1 = rf(ctx, parentID, tenantType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewBusinessTenantMappingService creates a new instance of BusinessTenantMappingService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBusinessTenantMappingService(t interface {
	mock.TestingT
	Cleanup(func())
}) *BusinessTenantMappingService {
	mock := &BusinessTenantMappingService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
