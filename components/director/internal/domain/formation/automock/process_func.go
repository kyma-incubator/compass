// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// ProcessFunc is an autogenerated mock type for the processFunc type
type ProcessFunc struct {
	mock.Mock
}

// ProcessScenarioFunc provides a mock function with given fields: _a0, _a1, _a2, _a3, _a4
func (_m *ProcessFunc) ProcessScenarioFunc(_a0 context.Context, _a1 string, _a2 string, _a3 graphql.FormationObjectType, _a4 model.Formation) (*model.Formation, error) {
	ret := _m.Called(_a0, _a1, _a2, _a3, _a4)

	if len(ret) == 0 {
		panic("no return value specified for ProcessScenarioFunc")
	}

	var r0 *model.Formation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation) (*model.Formation, error)); ok {
		return rf(_a0, _a1, _a2, _a3, _a4)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation) *model.Formation); ok {
		r0 = rf(_a0, _a1, _a2, _a3, _a4)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation) error); ok {
		r1 = rf(_a0, _a1, _a2, _a3, _a4)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewProcessFunc creates a new instance of ProcessFunc. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewProcessFunc(t interface {
	mock.TestingT
	Cleanup(func())
}) *ProcessFunc {
	mock := &ProcessFunc{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
