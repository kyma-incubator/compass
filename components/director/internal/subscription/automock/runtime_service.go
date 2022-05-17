// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	labelfilter "github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// RuntimeService is an autogenerated mock type for the RuntimeService type
type RuntimeService struct {
	mock.Mock
}

// ListByFilters provides a mock function with given fields: _a0, _a1
func (_m *RuntimeService) ListByFilters(_a0 context.Context, _a1 []*labelfilter.LabelFilter) ([]*model.Runtime, error) {
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

// NewRuntimeService creates a new instance of RuntimeService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewRuntimeService(t testing.TB) *RuntimeService {
	mock := &RuntimeService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
