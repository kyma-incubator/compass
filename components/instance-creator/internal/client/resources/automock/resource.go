// Code generated by mockery. DO NOT EDIT.

package automock

import (
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// Resource is an autogenerated mock type for the Resource type
type Resource struct {
	mock.Mock
}

// GetResourceID provides a mock function with given fields:
func (_m *Resource) GetResourceID() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetResourceType provides a mock function with given fields:
func (_m *Resource) GetResourceType() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// GetResourceURLPath provides a mock function with given fields:
func (_m *Resource) GetResourceURLPath() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// NewResource creates a new instance of Resource. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewResource(t testing.TB) *Resource {
	mock := &Resource{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
