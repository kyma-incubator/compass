// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	tombstone "github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"
	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// EntityConverter is an autogenerated mock type for the EntityConverter type
type EntityConverter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: entity
func (_m *EntityConverter) FromEntity(entity *tombstone.Entity) (*model.Tombstone, error) {
	ret := _m.Called(entity)

	var r0 *model.Tombstone
	if rf, ok := ret.Get(0).(func(*tombstone.Entity) *model.Tombstone); ok {
		r0 = rf(entity)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Tombstone)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*tombstone.Entity) error); ok {
		r1 = rf(entity)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToEntity provides a mock function with given fields: in
func (_m *EntityConverter) ToEntity(in *model.Tombstone) *tombstone.Entity {
	ret := _m.Called(in)

	var r0 *tombstone.Entity
	if rf, ok := ret.Get(0).(func(*model.Tombstone) *tombstone.Entity); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tombstone.Entity)
		}
	}

	return r0
}
