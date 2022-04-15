// Code generated by mockery v2.10.4. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// LabelDefService is an autogenerated mock type for the labelDefService type
type LabelDefService struct {
	mock.Mock
}

// CreateWithFormations provides a mock function with given fields: ctx, tnt, formations
func (_m *LabelDefService) CreateWithFormations(ctx context.Context, tnt string, formations []string) error {
	ret := _m.Called(ctx, tnt, formations)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) error); ok {
		r0 = rf(ctx, tnt, formations)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateAutomaticScenarioAssignmentAgainstSchema provides a mock function with given fields: ctx, schema, tenantID, key
func (_m *LabelDefService) ValidateAutomaticScenarioAssignmentAgainstSchema(ctx context.Context, schema interface{}, tenantID string, key string) error {
	ret := _m.Called(ctx, schema, tenantID, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, string, string) error); ok {
		r0 = rf(ctx, schema, tenantID, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ValidateExistingLabelsAgainstSchema provides a mock function with given fields: ctx, schema, tenant, key
func (_m *LabelDefService) ValidateExistingLabelsAgainstSchema(ctx context.Context, schema interface{}, tenant string, key string) error {
	ret := _m.Called(ctx, schema, tenant, key)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, interface{}, string, string) error); ok {
		r0 = rf(ctx, schema, tenant, key)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
