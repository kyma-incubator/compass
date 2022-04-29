// Code generated by mockery v2.12.1. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// TenantRepository is an autogenerated mock type for the TenantRepository type
type TenantRepository struct {
	mock.Mock
}

// GetByExternalTenant provides a mock function with given fields: ctx, externalTenant
func (_m *TenantRepository) GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error) {
	ret := _m.Called(ctx, externalTenant)

	var r0 *model.BusinessTenantMapping
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.BusinessTenantMapping); ok {
		r0 = rf(ctx, externalTenant)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BusinessTenantMapping)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, externalTenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewTenantRepository creates a new instance of TenantRepository. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewTenantRepository(t testing.TB) *TenantRepository {
	mock := &TenantRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
