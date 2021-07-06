// Code generated by mockery 2.9.0. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// BundleConverter is an autogenerated mock type for the BundleConverter type
type BundleConverter struct {
	mock.Mock
}

// ToGraphQL provides a mock function with given fields: in
func (_m *BundleConverter) ToGraphQL(in *model.Bundle) (*graphql.Bundle, error) {
	ret := _m.Called(in)

	var r0 *graphql.Bundle
	if rf, ok := ret.Get(0).(func(*model.Bundle) *graphql.Bundle); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.Bundle)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.Bundle) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
