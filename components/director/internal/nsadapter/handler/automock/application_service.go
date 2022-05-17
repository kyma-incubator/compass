// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	labelfilter "github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// ApplicationService is an autogenerated mock type for the applicationService type
type ApplicationService struct {
	mock.Mock
}

// CreateFromTemplate provides a mock function with given fields: ctx, in, appTemplateID
func (_m *ApplicationService) CreateFromTemplate(ctx context.Context, in model.ApplicationRegisterInput, appTemplateID *string) (string, error) {
	ret := _m.Called(ctx, in, appTemplateID)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, model.ApplicationRegisterInput, *string) string); ok {
		r0 = rf(ctx, in, appTemplateID)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.ApplicationRegisterInput, *string) error); ok {
		r1 = rf(ctx, in, appTemplateID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetLabel provides a mock function with given fields: ctx, applicationID, key
func (_m *ApplicationService) GetLabel(ctx context.Context, applicationID string, key string) (*model.Label, error) {
	ret := _m.Called(ctx, applicationID, key)

	var r0 *model.Label
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Label); ok {
		r0 = rf(ctx, applicationID, key)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Label)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, applicationID, key)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetSccSystem provides a mock function with given fields: ctx, sccSubaccount, locationID, virtualHost
func (_m *ApplicationService) GetSccSystem(ctx context.Context, sccSubaccount string, locationID string, virtualHost string) (*model.Application, error) {
	ret := _m.Called(ctx, sccSubaccount, locationID, virtualHost)

	var r0 *model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *model.Application); ok {
		r0 = rf(ctx, sccSubaccount, locationID, virtualHost)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, sccSubaccount, locationID, virtualHost)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListBySCC provides a mock function with given fields: ctx, filter
func (_m *ApplicationService) ListBySCC(ctx context.Context, filter *labelfilter.LabelFilter) ([]*model.ApplicationWithLabel, error) {
	ret := _m.Called(ctx, filter)

	var r0 []*model.ApplicationWithLabel
	if rf, ok := ret.Get(0).(func(context.Context, *labelfilter.LabelFilter) []*model.ApplicationWithLabel); ok {
		r0 = rf(ctx, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.ApplicationWithLabel)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *labelfilter.LabelFilter) error); ok {
		r1 = rf(ctx, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListSCCs provides a mock function with given fields: ctx
func (_m *ApplicationService) ListSCCs(ctx context.Context) ([]*model.SccMetadata, error) {
	ret := _m.Called(ctx)

	var r0 []*model.SccMetadata
	if rf, ok := ret.Get(0).(func(context.Context) []*model.SccMetadata); ok {
		r0 = rf(ctx)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.SccMetadata)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context) error); ok {
		r1 = rf(ctx)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// SetLabel provides a mock function with given fields: ctx, label
func (_m *ApplicationService) SetLabel(ctx context.Context, label *model.LabelInput) error {
	ret := _m.Called(ctx, label)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.LabelInput) error); ok {
		r0 = rf(ctx, label)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Update provides a mock function with given fields: ctx, id, in
func (_m *ApplicationService) Update(ctx context.Context, id string, in model.ApplicationUpdateInput) error {
	ret := _m.Called(ctx, id, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.ApplicationUpdateInput) error); ok {
		r0 = rf(ctx, id, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Upsert provides a mock function with given fields: ctx, in
func (_m *ApplicationService) Upsert(ctx context.Context, in model.ApplicationRegisterInput) error {
	ret := _m.Called(ctx, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.ApplicationRegisterInput) error); ok {
		r0 = rf(ctx, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// NewApplicationService creates a new instance of ApplicationService. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewApplicationService(t testing.TB) *ApplicationService {
	mock := &ApplicationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
