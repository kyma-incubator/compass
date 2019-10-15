// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/provisioner/internal/model"
import types "github.com/kyma-incubator/hydroform/types"

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// DeprovisionCluster provides a mock function with given fields: runtimeConfig, secretName
func (_m *Client) DeprovisionCluster(runtimeConfig model.RuntimeConfig, secretName string) error {
	ret := _m.Called(runtimeConfig, secretName)

	var r0 error
	if rf, ok := ret.Get(0).(func(model.RuntimeConfig, string) error); ok {
		r0 = rf(runtimeConfig, secretName)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ProvisionCluster provides a mock function with given fields: runtimeConfig, secretName
func (_m *Client) ProvisionCluster(runtimeConfig model.RuntimeConfig, secretName string) (types.ClusterStatus, string, error) {
	ret := _m.Called(runtimeConfig, secretName)

	var r0 types.ClusterStatus
	if rf, ok := ret.Get(0).(func(model.RuntimeConfig, string) types.ClusterStatus); ok {
		r0 = rf(runtimeConfig, secretName)
	} else {
		r0 = ret.Get(0).(types.ClusterStatus)
	}

	var r1 string
	if rf, ok := ret.Get(1).(func(model.RuntimeConfig, string) string); ok {
		r1 = rf(runtimeConfig, secretName)
	} else {
		r1 = ret.Get(1).(string)
	}

	var r2 error
	if rf, ok := ret.Get(2).(func(model.RuntimeConfig, string) error); ok {
		r2 = rf(runtimeConfig, secretName)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}
