// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// ApiClient is an autogenerated mock type for the ApiClient type
type ApiClient struct {
	mock.Mock
}

// Do provides a mock function with given fields: _a0, _a1
func (_m *ApiClient) Do(_a0 *http.Request, _a1 string) (*http.Response, error) {
	ret := _m.Called(_a0, _a1)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(*http.Request, string) *http.Response); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Request, string) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
