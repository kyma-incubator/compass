// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import context "context"
import labelfilter "github.com/kyma-incubator/compass/components/director/internal/labelfilter"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// RuntimeService is an autogenerated mock type for the RuntimeService type
type RuntimeService struct {
	mock.Mock
}

// AddAnnotation provides a mock function with given fields: ctx, runtimeID, key, value
func (_m *RuntimeService) AddAnnotation(ctx context.Context, runtimeID string, key string, value interface{}) error {
	ret := _m.Called(ctx, runtimeID, key, value)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, interface{}) error); ok {
		r0 = rf(ctx, runtimeID, key, value)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// AddLabel provides a mock function with given fields: ctx, runtimeID, key, values
func (_m *RuntimeService) AddLabel(ctx context.Context, runtimeID string, key string, values []string) error {
	ret := _m.Called(ctx, runtimeID, key, values)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, []string) error); ok {
		r0 = rf(ctx, runtimeID, key, values)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Create provides a mock function with given fields: ctx, in
func (_m *RuntimeService) Create(ctx context.Context, in model.RuntimeInput) (string, error) {
	ret := _m.Called(ctx, in)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, model.RuntimeInput) string); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.RuntimeInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *RuntimeService) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteAnnotation provides a mock function with given fields: ctx, runtimeID, key
func (_m *RuntimeService) DeleteAnnotation(ctx context.Context, runtimeID string, key string) error {
	ret := _m.Called(ctx, runtimeID, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, runtimeID, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteLabel provides a mock function with given fields: ctx, runtimeID, key, values
func (_m *RuntimeService) DeleteLabel(ctx context.Context, runtimeID string, key string, values []string) error {
	ret := _m.Called(ctx, runtimeID, key, values)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, []string) error); ok {
		r0 = rf(ctx, runtimeID, key, values)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, id
func (_m *RuntimeService) Get(ctx context.Context, id string) (*model.Runtime, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Runtime
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Runtime); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Runtime)
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

// List provides a mock function with given fields: ctx, filter, pageSize, cursor
func (_m *RuntimeService) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.RuntimePage, error) {
	ret := _m.Called(ctx, filter, pageSize, cursor)

	var r0 *model.RuntimePage
	if rf, ok := ret.Get(0).(func(context.Context, []*labelfilter.LabelFilter, *int, *string) *model.RuntimePage); ok {
		r0 = rf(ctx, filter, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RuntimePage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []*labelfilter.LabelFilter, *int, *string) error); ok {
		r1 = rf(ctx, filter, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, in
func (_m *RuntimeService) Update(ctx context.Context, id string, in model.RuntimeInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.RuntimeInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
