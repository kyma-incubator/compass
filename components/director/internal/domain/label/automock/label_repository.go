// Code generated by mockery v2.9.4. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// LabelRepository is an autogenerated mock type for the LabelRepository type
type LabelRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, tenant, _a2
func (_m *LabelRepository) Create(ctx context.Context, tenant string, _a2 *model.Label) error {
	ret := _m.Called(ctx, tenant, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Label) error); ok {
		r0 = rf(ctx, tenant, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByKey provides a mock function with given fields: ctx, tenant, objectType, objectID, key
func (_m *LabelRepository) GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) (*model.Label, error) {
	ret := _m.Called(ctx, tenant, objectType, objectID, key)

	var r0 *model.Label
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string, string) *model.Label); ok {
		r0 = rf(ctx, tenant, objectType, objectID, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Label)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, model.LabelableObject, string, string) error); ok {
		r1 = rf(ctx, tenant, objectType, objectID, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateWithVersion provides a mock function with given fields: ctx, tenant, _a2
func (_m *LabelRepository) UpdateWithVersion(ctx context.Context, tenant string, _a2 *model.Label) error {
	ret := _m.Called(ctx, tenant, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Label) error); ok {
		r0 = rf(ctx, tenant, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Upsert provides a mock function with given fields: ctx, tenant, _a2
func (_m *LabelRepository) Upsert(ctx context.Context, tenant string, _a2 *model.Label) error {
	ret := _m.Called(ctx, tenant, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Label) error); ok {
		r0 = rf(ctx, tenant, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
