// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	dberrors "github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/provisioner/internal/model"

	time "time"
)

// WriteSession is an autogenerated mock type for the WriteSession type
type WriteSession struct {
	mock.Mock
}

// DeleteCluster provides a mock function with given fields: runtimeID
func (_m *WriteSession) DeleteCluster(runtimeID string) dberrors.Error {
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
func (_m *WriteSession) InsertCluster(cluster model.Cluster) dberrors.Error {
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
func (_m *WriteSession) InsertGCPConfig(config model.GCPConfig) dberrors.Error {
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
func (_m *WriteSession) InsertGardenerConfig(config model.GardenerConfig) dberrors.Error {
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
func (_m *WriteSession) InsertKymaConfig(kymaConfig model.KymaConfig) dberrors.Error {
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
func (_m *WriteSession) InsertOperation(operation model.Operation) dberrors.Error {
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

// InsertRuntimeUpgrade provides a mock function with given fields: runtimeUpgrade
func (_m *WriteSession) InsertRuntimeUpgrade(runtimeUpgrade model.RuntimeUpgrade) dberrors.Error {
	ret := _m.Called(runtimeUpgrade)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(model.RuntimeUpgrade) dberrors.Error); ok {
		r0 = rf(runtimeUpgrade)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}

// MarkClusterAsDeleted provides a mock function with given fields: runtimeID
func (_m *WriteSession) MarkClusterAsDeleted(runtimeID string) dberrors.Error {
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

// SetActiveKymaConfig provides a mock function with given fields: runtimeID, kymaConfigId
func (_m *WriteSession) SetActiveKymaConfig(runtimeID string, kymaConfigId string) dberrors.Error {
	ret := _m.Called(runtimeID, kymaConfigId)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(string, string) dberrors.Error); ok {
		r0 = rf(runtimeID, kymaConfigId)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}

// TransitionOperation provides a mock function with given fields: operationID, message, stage, transitionTime
func (_m *WriteSession) TransitionOperation(operationID string, message string, stage model.OperationStage, transitionTime time.Time) dberrors.Error {
	ret := _m.Called(operationID, message, stage, transitionTime)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(string, string, model.OperationStage, time.Time) dberrors.Error); ok {
		r0 = rf(operationID, message, stage, transitionTime)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}

// UpdateCluster provides a mock function with given fields: runtimeID, kubeconfig, terraformState
func (_m *WriteSession) UpdateCluster(runtimeID string, kubeconfig string, terraformState []byte) dberrors.Error {
	ret := _m.Called(runtimeID, kubeconfig, terraformState)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(string, string, []byte) dberrors.Error); ok {
		r0 = rf(runtimeID, kubeconfig, terraformState)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}

// UpdateOperationState provides a mock function with given fields: operationID, message, state, endTime
func (_m *WriteSession) UpdateOperationState(operationID string, message string, state model.OperationState, endTime time.Time) dberrors.Error {
	ret := _m.Called(operationID, message, state, endTime)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(string, string, model.OperationState, time.Time) dberrors.Error); ok {
		r0 = rf(operationID, message, state, endTime)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}

// UpdateUpgradeState provides a mock function with given fields: operationID, upgradeState
func (_m *WriteSession) UpdateUpgradeState(operationID string, upgradeState model.UpgradeState) dberrors.Error {
	ret := _m.Called(operationID, upgradeState)

	var r0 dberrors.Error
	if rf, ok := ret.Get(0).(func(string, model.UpgradeState) dberrors.Error); ok {
		r0 = rf(operationID, upgradeState)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(dberrors.Error)
		}
	}

	return r0
}
