// Code generated by mockery v1.1.2. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// DocumentRepository is an autogenerated mock type for the DocumentRepository type
type DocumentRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, tenant, item
func (_m *DocumentRepository) Create(ctx context.Context, tenant string, item *model.Document) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Document) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, tenant, id
func (_m *DocumentRepository) Delete(ctx context.Context, tenant string, id string) error {
	ret := _m.Called(ctx, tenant, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: ctx, tenant, id
func (_m *DocumentRepository) Exists(ctx context.Context, tenant string, id string) (bool, error) {
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

// GetByID provides a mock function with given fields: ctx, tenant, id
func (_m *DocumentRepository) GetByID(ctx context.Context, tenant string, id string) (*model.Document, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 *model.Document
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Document); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Document)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetForBundle provides a mock function with given fields: ctx, tenant, id, bundleID
func (_m *DocumentRepository) GetForBundle(ctx context.Context, tenant string, id string, bundleID string) (*model.Document, error) {
	ret := _m.Called(ctx, tenant, id, bundleID)

	var r0 *model.Document
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *model.Document); ok {
		r0 = rf(ctx, tenant, id, bundleID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Document)
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

// ListByBundleIDs provides a mock function with given fields: ctx, tenantID, bundleIDs, pageSize, cursor
func (_m *DocumentRepository) ListByBundleIDs(ctx context.Context, tenantID string, bundleIDs []string, pageSize int, cursor string) ([]*model.DocumentPage, error) {
	ret := _m.Called(ctx, tenantID, bundleIDs, pageSize, cursor)

	var r0 []*model.DocumentPage
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, int, string) []*model.DocumentPage); ok {
		r0 = rf(ctx, tenantID, bundleIDs, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.DocumentPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string, int, string) error); ok {
		r1 = rf(ctx, tenantID, bundleIDs, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
