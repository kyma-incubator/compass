// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import context "context"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// Repository is an autogenerated mock type for the Repository type
type Repository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, item
func (_m *Repository) Create(ctx context.Context, item model.SystemAuth) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.SystemAuth) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, tenant, id, objectType
func (_m *Repository) Delete(ctx context.Context, tenant string, id string, objectType model.SystemAuthReferenceObjectType) error {
	ret := _m.Called(ctx, tenant, id, objectType)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, model.SystemAuthReferenceObjectType) error); ok {
		r0 = rf(ctx, tenant, id, objectType)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// ListForObject provides a mock function with given fields: ctx, tenant, objectType, objectID
func (_m *Repository) ListForObject(ctx context.Context, tenant string, objectType model.SystemAuthReferenceObjectType, objectID string) ([]model.SystemAuth, error) {
	ret := _m.Called(ctx, tenant, objectType, objectID)

	var r0 []model.SystemAuth
	if rf, ok := ret.Get(0).(func(context.Context, string, model.SystemAuthReferenceObjectType, string) []model.SystemAuth); ok {
		r0 = rf(ctx, tenant, objectType, objectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.SystemAuth)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, model.SystemAuthReferenceObjectType, string) error); ok {
		r1 = rf(ctx, tenant, objectType, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
