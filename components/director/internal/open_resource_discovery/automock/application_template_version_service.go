// Code generated by mockery. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

// ApplicationTemplateVersionService is an autogenerated mock type for the ApplicationTemplateVersionService type
type ApplicationTemplateVersionService struct {
	mock.Mock
}

type mockConstructorTestingTNewApplicationTemplateVersionService interface {
	mock.TestingT
	Cleanup(func())
}

// NewApplicationTemplateVersionService creates a new instance of ApplicationTemplateVersionService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewApplicationTemplateVersionService(t mockConstructorTestingTNewApplicationTemplateVersionService) *ApplicationTemplateVersionService {
	mock := &ApplicationTemplateVersionService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
