// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// ApplicationTemplateService is an autogenerated mock type for the ApplicationTemplateService type
type ApplicationTemplateService struct {
	mock.Mock
}

// GetLabel provides a mock function with given fields: ctx, appTemplateID, key
func (_m *ApplicationTemplateService) GetLabel(ctx context.Context, appTemplateID string, key string) (*model.Label, error) {
	ret := _m.Called(ctx, appTemplateID, key)

	var r0 *model.Label
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.Label, error)); ok {
		return rf(ctx, appTemplateID, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Label); ok {
		r0 = rf(ctx, appTemplateID, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Label)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, appTemplateID, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewApplicationTemplateService creates a new instance of ApplicationTemplateService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewApplicationTemplateService(t interface {
	mock.TestingT
	Cleanup(func())
}) *ApplicationTemplateService {
	mock := &ApplicationTemplateService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
