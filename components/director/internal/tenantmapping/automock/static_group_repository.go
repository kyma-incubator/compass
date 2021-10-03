// Code generated by mockery 2.9.0. DO NOT EDIT.

package automock

import (
	context "context"

	tenantmapping "github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	mock "github.com/stretchr/testify/mock"
)

// StaticGroupRepository is an autogenerated mock type for the StaticGroupRepository type
type StaticGroupRepository struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, groupnames
func (_m *StaticGroupRepository) Get(ctx context.Context, groupnames []string) tenantmapping.StaticGroups {
	ret := _m.Called(ctx, groupnames)

	var r0 tenantmapping.StaticGroups
	if rf, ok := ret.Get(0).(func(context.Context, []string) tenantmapping.StaticGroups); ok {
		r0 = rf(ctx, groupnames)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(tenantmapping.StaticGroups)
		}
	}

	return r0
}
