// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// ApplicationRepository is an autogenerated mock type for the applicationRepository type
type ApplicationRepository struct {
	mock.Mock
}

// GetByID provides a mock function with given fields: ctx, tenant, id
func (_m *ApplicationRepository) GetByID(ctx context.Context, tenant string, id string) (*model.Application, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 *model.Application
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (*model.Application, error)); ok {
		return rf(ctx, tenant, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Application); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Application)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByScenariosNoPaging provides a mock function with given fields: ctx, tenant, scenarios
func (_m *ApplicationRepository) ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error) {
	ret := _m.Called(ctx, tenant, scenarios)

	var r0 []*model.Application
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) ([]*model.Application, error)); ok {
		return rf(ctx, tenant, scenarios)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) []*model.Application); ok {
		r0 = rf(ctx, tenant, scenarios)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Application)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []string) error); ok {
		r1 = rf(ctx, tenant, scenarios)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OwnerExists provides a mock function with given fields: ctx, tenant, id
func (_m *ApplicationRepository) OwnerExists(ctx context.Context, tenant string, id string) (bool, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 bool
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) (bool, error)); ok {
		return rf(ctx, tenant, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewApplicationRepository creates a new instance of ApplicationRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewApplicationRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *ApplicationRepository {
	mock := &ApplicationRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
