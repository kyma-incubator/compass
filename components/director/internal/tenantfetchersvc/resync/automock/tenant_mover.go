// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// TenantMover is an autogenerated mock type for the TenantMover type
type TenantMover struct {
	mock.Mock
}

// MoveTenants provides a mock function with given fields: ctx, movedSubaccountMappings
func (_m *TenantMover) MoveTenants(ctx context.Context, movedSubaccountMappings []model.MovedSubaccountMappingInput) error {
	ret := _m.Called(ctx, movedSubaccountMappings)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []model.MovedSubaccountMappingInput) error); ok {
		r0 = rf(ctx, movedSubaccountMappings)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TenantsToMove provides a mock function with given fields: ctx, region, fromTimestamp
func (_m *TenantMover) TenantsToMove(ctx context.Context, region string, fromTimestamp string) ([]model.MovedSubaccountMappingInput, error) {
	ret := _m.Called(ctx, region, fromTimestamp)

	var r0 []model.MovedSubaccountMappingInput
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []model.MovedSubaccountMappingInput); ok {
		r0 = rf(ctx, region, fromTimestamp)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.MovedSubaccountMappingInput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, region, fromTimestamp)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewTenantMover creates a new instance of TenantMover. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewTenantMover(t testing.TB) *TenantMover {
	mock := &TenantMover{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
