// Code generated by mockery v1.0.0. DO NOT EDIT.
package automock

import graphql "github.com/kyma-incubator/compass/components/director/internal/graphql"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// ApplicationConverter is an autogenerated mock type for the ApplicationConverter type
type ApplicationConverter struct {
	mock.Mock
}

// InputFromGraphQL provides a mock function with given fields: in
func (_m *ApplicationConverter) InputFromGraphQL(in graphql.ApplicationInput) model.ApplicationInput {
	ret := _m.Called(in)

	var r0 model.ApplicationInput
	if rf, ok := ret.Get(0).(func(graphql.ApplicationInput) model.ApplicationInput); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(model.ApplicationInput)
	}

	return r0
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *ApplicationConverter) MultipleToGraphQL(in []*model.Application) []*graphql.Application {
	ret := _m.Called(in)

	var r0 []*graphql.Application
	if rf, ok := ret.Get(0).(func([]*model.Application) []*graphql.Application); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.Application)
		}
	}

	return r0
}

// ToGraphQL provides a mock function with given fields: in
func (_m *ApplicationConverter) ToGraphQL(in *model.Application) *graphql.Application {
	ret := _m.Called(in)

	var r0 *graphql.Application
	if rf, ok := ret.Get(0).(func(*model.Application) *graphql.Application); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.Application)
		}
	}

	return r0
}
