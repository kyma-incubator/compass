// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// FormationTemplateRepo is an autogenerated mock type for the formationTemplateRepo type
type FormationTemplateRepo struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, id
func (_m *FormationTemplateRepo) Get(ctx context.Context, id string) (*model.FormationTemplate, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.FormationTemplate
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.FormationTemplate); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationTemplate)
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

type mockConstructorTestingTNewFormationTemplateRepo interface {
	mock.TestingT
	Cleanup(func())
}

// NewFormationTemplateRepo creates a new instance of FormationTemplateRepo. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationTemplateRepo(t mockConstructorTestingTNewFormationTemplateRepo) *FormationTemplateRepo {
	mock := &FormationTemplateRepo{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
