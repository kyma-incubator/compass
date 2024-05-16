// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/pkg/model"
	mock "github.com/stretchr/testify/mock"
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

// NewOAuth20Service creates a new instance of OAuth20Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewOAuth20Service(t interface {
	mock.TestingT
	Cleanup(func())
}) *OAuth20Service {
	mock := &OAuth20Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
