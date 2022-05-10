// Code generated by mockery v2.10.5. DO NOT EDIT.

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
