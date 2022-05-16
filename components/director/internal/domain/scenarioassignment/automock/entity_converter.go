// Code generated by mockery. DO NOT EDIT.

package automock

import (
	scenarioassignment "github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"
	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// EntityConverter is an autogenerated mock type for the EntityConverter type
type EntityConverter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: assignment
func (_m *EntityConverter) FromEntity(assignment scenarioassignment.Entity) model.AutomaticScenarioAssignment {
	ret := _m.Called(assignment)

	var r0 model.AutomaticScenarioAssignment
	if rf, ok := ret.Get(0).(func(scenarioassignment.Entity) model.AutomaticScenarioAssignment); ok {
		r0 = rf(assignment)
	} else {
		r0 = ret.Get(0).(model.AutomaticScenarioAssignment)
	}

	return r0
}

// ToEntity provides a mock function with given fields: assignment
func (_m *EntityConverter) ToEntity(assignment model.AutomaticScenarioAssignment) scenarioassignment.Entity {
	ret := _m.Called(assignment)

	var r0 scenarioassignment.Entity
	if rf, ok := ret.Get(0).(func(model.AutomaticScenarioAssignment) scenarioassignment.Entity); ok {
		r0 = rf(assignment)
	} else {
		r0 = ret.Get(0).(scenarioassignment.Entity)
	}

	return r0
}

// NewEntityConverter creates a new instance of EntityConverter. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewEntityConverter(t testing.TB) *EntityConverter {
	mock := &EntityConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
