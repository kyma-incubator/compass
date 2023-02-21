// Code generated by mockery. DO NOT EDIT.

package automock

import (
	formationconstraint "github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// EntityConverter is an autogenerated mock type for the EntityConverter type
type EntityConverter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: entity
func (_m *EntityConverter) FromEntity(entity *formationconstraint.Entity) *model.FormationConstraint {
	ret := _m.Called(entity)

	var r0 *model.FormationConstraint
	if rf, ok := ret.Get(0).(func(*formationconstraint.Entity) *model.FormationConstraint); ok {
		r0 = rf(entity)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationConstraint)
		}
	}

	return r0
}

// ToEntity provides a mock function with given fields: in
func (_m *EntityConverter) ToEntity(in *model.FormationConstraint) *formationconstraint.Entity {
	ret := _m.Called(in)

	var r0 *formationconstraint.Entity
	if rf, ok := ret.Get(0).(func(*model.FormationConstraint) *formationconstraint.Entity); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*formationconstraint.Entity)
		}
	}

	return r0
}

type NewEntityConverterT interface {
	mock.TestingT
	Cleanup(func())
}

// NewEntityConverter creates a new instance of EntityConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewEntityConverter(t NewEntityConverterT) *EntityConverter {
	mock := &EntityConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
