// Code generated by mockery 2.9.0. DO NOT EDIT.

package mocks

import (
	mock "github.com/stretchr/testify/mock"
	corev1 "k8s.io/api/core/v1"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	watch "k8s.io/apimachinery/pkg/watch"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// Get provides a mock function with given fields: name, options
func (_m *Manager) Get(name string, options v1.GetOptions) (*corev1.ConfigMap, error) {
	ret := _m.Called(name, options)

	var r0 *corev1.ConfigMap
	if rf, ok := ret.Get(0).(func(string, v1.GetOptions) *corev1.ConfigMap); ok {
		r0 = rf(name, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, v1.GetOptions) error); ok {
		r1 = rf(name, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: configmap
func (_m *Manager) Update(configmap *corev1.ConfigMap) (*corev1.ConfigMap, error) {
	ret := _m.Called(configmap)

	var r0 *corev1.ConfigMap
	if rf, ok := ret.Get(0).(func(*corev1.ConfigMap) *corev1.ConfigMap); ok {
		r0 = rf(configmap)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*corev1.ConfigMap) error); ok {
		r1 = rf(configmap)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Watch provides a mock function with given fields: opts
func (_m *Manager) Watch(opts v1.ListOptions) (watch.Interface, error) {
	ret := _m.Called(opts)

	var r0 watch.Interface
	if rf, ok := ret.Get(0).(func(v1.ListOptions) watch.Interface); ok {
		r0 = rf(opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(watch.Interface)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(v1.ListOptions) error); ok {
		r1 = rf(opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
