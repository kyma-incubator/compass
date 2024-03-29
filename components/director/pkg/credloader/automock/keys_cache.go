// Code generated by mockery. DO NOT EDIT.

package automock

import (
	credloader "github.com/kyma-incubator/compass/components/director/pkg/credloader"
	mock "github.com/stretchr/testify/mock"
)

// KeysCache is an autogenerated mock type for the KeysCache type
type KeysCache struct {
	mock.Mock
}

// Get provides a mock function with given fields:
func (_m *KeysCache) Get() map[string]*credloader.KeyStore {
	ret := _m.Called()

	var r0 map[string]*credloader.KeyStore
	if rf, ok := ret.Get(0).(func() map[string]*credloader.KeyStore); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]*credloader.KeyStore)
		}
	}

	return r0
}

// NewKeysCache creates a new instance of KeysCache. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewKeysCache(t interface {
	mock.TestingT
	Cleanup(func())
}) *KeysCache {
	mock := &KeysCache{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
