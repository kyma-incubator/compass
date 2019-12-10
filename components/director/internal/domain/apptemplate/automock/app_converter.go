// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
import mock "github.com/stretchr/testify/mock"

// AppConverter is an autogenerated mock type for the AppConverter type
type AppConverter struct {
	mock.Mock
}

// CreateInputGQLToJSON provides a mock function with given fields: in
func (_m *AppConverter) CreateInputGQLToJSON(in *graphql.ApplicationCreateInput) (string, error) {
	ret := _m.Called(in)

	var r0 string
	if rf, ok := ret.Get(0).(func(*graphql.ApplicationCreateInput) string); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*graphql.ApplicationCreateInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
