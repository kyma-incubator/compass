// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import context "context"
import labelfilter "github.com/kyma-incubator/compass/components/director/internal/labelfilter"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// ApplicationRepository is an autogenerated mock type for the ApplicationRepository type
type ApplicationRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, item
func (_m *ApplicationRepository) Create(ctx context.Context, item *model.Application) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Application) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, item
func (_m *ApplicationRepository) Delete(ctx context.Context, item *model.Application) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Application) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: ctx, tenant, id
func (_m *ApplicationRepository) Exists(ctx context.Context, tenant string, id string) (bool, error) {
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

// GetByID provides a mock function with given fields: ctx, tenant, id
func (_m *ApplicationRepository) GetByID(ctx context.Context, tenant string, id string) (*model.Application, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 *model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Application); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Application)
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

// List provides a mock function with given fields: ctx, tenant, filter, pageSize, cursor
func (_m *ApplicationRepository) List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	ret := _m.Called(ctx, tenant, filter, pageSize, cursor)

	var r0 *model.ApplicationPage
	if rf, ok := ret.Get(0).(func(context.Context, string, []*labelfilter.LabelFilter, *int, *string) *model.ApplicationPage); ok {
		r0 = rf(ctx, tenant, filter, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []*labelfilter.LabelFilter, *int, *string) error); ok {
		r1 = rf(ctx, tenant, filter, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByRuntimeID provides a mock function with given fields: ctx, tenant, runtimeID, pageSize, cursor
func (_m *ApplicationRepository) ListByRuntimeID(ctx context.Context, tenant string, runtimeID string, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	ret := _m.Called(ctx, tenant, runtimeID, pageSize, cursor)

	var r0 *model.ApplicationPage
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *int, *string) *model.ApplicationPage); ok {
		r0 = rf(ctx, tenant, runtimeID, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, *int, *string) error); ok {
		r1 = rf(ctx, tenant, runtimeID, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, item
func (_m *ApplicationRepository) Update(ctx context.Context, item *model.Application) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Application) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
