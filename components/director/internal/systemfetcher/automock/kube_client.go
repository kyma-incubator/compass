// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// KubeClient is an autogenerated mock type for the KubeClient type
type KubeClient struct {
	mock.Mock
}

// GetSystemFetcherSecretData provides a mock function with given fields: ctx, secretName
func (_m *KubeClient) GetSystemFetcherSecretData(ctx context.Context, secretName string) ([]byte, error) {
	ret := _m.Called(ctx, secretName)

	var r0 []byte
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]byte, error)); ok {
		return rf(ctx, secretName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []byte); ok {
		r0 = rf(ctx, secretName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]byte)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, secretName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewKubeClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewKubeClient creates a new instance of KubeClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewKubeClient(t mockConstructorTestingTNewKubeClient) *KubeClient {
	mock := &KubeClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
