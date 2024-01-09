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

// GetByNameAndTenant provides a mock function with given fields: ctx, templateName, tenantID
func (_m *FormationTemplateRepository) GetByNameAndTenant(ctx context.Context, templateName string, tenantID string) (*model.FormationTemplate, error) {
	ret := _m.Called(ctx, templateName, tenantID)

	var r0 *model.FormationTemplate
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.FormationTemplate, error)); ok {
		return rf(ctx, templateName, tenantID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.FormationTemplate); ok {
		r0 = rf(ctx, templateName, tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationTemplate)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, templateName, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewFormationTemplateRepository creates a new instance of FormationTemplateRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFormationTemplateRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *FormationTemplateRepository {
	mock := &FormationTemplateRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
