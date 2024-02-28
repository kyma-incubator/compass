// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// ApplicationService is an autogenerated mock type for the ApplicationService type
type ApplicationService struct {
	mock.Mock
}

// UpdateBaseURLAndReadyState provides a mock function with given fields: ctx, appID, baseURL, ready
func (_m *ApplicationService) UpdateBaseURLAndReadyState(ctx context.Context, appID string, baseURL string, ready bool) error {
	ret := _m.Called(ctx, appID, baseURL, ready)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, bool) error); ok {
		r0 = rf(ctx, appID, baseURL, ready)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewApplicationService creates a new instance of ApplicationService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewApplicationService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ApplicationService {
	mock := &ApplicationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
