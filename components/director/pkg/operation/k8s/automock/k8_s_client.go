// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	testing "testing"

	v1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	v1alpha1 "github.com/kyma-incubator/compass/components/operations-controller/api/v1alpha1"
)

// K8SClient is an autogenerated mock type for the K8SClient type
type K8SClient struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, operation
func (_m *K8SClient) Create(ctx context.Context, operation *v1alpha1.Operation) (*v1alpha1.Operation, error) {
	ret := _m.Called(ctx, operation)

	var r0 *v1alpha1.Operation
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.Operation) *v1alpha1.Operation); ok {
		r0 = rf(ctx, operation)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.Operation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *v1alpha1.Operation) error); ok {
		r1 = rf(ctx, operation)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: ctx, name, options
func (_m *K8SClient) Get(ctx context.Context, name string, options v1.GetOptions) (*v1alpha1.Operation, error) {
	ret := _m.Called(ctx, name, options)

	var r0 *v1alpha1.Operation
	if rf, ok := ret.Get(0).(func(context.Context, string, v1.GetOptions) *v1alpha1.Operation); ok {
		r0 = rf(ctx, name, options)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.Operation)
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

// Update provides a mock function with given fields: ctx, operation
func (_m *K8SClient) Update(ctx context.Context, operation *v1alpha1.Operation) (*v1alpha1.Operation, error) {
	ret := _m.Called(ctx, operation)

	var r0 *v1alpha1.Operation
	if rf, ok := ret.Get(0).(func(context.Context, *v1alpha1.Operation) *v1alpha1.Operation); ok {
		r0 = rf(ctx, operation)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*v1alpha1.Operation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *v1alpha1.Operation) error); ok {
		r1 = rf(ctx, operation)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewK8SClient creates a new instance of K8SClient. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewK8SClient(t testing.TB) *K8SClient {
	mock := &K8SClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
