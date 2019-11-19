// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	eventapi "github.com/kyma-incubator/compass/components/director/internal/domain/eventapi"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// EventAPIDefinitionConverter is an autogenerated mock type for the EventAPIDefinitionConverter type
type EventAPIDefinitionConverter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: entity
func (_m *EventAPIDefinitionConverter) FromEntity(entity eventapi.Entity) (model.EventAPIDefinition, error) {
	ret := _m.Called(entity)

	var r0 model.EventAPIDefinition
	if rf, ok := ret.Get(0).(func(eventapi.Entity) model.EventAPIDefinition); ok {
		r0 = rf(entity)
	} else {
		r0 = ret.Get(0).(model.EventAPIDefinition)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(eventapi.Entity) error); ok {
		r1 = rf(entity)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToEntity provides a mock function with given fields: apiModel
func (_m *EventAPIDefinitionConverter) ToEntity(apiModel model.EventAPIDefinition) (eventapi.Entity, error) {
	ret := _m.Called(apiModel)

	var r0 eventapi.Entity
	if rf, ok := ret.Get(0).(func(model.EventAPIDefinition) eventapi.Entity); ok {
		r0 = rf(apiModel)
	} else {
		r0 = ret.Get(0).(eventapi.Entity)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.EventAPIDefinition) error); ok {
		r1 = rf(apiModel)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
