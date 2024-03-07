// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	mock "github.com/stretchr/testify/mock"
)

// ValidatorClient is an autogenerated mock type for the ValidatorClient type
type ValidatorClient struct {
	mock.Mock
}

// Validate provides a mock function with given fields: ctx, ruleset, requestBody
func (_m *ValidatorClient) Validate(ctx context.Context, ruleset string, requestBody string) ([]ord.ValidationResult, error) {
	ret := _m.Called(ctx, ruleset, requestBody)

	var r0 []ord.ValidationResult
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) ([]ord.ValidationResult, error)); ok {
		return rf(ctx, ruleset, requestBody)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []ord.ValidationResult); ok {
		r0 = rf(ctx, ruleset, requestBody)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]ord.ValidationResult)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, ruleset, requestBody)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewValidatorClient creates a new instance of ValidatorClient. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewValidatorClient(t interface {
	mock.TestingT
	Cleanup(func())
}) *ValidatorClient {
	mock := &ValidatorClient{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
