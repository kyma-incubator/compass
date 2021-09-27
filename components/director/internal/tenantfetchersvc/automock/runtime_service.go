// Code generated by mockery (devel). DO NOT EDIT.

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

// ListByFiltersGlobal provides a mock function with given fields: _a0, _a1
func (_m *RuntimeService) ListByFiltersGlobal(_a0 context.Context, _a1 []*labelfilter.LabelFilter) ([]*model.Runtime, error) {
	ret := _m.Called(_a0, _a1)

	var r0 []*model.Runtime
	if rf, ok := ret.Get(0).(func(context.Context, []*labelfilter.LabelFilter) []*model.Runtime); ok {
		r0 = rf(_a0, _a1)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Runtime)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []*labelfilter.LabelFilter) error); ok {
		r1 = rf(_a0, _a1)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetLabel provides a mock function with given fields: _a0, _a1
func (_m *RuntimeService) SetLabel(_a0 context.Context, _a1 *model.LabelInput) error {
	ret := _m.Called(_a0, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.LabelInput) error); ok {
		r0 = rf(_a0, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
