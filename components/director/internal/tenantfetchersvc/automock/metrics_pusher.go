// Code generated by mockery. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

// MetricsPusher is an autogenerated mock type for the MetricsPusher type
type MetricsPusher struct {
	mock.Mock
}

// Push provides a mock function with given fields:
func (_m *MetricsPusher) Push() {
	_m.Called()
}

// RecordEventingRequest provides a mock function with given fields: method, statusCode, desc
func (_m *MetricsPusher) RecordEventingRequest(method string, statusCode int, desc string) {
	_m.Called(method, statusCode, desc)
}

// RecordTenantsSyncJobFailure provides a mock function with given fields: method, statusCode, desc
func (_m *MetricsPusher) RecordTenantsSyncJobFailure(method string, statusCode int, desc string) {
	_m.Called(method, statusCode, desc)
}
