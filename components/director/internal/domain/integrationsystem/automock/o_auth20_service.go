// Code generated by mockery 2.9.0. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// OAuth20Service is an autogenerated mock type for the OAuth20Service type
type OAuth20Service struct {
	mock.Mock
}

// DeleteMultipleClientCredentials provides a mock function with given fields: ctx, auths
func (_m *OAuth20Service) DeleteMultipleClientCredentials(ctx context.Context, auths []model.SystemAuth) error {
	ret := _m.Called(ctx, auths)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []model.SystemAuth) error); ok {
		r0 = rf(ctx, auths)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
