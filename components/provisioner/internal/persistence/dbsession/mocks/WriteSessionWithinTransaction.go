// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import dberrors "github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"

import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/provisioner/internal/model"

// WriteSessionWithinTransaction is an autogenerated mock type for the WriteSessionWithinTransaction type
type WriteSessionWithinTransaction struct {
	mock.Mock
}

// Commit provides a mock function with given fields:
func (_m *WriteSessionWithinTransaction) Commit() dberrors.Error {
	ret := _m.Called()

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func() dberrors.Error); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}

// DeleteCluster provides a mock function with given fields: runtimeID
func (_m *WriteSessionWithinTransaction) DeleteCluster(runtimeID string) dberrors.Error {
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

// InsertCluster provides a mock function with given fields: cluster
func (_m *WriteSessionWithinTransaction) InsertCluster(cluster model.Cluster) dberrors.Error {
	ret := _m.Called(cluster)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(model.Cluster) dberrors.Error); ok {
		r0 = rf(cluster)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}

// InsertGCPConfig provides a mock function with given fields: config
func (_m *WriteSessionWithinTransaction) InsertGCPConfig(config model.GCPConfig) dberrors.Error {
	ret := _m.Called(config)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(model.GCPConfig) dberrors.Error); ok {
		r0 = rf(config)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}

// InsertGardenerConfig provides a mock function with given fields: config
func (_m *WriteSessionWithinTransaction) InsertGardenerConfig(config model.GardenerConfig) dberrors.Error {
	ret := _m.Called(config)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(model.GardenerConfig) dberrors.Error); ok {
		r0 = rf(config)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}

// InsertKymaConfig provides a mock function with given fields: kymaConfig
func (_m *WriteSessionWithinTransaction) InsertKymaConfig(kymaConfig model.KymaConfig) dberrors.Error {
	ret := _m.Called(kymaConfig)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(model.KymaConfig) dberrors.Error); ok {
		r0 = rf(kymaConfig)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}

// InsertOperation provides a mock function with given fields: operation
func (_m *WriteSessionWithinTransaction) InsertOperation(operation model.Operation) dberrors.Error {
	ret := _m.Called(operation)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(model.Operation) dberrors.Error); ok {
		r0 = rf(operation)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}

// RollbackUnlessCommitted provides a mock function with given fields:
func (_m *WriteSessionWithinTransaction) RollbackUnlessCommitted() {
	_m.Called()
}

// UpdateCluster provides a mock function with given fields: runtimeID, kubeconfig, terraformState
func (_m *WriteSessionWithinTransaction) UpdateCluster(runtimeID string, kubeconfig string, terraformState string) dberrors.Error {
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

// UpdateOperationState provides a mock function with given fields: operationID, message, state
func (_m *WriteSessionWithinTransaction) UpdateOperationState(operationID string, message string, state model.OperationState) dberrors.Error {
	ret := _m.Called(operationID, message, state)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(string, string, model.OperationState) dberrors.Error); ok {
		r0 = rf(operationID, message, state)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}
