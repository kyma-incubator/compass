// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// ScenariosDefService is an autogenerated mock type for the ScenariosDefService type
type ScenariosDefService struct {
	mock.Mock
}

// GetAvailableScenarios provides a mock function with given fields: ctx, tenantID
func (_m *ScenariosDefService) GetAvailableScenarios(ctx context.Context, tenantID string) ([]string, error) {
	ret := _m.Called(ctx, tenantID)

	var r0 []string
	if rf, ok := ret.Get(0).(func(context.Context, string) []string); ok {
		r0 = rf(ctx, tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewScenariosDefService interface {
	mock.TestingT
	Cleanup(func())
}

// NewScenariosDefService creates a new instance of ScenariosDefService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewScenariosDefService(t mockConstructorTestingTNewScenariosDefService) *ScenariosDefService {
	mock := &ScenariosDefService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
