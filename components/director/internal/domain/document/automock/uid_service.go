// Code generated by mockery v2.10.4. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

// UIDService is an autogenerated mock type for the UIDService type
type UIDService struct {
	mock.Mock
}

// Generate provides a mock function with given fields:
func (_m *UIDService) Generate() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
