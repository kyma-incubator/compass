// Code generated by mockery. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

// HealthCheckService is an autogenerated mock type for the HealthCheckService type
type HealthCheckService struct {
	mock.Mock
}

// NewHealthCheckService creates a new instance of HealthCheckService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewHealthCheckService(t interface {
	mock.TestingT
	Cleanup(func())
}) *HealthCheckService {
	mock := &HealthCheckService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
