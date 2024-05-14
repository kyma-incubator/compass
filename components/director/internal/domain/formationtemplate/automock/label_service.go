// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// LabelService is an autogenerated mock type for the labelService type
type LabelService struct {
	mock.Mock
}

// Delete provides a mock function with given fields: ctx, tenantID, objectType, objectID, key
func (_m *LabelService) Delete(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID string, key string) error {
	ret := _m.Called(ctx, tenantID, objectType, objectID, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string, string) error); ok {
		r0 = rf(ctx, tenantID, objectType, objectID, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByKey provides a mock function with given fields: ctx, tenantID, objectType, objectID, key
func (_m *LabelService) GetByKey(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID string, key string) (*model.Label, error) {
	ret := _m.Called(ctx, tenantID, objectType, objectID, key)

	var r0 *model.Label
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string, string) (*model.Label, error)); ok {
		return rf(ctx, tenantID, objectType, objectID, key)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string, string) *model.Label); ok {
		r0 = rf(ctx, tenantID, objectType, objectID, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Label)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, model.LabelableObject, string, string) error); ok {
		r1 = rf(ctx, tenantID, objectType, objectID, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListForObject provides a mock function with given fields: ctx, tenantID, objectType, objectID
func (_m *LabelService) ListForObject(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error) {
	ret := _m.Called(ctx, tenantID, objectType, objectID)

	var r0 map[string]*model.Label
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string) (map[string]*model.Label, error)); ok {
		return rf(ctx, tenantID, objectType, objectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string) map[string]*model.Label); ok {
		r0 = rf(ctx, tenantID, objectType, objectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]*model.Label)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, model.LabelableObject, string) error); ok {
		r1 = rf(ctx, tenantID, objectType, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpsertLabel provides a mock function with given fields: ctx, tenantID, labelInput
func (_m *LabelService) UpsertLabel(ctx context.Context, tenantID string, labelInput *model.LabelInput) error {
	ret := _m.Called(ctx, tenantID, labelInput)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.LabelInput) error); ok {
		r0 = rf(ctx, tenantID, labelInput)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpsertMultipleLabels provides a mock function with given fields: ctx, tenantID, objectType, objectID, labels
func (_m *LabelService) UpsertMultipleLabels(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID string, labels map[string]interface{}) error {
	ret := _m.Called(ctx, tenantID, objectType, objectID, labels)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string, map[string]interface{}) error); ok {
		r0 = rf(ctx, tenantID, objectType, objectID, labels)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewLabelService creates a new instance of LabelService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewLabelService(t interface {
	mock.TestingT
	Cleanup(func())
}) *LabelService {
	mock := &LabelService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
