// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// FormationTemplateConstraintReferenceRepository is an autogenerated mock type for the formationTemplateConstraintReferenceRepository type
type FormationTemplateConstraintReferenceRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, item
func (_m *FormationTemplateConstraintReferenceRepository) Create(ctx context.Context, item *model.FormationTemplateConstraintReference) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.FormationTemplateConstraintReference) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListByFormationTemplateID provides a mock function with given fields: ctx, formationTemplateID
func (_m *FormationTemplateConstraintReferenceRepository) ListByFormationTemplateID(ctx context.Context, formationTemplateID string) ([]*model.FormationTemplateConstraintReference, error) {
	ret := _m.Called(ctx, formationTemplateID)

	var r0 []*model.FormationTemplateConstraintReference
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.FormationTemplateConstraintReference); ok {
		r0 = rf(ctx, formationTemplateID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationTemplateConstraintReference)
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

type NewFormationTemplateConstraintReferenceRepositoryT interface {
	mock.TestingT
	Cleanup(func())
}

// NewFormationTemplateConstraintReferenceRepository creates a new instance of FormationTemplateConstraintReferenceRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationTemplateConstraintReferenceRepository(t NewFormationTemplateConstraintReferenceRepositoryT) *FormationTemplateConstraintReferenceRepository {
	mock := &FormationTemplateConstraintReferenceRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
