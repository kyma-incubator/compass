// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, in
func (_m *Service) Create(ctx context.Context, in model.AutomaticScenarioAssignment) (model.AutomaticScenarioAssignment, error) {
	ret := _m.Called(ctx, in)

	var r0 model.AutomaticScenarioAssignment
	if rf, ok := ret.Get(0).(func(context.Context, model.AutomaticScenarioAssignment) model.AutomaticScenarioAssignment); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(model.AutomaticScenarioAssignment)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.AutomaticScenarioAssignment) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
