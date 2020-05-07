// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gqlschema "github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	mock "github.com/stretchr/testify/mock"
)

// DirectorClient is an autogenerated mock type for the DirectorClient type
type DirectorClient struct {
	mock.Mock
}

// CreateRuntime provides a mock function with given fields: config, tenant
func (_m *DirectorClient) CreateRuntime(config *gqlschema.RuntimeInput, tenant string) (string, error) {
	ret := _m.Called(config, tenant)

	var r0 string
	if rf, ok := ret.Get(0).(func(*gqlschema.RuntimeInput, string) string); ok {
		r0 = rf(config, tenant)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*gqlschema.RuntimeInput, string) error); ok {
		r1 = rf(config, tenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// DeleteRuntime provides a mock function with given fields: id, tenant
func (_m *DirectorClient) DeleteRuntime(id string, tenant string) error {
	ret := _m.Called(id, tenant)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, string) error); ok {
		r0 = rf(id, tenant)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetConnectionToken provides a mock function with given fields: id, tenant
func (_m *DirectorClient) GetConnectionToken(id string, tenant string) (graphql.OneTimeTokenForRuntimeExt, error) {
	ret := _m.Called(id, tenant)

	var r0 graphql.OneTimeTokenForRuntimeExt
	if rf, ok := ret.Get(0).(func(string, string) graphql.OneTimeTokenForRuntimeExt); ok {
		r0 = rf(id, tenant)
	} else {
		r0 = ret.Get(0).(graphql.OneTimeTokenForRuntimeExt)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(id, tenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetRuntime provides a mock function with given fields: id, tenant
func (_m *DirectorClient) GetRuntime(id string, tenant string) (graphql.RuntimeExt, error) {
	ret := _m.Called(id, tenant)

	var r0 graphql.RuntimeExt
	if rf, ok := ret.Get(0).(func(string, string) graphql.RuntimeExt); ok {
		r0 = rf(id, tenant)
	} else {
		r0 = ret.Get(0).(graphql.RuntimeExt)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string, string) error); ok {
		r1 = rf(id, tenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetRuntimeStatusCondition provides a mock function with given fields: id, statusCondition, tenant
func (_m *DirectorClient) SetRuntimeStatusCondition(id string, statusCondition graphql.RuntimeStatusCondition, tenant string) error {
	ret := _m.Called(id, statusCondition, tenant)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, graphql.RuntimeStatusCondition, string) error); ok {
		r0 = rf(id, statusCondition, tenant)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateRuntime provides a mock function with given fields: id, config, tenant
func (_m *DirectorClient) UpdateRuntime(id string, config *graphql.RuntimeInput, tenant string) error {
	ret := _m.Called(id, config, tenant)

	var r0 error
	if rf, ok := ret.Get(0).(func(string, *graphql.RuntimeInput, string) error); ok {
		r0 = rf(id, config, tenant)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
