// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	labelfilter "github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// ApplicationTemplateService is an autogenerated mock type for the ApplicationTemplateService type
type ApplicationTemplateService struct {
	mock.Mock
}

// Exists provides a mock function with given fields: ctx, id
func (_m *ApplicationTemplateService) Exists(ctx context.Context, id string) (bool, error) {
	ret := _m.Called(ctx, id)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string) bool); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByFilters provides a mock function with given fields: ctx, filter
func (_m *ApplicationTemplateService) GetByFilters(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.ApplicationTemplate, error) {
	ret := _m.Called(ctx, filter)

	var r0 *model.ApplicationTemplate
	if rf, ok := ret.Get(0).(func(context.Context, []*labelfilter.LabelFilter) *model.ApplicationTemplate); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationTemplate)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []*labelfilter.LabelFilter) error); ok {
		r1 = rf(ctx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PrepareApplicationCreateInputJSON provides a mock function with given fields: appTemplate, values
func (_m *ApplicationTemplateService) PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, values model.ApplicationFromTemplateInputValues) (string, error) {
	ret := _m.Called(appTemplate, values)

	var r0 string
	if rf, ok := ret.Get(0).(func(*model.ApplicationTemplate, model.ApplicationFromTemplateInputValues) string); ok {
		r0 = rf(appTemplate, values)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.ApplicationTemplate, model.ApplicationFromTemplateInputValues) error); ok {
		r1 = rf(appTemplate, values)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewApplicationTemplateService creates a new instance of ApplicationTemplateService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewApplicationTemplateService(t testing.TB) *ApplicationTemplateService {
	mock := &ApplicationTemplateService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
