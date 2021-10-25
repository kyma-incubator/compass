// Code generated by mockery v2.9.4. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// ScenariosService is an autogenerated mock type for the ScenariosService type
type ScenariosService struct {
	mock.Mock
}

// AddDefaultScenarioIfEnabled provides a mock function with given fields: ctx, labels
func (_m *ScenariosService) AddDefaultScenarioIfEnabled(ctx context.Context, labels *map[string]interface{}) {
	_m.Called(ctx, labels)
}

// EnsureScenariosLabelDefinitionExists provides a mock function with given fields: ctx, tenant
func (_m *ScenariosService) EnsureScenariosLabelDefinitionExists(ctx context.Context, tenant string) error {
	ret := _m.Called(ctx, tenant)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, tenant)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
