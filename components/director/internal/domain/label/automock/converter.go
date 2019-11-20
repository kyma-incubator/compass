// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import label "github.com/kyma-incubator/compass/components/director/internal/domain/label"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// Converter is an autogenerated mock type for the Converter type
type Converter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: in
func (_m *Converter) FromEntity(in label.Entity) (model.Label, error) {
	ret := _m.Called(in)

	var r0 model.Label
	if rf, ok := ret.Get(0).(func(label.Entity) model.Label); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(model.Label)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(label.Entity) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToEntity provides a mock function with given fields: in
func (_m *Converter) ToEntity(in model.Label) (label.Entity, error) {
	ret := _m.Called(in)

	var r0 label.Entity
	if rf, ok := ret.Get(0).(func(model.Label) label.Entity); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(label.Entity)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.Label) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
