// Code generated by mockery v2.9.4. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// IntegrationSystemService is an autogenerated mock type for the IntegrationSystemService type
type IntegrationSystemService struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, in
func (_m *IntegrationSystemService) Create(ctx context.Context, in model.IntegrationSystemInput) (string, error) {
	ret := _m.Called(ctx, in)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, model.IntegrationSystemInput) string); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.IntegrationSystemInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *IntegrationSystemService) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, id
func (_m *IntegrationSystemService) Get(ctx context.Context, id string) (*model.IntegrationSystem, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.IntegrationSystem
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.IntegrationSystem); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.IntegrationSystem)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, pageSize, cursor
func (_m *IntegrationSystemService) List(ctx context.Context, pageSize int, cursor string) (model.IntegrationSystemPage, error) {
	ret := _m.Called(ctx, pageSize, cursor)

	var r0 model.IntegrationSystemPage
	if rf, ok := ret.Get(0).(func(context.Context, int, string) model.IntegrationSystemPage); ok {
		r0 = rf(ctx, pageSize, cursor)
	} else {
		r0 = ret.Get(0).(model.IntegrationSystemPage)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int, string) error); ok {
		r1 = rf(ctx, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, in
func (_m *IntegrationSystemService) Update(ctx context.Context, id string, in model.IntegrationSystemInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.IntegrationSystemInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
