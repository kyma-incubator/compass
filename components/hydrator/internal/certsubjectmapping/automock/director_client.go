// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"
)

// DirectorClient is an autogenerated mock type for the DirectorClient type
type DirectorClient struct {
	mock.Mock
}

// ListCertificateSubjectMappings provides a mock function with given fields: ctx, after
func (_m *DirectorClient) ListCertificateSubjectMappings(ctx context.Context, after string) (*graphql.CertificateSubjectMappingPage, error) {
	ret := _m.Called(ctx, after)

	var r0 *graphql.CertificateSubjectMappingPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*graphql.CertificateSubjectMappingPage, error)); ok {
		return rf(ctx, after)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *graphql.CertificateSubjectMappingPage); ok {
		r0 = rf(ctx, after)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.CertificateSubjectMappingPage)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, after)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewDirectorClient interface {
	mock.TestingT
	Cleanup(func())
}

// NewDirectorClient creates a new instance of DirectorClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDirectorClient(t mockConstructorTestingTNewDirectorClient) *DirectorClient {
	mock := &DirectorClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
