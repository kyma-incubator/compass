// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// AppTmplService is an autogenerated mock type for the appTmplService type
type AppTmplService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, in
func (_m *AppTmplService) Create(ctx context.Context, in model.ApplicationTemplateInput) (string, error) {
	ret := _m.Called(ctx, in)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, model.ApplicationTemplateInput) string); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.ApplicationTemplateInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByNameAndRegion provides a mock function with given fields: ctx, name, region
func (_m *AppTmplService) GetByNameAndRegion(ctx context.Context, name string, region interface{}) (*model.ApplicationTemplate, error) {
	ret := _m.Called(ctx, name, region)

	var r0 *model.ApplicationTemplate
	if rf, ok := ret.Get(0).(func(context.Context, string, interface{}) *model.ApplicationTemplate); ok {
		r0 = rf(ctx, name, region)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationTemplate)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, interface{}) error); ok {
		r1 = rf(ctx, name, region)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, in
func (_m *AppTmplService) Update(ctx context.Context, id string, in model.ApplicationTemplateUpdateInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.ApplicationTemplateUpdateInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewAppTmplService creates a new instance of AppTmplService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewAppTmplService(t testing.TB) *AppTmplService {
	mock := &AppTmplService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
