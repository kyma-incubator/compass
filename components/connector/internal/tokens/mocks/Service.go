// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	context "context"

	apperrors "github.com/kyma-incubator/compass/components/connector/internal/apperrors"

	mock "github.com/stretchr/testify/mock"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// GetToken provides a mock function with given fields: ctx, clientId
func (_m *Service) GetToken(ctx context.Context, clientId string) (string, apperrors.AppError) {
	ret := _m.Called(ctx, clientId)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string) string); ok {
		r0 = rf(ctx, clientId)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(context.Context, string) apperrors.AppError); ok {
		r1 = rf(ctx, clientId)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
