// Code generated by mockery v1.1.2. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// SpecRepository is an autogenerated mock type for the SpecRepository type
type SpecRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, tenant, item
func (_m *SpecRepository) Create(ctx context.Context, tenant string, item *model.Spec) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Spec) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, tenant, id, objectType
func (_m *SpecRepository) Delete(ctx context.Context, tenant string, id string, objectType model.SpecReferenceObjectType) error {
	ret := _m.Called(ctx, tenant, id, objectType)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.SpecReferenceObjectType) error); ok {
		r0 = rf(ctx, tenant, id, objectType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteByReferenceObjectID provides a mock function with given fields: ctx, tenant, objectType, objectID
func (_m *SpecRepository) DeleteByReferenceObjectID(ctx context.Context, tenant string, objectType model.SpecReferenceObjectType, objectID string) error {
	ret := _m.Called(ctx, tenant, objectType, objectID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.SpecReferenceObjectType, string) error); ok {
		r0 = rf(ctx, tenant, objectType, objectID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: ctx, tenantID, id, objectType
func (_m *SpecRepository) Exists(ctx context.Context, tenantID string, id string, objectType model.SpecReferenceObjectType) (bool, error) {
	ret := _m.Called(ctx, tenantID, id, objectType)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.SpecReferenceObjectType) bool); ok {
		r0 = rf(ctx, tenantID, id, objectType)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, model.SpecReferenceObjectType) error); ok {
		r1 = rf(ctx, tenantID, id, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: ctx, tenantID, id, objectType
func (_m *SpecRepository) GetByID(ctx context.Context, tenantID string, id string, objectType model.SpecReferenceObjectType) (*model.Spec, error) {
	ret := _m.Called(ctx, tenantID, id, objectType)

	var r0 *model.Spec
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.SpecReferenceObjectType) *model.Spec); ok {
		r0 = rf(ctx, tenantID, id, objectType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Spec)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, model.SpecReferenceObjectType) error); ok {
		r1 = rf(ctx, tenantID, id, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByReferenceObjectID provides a mock function with given fields: ctx, tenant, objectType, objectID
func (_m *SpecRepository) ListByReferenceObjectID(ctx context.Context, tenant string, objectType model.SpecReferenceObjectType, objectID string) ([]*model.Spec, error) {
	ret := _m.Called(ctx, tenant, objectType, objectID)

	var r0 []*model.Spec
	if rf, ok := ret.Get(0).(func(context.Context, string, model.SpecReferenceObjectType, string) []*model.Spec); ok {
		r0 = rf(ctx, tenant, objectType, objectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Spec)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, model.SpecReferenceObjectType, string) error); ok {
		r1 = rf(ctx, tenant, objectType, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByReferenceObjectIDs provides a mock function with given fields: ctx, tenant, objectType, objectIDs
func (_m *SpecRepository) ListByReferenceObjectIDs(ctx context.Context, tenant string, objectType model.SpecReferenceObjectType, objectIDs []string) ([]*model.Spec, error) {
	ret := _m.Called(ctx, tenant, objectType, objectIDs)

	var r0 []*model.Spec
	if rf, ok := ret.Get(0).(func(context.Context, string, model.SpecReferenceObjectType, []string) []*model.Spec); ok {
		r0 = rf(ctx, tenant, objectType, objectIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Spec)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, model.SpecReferenceObjectType, []string) error); ok {
		r1 = rf(ctx, tenant, objectType, objectIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, tenant, item
func (_m *SpecRepository) Update(ctx context.Context, tenant string, item *model.Spec) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Spec) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
