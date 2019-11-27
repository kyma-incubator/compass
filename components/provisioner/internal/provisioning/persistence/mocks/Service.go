// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	dberrors "github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/provisioner/internal/model"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// CleanupClusterData provides a mock function with given fields: runtimeID
func (_m *Service) CleanupClusterData(runtimeID string) dberrors.Error {
	ret := _m.Called(runtimeID)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(string) dberrors.Error); ok {
		r0 = rf(runtimeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}

// GetOperation provides a mock function with given fields: operationID
func (_m *Service) GetOperation(operationID string) (model.Operation, error) {
	ret := _m.Called(operationID)

	var r0 model.Operation
	if rf, ok := ret.Get(0).(func(string) model.Operation); ok {
		r0 = rf(operationID)
	} else {
		r0 = ret.Get(0).(model.Operation)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(operationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetClusterData provides a mock function with given fields: runtimeID
func (_m *Service) GetClusterData(runtimeID string) (model.Cluster, dberrors.Error) {
	ret := _m.Called(runtimeID)

	var r0 model.Cluster
	if rf, ok := ret.Get(0).(func(string) model.Cluster); ok {
		r0 = rf(runtimeID)
	} else {
		r0 = ret.Get(0).(model.Cluster)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(runtimeID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// GetLastOperation provides a mock function with given fields: runtimeID
func (_m *Service) GetLastOperation(runtimeID string) (model.Operation, dberrors.Error) {
	ret := _m.Called(runtimeID)

	var r0 model.Operation
	if rf, ok := ret.Get(0).(func(string) model.Operation); ok {
		r0 = rf(runtimeID)
	} else {
		r0 = ret.Get(0).(model.Operation)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(runtimeID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// GetRuntimeStatus provides a mock function with given fields: runtimeID
func (_m *Service) GetStatus(runtimeID string) (model.RuntimeStatus, dberrors.Error) {
	ret := _m.Called(runtimeID)

	var r0 model.RuntimeStatus
	if rf, ok := ret.Get(0).(func(string) model.RuntimeStatus); ok {
		r0 = rf(runtimeID)
	} else {
		r0 = ret.Get(0).(model.RuntimeStatus)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(runtimeID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// SetOperationAsFailed provides a mock function with given fields: operationID, message
func (_m *Service) SetAsFailed(operationID string, message string) error {
	ret := _m.Called(operationID, message)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(operationID, message)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetOperationAsSucceeded provides a mock function with given fields: operationID
func (_m *Service) SetAsSucceeded(operationID string) error {
	ret := _m.Called(operationID)

	var r0 error
	if rf, ok := ret.Get(0).(func(string) error); ok {
		r0 = rf(operationID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// SetDeprovisioningStarted provides a mock function with given fields: runtimeID
func (_m *Service) SetDeprovisioningStarted(runtimeID string) (model.Operation, dberrors.Error) {
	ret := _m.Called(runtimeID)

	var r0 model.Operation
	if rf, ok := ret.Get(0).(func(string) model.Operation); ok {
		r0 = rf(runtimeID)
	} else {
		r0 = ret.Get(0).(model.Operation)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(runtimeID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// SetProvisioningStarted provides a mock function with given fields: runtimeID, runtimeConfig
func (_m *Service) SetProvisioningStarted(runtimeID string, runtimeConfig model.RuntimeConfig) (model.Operation, dberrors.Error) {
	ret := _m.Called(runtimeID, runtimeConfig)

	var r0 model.Operation
	if rf, ok := ret.Get(0).(func(string, model.RuntimeConfig) model.Operation); ok {
		r0 = rf(runtimeID, runtimeConfig)
	} else {
		r0 = ret.Get(0).(model.Operation)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string, model.RuntimeConfig) dberrors.Error); ok {
		r1 = rf(runtimeID, runtimeConfig)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// SetUpgradeStarted provides a mock function with given fields: runtimeID
func (_m *Service) SetUpgradeStarted(runtimeID string) (model.Operation, dberrors.Error) {
	ret := _m.Called(runtimeID)

	var r0 model.Operation
	if rf, ok := ret.Get(0).(func(string) model.Operation); ok {
		r0 = rf(runtimeID)
	} else {
		r0 = ret.Get(0).(model.Operation)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(runtimeID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}

// UpdateClusterData provides a mock function with given fields: runtimeID, kubeconfig, terraformState
func (_m *Service) UpdateClusterData(runtimeID string, kubeconfig string, terraformState string) dberrors.Error {
	ret := _m.Called(runtimeID, kubeconfig, terraformState)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(string, string, string) dberrors.Error); ok {
		r0 = rf(runtimeID, kubeconfig, terraformState)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}
