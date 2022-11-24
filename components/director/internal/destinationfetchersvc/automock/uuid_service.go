// Code generated by mockery. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

// UUIDService is an autogenerated mock type for the UUIDService type
type UUIDService struct {
	mock.Mock
}

// Generate provides a mock function with given fields:
func (_m *UUIDService) Generate() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

type mockConstructorTestingTNewUUIDService interface {
	mock.TestingT
	Cleanup(func())
}

// NewUUIDService creates a new instance of UUIDService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUUIDService(t mockConstructorTestingTNewUUIDService) *UUIDService {
	mock := &UUIDService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
