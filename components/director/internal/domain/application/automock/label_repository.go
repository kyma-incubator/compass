// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// LabelRepository is an autogenerated mock type for the LabelRepository type
type LabelRepository struct {
	mock.Mock
}

// Delete provides a mock function with given fields: ctx, tenant, objectType, objectID, key
func (_m *LabelRepository) Delete(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) error {
	ret := _m.Called(ctx, tenant, objectType, objectID, key)

	if len(ret) == 0 {
		panic("no return value specified for Delete")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string, string) error); ok {
		r0 = rf(ctx, tenant, objectType, objectID, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteAll provides a mock function with given fields: ctx, tenant, objectType, objectID
func (_m *LabelRepository) DeleteAll(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) error {
	ret := _m.Called(ctx, tenant, objectType, objectID)

	if len(ret) == 0 {
		panic("no return value specified for DeleteAll")
	}

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string) error); ok {
		r0 = rf(ctx, tenant, objectType, objectID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByKey provides a mock function with given fields: ctx, tenant, objectType, objectID, key
func (_m *LabelRepository) GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) (*model.Label, error) {
	ret := _m.Called(ctx, tenant, objectType, objectID, key)

	if len(ret) == 0 {
		panic("no return value specified for GetByKey")
	}

	var r0 *model.Label
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string, string) (*model.Label, error)); ok {
		return rf(ctx, tenant, objectType, objectID, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string, string) *model.Label); ok {
		r0 = rf(ctx, tenant, objectType, objectID, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Label)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, model.LabelableObject, string, string) error); ok {
		r1 = rf(ctx, tenant, objectType, objectID, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListForGlobalObject provides a mock function with given fields: ctx, objectType, objectID
func (_m *LabelRepository) ListForGlobalObject(ctx context.Context, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error) {
	ret := _m.Called(ctx, objectType, objectID)

	if len(ret) == 0 {
		panic("no return value specified for ListForGlobalObject")
	}

	var r0 map[string]*model.Label
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.LabelableObject, string) (map[string]*model.Label, error)); ok {
		return rf(ctx, objectType, objectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.LabelableObject, string) map[string]*model.Label); ok {
		r0 = rf(ctx, objectType, objectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]*model.Label)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.LabelableObject, string) error); ok {
		r1 = rf(ctx, objectType, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListForObject provides a mock function with given fields: ctx, tenant, objectType, objectID
func (_m *LabelRepository) ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error) {
	ret := _m.Called(ctx, tenant, objectType, objectID)

	if len(ret) == 0 {
		panic("no return value specified for ListForObject")
	}

	var r0 map[string]*model.Label
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string) (map[string]*model.Label, error)); ok {
		return rf(ctx, tenant, objectType, objectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string) map[string]*model.Label); ok {
		r0 = rf(ctx, tenant, objectType, objectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]*model.Label)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, model.LabelableObject, string) error); ok {
		r1 = rf(ctx, tenant, objectType, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListGlobalByKey provides a mock function with given fields: ctx, key
func (_m *LabelRepository) ListGlobalByKey(ctx context.Context, key string) ([]*model.Label, error) {
	ret := _m.Called(ctx, key)

	if len(ret) == 0 {
		panic("no return value specified for ListGlobalByKey")
	}

	var r0 []*model.Label
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) ([]*model.Label, error)); ok {
		return rf(ctx, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.Label); ok {
		r0 = rf(ctx, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Label)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListGlobalByKeyAndObjects provides a mock function with given fields: ctx, objectType, objectIDs, key
func (_m *LabelRepository) ListGlobalByKeyAndObjects(ctx context.Context, objectType model.LabelableObject, objectIDs []string, key string) ([]*model.Label, error) {
	ret := _m.Called(ctx, objectType, objectIDs, key)

	if len(ret) == 0 {
		panic("no return value specified for ListGlobalByKeyAndObjects")
	}

	var r0 []*model.Label
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.LabelableObject, []string, string) ([]*model.Label, error)); ok {
		return rf(ctx, objectType, objectIDs, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.LabelableObject, []string, string) []*model.Label); ok {
		r0 = rf(ctx, objectType, objectIDs, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Label)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.LabelableObject, []string, string) error); ok {
		r1 = rf(ctx, objectType, objectIDs, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewLabelRepository creates a new instance of LabelRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLabelRepository(t interface {
	mock.TestingT
	Cleanup(func())
}) *LabelRepository {
	mock := &LabelRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
