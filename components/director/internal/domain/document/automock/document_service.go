// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// DocumentService is an autogenerated mock type for the DocumentService type
type DocumentService struct {
	mock.Mock
}

// CreateInBundle provides a mock function with given fields: ctx, resourceType, resourceID, bundleID, in
func (_m *DocumentService) CreateInBundle(ctx context.Context, resourceType resource.Type, resourceID string, bundleID string, in model.DocumentInput) (string, error) {
	ret := _m.Called(ctx, resourceType, resourceID, bundleID, in)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, string, model.DocumentInput) (string, error)); ok {
		return rf(ctx, resourceType, resourceID, bundleID, in)
	}
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, string, model.DocumentInput) string); ok {
		r0 = rf(ctx, resourceType, resourceID, bundleID, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string, string, model.DocumentInput) error); ok {
		r1 = rf(ctx, resourceType, resourceID, bundleID, in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *DocumentService) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Get provides a mock function with given fields: ctx, id
func (_m *DocumentService) Get(ctx context.Context, id string) (*model.Document, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Document
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Document, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Document); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Document)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListFetchRequests provides a mock function with given fields: ctx, documentIDs
func (_m *DocumentService) ListFetchRequests(ctx context.Context, documentIDs []string) ([]*model.FetchRequest, error) {
	ret := _m.Called(ctx, documentIDs)

	var r0 []*model.FetchRequest
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []string) ([]*model.FetchRequest, error)); ok {
		return rf(ctx, documentIDs)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []string) []*model.FetchRequest); ok {
		r0 = rf(ctx, documentIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FetchRequest)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, documentIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewDocumentService creates a new instance of DocumentService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDocumentService(t interface {
	mock.TestingT
	Cleanup(func())
}) *DocumentService {
	mock := &DocumentService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
