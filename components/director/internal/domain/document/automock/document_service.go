// Code generated by mockery v2.5.1. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// DocumentService is an autogenerated mock type for the DocumentService type
type DocumentService struct {
	mock.Mock
}

// CreateInBundle provides a mock function with given fields: ctx, bundleID, in
func (_m *DocumentService) CreateInBundle(ctx context.Context, bundleID string, in model.DocumentInput) (string, error) {
	ret := _m.Called(ctx, bundleID, in)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, model.DocumentInput) string); ok {
		r0 = rf(ctx, bundleID, in)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, model.DocumentInput) error); ok {
		r1 = rf(ctx, bundleID, in)
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
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Document); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Document)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetFetchRequest provides a mock function with given fields: ctx, documentID
func (_m *DocumentService) GetFetchRequest(ctx context.Context, documentID string) (*model.FetchRequest, error) {
	ret := _m.Called(ctx, documentID)

	var r0 *model.FetchRequest
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.FetchRequest); ok {
		r0 = rf(ctx, documentID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FetchRequest)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, documentID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
