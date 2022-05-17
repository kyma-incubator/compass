// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// BundleConverter is an autogenerated mock type for the BundleConverter type
type BundleConverter struct {
	mock.Mock
}

// MultipleCreateInputFromGraphQL provides a mock function with given fields: in
func (_m *BundleConverter) MultipleCreateInputFromGraphQL(in []*graphql.BundleCreateInput) ([]*model.BundleCreateInput, error) {
	ret := _m.Called(in)

	var r0 []*model.BundleCreateInput
	if rf, ok := ret.Get(0).(func([]*graphql.BundleCreateInput) []*model.BundleCreateInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.BundleCreateInput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*graphql.BundleCreateInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *BundleConverter) MultipleToGraphQL(in []*model.Bundle) ([]*graphql.Bundle, error) {
	ret := _m.Called(in)

	var r0 []*graphql.Bundle
	if rf, ok := ret.Get(0).(func([]*model.Bundle) []*graphql.Bundle); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.Bundle)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*model.Bundle) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
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

// NewBundleConverter creates a new instance of BundleConverter. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewBundleConverter(t testing.TB) *BundleConverter {
	mock := &BundleConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
