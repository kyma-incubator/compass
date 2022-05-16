// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// Converter is an autogenerated mock type for the Converter type
type Converter struct {
	mock.Mock
}

// RequestInputFromGraphQL provides a mock function with given fields: in
func (_m *Converter) RequestInputFromGraphQL(in graphql.BundleInstanceAuthRequestInput) model.BundleInstanceAuthRequestInput {
	ret := _m.Called(in)

	var r0 model.BundleInstanceAuthRequestInput
	if rf, ok := ret.Get(0).(func(graphql.BundleInstanceAuthRequestInput) model.BundleInstanceAuthRequestInput); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(model.BundleInstanceAuthRequestInput)
	}

	return r0
}

// SetInputFromGraphQL provides a mock function with given fields: in
func (_m *Converter) SetInputFromGraphQL(in graphql.BundleInstanceAuthSetInput) (model.BundleInstanceAuthSetInput, error) {
	ret := _m.Called(in)

	var r0 model.BundleInstanceAuthSetInput
	if rf, ok := ret.Get(0).(func(graphql.BundleInstanceAuthSetInput) model.BundleInstanceAuthSetInput); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(model.BundleInstanceAuthSetInput)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(graphql.BundleInstanceAuthSetInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGraphQL provides a mock function with given fields: in
func (_m *Converter) ToGraphQL(in *model.BundleInstanceAuth) (*graphql.BundleInstanceAuth, error) {
	ret := _m.Called(in)

	var r0 *graphql.BundleInstanceAuth
	if rf, ok := ret.Get(0).(func(*model.BundleInstanceAuth) *graphql.BundleInstanceAuth); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.BundleInstanceAuth)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.BundleInstanceAuth) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewConverter creates a new instance of Converter. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewConverter(t testing.TB) *Converter {
	mock := &Converter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
