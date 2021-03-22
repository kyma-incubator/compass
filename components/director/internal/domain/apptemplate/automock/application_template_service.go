// Code generated by mockery v1.0.0. DO NOT EDIT.

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

// Create provides a mock function with given fields: ctx, in
func (_m *ApplicationTemplateService) Create(ctx context.Context, in model.ApplicationTemplateInput) (string, error) {
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

// Delete provides a mock function with given fields: ctx, id
func (_m *ApplicationTemplateService) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, id
func (_m *ApplicationTemplateService) Get(ctx context.Context, id string) (*model.ApplicationTemplate, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.ApplicationTemplate
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.ApplicationTemplate); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationTemplate)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByName provides a mock function with given fields: ctx, name
func (_m *ApplicationTemplateService) GetByName(ctx context.Context, name string) (*model.ApplicationTemplate, error) {
	ret := _m.Called(ctx, name)

	var r0 *model.ApplicationTemplate
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.ApplicationTemplate); ok {
		r0 = rf(ctx, name)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationTemplate)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, name)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, pageSize, cursor
func (_m *ApplicationTemplateService) List(ctx context.Context, pageSize int, cursor string) (model.ApplicationTemplatePage, error) {
	ret := _m.Called(ctx, pageSize, cursor)

	var r0 model.ApplicationTemplatePage
	if rf, ok := ret.Get(0).(func(context.Context, int, string) model.ApplicationTemplatePage); ok {
		r0 = rf(ctx, pageSize, cursor)
	} else {
		r0 = ret.Get(0).(model.ApplicationTemplatePage)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int, string) error); ok {
		r1 = rf(ctx, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PrepareApplicationCreateInputJSON provides a mock function with given fields: appTemplate, values
func (_m *ApplicationTemplateService) PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, values model.ApplicationFromTemplateInputValues) (string, error) {
	ret := _m.Called(appTemplate, values)

	var r0 string
	if rf, ok := ret.Get(0).(func(*model.ApplicationTemplate, model.ApplicationFromTemplateInputValues) string); ok {
		r0 = rf(appTemplate, values)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.ApplicationTemplate, model.ApplicationFromTemplateInputValues) error); ok {
		r1 = rf(appTemplate, values)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, in
func (_m *ApplicationTemplateService) Update(ctx context.Context, id string, in model.ApplicationTemplateUpdateInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.ApplicationTemplateUpdateInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
