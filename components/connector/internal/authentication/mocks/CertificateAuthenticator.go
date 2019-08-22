// Code generated by mockery v1.0.0. DO NOT EDIT.
package mocks

import authentication "github.com/kyma-incubator/compass/components/connector/internal/authentication"
import context "context"
import mock "github.com/stretchr/testify/mock"

// CertificateAuthenticator is an autogenerated mock type for the CertificateAuthenticator type
type CertificateAuthenticator struct {
	mock.Mock
}

// AuthenticateCertificate provides a mock function with given fields: _a0
func (_m *CertificateAuthenticator) AuthenticateCertificate(_a0 context.Context) (authentication.CertificateData, error) {
	ret := _m.Called(_a0)

	var r0 authentication.CertificateData
	if rf, ok := ret.Get(0).(func(context.Context) authentication.CertificateData); ok {
		r0 = rf(_a0)
	} else {
		r0 = ret.Get(0).(authentication.CertificateData)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(_a0)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
