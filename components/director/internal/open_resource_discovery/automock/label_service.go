// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// LabelService is an autogenerated mock type for the LabelService type
type LabelService struct {
	mock.Mock
}

// GetByKey provides a mock function with given fields: ctx, tenant, objectType, objectID, key
func (_m *LabelService) GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, key string) (*model.Label, error) {
	ret := _m.Called(ctx, tenant, objectType, objectID, key)

	var r0 *model.Label
	if rf, ok := ret.Get(0).(func(context.Context, string, model.LabelableObject, string, string) *model.Label); ok {
		r0 = rf(ctx, tenant, objectType, objectID, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Label)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, model.LabelableObject, string, string) error); ok {
		r1 = rf(ctx, tenant, objectType, objectID, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewLabelService interface {
	mock.TestingT
	Cleanup(func())
}

// NewLabelService creates a new instance of LabelService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewLabelService(t mockConstructorTestingTNewLabelService) *LabelService {
	mock := &LabelService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
