// Code generated by mockery v1.0.0
package automock

import context "context"
import labelfilter "github.com/kyma-incubator/compass/components/director/internal/labelfilter"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// ApplicationService is an autogenerated mock type for the ApplicationService type
type ApplicationService struct {
	mock.Mock
}

// AddAnnotation provides a mock function with given fields: ctx, runtimeID, key, value
func (_m *ApplicationService) AddAnnotation(ctx context.Context, runtimeID string, key string, value interface{}) error {
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
func (_m *ApplicationService) AddLabel(ctx context.Context, runtimeID string, key string, values []string) error {
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
func (_m *ApplicationService) Create(ctx context.Context, in model.ApplicationInput) (string, error) {
	ret := _m.Called(ctx, in)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, model.ApplicationInput) string); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.ApplicationInput) error); ok {
		r1 = rf(ctx, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *ApplicationService) Delete(ctx context.Context, id string) error {
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
func (_m *ApplicationService) DeleteAnnotation(ctx context.Context, runtimeID string, key string) error {
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
func (_m *ApplicationService) DeleteLabel(ctx context.Context, runtimeID string, key string, values []string) error {
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
func (_m *ApplicationService) Get(ctx context.Context, id string) (*model.Application, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Application); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Application)
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
func (_m *ApplicationService) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	ret := _m.Called(ctx, filter, pageSize, cursor)

	var r0 *model.ApplicationPage
	if rf, ok := ret.Get(0).(func(context.Context, []*labelfilter.LabelFilter, *int, *string) *model.ApplicationPage); ok {
		r0 = rf(ctx, filter, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationPage)
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
func (_m *ApplicationService) Update(ctx context.Context, id string, in model.ApplicationInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.ApplicationInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
