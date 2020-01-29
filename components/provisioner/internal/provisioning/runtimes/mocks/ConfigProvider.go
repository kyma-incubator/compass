// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import runtimes "github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/runtimes"
import v1 "k8s.io/api/core/v1"

// ConfigProvider is an autogenerated mock type for the ConfigProvider type
type ConfigProvider struct {
	mock.Mock
}

// CreateConfigMapForRuntime provides a mock function with given fields: runtimeConfig, kubeconfigRaw
func (_m *ConfigProvider) CreateConfigMapForRuntime(runtimeConfig runtimes.RuntimeConfig, kubeconfigRaw string) (*v1.ConfigMap, error) {
	ret := _m.Called(runtimeConfig, kubeconfigRaw)

	var r0 *v1.ConfigMap
	if rf, ok := ret.Get(0).(func(runtimes.RuntimeConfig, string) *v1.ConfigMap); ok {
		r0 = rf(runtimeConfig, kubeconfigRaw)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(runtimes.RuntimeConfig, string) error); ok {
		r1 = rf(runtimeConfig, kubeconfigRaw)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
