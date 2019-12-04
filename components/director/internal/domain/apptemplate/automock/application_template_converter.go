// Code generated by mockery v1.0.0. DO NOT EDIT.

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
