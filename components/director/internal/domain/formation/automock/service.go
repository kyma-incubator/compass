// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// AssignFormation provides a mock function with given fields: ctx, tnt, objectID, objectType, _a4
func (_m *Service) AssignFormation(ctx context.Context, tnt string, objectID string, objectType graphql.FormationObjectType, _a4 model.Formation) (*model.Formation, error) {
	ret := _m.Called(ctx, tnt, objectID, objectType, _a4)

	var r0 *model.Formation
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation) *model.Formation); ok {
		r0 = rf(ctx, tnt, objectID, objectType, _a4)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation) error); ok {
		r1 = rf(ctx, tnt, objectID, objectType, _a4)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateFormation provides a mock function with given fields: ctx, tnt, _a2, templateName
func (_m *Service) CreateFormation(ctx context.Context, tnt string, _a2 model.Formation, templateName *string) (*model.Formation, error) {
	ret := _m.Called(ctx, tnt, _a2, templateName)

	var r0 *model.Formation
	if rf, ok := ret.Get(0).(func(context.Context, string, model.Formation, *string) *model.Formation); ok {
		r0 = rf(ctx, tnt, _a2, templateName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, model.Formation, *string) error); ok {
		r1 = rf(ctx, tnt, _a2, templateName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteFormation provides a mock function with given fields: ctx, tnt, _a2
func (_m *Service) DeleteFormation(ctx context.Context, tnt string, _a2 model.Formation) (*model.Formation, error) {
	ret := _m.Called(ctx, tnt, _a2)

	var r0 *model.Formation
	if rf, ok := ret.Get(0).(func(context.Context, string, model.Formation) *model.Formation); ok {
		r0 = rf(ctx, tnt, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, model.Formation) error); ok {
		r1 = rf(ctx, tnt, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UnassignFormation provides a mock function with given fields: ctx, tnt, objectID, objectType, _a4
func (_m *Service) UnassignFormation(ctx context.Context, tnt string, objectID string, objectType graphql.FormationObjectType, _a4 model.Formation) (*model.Formation, error) {
	ret := _m.Called(ctx, tnt, objectID, objectType, _a4)

	var r0 *model.Formation
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation) *model.Formation); ok {
		r0 = rf(ctx, tnt, objectID, objectType, _a4)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation) error); ok {
		r1 = rf(ctx, tnt, objectID, objectType, _a4)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type NewServiceT interface {
	mock.TestingT
	Cleanup(func())
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewService(t NewServiceT) *Service {
	mock := &Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
