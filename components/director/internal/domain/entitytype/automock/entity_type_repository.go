// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// EntityTypeRepository is an autogenerated mock type for the EntityTypeRepository type
type EntityTypeRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, tenant, item
func (_m *EntityTypeRepository) Create(ctx context.Context, tenant string, item *model.EntityType) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.EntityType) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateGlobal provides a mock function with given fields: ctx, _a1
func (_m *EntityTypeRepository) CreateGlobal(ctx context.Context, _a1 *model.EntityType) error {
	ret := _m.Called(ctx, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.EntityType) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, tenant, id
func (_m *EntityTypeRepository) Delete(ctx context.Context, tenant string, id string) error {
	ret := _m.Called(ctx, tenant, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteGlobal provides a mock function with given fields: ctx, id
func (_m *EntityTypeRepository) DeleteGlobal(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: ctx, tenant, id
func (_m *EntityTypeRepository) Exists(ctx context.Context, tenant string, id string) (bool, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (bool, error)); ok {
		return rf(ctx, tenant, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: ctx, tenant, id
func (_m *EntityTypeRepository) GetByID(ctx context.Context, tenant string, id string) (*model.EntityType, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 *model.EntityType
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.EntityType, error)); ok {
		return rf(ctx, tenant, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.EntityType); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.EntityType)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByIDGlobal provides a mock function with given fields: ctx, id
func (_m *EntityTypeRepository) GetByIDGlobal(ctx context.Context, id string) (*model.EntityType, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.EntityType
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.EntityType, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.EntityType); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.EntityType)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByResourceID provides a mock function with given fields: ctx, tenantID, resourceID, resourceType
func (_m *EntityTypeRepository) ListByResourceID(ctx context.Context, tenantID string, resourceID string, resourceType resource.Type) ([]*model.EntityType, error) {
	ret := _m.Called(ctx, tenantID, resourceID, resourceType)

	var r0 []*model.EntityType
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, resource.Type) ([]*model.EntityType, error)); ok {
		return rf(ctx, tenantID, resourceID, resourceType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, resource.Type) []*model.EntityType); ok {
		r0 = rf(ctx, tenantID, resourceID, resourceType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.EntityType)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, resource.Type) error); ok {
		r1 = rf(ctx, tenantID, resourceID, resourceType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, tenant, item
func (_m *EntityTypeRepository) Update(ctx context.Context, tenant string, item *model.EntityType) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.EntityType) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateGlobal provides a mock function with given fields: ctx, _a1
func (_m *EntityTypeRepository) UpdateGlobal(ctx context.Context, _a1 *model.EntityType) error {
	ret := _m.Called(ctx, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.EntityType) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewEntityTypeRepository creates a new instance of EntityTypeRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEntityTypeRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *EntityTypeRepository {
	mock := &EntityTypeRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
