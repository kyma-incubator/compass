// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	context "context"

	corev1 "k8s.io/api/core/v1"

	mock "github.com/stretchr/testify/mock"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	watch "k8s.io/apimachinery/pkg/watch"
)

// Manager is an autogenerated mock type for the Manager type
type Manager struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, name, options
func (_m *Manager) Get(ctx context.Context, name string, options v1.GetOptions) (*corev1.ConfigMap, error) {
	ret := _m.Called(ctx, name, options)

	var r0 *corev1.ConfigMap
	if rf, ok := ret.Get(0).(func(context.Context, string, v1.GetOptions) *corev1.ConfigMap); ok {
		r0 = rf(ctx, name, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, v1.GetOptions) error); ok {
		r1 = rf(ctx, name, options)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, configMap, opts
func (_m *Manager) Update(ctx context.Context, configMap *corev1.ConfigMap, opts v1.UpdateOptions) (*corev1.ConfigMap, error) {
	ret := _m.Called(ctx, configMap, opts)

	var r0 *corev1.ConfigMap
	if rf, ok := ret.Get(0).(func(context.Context, *corev1.ConfigMap, v1.UpdateOptions) *corev1.ConfigMap); ok {
		r0 = rf(ctx, configMap, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*corev1.ConfigMap)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *corev1.ConfigMap, v1.UpdateOptions) error); ok {
		r1 = rf(ctx, configMap, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Watch provides a mock function with given fields: ctx, opts
func (_m *Manager) Watch(ctx context.Context, opts v1.ListOptions) (watch.Interface, error) {
	ret := _m.Called(ctx, opts)

	var r0 watch.Interface
	if rf, ok := ret.Get(0).(func(context.Context, v1.ListOptions) watch.Interface); ok {
		r0 = rf(ctx, opts)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(watch.Interface)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, v1.ListOptions) error); ok {
		r1 = rf(ctx, opts)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
