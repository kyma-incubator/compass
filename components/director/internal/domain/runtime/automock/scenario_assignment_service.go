// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// ScenarioAssignmentService is an autogenerated mock type for the ScenarioAssignmentService type
type ScenarioAssignmentService struct {
	mock.Mock
}

// GetForScenarioName provides a mock function with given fields: ctx, scenarioName
func (_m *ScenarioAssignmentService) GetForScenarioName(ctx context.Context, scenarioName string) (*model.AutomaticScenarioAssignment, error) {
	ret := _m.Called(ctx, scenarioName)

	if len(ret) == 0 {
		panic("no return value specified for GetForScenarioName")
	}

	var r0 *model.AutomaticScenarioAssignment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.AutomaticScenarioAssignment, error)); ok {
		return rf(ctx, scenarioName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.AutomaticScenarioAssignment); ok {
		r0 = rf(ctx, scenarioName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AutomaticScenarioAssignment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, scenarioName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListForScenarioNames provides a mock function with given fields: ctx, scenarioNames
func (_m *ScenarioAssignmentService) ListForScenarioNames(ctx context.Context, scenarioNames []string) ([]*model.AutomaticScenarioAssignment, error) {
	ret := _m.Called(ctx, scenarioNames)

	if len(ret) == 0 {
		panic("no return value specified for ListForScenarioNames")
	}

	var r0 []*model.AutomaticScenarioAssignment
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([]*model.AutomaticScenarioAssignment, error)); ok {
		return rf(ctx, scenarioNames)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) []*model.AutomaticScenarioAssignment); ok {
		r0 = rf(ctx, scenarioNames)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.AutomaticScenarioAssignment)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, scenarioNames)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewScenarioAssignmentService creates a new instance of ScenarioAssignmentService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewScenarioAssignmentService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ScenarioAssignmentService {
	mock := &ScenarioAssignmentService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
