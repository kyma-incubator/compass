// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	pkgformationconstraint "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
)

// FormationConstraintRepository is an autogenerated mock type for the formationConstraintRepository type
type FormationConstraintRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, item
func (_m *FormationConstraintRepository) Create(ctx context.Context, item *model.FormationConstraint) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.FormationConstraint) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, id
func (_m *FormationConstraintRepository) Delete(ctx context.Context, id string) error {
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
func (_m *FormationConstraintRepository) Get(ctx context.Context, id string) (*model.FormationConstraint, error) {
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

// ListAll provides a mock function with given fields: ctx
func (_m *FormationConstraintRepository) ListAll(ctx context.Context) ([]*model.FormationConstraint, error) {
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

// ListByIDs provides a mock function with given fields: ctx, formationConstraintIDs
func (_m *FormationConstraintRepository) ListByIDs(ctx context.Context, formationConstraintIDs []string) ([]*model.FormationConstraint, error) {
	ret := _m.Called(ctx, formationConstraintIDs)

	var r0 []*model.FormationConstraint
	if rf, ok := ret.Get(0).(func(context.Context, []string) []*model.FormationConstraint); ok {
		r0 = rf(ctx, formationConstraintIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationConstraint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, formationConstraintIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByIDsAndGlobal provides a mock function with given fields: ctx, formationConstraintIDs
func (_m *FormationConstraintRepository) ListByIDsAndGlobal(ctx context.Context, formationConstraintIDs []string) ([]*model.FormationConstraint, error) {
	ret := _m.Called(ctx, formationConstraintIDs)

	var r0 []*model.FormationConstraint
	if rf, ok := ret.Get(0).(func(context.Context, []string) []*model.FormationConstraint); ok {
		r0 = rf(ctx, formationConstraintIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationConstraint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, formationConstraintIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListMatchingFormationConstraints provides a mock function with given fields: ctx, formationConstraintIDs, location, details
func (_m *FormationConstraintRepository) ListMatchingFormationConstraints(ctx context.Context, formationConstraintIDs []string, location pkgformationconstraint.JoinPointLocation, details pkgformationconstraint.MatchingDetails) ([]*model.FormationConstraint, error) {
	ret := _m.Called(ctx, formationConstraintIDs, location, details)

	var r0 []*model.FormationConstraint
	if rf, ok := ret.Get(0).(func(context.Context, []string, pkgformationconstraint.JoinPointLocation, pkgformationconstraint.MatchingDetails) []*model.FormationConstraint); ok {
		r0 = rf(ctx, formationConstraintIDs, location, details)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationConstraint)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string, pkgformationconstraint.JoinPointLocation, pkgformationconstraint.MatchingDetails) error); ok {
		r1 = rf(ctx, formationConstraintIDs, location, details)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, _a1
func (_m *FormationConstraintRepository) Update(ctx context.Context, _a1 *model.FormationConstraint) error {
	ret := _m.Called(ctx, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.FormationConstraint) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewFormationConstraintRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewFormationConstraintRepository creates a new instance of FormationConstraintRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationConstraintRepository(t mockConstructorTestingTNewFormationConstraintRepository) *FormationConstraintRepository {
	mock := &FormationConstraintRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
