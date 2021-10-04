// Code generated by mockery (devel). DO NOT EDIT.

package automock

import (
	context "context"

	tenantfetchersvc "github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc"
	mock "github.com/stretchr/testify/mock"
)

// TenantProvisioner is an autogenerated mock type for the TenantProvisioner type
type TenantProvisioner struct {
	mock.Mock
}

// ProvisionTenants provides a mock function with given fields: _a0, _a1, _a2
func (_m *TenantProvisioner) ProvisionTenants(_a0 context.Context, _a1 *tenantfetchersvc.TenantSubscriptionRequest, _a2 string) error {
	ret := _m.Called(_a0, _a1, _a2)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *tenantfetchersvc.TenantSubscriptionRequest, string) error); ok {
		r0 = rf(_a0, _a1, _a2)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
