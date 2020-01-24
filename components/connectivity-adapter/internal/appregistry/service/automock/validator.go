// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import apperrors "github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"

// Validator is an autogenerated mock type for the Validator type
type Validator struct {
	mock.Mock
}

// Validate provides a mock function with given fields: details
func (_m *Validator) Validate(details model.ServiceDetails) apperrors.AppError {
	ret := _m.Called(details)

	var r0 apperrors.AppError
	if rf, ok := ret.Get(0).(func(model.ServiceDetails) apperrors.AppError); ok {
		r0 = rf(details)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(apperrors.AppError)
		}
	}

	return r0
}
