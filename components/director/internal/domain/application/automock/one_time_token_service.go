// Code generated by mockery v2.9.4. DO NOT EDIT.

package automock

import (
	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// OneTimeTokenService is an autogenerated mock type for the OneTimeTokenService type
type OneTimeTokenService struct {
	mock.Mock
}

// IsTokenValid provides a mock function with given fields: systemAuth
func (_m *OneTimeTokenService) IsTokenValid(systemAuth *model.SystemAuth) (bool, error) {
	ret := _m.Called(systemAuth)

	var r0 bool
	if rf, ok := ret.Get(0).(func(*model.SystemAuth) bool); ok {
		r0 = rf(systemAuth)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.SystemAuth) error); ok {
		r1 = rf(systemAuth)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
