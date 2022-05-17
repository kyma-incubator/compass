// Code generated by mockery. DO NOT EDIT.

package automock

import (
	http "net/http"

	certresolver "github.com/kyma-incubator/compass/components/hydrator/internal/certresolver"

	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// CertificateHeaderParser is an autogenerated mock type for the CertificateHeaderParser type
type CertificateHeaderParser struct {
	mock.Mock
}

// GetCertificateData provides a mock function with given fields: _a0
func (_m *CertificateHeaderParser) GetCertificateData(_a0 *http.Request) *certresolver.CertificateData {
	ret := _m.Called(_a0)

	var r0 *certresolver.CertificateData
	if rf, ok := ret.Get(0).(func(*http.Request) *certresolver.CertificateData); ok {
		r0 = rf(_a0)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*certresolver.CertificateData)
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

// NewCertificateHeaderParser creates a new instance of CertificateHeaderParser. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewCertificateHeaderParser(t testing.TB) *CertificateHeaderParser {
	mock := &CertificateHeaderParser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
