// Code generated by mockery v2.5.1. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// LabelDefinitionService is an autogenerated mock type for the LabelDefinitionService type
type LabelDefinitionService struct {
	mock.Mock
}

// Upsert provides a mock function with given fields: ctx, def
func (_m *LabelDefinitionService) Upsert(ctx context.Context, def model.LabelDefinition) error {
	ret := _m.Called(ctx, def)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.LabelDefinition) error); ok {
		r0 = rf(ctx, def)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
