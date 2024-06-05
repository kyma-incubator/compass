// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// FormationService is an autogenerated mock type for the FormationService type
type FormationService struct {
	mock.Mock
}

// ListFormationsForObject provides a mock function with given fields: ctx, objectID
func (_m *FormationService) ListFormationsForObject(ctx context.Context, objectID string) ([]*model.Formation, error) {
	ret := _m.Called(ctx, objectID)

	if len(ret) == 0 {
		panic("no return value specified for ListFormationsForObject")
	}

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
