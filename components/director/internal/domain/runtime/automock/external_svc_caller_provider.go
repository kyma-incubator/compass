// Code generated by mockery. DO NOT EDIT.

package automock

import (
	config "github.com/kyma-incubator/compass/components/director/pkg/config"
	mock "github.com/stretchr/testify/mock"

	runtime "github.com/kyma-incubator/compass/components/director/internal/domain/runtime"

	testing "testing"
)

// ExternalSvcCallerProvider is an autogenerated mock type for the ExternalSvcCallerProvider type
type ExternalSvcCallerProvider struct {
	mock.Mock
}

// GetCaller provides a mock function with given fields: _a0, _a1
func (_m *ExternalSvcCallerProvider) GetCaller(_a0 config.SelfRegConfig, _a1 string) (runtime.ExternalSvcCaller, error) {
	ret := _m.Called(_a0, _a1)

	var r0 runtime.ExternalSvcCaller
	if rf, ok := ret.Get(0).(func(config.SelfRegConfig, string) runtime.ExternalSvcCaller); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(runtime.ExternalSvcCaller)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(config.SelfRegConfig, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewExternalSvcCallerProvider creates a new instance of ExternalSvcCallerProvider. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewExternalSvcCallerProvider(t testing.TB) *ExternalSvcCallerProvider {
	mock := &ExternalSvcCallerProvider{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
