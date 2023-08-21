// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"
	http "net/http"

	mock "github.com/stretchr/testify/mock"

	sync "sync"
)

// Executor is an autogenerated mock type for the Executor type
type Executor struct {
	mock.Mock
}

// Execute provides a mock function with given fields: ctx, client, url, tnt, additionalHeaders
func (_m *Executor) Execute(ctx context.Context, client *http.Client, url string, tnt string, additionalHeaders sync.Map) (*http.Response, error) {
	ret := _m.Called(ctx, client, url, tnt, additionalHeaders)

	var r0 *http.Response
	if rf, ok := ret.Get(0).(func(context.Context, *http.Client, string, string, sync.Map) *http.Response); ok {
		r0 = rf(ctx, client, url, tnt, additionalHeaders)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*http.Response)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *http.Client, string, string, sync.Map) error); ok {
		r1 = rf(ctx, client, url, tnt, additionalHeaders)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewExecutor interface {
	mock.TestingT
	Cleanup(func())
}

// NewExecutor creates a new instance of Executor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewExecutor(t mockConstructorTestingTNewExecutor) *Executor {
	mock := &Executor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
