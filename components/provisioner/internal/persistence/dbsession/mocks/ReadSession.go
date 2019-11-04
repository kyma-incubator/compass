// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import dberrors "github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"

import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/provisioner/internal/model"

// ReadSession is an autogenerated mock type for the ReadSession type
type ReadSession struct {
	mock.Mock
}

// GetCluster provides a mock function with given fields: runtimeID
func (_m *ReadSession) GetCluster(runtimeID string) (model.Cluster, dberrors.Error) {
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

// GetClusterConfig provides a mock function with given fields: runtimeID
func (_m *ReadSession) GetClusterConfig(runtimeID string) (interface{}, dberrors.Error) {
	ret := _m.Called(runtimeID)

	var r0 interface{}
	if rf, ok := ret.Get(0).(func(string) interface{}); ok {
		r0 = rf(runtimeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(interface{})
		}
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

// GetKymaConfig provides a mock function with given fields: runtimeID
func (_m *ReadSession) GetKymaConfig(runtimeID string) (model.KymaConfig, dberrors.Error) {
	ret := _m.Called(runtimeID)

	var r0 model.KymaConfig
	if rf, ok := ret.Get(0).(func(string) model.KymaConfig); ok {
		r0 = rf(runtimeID)
	} else {
		r0 = ret.Get(0).(model.KymaConfig)
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
func (_m *ReadSession) GetLastOperation(runtimeID string) (model.Operation, dberrors.Error) {
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

// GetOperation provides a mock function with given fields: operationID
func (_m *ReadSession) GetOperation(operationID string) (model.Operation, dberrors.Error) {
	ret := _m.Called(operationID)

	var r0 model.Operation
	if rf, ok := ret.Get(0).(func(string) model.Operation); ok {
		r0 = rf(operationID)
	} else {
		r0 = ret.Get(0).(model.Operation)
	}

	var r1 dberrors.Error
	if rf, ok := ret.Get(1).(func(string) dberrors.Error); ok {
		r1 = rf(operationID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(dberrors.Error)
		}
	}

	return r0, r1
}
