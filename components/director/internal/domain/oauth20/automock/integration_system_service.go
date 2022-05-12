// Code generated by mockery v2.12.1. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// IntegrationSystemService is an autogenerated mock type for the IntegrationSystemService type
type IntegrationSystemService struct {
	mock.Mock
}

// Exists provides a mock function with given fields: ctx, id
func (_m *IntegrationSystemService) Exists(ctx context.Context, id string) (bool, error) {
	ret := _m.Called(ctx, id)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewIntegrationSystemService creates a new instance of IntegrationSystemService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewIntegrationSystemService(t testing.TB) *IntegrationSystemService {
	mock := &IntegrationSystemService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
