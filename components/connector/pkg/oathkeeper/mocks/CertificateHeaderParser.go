// Code generated by mockery v2.5.1. DO NOT EDIT.

package mocks

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// CertificateHeaderParser is an autogenerated mock type for the CertificateHeaderParser type
type CertificateHeaderParser struct {
	mock.Mock
}

// GetCertificateData provides a mock function with given fields: r
func (_m *CertificateHeaderParser) GetCertificateData(r *http.Request) (string, string, bool) {
	ret := _m.Called(r)

	var r0 string
	if rf, ok := ret.Get(0).(func(*http.Request) string); ok {
		r0 = rf(r)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 string
	if rf, ok := ret.Get(1).(func(*http.Request) string); ok {
		r1 = rf(r)
	} else {
		r1 = ret.Get(1).(string)
	}

	var r2 bool
	if rf, ok := ret.Get(2).(func(*http.Request) bool); ok {
		r2 = rf(r)
	} else {
		r2 = ret.Get(2).(bool)
	}

	return r0, r1, r2
}

// GetIssuer provides a mock function with given fields:
func (_m *CertificateHeaderParser) GetIssuer() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
