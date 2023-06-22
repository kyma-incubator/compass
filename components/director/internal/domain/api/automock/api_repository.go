// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// APIRepository is an autogenerated mock type for the APIRepository type
type APIRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, tenant, item
func (_m *APIRepository) Create(ctx context.Context, tenant string, item *model.APIDefinition) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.APIDefinition) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateGlobal provides a mock function with given fields: ctx, item
func (_m *APIRepository) CreateGlobal(ctx context.Context, item *model.APIDefinition) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.APIDefinition) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateMany provides a mock function with given fields: ctx, tenant, item
func (_m *APIRepository) CreateMany(ctx context.Context, tenant string, item []*model.APIDefinition) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []*model.APIDefinition) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, tenantID, id
func (_m *APIRepository) Delete(ctx context.Context, tenantID string, id string) error {
	ret := _m.Called(ctx, tenantID, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, tenantID, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteAllByBundleID provides a mock function with given fields: ctx, tenantID, bundleID
func (_m *APIRepository) DeleteAllByBundleID(ctx context.Context, tenantID string, bundleID string) error {
	ret := _m.Called(ctx, tenantID, bundleID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, tenantID, bundleID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteGlobal provides a mock function with given fields: ctx, id
func (_m *APIRepository) DeleteGlobal(ctx context.Context, id string) error {
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
func (_m *APIRepository) Exists(ctx context.Context, tenant string, id string) (bool, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: ctx, tenantID, id
func (_m *APIRepository) GetByID(ctx context.Context, tenantID string, id string) (*model.APIDefinition, error) {
	ret := _m.Called(ctx, tenantID, id)

	var r0 *model.APIDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.APIDefinition); ok {
		r0 = rf(ctx, tenantID, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.APIDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenantID, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByIDGlobal provides a mock function with given fields: ctx, id
func (_m *APIRepository) GetByIDGlobal(ctx context.Context, id string) (*model.APIDefinition, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.APIDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.APIDefinition); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.APIDefinition)
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

// GetForBundle provides a mock function with given fields: ctx, tenant, id, bundleID
func (_m *APIRepository) GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.APIDefinition, error) {
	ret := _m.Called(ctx, tenant, id, bundleID)

	var r0 *model.APIDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *model.APIDefinition); ok {
		r0 = rf(ctx, tenant, id, bundleID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.APIDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, tenant, id, bundleID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByBundleIDs provides a mock function with given fields: ctx, tenantID, bundleIDs, bundleRefs, counts, pageSize, cursor
func (_m *APIRepository) ListByBundleIDs(ctx context.Context, tenantID string, bundleIDs []string, bundleRefs []*model.BundleReference, counts map[string]int, pageSize int, cursor string) ([]*model.APIDefinitionPage, error) {
	ret := _m.Called(ctx, tenantID, bundleIDs, bundleRefs, counts, pageSize, cursor)

	var r0 []*model.APIDefinitionPage
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, []*model.BundleReference, map[string]int, int, string) []*model.APIDefinitionPage); ok {
		r0 = rf(ctx, tenantID, bundleIDs, bundleRefs, counts, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.APIDefinitionPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string, []*model.BundleReference, map[string]int, int, string) error); ok {
		r1 = rf(ctx, tenantID, bundleIDs, bundleRefs, counts, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByResourceID provides a mock function with given fields: ctx, tenantID, resourceType, resourceID
func (_m *APIRepository) ListByResourceID(ctx context.Context, tenantID string, resourceType resource.Type, resourceID string) ([]*model.APIDefinition, error) {
	ret := _m.Called(ctx, tenantID, resourceType, resourceID)

	var r0 []*model.APIDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string, resource.Type, string) []*model.APIDefinition); ok {
		r0 = rf(ctx, tenantID, resourceType, resourceID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.APIDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, resource.Type, string) error); ok {
		r1 = rf(ctx, tenantID, resourceType, resourceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, tenant, item
func (_m *APIRepository) Update(ctx context.Context, tenant string, item *model.APIDefinition) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.APIDefinition) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateGlobal provides a mock function with given fields: ctx, item
func (_m *APIRepository) UpdateGlobal(ctx context.Context, item *model.APIDefinition) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.APIDefinition) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewAPIRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewAPIRepository creates a new instance of APIRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAPIRepository(t mockConstructorTestingTNewAPIRepository) *APIRepository {
	mock := &APIRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
