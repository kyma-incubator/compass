// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// BusinessTenantMappingService is an autogenerated mock type for the BusinessTenantMappingService type
type BusinessTenantMappingService struct {
	mock.Mock
}

// CreateTenantAccessForResourceRecursively provides a mock function with given fields: ctx, tenantAccess
func (_m *BusinessTenantMappingService) CreateTenantAccessForResourceRecursively(ctx context.Context, tenantAccess *model.TenantAccess) error {
	ret := _m.Called(ctx, tenantAccess)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.TenantAccess) error); ok {
		r0 = rf(ctx, tenantAccess)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteMany provides a mock function with given fields: ctx, externalTenantIDs
func (_m *BusinessTenantMappingService) DeleteMany(ctx context.Context, externalTenantIDs []string) error {
	ret := _m.Called(ctx, externalTenantIDs)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) error); ok {
		r0 = rf(ctx, externalTenantIDs)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteTenantAccessForResourceRecursively provides a mock function with given fields: ctx, tenantAccess
func (_m *BusinessTenantMappingService) DeleteTenantAccessForResourceRecursively(ctx context.Context, tenantAccess *model.TenantAccess) error {
	ret := _m.Called(ctx, tenantAccess)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.TenantAccess) error); ok {
		r0 = rf(ctx, tenantAccess)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetInternalTenant provides a mock function with given fields: ctx, externalTenant
func (_m *BusinessTenantMappingService) GetInternalTenant(ctx context.Context, externalTenant string) (string, error) {
	ret := _m.Called(ctx, externalTenant)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (string, error)); ok {
		return rf(ctx, externalTenant)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, externalTenant)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, externalTenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLowestOwnerForResource provides a mock function with given fields: ctx, resourceType, objectID
func (_m *BusinessTenantMappingService) GetLowestOwnerForResource(ctx context.Context, resourceType resource.Type, objectID string) (string, error) {
	ret := _m.Called(ctx, resourceType, objectID)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string) (string, error)); ok {
		return rf(ctx, resourceType, objectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string) string); ok {
		r0 = rf(ctx, resourceType, objectID)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string) error); ok {
		r1 = rf(ctx, resourceType, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetParentsRecursivelyByExternalTenant provides a mock function with given fields: ctx, externalTenant
func (_m *BusinessTenantMappingService) GetParentsRecursivelyByExternalTenant(ctx context.Context, externalTenant string) ([]*model.BusinessTenantMapping, error) {
	ret := _m.Called(ctx, externalTenant)

	var r0 []*model.BusinessTenantMapping
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*model.BusinessTenantMapping, error)); ok {
		return rf(ctx, externalTenant)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.BusinessTenantMapping); ok {
		r0 = rf(ctx, externalTenant)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.BusinessTenantMapping)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, externalTenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTenantAccessForResource provides a mock function with given fields: ctx, tenantID, resourceID, resourceType
func (_m *BusinessTenantMappingService) GetTenantAccessForResource(ctx context.Context, tenantID string, resourceID string, resourceType resource.Type) (*model.TenantAccess, error) {
	ret := _m.Called(ctx, tenantID, resourceID, resourceType)

	var r0 *model.TenantAccess
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, resource.Type) (*model.TenantAccess, error)); ok {
		return rf(ctx, tenantID, resourceID, resourceType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, resource.Type) *model.TenantAccess); ok {
		r0 = rf(ctx, tenantID, resourceID, resourceType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.TenantAccess)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, resource.Type) error); ok {
		r1 = rf(ctx, tenantID, resourceID, resourceType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetTenantByExternalID provides a mock function with given fields: ctx, externalID
func (_m *BusinessTenantMappingService) GetTenantByExternalID(ctx context.Context, externalID string) (*model.BusinessTenantMapping, error) {
	ret := _m.Called(ctx, externalID)

	var r0 *model.BusinessTenantMapping
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.BusinessTenantMapping, error)); ok {
		return rf(ctx, externalID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.BusinessTenantMapping); ok {
		r0 = rf(ctx, externalID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BusinessTenantMapping)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, externalID)
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

// List provides a mock function with given fields: ctx
func (_m *BusinessTenantMappingService) List(ctx context.Context) ([]*model.BusinessTenantMapping, error) {
	ret := _m.Called(ctx)

	var r0 []*model.BusinessTenantMapping
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context) ([]*model.BusinessTenantMapping, error)); ok {
		return rf(ctx)
	}
	if rf, ok := ret.Get(0).(func(context.Context) []*model.BusinessTenantMapping); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.BusinessTenantMapping)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListLabels provides a mock function with given fields: ctx, tenantID
