// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// FormationTemplateConverter is an autogenerated mock type for the FormationTemplateConverter type
type FormationTemplateConverter struct {
	mock.Mock
}

// FromInputGraphQL provides a mock function with given fields: in
func (_m *FormationTemplateConverter) FromInputGraphQL(in *graphql.FormationTemplateInput) *model.FormationTemplateInput {
	ret := _m.Called(in)

	var r0 *model.FormationTemplateInput
	if rf, ok := ret.Get(0).(func(*graphql.FormationTemplateInput) *model.FormationTemplateInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationTemplateInput)
		}
	}

	return r0
}

// FromModelInputToModel provides a mock function with given fields: in, id
func (_m *FormationTemplateConverter) FromModelInputToModel(in *model.FormationTemplateInput, id string) *model.FormationTemplate {
	ret := _m.Called(in, id)

	var r0 *model.FormationTemplate
	if rf, ok := ret.Get(0).(func(*model.FormationTemplateInput, string) *model.FormationTemplate); ok {
		r0 = rf(in, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationTemplate)
		}
	}

	return r0
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *FormationTemplateConverter) MultipleToGraphQL(in []*model.FormationTemplate) []*graphql.FormationTemplate {
	ret := _m.Called(in)

	var r0 []*graphql.FormationTemplate
	if rf, ok := ret.Get(0).(func([]*model.FormationTemplate) []*graphql.FormationTemplate); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.FormationTemplate)
		}
	}

	return r0
}

// ToGraphQL provides a mock function with given fields: in
func (_m *FormationTemplateConverter) ToGraphQL(in *model.FormationTemplate) *graphql.FormationTemplate {
	ret := _m.Called(in)

	var r0 *graphql.FormationTemplate
	if rf, ok := ret.Get(0).(func(*model.FormationTemplate) *graphql.FormationTemplate); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.FormationTemplate)
		}
	}

	return r0
}

// NewFormationTemplateConverter creates a new instance of FormationTemplateConverter. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationTemplateConverter(t testing.TB) *FormationTemplateConverter {
	mock := &FormationTemplateConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
