// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// CertSubjectMappingService is an autogenerated mock type for the CertSubjectMappingService type
type CertSubjectMappingService struct {
	mock.Mock
}

// DeleteByConsumerID provides a mock function with given fields: ctx, consumerID
func (_m *CertSubjectMappingService) DeleteByConsumerID(ctx context.Context, consumerID string) error {
	ret := _m.Called(ctx, consumerID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, consumerID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewCertSubjectMappingService creates a new instance of CertSubjectMappingService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewCertSubjectMappingService(t interface {
	mock.TestingT
	Cleanup(func())
}) *CertSubjectMappingService {
	mock := &CertSubjectMappingService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