func (_m *BusinessTenantMappingService) ListLabels(ctx context.Context, tenantID string) (map[string]*model.Label, error) {
	ret := _m.Called(ctx, tenantID)

	var r0 map[string]*model.Label
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (map[string]*model.Label, error)); ok {
		return rf(ctx, tenantID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) map[string]*model.Label); ok {
		r0 = rf(ctx, tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]*model.Label)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListPageBySearchTerm provides a mock function with given fields: ctx, searchTerm, pageSize, cursor
func (_m *BusinessTenantMappingService) ListPageBySearchTerm(ctx context.Context, searchTerm string, pageSize int, cursor string) (*model.BusinessTenantMappingPage, error) {
	ret := _m.Called(ctx, searchTerm, pageSize, cursor)

	var r0 *model.BusinessTenantMappingPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int, string) (*model.BusinessTenantMappingPage, error)); ok {
		return rf(ctx, searchTerm, pageSize, cursor)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, int, string) *model.BusinessTenantMappingPage); ok {
		r0 = rf(ctx, searchTerm, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BusinessTenantMappingPage)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, int, string) error); ok {
		r1 = rf(ctx, searchTerm, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, tenantInput
func (_m *BusinessTenantMappingService) Update(ctx context.Context, id string, tenantInput model.BusinessTenantMappingInput) error {
	ret := _m.Called(ctx, id, tenantInput)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.BusinessTenantMappingInput) error); ok {
		r0 = rf(ctx, id, tenantInput)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpsertLabel provides a mock function with given fields: ctx, tenantID, key, value
func (_m *BusinessTenantMappingService) UpsertLabel(ctx context.Context, tenantID string, key string, value interface{}) error {
	ret := _m.Called(ctx, tenantID, key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, interface{}) error); ok {
		r0 = rf(ctx, tenantID, key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpsertMany provides a mock function with given fields: ctx, tenantInputs
func (_m *BusinessTenantMappingService) UpsertMany(ctx context.Context, tenantInputs ...model.BusinessTenantMappingInput) ([]string, error) {
	_va := make([]interface{}, len(tenantInputs))
	for _i := range tenantInputs {
		_va[_i] = tenantInputs[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, ctx)
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, ...model.BusinessTenantMappingInput) ([]string, error)); ok {
		return rf(ctx, tenantInputs...)
	}
	if rf, ok := ret.Get(0).(func(context.Context, ...model.BusinessTenantMappingInput) []string); ok {
		r0 = rf(ctx, tenantInputs...)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, ...model.BusinessTenantMappingInput) error); ok {
		r1 = rf(ctx, tenantInputs...)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpsertSingle provides a mock function with given fields: ctx, tenantInput
func (_m *BusinessTenantMappingService) UpsertSingle(ctx context.Context, tenantInput model.BusinessTenantMappingInput) (string, error) {
	ret := _m.Called(ctx, tenantInput)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.BusinessTenantMappingInput) (string, error)); ok {
		return rf(ctx, tenantInput)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.BusinessTenantMappingInput) string); ok {
		r0 = rf(ctx, tenantInput)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.BusinessTenantMappingInput) error); ok {
		r1 = rf(ctx, tenantInput)
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
