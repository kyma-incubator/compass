// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"
import proxy "github.com/kyma-incubator/compass/components/gateway/pkg/proxy"

// AuditlogService is an autogenerated mock type for the AuditlogService type
type AuditlogService struct {
	mock.Mock
}

// Log provides a mock function with given fields: request, resposne, claims
func (_m *AuditlogService) Log(request string, resposne string, claims proxy.Claims) error {
	ret := _m.Called(request, resposne, claims)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string, proxy.Claims) error); ok {
		r0 = rf(request, resposne, claims)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
