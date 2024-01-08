// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// FormationService is an autogenerated mock type for the formationService type
type FormationService struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, id
func (_m *FormationService) Get(ctx context.Context, id string) (*model.Formation, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Formation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Formation, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Formation); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGlobalByID provides a mock function with given fields: ctx, id
func (_m *FormationService) GetGlobalByID(ctx context.Context, id string) (*model.Formation, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Formation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Formation, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Formation); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResynchronizeFormationNotifications provides a mock function with given fields: ctx, formationID, reset
func (_m *FormationService) ResynchronizeFormationNotifications(ctx context.Context, formationID string, reset bool) (*model.Formation, error) {
	ret := _m.Called(ctx, formationID, reset)

	var r0 *model.Formation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, bool) (*model.Formation, error)); ok {
		return rf(ctx, formationID, reset)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, bool) *model.Formation); ok {
		r0 = rf(ctx, formationID, reset)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, bool) error); ok {
		r1 = rf(ctx, formationID, reset)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnassignFromScenarioLabel provides a mock function with given fields: ctx, tnt, objectID, objectType, formation
func (_m *FormationService) UnassignFromScenarioLabel(ctx context.Context, tnt string, objectID string, objectType graphql.FormationObjectType, formation *model.Formation) error {
	ret := _m.Called(ctx, tnt, objectID, objectType, formation)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType, *model.Formation) error); ok {
		r0 = rf(ctx, tnt, objectID, objectType, formation)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewFormationService creates a new instance of FormationService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewFormationService(t interface {
	mock.TestingT
	Cleanup(func())
}) *FormationService {
	mock := &FormationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
