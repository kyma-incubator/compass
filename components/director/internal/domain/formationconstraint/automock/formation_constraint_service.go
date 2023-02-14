// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// FormationConstraintService is an autogenerated mock type for the formationConstraintService type
type FormationConstraintService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, in
func (_m *FormationConstraintService) Create(ctx context.Context, in *model.FormationConstraintInput) (string, error) {
	ret := _m.Called(ctx, in)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, *model.FormationConstraintInput) string); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *model.FormationConstraintInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *FormationConstraintService) Delete(ctx context.Context, id string) error {
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
func (_m *FormationConstraintService) Get(ctx context.Context, id string) (*model.FormationConstraint, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.FormationConstraint
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.FormationConstraint); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationConstraint)
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

// List provides a mock function with given fields: ctx
func (_m *FormationConstraintService) List(ctx context.Context) ([]*model.FormationConstraint, error) {
	ret := _m.Called(ctx)

	var r0 []*model.FormationConstraint
	if rf, ok := ret.Get(0).(func(context.Context) []*model.FormationConstraint); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationConstraint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByFormationTemplateID provides a mock function with given fields: ctx, formationTemplateID
func (_m *FormationConstraintService) ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.FormationConstraint, error) {
	ret := _m.Called(ctx, formationTemplateID)

	var r0 []*model.FormationConstraint
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.FormationConstraint); ok {
		r0 = rf(ctx, formationTemplateID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationConstraint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, formationTemplateID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewFormationConstraintService creates a new instance of FormationConstraintService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationConstraintService(t testing.TB) *FormationConstraintService {
	mock := &FormationConstraintService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
