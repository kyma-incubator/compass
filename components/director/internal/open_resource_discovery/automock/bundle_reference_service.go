// Code generated by mockery v2.2.1. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// BundleReferenceService is an autogenerated mock type for the BundleReferenceService type
type BundleReferenceService struct {
	mock.Mock
}

// GetBundleIDsForObject provides a mock function with given fields: ctx, objectType, objectID
func (_m *BundleReferenceService) GetBundleIDsForObject(ctx context.Context, objectType model.BundleReferenceObjectType, objectID *string) ([]string, error) {
	ret := _m.Called(ctx, objectType, objectID)

	var r0 []string
	if rf, ok := ret.Get(0).(func(context.Context, model.BundleReferenceObjectType, *string) []string); ok {
		r0 = rf(ctx, objectType, objectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.BundleReferenceObjectType, *string) error); ok {
		r1 = rf(ctx, objectType, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
