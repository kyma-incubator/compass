// Code generated by mockery 2.9.0. DO NOT EDIT.

package automock

import (
	http "net/http"

	mock "github.com/stretchr/testify/mock"
)

// CertificateHeaderParser is an autogenerated mock type for the certificateHeaderParser type
type CertificateHeaderParser struct {
	mock.Mock
}

// GetCertificateData provides a mock function with given fields: _a0
func (_m *CertificateHeaderParser) GetCertificateData(_a0 *http.Request) *oathkeeper.certificateData {
	ret := _m.Called(_a0)

	var r0 *oathkeeper.certificateData
	if rf, ok := ret.Get(0).(func(*http.Request) *oathkeeper.certificateData); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*oathkeeper.certificateData)
		}
	}

	return r0
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
