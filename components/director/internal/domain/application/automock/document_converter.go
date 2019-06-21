// Code generated by mockery v1.0.0. DO NOT EDIT.
package automock

import graphql "github.com/kyma-incubator/compass/components/director/internal/graphql"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// DocumentConverter is an autogenerated mock type for the DocumentConverter type
type DocumentConverter struct {
	mock.Mock
}

// MultipleInputFromGraphQL provides a mock function with given fields: in
func (_m *DocumentConverter) MultipleInputFromGraphQL(in []*graphql.DocumentInput) []*model.DocumentInput {
	ret := _m.Called(in)

	var r0 []*model.DocumentInput
	if rf, ok := ret.Get(0).(func([]*graphql.DocumentInput) []*model.DocumentInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.DocumentInput)
		}
	}

	return r0
}
