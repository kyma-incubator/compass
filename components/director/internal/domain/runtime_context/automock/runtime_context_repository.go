// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	labelfilter "github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// RuntimeContextRepository is an autogenerated mock type for the RuntimeContextRepository type
type RuntimeContextRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, tenant, item
func (_m *RuntimeContextRepository) Create(ctx context.Context, tenant string, item *model.RuntimeContext) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.RuntimeContext) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, tenant, id
func (_m *RuntimeContextRepository) Delete(ctx context.Context, tenant string, id string) error {
	ret := _m.Called(ctx, tenant, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: ctx, tenant, id
func (_m *RuntimeContextRepository) Exists(ctx context.Context, tenant string, id string) (bool, error) {
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

// GetByFiltersGlobal provides a mock function with given fields: ctx, filter
func (_m *RuntimeContextRepository) GetByFiltersGlobal(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.RuntimeContext, error) {
	ret := _m.Called(ctx, filter)

	var r0 *model.RuntimeContext
	if rf, ok := ret.Get(0).(func(context.Context, []*labelfilter.LabelFilter) *model.RuntimeContext); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RuntimeContext)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []*labelfilter.LabelFilter) error); ok {
		r1 = rf(ctx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: ctx, tenant, id
func (_m *RuntimeContextRepository) GetByID(ctx context.Context, tenant string, id string) (*model.RuntimeContext, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 *model.RuntimeContext
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.RuntimeContext); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RuntimeContext)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetForRuntime provides a mock function with given fields: ctx, tenant, id, runtimeID
func (_m *RuntimeContextRepository) GetForRuntime(ctx context.Context, tenant string, id string, runtimeID string) (*model.RuntimeContext, error) {
	ret := _m.Called(ctx, tenant, id, runtimeID)

	var r0 *model.RuntimeContext
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *model.RuntimeContext); ok {
		r0 = rf(ctx, tenant, id, runtimeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RuntimeContext)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, tenant, id, runtimeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, runtimeID, tenant, filter, pageSize, cursor
func (_m *RuntimeContextRepository) List(ctx context.Context, runtimeID string, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimeContextPage, error) {
	ret := _m.Called(ctx, runtimeID, tenant, filter, pageSize, cursor)

	var r0 *model.RuntimeContextPage
	if rf, ok := ret.Get(0).(func(context.Context, string, string, []*labelfilter.LabelFilter, int, string) *model.RuntimeContextPage); ok {
		r0 = rf(ctx, runtimeID, tenant, filter, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RuntimeContextPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, []*labelfilter.LabelFilter, int, string) error); ok {
		r1 = rf(ctx, runtimeID, tenant, filter, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllForRuntime provides a mock function with given fields: ctx, tenant, runtimeID
func (_m *RuntimeContextRepository) ListAllForRuntime(ctx context.Context, tenant string, runtimeID string) ([]*model.RuntimeContext, error) {
	ret := _m.Called(ctx, tenant, runtimeID)

	var r0 []*model.RuntimeContext
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []*model.RuntimeContext); ok {
		r0 = rf(ctx, tenant, runtimeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.RuntimeContext)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, runtimeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByRuntimeIDs provides a mock function with given fields: ctx, tenantID, runtimeIDs, pageSize, cursor
func (_m *RuntimeContextRepository) ListByRuntimeIDs(ctx context.Context, tenantID string, runtimeIDs []string, pageSize int, cursor string) ([]*model.RuntimeContextPage, error) {
	ret := _m.Called(ctx, tenantID, runtimeIDs, pageSize, cursor)

	var r0 []*model.RuntimeContextPage
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, int, string) []*model.RuntimeContextPage); ok {
		r0 = rf(ctx, tenantID, runtimeIDs, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.RuntimeContextPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string, int, string) error); ok {
		r1 = rf(ctx, tenantID, runtimeIDs, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, tenant, item
func (_m *RuntimeContextRepository) Update(ctx context.Context, tenant string, item *model.RuntimeContext) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.RuntimeContext) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewRuntimeContextRepository creates a new instance of RuntimeContextRepository. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewRuntimeContextRepository(t testing.TB) *RuntimeContextRepository {
	mock := &RuntimeContextRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
