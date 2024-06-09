// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	formation "github.com/kyma-incubator/compass/components/director/internal/domain/formation"
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// AsaEngine is an autogenerated mock type for the asaEngine type
type AsaEngine struct {
	mock.Mock
}

// EnsureScenarioAssigned provides a mock function with given fields: ctx, in, processScenarioFunc
func (_m *AsaEngine) EnsureScenarioAssigned(ctx context.Context, in *model.AutomaticScenarioAssignment, processScenarioFunc formation.ProcessScenarioFunc) error {
	ret := _m.Called(ctx, in, processScenarioFunc)

	if len(ret) == 0 {
		panic("no return value specified for EnsureScenarioAssigned")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.AutomaticScenarioAssignment, formation.ProcessScenarioFunc) error); ok {
		r0 = rf(ctx, in, processScenarioFunc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetScenariosFromMatchingASAs provides a mock function with given fields: ctx, objectID, objType
func (_m *AsaEngine) GetScenariosFromMatchingASAs(ctx context.Context, objectID string, objType graphql.FormationObjectType) ([]string, error) {
	ret := _m.Called(ctx, objectID, objType)

	if len(ret) == 0 {
		panic("no return value specified for GetScenariosFromMatchingASAs")
	}

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, graphql.FormationObjectType) ([]string, error)); ok {
		return rf(ctx, objectID, objType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, graphql.FormationObjectType) []string); ok {
		r0 = rf(ctx, objectID, objType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, graphql.FormationObjectType) error); ok {
		r1 = rf(ctx, objectID, objType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// IsFormationComingFromASA provides a mock function with given fields: ctx, objectID, _a2, objectType
func (_m *AsaEngine) IsFormationComingFromASA(ctx context.Context, objectID string, _a2 string, objectType graphql.FormationObjectType) (bool, error) {
	ret := _m.Called(ctx, objectID, _a2, objectType)

	if len(ret) == 0 {
		panic("no return value specified for IsFormationComingFromASA")
	}

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType) (bool, error)); ok {
		return rf(ctx, objectID, _a2, objectType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType) bool); ok {
		r0 = rf(ctx, objectID, _a2, objectType)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, graphql.FormationObjectType) error); ok {
		r1 = rf(ctx, objectID, _a2, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnassignFormationComingFromASA provides a mock function with given fields: ctx, in, processScenarioFunc
func (_m *AsaEngine) UnassignFormationComingFromASA(ctx context.Context, in *model.AutomaticScenarioAssignment, processScenarioFunc formation.ProcessScenarioFunc) error {
	ret := _m.Called(ctx, in, processScenarioFunc)

	if len(ret) == 0 {
		panic("no return value specified for UnassignFormationComingFromASA")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.AutomaticScenarioAssignment, formation.ProcessScenarioFunc) error); ok {
		r0 = rf(ctx, in, processScenarioFunc)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewAsaEngine creates a new instance of AsaEngine. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAsaEngine(t interface {
	mock.TestingT
	Cleanup(func())
}) *AsaEngine {
	mock := &AsaEngine{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
