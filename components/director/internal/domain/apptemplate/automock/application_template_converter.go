// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// ApplicationTemplateConverter is an autogenerated mock type for the ApplicationTemplateConverter type
type ApplicationTemplateConverter struct {
	mock.Mock
}

// ApplicationFromTemplateInputFromGraphQL provides a mock function with given fields: appTemplate, in
func (_m *ApplicationTemplateConverter) ApplicationFromTemplateInputFromGraphQL(appTemplate *model.ApplicationTemplate, in graphql.ApplicationFromTemplateInput) (model.ApplicationFromTemplateInput, error) {
	ret := _m.Called(appTemplate, in)

	var r0 model.ApplicationFromTemplateInput
	if rf, ok := ret.Get(0).(func(*model.ApplicationTemplate, graphql.ApplicationFromTemplateInput) model.ApplicationFromTemplateInput); ok {
		r0 = rf(appTemplate, in)
	} else {
		r0 = ret.Get(0).(model.ApplicationFromTemplateInput)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.ApplicationTemplate, graphql.ApplicationFromTemplateInput) error); ok {
		r1 = rf(appTemplate, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// InputFromGraphQL provides a mock function with given fields: in
func (_m *ApplicationTemplateConverter) InputFromGraphQL(in graphql.ApplicationTemplateInput) (model.ApplicationTemplateInput, error) {
	ret := _m.Called(in)

	var r0 model.ApplicationTemplateInput
	if rf, ok := ret.Get(0).(func(graphql.ApplicationTemplateInput) model.ApplicationTemplateInput); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(model.ApplicationTemplateInput)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(graphql.ApplicationTemplateInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *ApplicationTemplateConverter) MultipleToGraphQL(in []*model.ApplicationTemplate) ([]*graphql.ApplicationTemplate, error) {
	ret := _m.Called(in)

	var r0 []*graphql.ApplicationTemplate
	if rf, ok := ret.Get(0).(func([]*model.ApplicationTemplate) []*graphql.ApplicationTemplate); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.ApplicationTemplate)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*model.ApplicationTemplate) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGraphQL provides a mock function with given fields: in
func (_m *ApplicationTemplateConverter) ToGraphQL(in *model.ApplicationTemplate) (*graphql.ApplicationTemplate, error) {
	ret := _m.Called(in)

	var r0 *graphql.ApplicationTemplate
	if rf, ok := ret.Get(0).(func(*model.ApplicationTemplate) *graphql.ApplicationTemplate); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.ApplicationTemplate)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.ApplicationTemplate) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateInputFromGraphQL provides a mock function with given fields: in
func (_m *ApplicationTemplateConverter) UpdateInputFromGraphQL(in graphql.ApplicationTemplateInput) (model.ApplicationTemplateInput, error) {
	ret := _m.Called(in)

	var r0 model.ApplicationTemplateInput
	if rf, ok := ret.Get(0).(func(graphql.ApplicationTemplateInput) model.ApplicationTemplateInput); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(model.ApplicationTemplateInput)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(graphql.ApplicationTemplateInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewApplicationTemplateConverter interface {
	mock.TestingT
	Cleanup(func())
}

// NewApplicationTemplateConverter creates a new instance of ApplicationTemplateConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewApplicationTemplateConverter(t mockConstructorTestingTNewApplicationTemplateConverter) *ApplicationTemplateConverter {
	mock := &ApplicationTemplateConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
