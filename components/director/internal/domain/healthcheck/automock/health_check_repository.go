// Code generated by mockery v2.12.1. DO NOT EDIT.

package automock

import (
	testing "testing"

	mock "github.com/stretchr/testify/mock"
)

// HealthCheckRepository is an autogenerated mock type for the HealthCheckRepository type
type HealthCheckRepository struct {
	mock.Mock
}

// NewHealthCheckRepository creates a new instance of HealthCheckRepository. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewHealthCheckRepository(t testing.TB) *HealthCheckRepository {
	mock := &HealthCheckRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
