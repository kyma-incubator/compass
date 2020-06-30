// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import document "github.com/kyma-incubator/compass/components/director/internal/domain/document"
import graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// DocumentConverter is an autogenerated mock type for the DocumentConverter type
type DocumentConverter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: in
func (_m *DocumentConverter) FromEntity(in document.Entity) (model.Document, error) {
	ret := _m.Called(in)

	var r0 model.Document
	if rf, ok := ret.Get(0).(func(document.Entity) model.Document); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(model.Document)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(document.Entity) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InputFromGraphQL provides a mock function with given fields: in
func (_m *DocumentConverter) InputFromGraphQL(in *graphql.DocumentInput) (*model.DocumentInput, error) {
	ret := _m.Called(in)

	var r0 *model.DocumentInput
	if rf, ok := ret.Get(0).(func(*graphql.DocumentInput) *model.DocumentInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.DocumentInput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*graphql.DocumentInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToEntity provides a mock function with given fields: in
func (_m *DocumentConverter) ToEntity(in model.Document) (document.Entity, error) {
	ret := _m.Called(in)

	var r0 document.Entity
	if rf, ok := ret.Get(0).(func(model.Document) document.Entity); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(document.Entity)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.Document) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGraphQL provides a mock function with given fields: in
func (_m *DocumentConverter) ToGraphQL(in *model.Document) *graphql.Document {
	ret := _m.Called(in)

	var r0 *graphql.Document
	if rf, ok := ret.Get(0).(func(*model.Document) *graphql.Document); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.Document)
		}
	}

	return r0
}
