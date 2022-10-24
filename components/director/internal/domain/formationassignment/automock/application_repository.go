// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// ApplicationRepository is an autogenerated mock type for the applicationRepository type
type ApplicationRepository struct {
	mock.Mock
}

// ListByScenariosNoPaging provides a mock function with given fields: ctx, tenant, scenarios
func (_m *ApplicationRepository) ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error) {
	ret := _m.Called(ctx, tenant, scenarios)

	var r0 []*model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) []*model.Application); ok {
		r0 = rf(ctx, tenant, scenarios)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string) error); ok {
		r1 = rf(ctx, tenant, scenarios)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewApplicationRepository creates a new instance of ApplicationRepository. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewApplicationRepository(t testing.TB) *ApplicationRepository {
	mock := &ApplicationRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
