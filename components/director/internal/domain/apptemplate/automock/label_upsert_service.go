// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// LabelUpsertService is an autogenerated mock type for the LabelUpsertService type
type LabelUpsertService struct {
	mock.Mock
}

// UpsertLabel provides a mock function with given fields: ctx, tenant, labelInput
func (_m *LabelUpsertService) UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error {
	ret := _m.Called(ctx, tenant, labelInput)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.LabelInput) error); ok {
		r0 = rf(ctx, tenant, labelInput)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpsertMultipleLabels provides a mock function with given fields: ctx, tenant, objectType, objectID, labels
func (_m *LabelUpsertService) UpsertMultipleLabels(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labels map[string]interface{}) error {
	ret := _m.Called(ctx, tenant, objectType, objectID, labels)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string, map[string]interface{}) error); ok {
		r0 = rf(ctx, tenant, objectType, objectID, labels)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewLabelUpsertService interface {
	mock.TestingT
	Cleanup(func())
}

// NewLabelUpsertService creates a new instance of LabelUpsertService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewLabelUpsertService(t mockConstructorTestingTNewLabelUpsertService) *LabelUpsertService {
	mock := &LabelUpsertService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
