// Code generated by mockery v2.12.1. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// DirectorClient is an autogenerated mock type for the directorClient type
type DirectorClient struct {
	mock.Mock
}

// DeleteSystemAsync provides a mock function with given fields: ctx, id, tenant
func (_m *DirectorClient) DeleteSystemAsync(ctx context.Context, id string, tenant string) error {
	ret := _m.Called(ctx, id, tenant)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, id, tenant)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewDirectorClient creates a new instance of DirectorClient. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewDirectorClient(t testing.TB) *DirectorClient {
	mock := &DirectorClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
