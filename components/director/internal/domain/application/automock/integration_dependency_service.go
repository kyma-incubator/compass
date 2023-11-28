// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// IntegrationDependencyService is an autogenerated mock type for the IntegrationDependencyService type
type IntegrationDependencyService struct {
	mock.Mock
}

// ListByApplicationIDs provides a mock function with given fields: ctx, applicationIDs, pageSize, cursor
func (_m *IntegrationDependencyService) ListByApplicationIDs(ctx context.Context, applicationIDs []string, pageSize int, cursor string) ([]*model.IntegrationDependencyPage, error) {
	ret := _m.Called(ctx, applicationIDs, pageSize, cursor)

	var r0 []*model.IntegrationDependencyPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string, int, string) ([]*model.IntegrationDependencyPage, error)); ok {
		return rf(ctx, applicationIDs, pageSize, cursor)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string, int, string) []*model.IntegrationDependencyPage); ok {
		r0 = rf(ctx, applicationIDs, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.IntegrationDependencyPage)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string, int, string) error); ok {
		r1 = rf(ctx, applicationIDs, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewIntegrationDependencyService creates a new instance of IntegrationDependencyService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewIntegrationDependencyService(t interface {
	mock.TestingT
	Cleanup(func())
}) *IntegrationDependencyService {
	mock := &IntegrationDependencyService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
