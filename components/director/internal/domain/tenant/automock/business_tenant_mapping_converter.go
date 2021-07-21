// Code generated by mockery 2.9.0. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// BusinessTenantMappingConverter is an autogenerated mock type for the BusinessTenantMappingConverter type
type BusinessTenantMappingConverter struct {
	mock.Mock
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *BusinessTenantMappingConverter) MultipleToGraphQL(in []*model.BusinessTenantMapping) []*graphql.Tenant {
	ret := _m.Called(in)

	var r0 []*graphql.Tenant
	if rf, ok := ret.Get(0).(func([]*model.BusinessTenantMapping) []*graphql.Tenant); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.Tenant)
		}
	}

	return r0
}
