// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

// GetForScenarioName provides a mock function with given fields: ctx, tenantID, scenarioName
func (_m *Repository) GetForScenarioName(ctx context.Context, tenantID string, scenarioName string) (model.AutomaticScenarioAssignment, error) {
	ret := _m.Called(ctx, tenantID, scenarioName)

	var r0 model.AutomaticScenarioAssignment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (model.AutomaticScenarioAssignment, error)); ok {
		return rf(ctx, tenantID, scenarioName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) model.AutomaticScenarioAssignment); ok {
		r0 = rf(ctx, tenantID, scenarioName)
	} else {
		r0 = ret.Get(0).(model.AutomaticScenarioAssignment)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenantID, scenarioName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, tenant, pageSize, cursor
func (_m *Repository) List(ctx context.Context, tenant string, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error) {
	ret := _m.Called(ctx, tenant, pageSize, cursor)

	var r0 *model.AutomaticScenarioAssignmentPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, int, string) (*model.AutomaticScenarioAssignmentPage, error)); ok {
		return rf(ctx, tenant, pageSize, cursor)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, int, string) *model.AutomaticScenarioAssignmentPage); ok {
		r0 = rf(ctx, tenant, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AutomaticScenarioAssignmentPage)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, int, string) error); ok {
		r1 = rf(ctx, tenant, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListForTargetTenant provides a mock function with given fields: ctx, tenantID, targetTenantID
func (_m *Repository) ListForTargetTenant(ctx context.Context, tenantID string, targetTenantID string) ([]*model.AutomaticScenarioAssignment, error) {
	ret := _m.Called(ctx, tenantID, targetTenantID)

	var r0 []*model.AutomaticScenarioAssignment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) ([]*model.AutomaticScenarioAssignment, error)); ok {
		return rf(ctx, tenantID, targetTenantID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []*model.AutomaticScenarioAssignment); ok {
		r0 = rf(ctx, tenantID, targetTenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.AutomaticScenarioAssignment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenantID, targetTenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewRepository creates a new instance of Repository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *Repository {
	mock := &Repository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
