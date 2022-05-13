// Code generated by mockery. DO NOT EDIT.

package automock

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// HttpClient is an autogenerated mock type for the HttpClient type
type HttpClient struct {
	mock.Mock
}

// Do provides a mock function with given fields: request
func (_m *HttpClient) Do(request *http.Request) (*http.Response, error) {
	ret := _m.Called(request)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(*http.Request) *http.Response); ok {
		r0 = rf(request)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Request) error); ok {
		r1 = rf(request)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
