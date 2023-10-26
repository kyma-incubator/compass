// Code generated by mockery. DO NOT EDIT.

package automock

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// APIClient is an autogenerated mock type for the APIClient type
type APIClient struct {
	mock.Mock
}

// Do provides a mock function with given fields: _a0, _a1
func (_m *APIClient) Do(_a0 *http.Request, _a1 string) (*http.Response, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *http.Response
	var r1 error
	if rf, ok := ret.Get(0).(func(*http.Request, string) (*http.Response, error)); ok {
		return rf(_a0, _a1)
	}
	if rf, ok := ret.Get(0).(func(*http.Request, string) *http.Response); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	if rf, ok := ret.Get(1).(func(*http.Request, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewAPIClient creates a new instance of APIClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewAPIClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *APIClient {
	mock := &APIClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
