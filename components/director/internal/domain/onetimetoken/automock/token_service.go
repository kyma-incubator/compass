// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	internalmodel "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/pkg/model"
)

// TokenService is an autogenerated mock type for the TokenService type
type TokenService struct {
	mock.Mock
}

// GenerateOneTimeToken provides a mock function with given fields: ctx, runtimeID, tokenType
func (_m *TokenService) GenerateOneTimeToken(ctx context.Context, runtimeID string, tokenType model.SystemAuthReferenceObjectType) (*internalmodel.OneTimeToken, error) {
	ret := _m.Called(ctx, runtimeID, tokenType)

	var r0 *internalmodel.OneTimeToken
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.SystemAuthReferenceObjectType) (*internalmodel.OneTimeToken, error)); ok {
		return rf(ctx, runtimeID, tokenType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, model.SystemAuthReferenceObjectType) *internalmodel.OneTimeToken); ok {
		r0 = rf(ctx, runtimeID, tokenType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internalmodel.OneTimeToken)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, model.SystemAuthReferenceObjectType) error); ok {
		r1 = rf(ctx, runtimeID, tokenType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RegenerateOneTimeToken provides a mock function with given fields: ctx, sysAuthID
func (_m *TokenService) RegenerateOneTimeToken(ctx context.Context, sysAuthID string) (*internalmodel.OneTimeToken, error) {
	ret := _m.Called(ctx, sysAuthID)

	var r0 *internalmodel.OneTimeToken
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*internalmodel.OneTimeToken, error)); ok {
		return rf(ctx, sysAuthID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *internalmodel.OneTimeToken); ok {
		r0 = rf(ctx, sysAuthID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*internalmodel.OneTimeToken)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, sysAuthID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewTokenService creates a new instance of TokenService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewTokenService(t interface {
	mock.TestingT
	Cleanup(func())
}) *TokenService {
	mock := &TokenService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
