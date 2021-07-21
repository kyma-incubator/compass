// Code generated by mockery 2.9.0. DO NOT EDIT.

package automock

import (
	model "github.com/kyma-incubator/compass/components/director/internal/model"
	tenant "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	mock "github.com/stretchr/testify/mock"
)

// Converter is an autogenerated mock type for the Converter type
type Converter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: in
func (_m *Converter) FromEntity(in *tenant.Entity) *model.BusinessTenantMapping {
	ret := _m.Called(in)

	var r0 *model.BusinessTenantMapping
	if rf, ok := ret.Get(0).(func(*tenant.Entity) *model.BusinessTenantMapping); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BusinessTenantMapping)
		}
	}

	return r0
}

// ToEntity provides a mock function with given fields: in
func (_m *Converter) ToEntity(in *model.BusinessTenantMapping) *tenant.Entity {
	ret := _m.Called(in)

	var r0 *tenant.Entity
	if rf, ok := ret.Get(0).(func(*model.BusinessTenantMapping) *tenant.Entity); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*tenant.Entity)
		}
	}

	return r0
}
