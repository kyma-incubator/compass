// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// BundleService is an autogenerated mock type for the BundleService type
type BundleService struct {
	mock.Mock
}

// CreateMultiple provides a mock function with given fields: ctx, applicationID, in
func (_m *BundleService) CreateMultiple(ctx context.Context, applicationID string, in []*model.BundleCreateInput) error {
	ret := _m.Called(ctx, applicationID, in)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []*model.BundleCreateInput) error); ok {
		r0 = rf(ctx, applicationID, in)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetForApplication provides a mock function with given fields: ctx, id, applicationID
func (_m *BundleService) GetForApplication(ctx context.Context, id string, applicationID string) (*model.Bundle, error) {
	ret := _m.Called(ctx, id, applicationID)

	var r0 *model.Bundle
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Bundle); ok {
		r0 = rf(ctx, id, applicationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Bundle)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, applicationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByApplicationIDs provides a mock function with given fields: ctx, applicationIDs, pageSize, cursor
func (_m *BundleService) ListByApplicationIDs(ctx context.Context, applicationIDs []string, pageSize int, cursor string) ([]*model.BundlePage, error) {
	ret := _m.Called(ctx, applicationIDs, pageSize, cursor)

	var r0 []*model.BundlePage
	if rf, ok := ret.Get(0).(func(context.Context, []string, int, string) []*model.BundlePage); ok {
		r0 = rf(ctx, applicationIDs, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.BundlePage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string, int, string) error); ok {
		r1 = rf(ctx, applicationIDs, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewBundleService interface {
	mock.TestingT
	Cleanup(func())
}

// NewBundleService creates a new instance of BundleService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewBundleService(t mockConstructorTestingTNewBundleService) *BundleService {
	mock := &BundleService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
