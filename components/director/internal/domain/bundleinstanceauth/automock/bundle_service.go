// Code generated by mockery 2.9.0. DO NOT EDIT.

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

// Get provides a mock function with given fields: ctx, id
func (_m *BundleService) Get(ctx context.Context, id string) (*model.Bundle, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Bundle
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Bundle); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Bundle)
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

// GetByInstanceAuthID provides a mock function with given fields: ctx, instanceAuthID
func (_m *BundleService) GetByInstanceAuthID(ctx context.Context, instanceAuthID string) (*model.Bundle, error) {
	ret := _m.Called(ctx, instanceAuthID)

	var r0 *model.Bundle
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Bundle); ok {
		r0 = rf(ctx, instanceAuthID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Bundle)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, instanceAuthID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
