// Code generated by mockery v2.5.1. DO NOT EDIT.

package automock

import (
	mock "github.com/stretchr/testify/mock"

	time "time"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// Now provides a mock function with given fields:
func (_m *Service) Now() time.Time {
	ret := _m.Called()

	var r0 time.Time
	if rf, ok := ret.Get(0).(func() time.Time); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(time.Time)
	}

	return r0
}
