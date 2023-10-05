// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// LabelInputBuilder is an autogenerated mock type for the labelInputBuilder type
type LabelInputBuilder struct {
	mock.Mock
}

// GetLabelsForObject provides a mock function with given fields: ctx, tenant, objectID, objectType
func (_m *LabelInputBuilder) GetLabelsForObject(ctx context.Context, tenant string, objectID string, objectType model.LabelableObject) (map[string]string, error) {
	ret := _m.Called(ctx, tenant, objectID, objectType)

	var r0 map[string]string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.LabelableObject) (map[string]string, error)); ok {
		return rf(ctx, tenant, objectID, objectType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.LabelableObject) map[string]string); ok {
		r0 = rf(ctx, tenant, objectID, objectType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string, model.LabelableObject) error); ok {
		r1 = rf(ctx, tenant, objectID, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLabelsForObjects provides a mock function with given fields: ctx, tenant, objectIDs, objectType
func (_m *LabelInputBuilder) GetLabelsForObjects(ctx context.Context, tenant string, objectIDs []string, objectType model.LabelableObject) (map[string]map[string]string, error) {
	ret := _m.Called(ctx, tenant, objectIDs, objectType)

	var r0 map[string]map[string]string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, model.LabelableObject) (map[string]map[string]string, error)); ok {
		return rf(ctx, tenant, objectIDs, objectType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, model.LabelableObject) map[string]map[string]string); ok {
		r0 = rf(ctx, tenant, objectIDs, objectType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]map[string]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []string, model.LabelableObject) error); ok {
		r1 = rf(ctx, tenant, objectIDs, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewLabelInputBuilder creates a new instance of LabelInputBuilder. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLabelInputBuilder(t interface {
	mock.TestingT
	Cleanup(func())
}) *LabelInputBuilder {
	mock := &LabelInputBuilder{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
