// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import "github.com/kyma-incubator/compass/components/director/pkg/graphql"
import "github.com/stretchr/testify/mock"
import "github.com/kyma-incubator/compass/components/director/internal/model"

// RuntimeConverter is an autogenerated mock type for the RuntimeConverter type
type RuntimeConverter struct {
	mock.Mock
}

// InputFromGraphQL provides a mock function with given fields: in
func (_m *RuntimeConverter) InputFromGraphQL(in graphql.RuntimeInput) model.RuntimeInput {
	ret := _m.Called(in)

	var r0 model.RuntimeInput
	if rf, ok := ret.Get(0).(func(graphql.RuntimeInput) model.RuntimeInput); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(model.RuntimeInput)
	}

	return r0
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *RuntimeConverter) MultipleToGraphQL(in []*model.Runtime) []*graphql.Runtime {
	ret := _m.Called(in)

	var r0 []*graphql.Runtime
	if rf, ok := ret.Get(0).(func([]*model.Runtime) []*graphql.Runtime); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.Runtime)
		}
	}

	return r0
}

// ToGraphQL provides a mock function with given fields: in
func (_m *RuntimeConverter) ToGraphQL(in *model.Runtime) *graphql.Runtime {
	ret := _m.Called(in)

	var r0 *graphql.Runtime
	if rf, ok := ret.Get(0).(func(*model.Runtime) *graphql.Runtime); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.Runtime)
		}
	}

	return r0
}
