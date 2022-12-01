// Code generated by mockery. DO NOT EDIT.

package automock

import (
	eventdef "github.com/kyma-incubator/compass/components/director/internal/domain/eventdef"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// EventAPIDefinitionConverter is an autogenerated mock type for the EventAPIDefinitionConverter type
type EventAPIDefinitionConverter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: entity
func (_m *EventAPIDefinitionConverter) FromEntity(entity *eventdef.Entity) *model.EventDefinition {
	ret := _m.Called(entity)

	var r0 *model.EventDefinition
	if rf, ok := ret.Get(0).(func(*eventdef.Entity) *model.EventDefinition); ok {
		r0 = rf(entity)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.EventDefinition)
		}
	}

	return r0
}

// ToEntity provides a mock function with given fields: apiModel
func (_m *EventAPIDefinitionConverter) ToEntity(apiModel *model.EventDefinition) *eventdef.Entity {
	ret := _m.Called(apiModel)

	var r0 *eventdef.Entity
	if rf, ok := ret.Get(0).(func(*model.EventDefinition) *eventdef.Entity); ok {
		r0 = rf(apiModel)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*eventdef.Entity)
		}
	}

	return r0
}

type mockConstructorTestingTNewEventAPIDefinitionConverter interface {
	mock.TestingT
	Cleanup(func())
}

// NewEventAPIDefinitionConverter creates a new instance of EventAPIDefinitionConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewEventAPIDefinitionConverter(t mockConstructorTestingTNewEventAPIDefinitionConverter) *EventAPIDefinitionConverter {
	mock := &EventAPIDefinitionConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
