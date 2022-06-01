// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// FormationService is an autogenerated mock type for the formationService type
type FormationService struct {
	mock.Mock
}

// DeleteAutomaticScenarioAssignment provides a mock function with given fields: ctx, in
func (_m *FormationService) DeleteAutomaticScenarioAssignment(ctx context.Context, in model.AutomaticScenarioAssignment) error {
	ret := _m.Called(ctx, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.AutomaticScenarioAssignment) error); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewFormationService creates a new instance of FormationService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationService(t testing.TB) *FormationService {
	mock := &FormationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
