// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// Service is an autogenerated mock type for the Service type
type Service struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, tenant, key
func (_m *Service) Get(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error) {
	ret := _m.Called(ctx, tenant, key)

	var r0 *model.LabelDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.LabelDefinition); ok {
		r0 = rf(ctx, tenant, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.LabelDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, tenant
func (_m *Service) List(ctx context.Context, tenant string) ([]model.LabelDefinition, error) {
	ret := _m.Called(ctx, tenant)

	var r0 []model.LabelDefinition
	if rf, ok := ret.Get(0).(func(context.Context, string) []model.LabelDefinition); ok {
		r0 = rf(ctx, tenant)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.LabelDefinition)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewService creates a new instance of Service. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewService(t testing.TB) *Service {
	mock := &Service{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
