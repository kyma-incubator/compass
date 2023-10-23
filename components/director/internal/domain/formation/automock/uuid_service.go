// Code generated by mockery. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

// UuidService is an autogenerated mock type for the uuidService type
type UuidService struct {
	mock.Mock
}

// Generate provides a mock function with given fields:
func (_m *UuidService) Generate() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// NewUuidService creates a new instance of UuidService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewUuidService(t interface {
	mock.TestingT
	Cleanup(func())
}) *UuidService {
	mock := &UuidService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
