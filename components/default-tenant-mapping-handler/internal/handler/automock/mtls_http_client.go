// Code generated by mockery. DO NOT EDIT.

package automock

import (
	http "net/http"
	testing "testing"

	mock "github.com/stretchr/testify/mock"
)

// MtlsHTTPClient is an autogenerated mock type for the mtlsHTTPClient type
type MtlsHTTPClient struct {
	mock.Mock
}

// Do provides a mock function with given fields: request
func (_m *MtlsHTTPClient) Do(request *http.Request) (*http.Response, error) {
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

// NewMtlsHTTPClient creates a new instance of MtlsHTTPClient. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewMtlsHTTPClient(t testing.TB) *MtlsHTTPClient {
	mock := &MtlsHTTPClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
