// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// PackageInstanceAuthConverter is an autogenerated mock type for the PackageInstanceAuthConverter type
type PackageInstanceAuthConverter struct {
	mock.Mock
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *PackageInstanceAuthConverter) MultipleToGraphQL(in []*model.PackageInstanceAuth) ([]*graphql.PackageInstanceAuth, error) {
	ret := _m.Called(in)

	var r0 []*graphql.PackageInstanceAuth
	if rf, ok := ret.Get(0).(func([]*model.PackageInstanceAuth) []*graphql.PackageInstanceAuth); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.PackageInstanceAuth)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*model.PackageInstanceAuth) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGraphQL provides a mock function with given fields: in
func (_m *PackageInstanceAuthConverter) ToGraphQL(in *model.PackageInstanceAuth) (*graphql.PackageInstanceAuth, error) {
	ret := _m.Called(in)

	var r0 *graphql.PackageInstanceAuth
	if rf, ok := ret.Get(0).(func(*model.PackageInstanceAuth) *graphql.PackageInstanceAuth); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.PackageInstanceAuth)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.PackageInstanceAuth) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
