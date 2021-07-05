// Code generated by mockery v0.0.0-dev. DO NOT EDIT.

package automock

import (
	tenant "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	model "github.com/kyma-incubator/compass/components/tenant-fetcher/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// Converter is an autogenerated mock type for the Converter type
type Converter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: in
func (_m *Converter) FromEntity(in tenant.Entity) model.TenantModel {
	ret := _m.Called(in)

	var r0 model.TenantModel
	if rf, ok := ret.Get(0).(func(tenant.Entity) model.TenantModel); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(model.TenantModel)
	}

	return r0
}

// ToEntity provides a mock function with given fields: in
func (_m *Converter) ToEntity(in model.TenantModel) tenant.Entity {
	ret := _m.Called(in)

	var r0 tenant.Entity
	if rf, ok := ret.Get(0).(func(model.TenantModel) tenant.Entity); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(tenant.Entity)
	}

	return r0
}
