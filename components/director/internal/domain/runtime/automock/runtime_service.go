// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	labelfilter "github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// RuntimeService is an autogenerated mock type for the RuntimeService type
type RuntimeService struct {
	mock.Mock
}

// CreateWithMandatoryLabels provides a mock function with given fields: ctx, in, id, mandatoryLabels
func (_m *RuntimeService) CreateWithMandatoryLabels(ctx context.Context, in model.RuntimeRegisterInput, id string, mandatoryLabels map[string]interface{}) error {
	ret := _m.Called(ctx, in, id, mandatoryLabels)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.RuntimeRegisterInput, string, map[string]interface{}) error); ok {
		r0 = rf(ctx, in, id, mandatoryLabels)
	} else {
		r0 = ret.Error(0)
	}

	return r0
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

// DeleteLabel provides a mock function with given fields: ctx, runtimeID, key
func (_m *RuntimeService) DeleteLabel(ctx context.Context, runtimeID string, key string) error {
	ret := _m.Called(ctx, runtimeID, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, runtimeID, key)
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

// GetByFilters provides a mock function with given fields: ctx, filters
func (_m *RuntimeService) GetByFilters(ctx context.Context, filters []*labelfilter.LabelFilter) (*model.Runtime, error) {
	ret := _m.Called(ctx, filters)

	var r0 *model.Runtime
	if rf, ok := ret.Get(0).(func(context.Context, []*labelfilter.LabelFilter) *model.Runtime); ok {
		r0 = rf(ctx, filters)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Runtime)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []*labelfilter.LabelFilter) error); ok {
		r1 = rf(ctx, filters)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByTokenIssuer provides a mock function with given fields: ctx, issuer
func (_m *RuntimeService) GetByTokenIssuer(ctx context.Context, issuer string) (*model.Runtime, error) {
	ret := _m.Called(ctx, issuer)

	var r0 *model.Runtime
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Runtime); ok {
		r0 = rf(ctx, issuer)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Runtime)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, issuer)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLabel provides a mock function with given fields: ctx, runtimeID, key
func (_m *RuntimeService) GetLabel(ctx context.Context, runtimeID string, key string) (*model.Label, error) {
	ret := _m.Called(ctx, runtimeID, key)

	var r0 *model.Label
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Label); ok {
		r0 = rf(ctx, runtimeID, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Label)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, runtimeID, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, filter, pageSize, cursor
func (_m *RuntimeService) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.RuntimePage, error) {
	ret := _m.Called(ctx, filter, pageSize, cursor)

	var r0 *model.RuntimePage
	if rf, ok := ret.Get(0).(func(context.Context, []*labelfilter.LabelFilter, int, string) *model.RuntimePage); ok {
		r0 = rf(ctx, filter, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.RuntimePage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []*labelfilter.LabelFilter, int, string) error); ok {
		r1 = rf(ctx, filter, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListLabels provides a mock function with given fields: ctx, runtimeID
func (_m *RuntimeService) ListLabels(ctx context.Context, runtimeID string) (map[string]*model.Label, error) {
	ret := _m.Called(ctx, runtimeID)

	var r0 map[string]*model.Label
	if rf, ok := ret.Get(0).(func(context.Context, string) map[string]*model.Label); ok {
		r0 = rf(ctx, runtimeID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]*model.Label)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, runtimeID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetLabel provides a mock function with given fields: ctx, label
func (_m *RuntimeService) SetLabel(ctx context.Context, label *model.LabelInput) error {
	ret := _m.Called(ctx, label)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.LabelInput) error); ok {
		r0 = rf(ctx, label)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UnsafeExtractModifiableLabels provides a mock function with given fields: labels
func (_m *RuntimeService) UnsafeExtractModifiableLabels(labels map[string]interface{}) (map[string]interface{}, error) {
	ret := _m.Called(labels)

	var r0 map[string]interface{}
	if rf, ok := ret.Get(0).(func(map[string]interface{}) map[string]interface{}); ok {
		r0 = rf(labels)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string]interface{})
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(map[string]interface{}) error); ok {
		r1 = rf(labels)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, id, in
func (_m *RuntimeService) Update(ctx context.Context, id string, in model.RuntimeUpdateInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.RuntimeUpdateInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewRuntimeService interface {
	mock.TestingT
	Cleanup(func())
}

// NewRuntimeService creates a new instance of RuntimeService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewRuntimeService(t mockConstructorTestingTNewRuntimeService) *RuntimeService {
	mock := &RuntimeService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
