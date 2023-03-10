// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// AsaService is an autogenerated mock type for the asaService type
type AsaService struct {
	mock.Mock
}

// GetForScenarioName provides a mock function with given fields: ctx, scenarioName
func (_m *AsaService) GetForScenarioName(ctx context.Context, scenarioName string) (model.AutomaticScenarioAssignment, error) {
	ret := _m.Called(ctx, scenarioName)

	var r0 model.AutomaticScenarioAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string) model.AutomaticScenarioAssignment); ok {
		r0 = rf(ctx, scenarioName)
	} else {
		r0 = ret.Get(0).(model.AutomaticScenarioAssignment)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, scenarioName)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, pageSize, cursor
func (_m *AsaService) List(ctx context.Context, pageSize int, cursor string) (*model.AutomaticScenarioAssignmentPage, error) {
	ret := _m.Called(ctx, pageSize, cursor)

	var r0 *model.AutomaticScenarioAssignmentPage
	if rf, ok := ret.Get(0).(func(context.Context, int, string) *model.AutomaticScenarioAssignmentPage); ok {
		r0 = rf(ctx, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.AutomaticScenarioAssignmentPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int, string) error); ok {
		r1 = rf(ctx, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListForTargetTenant provides a mock function with given fields: ctx, targetTenantInternalID
func (_m *AsaService) ListForTargetTenant(ctx context.Context, targetTenantInternalID string) ([]*model.AutomaticScenarioAssignment, error) {
	ret := _m.Called(ctx, targetTenantInternalID)

	var r0 []*model.AutomaticScenarioAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.AutomaticScenarioAssignment); ok {
		r0 = rf(ctx, targetTenantInternalID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.AutomaticScenarioAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, targetTenantInternalID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewAsaService interface {
	mock.TestingT
	Cleanup(func())
}

// NewAsaService creates a new instance of AsaService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewAsaService(t mockConstructorTestingTNewAsaService) *AsaService {
	mock := &AsaService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
