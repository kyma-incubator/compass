// Code generated by mockery. DO NOT EDIT.

package automock

import (
	testing "testing"

	mock "github.com/stretchr/testify/mock"
)

// HealthCheckConverter is an autogenerated mock type for the HealthCheckConverter type
type HealthCheckConverter struct {
	mock.Mock
}

// NewHealthCheckConverter creates a new instance of HealthCheckConverter. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewHealthCheckConverter(t testing.TB) *HealthCheckConverter {
	mock := &HealthCheckConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
