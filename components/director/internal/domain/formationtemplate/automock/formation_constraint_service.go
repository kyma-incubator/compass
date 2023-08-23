// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// FormationConstraintService is an autogenerated mock type for the FormationConstraintService type
type FormationConstraintService struct {
	mock.Mock
}

// ListByFormationTemplateIDs provides a mock function with given fields: ctx, formationTemplateIDs
func (_m *FormationConstraintService) ListByFormationTemplateIDs(ctx context.Context, formationTemplateIDs []string) ([][]*model.FormationConstraint, error) {
	ret := _m.Called(ctx, formationTemplateIDs)

	var r0 [][]*model.FormationConstraint
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([][]*model.FormationConstraint, error)); ok {
		return rf(ctx, formationTemplateIDs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) [][]*model.FormationConstraint); ok {
		r0 = rf(ctx, formationTemplateIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]*model.FormationConstraint)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, formationTemplateIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewFormationConstraintService interface {
	mock.TestingT
	Cleanup(func())
}

// NewFormationConstraintService creates a new instance of FormationConstraintService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationConstraintService(t mockConstructorTestingTNewFormationConstraintService) *FormationConstraintService {
	mock := &FormationConstraintService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
