// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// ConstraintReferenceService is an autogenerated mock type for the constraintReferenceService type
type ConstraintReferenceService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, in
func (_m *ConstraintReferenceService) Create(ctx context.Context, in *model.FormationTemplateConstraintReference) error {
	ret := _m.Called(ctx, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.FormationTemplateConstraintReference) error); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, constraintID, formationTemplateID
func (_m *ConstraintReferenceService) Delete(ctx context.Context, constraintID string, formationTemplateID string) error {
	ret := _m.Called(ctx, constraintID, formationTemplateID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, constraintID, formationTemplateID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewConstraintReferenceService creates a new instance of ConstraintReferenceService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewConstraintReferenceService(t testing.TB) *ConstraintReferenceService {
	mock := &ConstraintReferenceService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
