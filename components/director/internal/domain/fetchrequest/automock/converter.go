// Code generated by mockery v2.9.4. DO NOT EDIT.

package automock

import (
	fetchrequest "github.com/kyma-incubator/compass/components/director/internal/domain/fetchrequest"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// Converter is an autogenerated mock type for the Converter type
type Converter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: in, objectType
func (_m *Converter) FromEntity(in *fetchrequest.Entity, objectType model.FetchRequestReferenceObjectType) (*model.FetchRequest, error) {
	ret := _m.Called(in, objectType)

	var r0 *model.FetchRequest
	if rf, ok := ret.Get(0).(func(*fetchrequest.Entity, model.FetchRequestReferenceObjectType) *model.FetchRequest); ok {
		r0 = rf(in, objectType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FetchRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*fetchrequest.Entity, model.FetchRequestReferenceObjectType) error); ok {
		r1 = rf(in, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToEntity provides a mock function with given fields: in
func (_m *Converter) ToEntity(in *model.FetchRequest) (*fetchrequest.Entity, error) {
	ret := _m.Called(in)

	var r0 *fetchrequest.Entity
	if rf, ok := ret.Get(0).(func(*model.FetchRequest) *fetchrequest.Entity); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*fetchrequest.Entity)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.FetchRequest) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
