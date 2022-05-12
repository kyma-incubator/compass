// Code generated by mockery v2.12.1. DO NOT EDIT.

package automock

import (
	context "context"
	testing "testing"

	systemfetcher "github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	mock "github.com/stretchr/testify/mock"
)

// SystemsAPIClient is an autogenerated mock type for the systemsAPIClient type
type SystemsAPIClient struct {
	mock.Mock
}

// FetchSystemsForTenant provides a mock function with given fields: ctx, tenant
func (_m *SystemsAPIClient) FetchSystemsForTenant(ctx context.Context, tenant string) ([]systemfetcher.System, error) {
	ret := _m.Called(ctx, tenant)

	var r0 []systemfetcher.System
	if rf, ok := ret.Get(0).(func(context.Context, string) []systemfetcher.System); ok {
		r0 = rf(ctx, tenant)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]systemfetcher.System)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewSystemsAPIClient creates a new instance of SystemsAPIClient. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewSystemsAPIClient(t testing.TB) *SystemsAPIClient {
	mock := &SystemsAPIClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
