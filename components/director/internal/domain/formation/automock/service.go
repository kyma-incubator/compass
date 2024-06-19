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

// AssignFormation provides a mock function with given fields: ctx, tnt, objectID, objectType, _a4, initialConfigurations
func (_m *Service) AssignFormation(ctx context.Context, tnt string, objectID string, objectType graphql.FormationObjectType, _a4 model.Formation, initialConfigurations model.InitialConfigurations) (*model.Formation, error) {
	ret := _m.Called(ctx, tnt, objectID, objectType, _a4, initialConfigurations)

	var r0 *model.Formation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation, model.InitialConfigurations) (*model.Formation, error)); ok {
		return rf(ctx, tnt, objectID, objectType, _a4, initialConfigurations)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation, model.InitialConfigurations) *model.Formation); ok {
		r0 = rf(ctx, tnt, objectID, objectType, _a4, initialConfigurations)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation, model.InitialConfigurations) error); ok {
		r1 = rf(ctx, tnt, objectID, objectType, _a4, initialConfigurations)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateFormation provides a mock function with given fields: ctx, tnt, _a2, templateName
func (_m *Service) CreateFormation(ctx context.Context, tnt string, _a2 model.Formation, templateName string) (*model.Formation, error) {
	ret := _m.Called(ctx, tnt, _a2, templateName)

	var r0 *model.Formation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.Formation, string) (*model.Formation, error)); ok {
		return rf(ctx, tnt, _a2, templateName)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, model.Formation, string) *model.Formation); ok {
		r0 = rf(ctx, tnt, _a2, templateName)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, model.Formation, string) error); ok {
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
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.Formation) (*model.Formation, error)); ok {
		return rf(ctx, tnt, _a2)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, model.Formation) *model.Formation); ok {
		r0 = rf(ctx, tnt, _a2)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, model.Formation) error); ok {
		r1 = rf(ctx, tnt, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// FinalizeDraftFormation provides a mock function with given fields: ctx, formationID
func (_m *Service) FinalizeDraftFormation(ctx context.Context, formationID string) (*model.Formation, error) {
	ret := _m.Called(ctx, formationID)

	var r0 *model.Formation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Formation, error)); ok {
		return rf(ctx, formationID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Formation); ok {
		r0 = rf(ctx, formationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, formationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: ctx, id
func (_m *Service) Get(ctx context.Context, id string) (*model.Formation, error) {
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

// GetFormationByName provides a mock function with given fields: ctx, formationName, tnt
func (_m *Service) GetFormationByName(ctx context.Context, formationName string, tnt string) (*model.Formation, error) {
	ret := _m.Called(ctx, formationName, tnt)

	var r0 *model.Formation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.Formation, error)); ok {
		return rf(ctx, formationName, tnt)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Formation); ok {
		r0 = rf(ctx, formationName, tnt)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, formationName, tnt)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGlobalByID provides a mock function with given fields: ctx, id
func (_m *Service) GetGlobalByID(ctx context.Context, id string) (*model.Formation, error) {
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

// List provides a mock function with given fields: ctx, pageSize, cursor
func (_m *Service) List(ctx context.Context, pageSize int, cursor string) (*model.FormationPage, error) {
	ret := _m.Called(ctx, pageSize, cursor)

	var r0 *model.FormationPage
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, int, string) (*model.FormationPage, error)); ok {
		return rf(ctx, pageSize, cursor)
	}
	if rf, ok := ret.Get(0).(func(context.Context, int, string) *model.FormationPage); ok {
		r0 = rf(ctx, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationPage)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, int, string) error); ok {
		r1 = rf(ctx, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListFormationsForObjectGlobal provides a mock function with given fields: ctx, objectID
func (_m *Service) ListFormationsForObjectGlobal(ctx context.Context, objectID string) ([]*model.Formation, error) {
	ret := _m.Called(ctx, objectID)

	var r0 []*model.Formation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*model.Formation, error)); ok {
		return rf(ctx, objectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.Formation); ok {
		r0 = rf(ctx, objectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Formation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ResynchronizeFormationNotifications provides a mock function with given fields: ctx, formationID, reset
func (_m *Service) ResynchronizeFormationNotifications(ctx context.Context, formationID string, reset bool) (*model.Formation, error) {
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

// UnassignFormation provides a mock function with given fields: ctx, tnt, objectID, objectType, _a4, ignoreASA
func (_m *Service) UnassignFormation(ctx context.Context, tnt string, objectID string, objectType graphql.FormationObjectType, _a4 model.Formation, ignoreASA bool) (*model.Formation, error) {
	ret := _m.Called(ctx, tnt, objectID, objectType, _a4, ignoreASA)

	var r0 *model.Formation
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation, bool) (*model.Formation, error)); ok {
		return rf(ctx, tnt, objectID, objectType, _a4, ignoreASA)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation, bool) *model.Formation); ok {
		r0 = rf(ctx, tnt, objectID, objectType, _a4, ignoreASA)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, graphql.FormationObjectType, model.Formation, bool) error); ok {
		r1 = rf(ctx, tnt, objectID, objectType, _a4, ignoreASA)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewService creates a new instance of Service. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewService(t interface {
	mock.TestingT
	Cleanup(func())
}) *Service {
	mock := &Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
