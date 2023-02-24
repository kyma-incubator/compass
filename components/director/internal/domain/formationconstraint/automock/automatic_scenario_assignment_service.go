// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// AutomaticScenarioAssignmentService is an autogenerated mock type for the automaticScenarioAssignmentService type
type AutomaticScenarioAssignmentService struct {
	mock.Mock
}

// ListForTargetTenant provides a mock function with given fields: ctx, targetTenantInternalID
func (_m *AutomaticScenarioAssignmentService) ListForTargetTenant(ctx context.Context, targetTenantInternalID string) ([]*model.AutomaticScenarioAssignment, error) {
	ret := _m.Called(ctx, targetTenantInternalID)

	var r0 []*model.AutomaticScenarioAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.AutomaticScenarioAssignment); ok {
		r0 = rf(ctx, targetTenantInternalID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.AutomaticScenarioAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, targetTenantInternalID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewAutomaticScenarioAssignmentService interface {
	mock.TestingT
	Cleanup(func())
}

// NewAutomaticScenarioAssignmentService creates a new instance of AutomaticScenarioAssignmentService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAutomaticScenarioAssignmentService(t mockConstructorTestingTNewAutomaticScenarioAssignmentService) *AutomaticScenarioAssignmentService {
	mock := &AutomaticScenarioAssignmentService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
