// Code generated by mockery 2.9.0. DO NOT EDIT.

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

// GetForBundle provides a mock function with given fields: ctx, objectType, objectID, bundleID
func (_m *BundleReferenceService) GetForBundle(ctx context.Context, objectType model.BundleReferenceObjectType, objectID *string, bundleID *string) (*model.BundleReference, error) {
	ret := _m.Called(ctx, objectType, objectID, bundleID)

	var r0 *model.BundleReference
	if rf, ok := ret.Get(0).(func(context.Context, model.BundleReferenceObjectType, *string, *string) *model.BundleReference); ok {
		r0 = rf(ctx, objectType, objectID, bundleID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.BundleReference)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, model.BundleReferenceObjectType, *string, *string) error); ok {
		r1 = rf(ctx, objectType, objectID, bundleID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
