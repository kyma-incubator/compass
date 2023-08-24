// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// FormationTemplateRepository is an autogenerated mock type for the FormationTemplateRepository type
type FormationTemplateRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, item
func (_m *FormationTemplateRepository) Create(ctx context.Context, item *model.FormationTemplate) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.FormationTemplate) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, id, tenantID
func (_m *FormationTemplateRepository) Delete(ctx context.Context, id string, tenantID string) error {
	ret := _m.Called(ctx, id, tenantID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, id, tenantID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: ctx, id
func (_m *FormationTemplateRepository) Exists(ctx context.Context, id string) (bool, error) {
	ret := _m.Called(ctx, id)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (bool, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: ctx, id
func (_m *FormationTemplateRepository) Get(ctx context.Context, id string) (*model.FormationTemplate, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.FormationTemplate
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.FormationTemplate, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.FormationTemplate); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationTemplate)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, name, tenantID, pageSize, cursor
func (_m *FormationTemplateRepository) List(ctx context.Context, name *string, tenantID string, pageSize int, cursor string) (*model.FormationTemplatePage, error) {
	ret := _m.Called(ctx, name, tenantID, pageSize, cursor)

	var r0 *model.FormationTemplatePage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, *string, string, int, string) (*model.FormationTemplatePage, error)); ok {
		return rf(ctx, name, tenantID, pageSize, cursor)
	}
	if rf, ok := ret.Get(0).(func(context.Context, *string, string, int, string) *model.FormationTemplatePage); ok {
		r0 = rf(ctx, name, tenantID, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationTemplatePage)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, *string, string, int, string) error); ok {
		r1 = rf(ctx, name, tenantID, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, _a1
func (_m *FormationTemplateRepository) Update(ctx context.Context, _a1 *model.FormationTemplate) error {
	ret := _m.Called(ctx, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.FormationTemplate) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewFormationTemplateRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewFormationTemplateRepository creates a new instance of FormationTemplateRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationTemplateRepository(t mockConstructorTestingTNewFormationTemplateRepository) *FormationTemplateRepository {
	mock := &FormationTemplateRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
