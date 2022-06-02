// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// AutomaticFormationAssignmentRepository is an autogenerated mock type for the automaticFormationAssignmentRepository type
type AutomaticFormationAssignmentRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, _a1
func (_m *AutomaticFormationAssignmentRepository) Create(ctx context.Context, _a1 model.AutomaticScenarioAssignment) error {
	ret := _m.Called(ctx, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.AutomaticScenarioAssignment) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteForScenarioName provides a mock function with given fields: ctx, tenantID, scenarioName
func (_m *AutomaticFormationAssignmentRepository) DeleteForScenarioName(ctx context.Context, tenantID string, scenarioName string) error {
	ret := _m.Called(ctx, tenantID, scenarioName)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, tenantID, scenarioName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteForTargetTenant provides a mock function with given fields: ctx, tenantID, targetTenantID
func (_m *AutomaticFormationAssignmentRepository) DeleteForTargetTenant(ctx context.Context, tenantID string, targetTenantID string) error {
	ret := _m.Called(ctx, tenantID, targetTenantID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, tenantID, targetTenantID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListAll provides a mock function with given fields: ctx, tenantID
func (_m *AutomaticFormationAssignmentRepository) ListAll(ctx context.Context, tenantID string) ([]*model.AutomaticScenarioAssignment, error) {
	ret := _m.Called(ctx, tenantID)

	var r0 []*model.AutomaticScenarioAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.AutomaticScenarioAssignment); ok {
		r0 = rf(ctx, tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.AutomaticScenarioAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
