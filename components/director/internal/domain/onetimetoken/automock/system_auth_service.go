// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	internalmodel "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/pkg/model"

	testing "testing"
)

// SystemAuthService is an autogenerated mock type for the SystemAuthService type
type SystemAuthService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, objectType, objectID, authInput
func (_m *SystemAuthService) Create(ctx context.Context, objectType model.SystemAuthReferenceObjectType, objectID string, authInput *internalmodel.AuthInput) (string, error) {
	ret := _m.Called(ctx, objectType, objectID, authInput)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, model.SystemAuthReferenceObjectType, string, *internalmodel.AuthInput) string); ok {
		r0 = rf(ctx, objectType, objectID, authInput)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.SystemAuthReferenceObjectType, string, *internalmodel.AuthInput) error); ok {
		r1 = rf(ctx, objectType, objectID, authInput)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByToken provides a mock function with given fields: ctx, token
func (_m *SystemAuthService) GetByToken(ctx context.Context, token string) (*model.SystemAuth, error) {
	ret := _m.Called(ctx, token)

	var r0 *model.SystemAuth
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.SystemAuth); ok {
		r0 = rf(ctx, token)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.SystemAuth)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, token)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGlobal provides a mock function with given fields: ctx, authID
func (_m *SystemAuthService) GetGlobal(ctx context.Context, authID string) (*model.SystemAuth, error) {
	ret := _m.Called(ctx, authID)

	var r0 *model.SystemAuth
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.SystemAuth); ok {
		r0 = rf(ctx, authID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.SystemAuth)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, authID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, item
func (_m *SystemAuthService) Update(ctx context.Context, item *model.SystemAuth) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.SystemAuth) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewSystemAuthService creates a new instance of SystemAuthService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewSystemAuthService(t testing.TB) *SystemAuthService {
	mock := &SystemAuthService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
