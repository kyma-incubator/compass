// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// FormationService is an autogenerated mock type for the formationService type
type FormationService struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, id
func (_m *FormationService) Get(ctx context.Context, id string) (*model.Formation, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Formation
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Formation); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
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

// UnassignFormation provides a mock function with given fields: ctx, tnt, objectID, objectType, formation
func (_m *FormationService) UnassignFormation(ctx context.Context, tnt string, objectID string, objectType graphql.FormationObjectType, formation model.Formation) (*model.Formation, error) {
	ret := _m.Called(ctx, tnt, objectID, objectType, formation)

	var r0 *model.Formation
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation) *model.Formation); ok {
		r0 = rf(ctx, tnt, objectID, objectType, formation)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation) error); ok {
		r1 = rf(ctx, tnt, objectID, objectType, formation)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewFormationService creates a new instance of FormationService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationService(t testing.TB) *FormationService {
	mock := &FormationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
