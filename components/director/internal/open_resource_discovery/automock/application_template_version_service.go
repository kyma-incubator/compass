// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// ApplicationTemplateVersionService is an autogenerated mock type for the ApplicationTemplateVersionService type
type ApplicationTemplateVersionService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, appTemplateID, item
func (_m *ApplicationTemplateVersionService) Create(ctx context.Context, appTemplateID string, item *model.ApplicationTemplateVersionInput) (string, error) {
	ret := _m.Called(ctx, appTemplateID, item)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.ApplicationTemplateVersionInput) string); ok {
		r0 = rf(ctx, appTemplateID, item)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *model.ApplicationTemplateVersionInput) error); ok {
		r1 = rf(ctx, appTemplateID, item)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByAppTemplateIDAndVersion provides a mock function with given fields: ctx, id, version
func (_m *ApplicationTemplateVersionService) GetByAppTemplateIDAndVersion(ctx context.Context, id string, version string) (*model.ApplicationTemplateVersion, error) {
	ret := _m.Called(ctx, id, version)

	var r0 *model.ApplicationTemplateVersion
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.ApplicationTemplateVersion); ok {
		r0 = rf(ctx, id, version)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationTemplateVersion)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByAppTemplateID provides a mock function with given fields: ctx, appTemplateID
func (_m *ApplicationTemplateVersionService) ListByAppTemplateID(ctx context.Context, appTemplateID string) ([]*model.ApplicationTemplateVersion, error) {
	ret := _m.Called(ctx, appTemplateID)

	var r0 []*model.ApplicationTemplateVersion
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.ApplicationTemplateVersion); ok {
		r0 = rf(ctx, appTemplateID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.ApplicationTemplateVersion)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, appTemplateID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, appTemplateID, in
func (_m *ApplicationTemplateVersionService) Update(ctx context.Context, id string, appTemplateID string, in *model.ApplicationTemplateVersionInput) error {
	ret := _m.Called(ctx, id, appTemplateID, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *model.ApplicationTemplateVersionInput) error); ok {
		r0 = rf(ctx, id, appTemplateID, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
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
