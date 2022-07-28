// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	labelfilter "github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// RuntimeRepository is an autogenerated mock type for the runtimeRepository type
type RuntimeRepository struct {
	mock.Mock
}

// Exists provides a mock function with given fields: ctx, tenant, id
func (_m *RuntimeRepository) Exists(ctx context.Context, tenant string, id string) (bool, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAll provides a mock function with given fields: ctx, tenant, filter
func (_m *RuntimeRepository) ListAll(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error) {
	ret := _m.Called(ctx, tenant, filter)

	var r0 []*model.Runtime
	if rf, ok := ret.Get(0).(func(context.Context, string, []*labelfilter.LabelFilter) []*model.Runtime); ok {
		r0 = rf(ctx, tenant, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Runtime)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []*labelfilter.LabelFilter) error); ok {
		r1 = rf(ctx, tenant, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListOwnedRuntimes provides a mock function with given fields: ctx, tenant, filter
func (_m *RuntimeRepository) ListOwnedRuntimes(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Runtime, error) {
	ret := _m.Called(ctx, tenant, filter)

	var r0 []*model.Runtime
	if rf, ok := ret.Get(0).(func(context.Context, string, []*labelfilter.LabelFilter) []*model.Runtime); ok {
		r0 = rf(ctx, tenant, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Runtime)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []*labelfilter.LabelFilter) error); ok {
		r1 = rf(ctx, tenant, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// OwnerExistsByFiltersAndID provides a mock function with given fields: ctx, tenant, id, filter
func (_m *RuntimeRepository) OwnerExistsByFiltersAndID(ctx context.Context, tenant string, id string, filter []*labelfilter.LabelFilter) (bool, error) {
	ret := _m.Called(ctx, tenant, id, filter)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string, []*labelfilter.LabelFilter) bool); ok {
		r0 = rf(ctx, tenant, id, filter)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, []*labelfilter.LabelFilter) error); ok {
		r1 = rf(ctx, tenant, id, filter)
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
