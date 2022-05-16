// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// FetchRequestService is an autogenerated mock type for the FetchRequestService type
type FetchRequestService struct {
	mock.Mock
}

// HandleSpec provides a mock function with given fields: ctx, fr
func (_m *FetchRequestService) HandleSpec(ctx context.Context, fr *model.FetchRequest) *string {
	ret := _m.Called(ctx, fr)

	var r0 *string
	if rf, ok := ret.Get(0).(func(context.Context, *model.FetchRequest) *string); ok {
		r0 = rf(ctx, fr)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*string)
		}
	}

	return r0
}

// NewFetchRequestService creates a new instance of FetchRequestService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewFetchRequestService(t testing.TB) *FetchRequestService {
	mock := &FetchRequestService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
