// Code generated by mockery v2.10.4. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// TenantMappingRepository is an autogenerated mock type for the TenantMappingRepository type
type TenantMappingRepository struct {
	mock.Mock
}

// DeleteByExternalTenant provides a mock function with given fields: ctx, externalTenant
func (_m *TenantMappingRepository) DeleteByExternalTenant(ctx context.Context, externalTenant string) error {
	ret := _m.Called(ctx, externalTenant)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, externalTenant)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: ctx, id
func (_m *TenantMappingRepository) Exists(ctx context.Context, id string) (bool, error) {
	ret := _m.Called(ctx, id)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ExistsByExternalTenant provides a mock function with given fields: ctx, externalTenant
func (_m *TenantMappingRepository) ExistsByExternalTenant(ctx context.Context, externalTenant string) (bool, error) {
	ret := _m.Called(ctx, externalTenant)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, externalTenant)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, externalTenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: ctx, id
func (_m *TenantMappingRepository) Get(ctx context.Context, id string) (*model.BusinessTenantMapping, error) {
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

// GetByExternalTenant provides a mock function with given fields: ctx, externalTenant
func (_m *TenantMappingRepository) GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error) {
	ret := _m.Called(ctx, externalTenant)

	var r0 *model.BusinessTenantMapping
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.BusinessTenantMapping); ok {
		r0 = rf(ctx, externalTenant)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BusinessTenantMapping)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, externalTenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLowestOwnerForResource provides a mock function with given fields: ctx, resourceType, objectID
func (_m *TenantMappingRepository) GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error) {
	ret := _m.Called(ctx, resourceType, objectID)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string) string); ok {
		r0 = rf(ctx, resourceType, objectID)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string) error); ok {
		r1 = rf(ctx, resourceType, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx
func (_m *TenantMappingRepository) List(ctx context.Context) ([]*model.BusinessTenantMapping, error) {
	ret := _m.Called(ctx)

	var r0 []*model.BusinessTenantMapping
	if rf, ok := ret.Get(0).(func(context.Context) []*model.BusinessTenantMapping); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.BusinessTenantMapping)
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

// ListByExternalTenants provides a mock function with given fields: ctx, externalTenant
func (_m *TenantMappingRepository) ListByExternalTenants(ctx context.Context, externalTenant []string) ([]*model.BusinessTenantMapping, error) {
	ret := _m.Called(ctx, externalTenant)

	var r0 []*model.BusinessTenantMapping
	if rf, ok := ret.Get(0).(func(context.Context, []string) []*model.BusinessTenantMapping); ok {
		r0 = rf(ctx, externalTenant)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.BusinessTenantMapping)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, externalTenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListPageBySearchTerm provides a mock function with given fields: ctx, searchTerm, pageSize, cursor
func (_m *TenantMappingRepository) ListPageBySearchTerm(ctx context.Context, searchTerm string, pageSize int, cursor string) (*model.BusinessTenantMappingPage, error) {
	ret := _m.Called(ctx, searchTerm, pageSize, cursor)

	var r0 *model.BusinessTenantMappingPage
	if rf, ok := ret.Get(0).(func(context.Context, string, int, string) *model.BusinessTenantMappingPage); ok {
		r0 = rf(ctx, searchTerm, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BusinessTenantMappingPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, int, string) error); ok {
		r1 = rf(ctx, searchTerm, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnsafeCreate provides a mock function with given fields: ctx, item
func (_m *TenantMappingRepository) UnsafeCreate(ctx context.Context, item model.BusinessTenantMapping) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.BusinessTenantMapping) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: ctx, _a1
func (_m *TenantMappingRepository) Update(ctx context.Context, _a1 *model.BusinessTenantMapping) error {
	ret := _m.Called(ctx, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.BusinessTenantMapping) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Upsert provides a mock function with given fields: ctx, item
func (_m *TenantMappingRepository) Upsert(ctx context.Context, item model.BusinessTenantMapping) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.BusinessTenantMapping) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
