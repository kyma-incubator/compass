// Code generated by mockery v2.10.5. DO NOT EDIT.

package automock

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// HTTPDoer is an autogenerated mock type for the HTTPDoer type
type HTTPDoer struct {
	mock.Mock
}

// Do provides a mock function with given fields: req
func (_m *HTTPDoer) Do(req *http.Request) (*http.Response, error) {
	ret := _m.Called(req)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(*http.Request) *http.Response); ok {
		r0 = rf(req)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Request) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
