// Code generated by mockery v1.1.2. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// FetchRequestConverter is an autogenerated mock type for the FetchRequestConverter type
type FetchRequestConverter struct {
	mock.Mock
}

// ToGraphQL provides a mock function with given fields: in
func (_m *FetchRequestConverter) ToGraphQL(in *model.FetchRequest) (*graphql.FetchRequest, error) {
	ret := _m.Called(in)

	var r0 *graphql.FetchRequest
	if rf, ok := ret.Get(0).(func(*model.FetchRequest) *graphql.FetchRequest); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.FetchRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.FetchRequest) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
