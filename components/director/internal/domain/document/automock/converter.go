// Code generated by mockery v2.12.1. DO NOT EDIT.

package automock

import (
	document "github.com/kyma-incubator/compass/components/director/internal/domain/document"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// Converter is an autogenerated mock type for the Converter type
type Converter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: in
func (_m *Converter) FromEntity(in *document.Entity) (*model.Document, error) {
	ret := _m.Called(in)

	var r0 *model.Document
	if rf, ok := ret.Get(0).(func(*document.Entity) *model.Document); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Document)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*document.Entity) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToEntity provides a mock function with given fields: in
func (_m *Converter) ToEntity(in *model.Document) (*document.Entity, error) {
	ret := _m.Called(in)

	var r0 *document.Entity
	if rf, ok := ret.Get(0).(func(*model.Document) *document.Entity); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*document.Entity)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.Document) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewConverter creates a new instance of Converter. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewConverter(t testing.TB) *Converter {
	mock := &Converter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
