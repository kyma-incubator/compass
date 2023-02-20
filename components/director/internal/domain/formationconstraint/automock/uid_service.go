// Code generated by mockery. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

// UidService is an autogenerated mock type for the uidService type
type UidService struct {
	mock.Mock
}

// Generate provides a mock function with given fields:
func (_m *UidService) Generate() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

type NewUidServiceT interface {
	mock.TestingT
	Cleanup(func())
}

// NewUidService creates a new instance of UidService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewUidService(t NewUidServiceT) *UidService {
	mock := &UidService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
