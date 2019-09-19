// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// SystemAuthConverter is an autogenerated mock type for the SystemAuthConverter type
type SystemAuthConverter struct {
	mock.Mock
}

// ToGraphQL provides a mock function with given fields: in
func (_m *SystemAuthConverter) ToGraphQL(in *model.SystemAuth) *graphql.SystemAuth {
	ret := _m.Called(in)

	var r0 *graphql.SystemAuth
	if rf, ok := ret.Get(0).(func(*model.SystemAuth) *graphql.SystemAuth); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.SystemAuth)
		}
	}

	return r0
}
