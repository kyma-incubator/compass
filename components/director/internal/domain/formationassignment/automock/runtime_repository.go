// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// RuntimeRepository is an autogenerated mock type for the runtimeRepository type
type RuntimeRepository struct {
	mock.Mock
}

// ListByScenarios provides a mock function with given fields: ctx, tenant, scenarios
func (_m *RuntimeRepository) ListByScenarios(ctx context.Context, tenant string, scenarios []string) ([]*model.Runtime, error) {
	ret := _m.Called(ctx, tenant, scenarios)

	var r0 []*model.Runtime
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) []*model.Runtime); ok {
		r0 = rf(ctx, tenant, scenarios)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Runtime)
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

// NewRuntimeRepository creates a new instance of RuntimeRepository. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewRuntimeRepository(t testing.TB) *RuntimeRepository {
	mock := &RuntimeRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
