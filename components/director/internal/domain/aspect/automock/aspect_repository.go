// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// AspectRepository is an autogenerated mock type for the AspectRepository type
type AspectRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, tenant, item
func (_m *AspectRepository) Create(ctx context.Context, tenant string, item *model.Aspect) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Aspect) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteByIntegrationDependencyID provides a mock function with given fields: ctx, tenant, integrationDependencyID
func (_m *AspectRepository) DeleteByIntegrationDependencyID(ctx context.Context, tenant string, integrationDependencyID string) error {
	ret := _m.Called(ctx, tenant, integrationDependencyID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, tenant, integrationDependencyID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListByApplicationIDs provides a mock function with given fields: ctx, tenantID, applicationIDs, pageSize, cursor
func (_m *AspectRepository) ListByApplicationIDs(ctx context.Context, tenantID string, applicationIDs []string, pageSize int, cursor string) ([]*model.Aspect, map[string]int, error) {
	ret := _m.Called(ctx, tenantID, applicationIDs, pageSize, cursor)

	var r0 []*model.Aspect
	var r1 map[string]int
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, int, string) ([]*model.Aspect, map[string]int, error)); ok {
		return rf(ctx, tenantID, applicationIDs, pageSize, cursor)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, int, string) []*model.Aspect); ok {
		r0 = rf(ctx, tenantID, applicationIDs, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Aspect)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []string, int, string) map[string]int); ok {
		r1 = rf(ctx, tenantID, applicationIDs, pageSize, cursor)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(map[string]int)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, string, []string, int, string) error); ok {
		r2 = rf(ctx, tenantID, applicationIDs, pageSize, cursor)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// ListByIntegrationDependencyID provides a mock function with given fields: ctx, tenant, integrationDependencyID
func (_m *AspectRepository) ListByIntegrationDependencyID(ctx context.Context, tenant string, integrationDependencyID string) ([]*model.Aspect, error) {
	ret := _m.Called(ctx, tenant, integrationDependencyID)

	var r0 []*model.Aspect
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) ([]*model.Aspect, error)); ok {
		return rf(ctx, tenant, integrationDependencyID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []*model.Aspect); ok {
		r0 = rf(ctx, tenant, integrationDependencyID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Aspect)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, integrationDependencyID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewAspectRepository creates a new instance of AspectRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAspectRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *AspectRepository {
	mock := &AspectRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
