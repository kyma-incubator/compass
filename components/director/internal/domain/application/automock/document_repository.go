// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// DocumentRepository is an autogenerated mock type for the DocumentRepository type
type DocumentRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, item
func (_m *DocumentRepository) Create(ctx context.Context, item *model.Document) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Document) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteAllByApplicationID provides a mock function with given fields: ctx, tenant, applicationID
func (_m *DocumentRepository) DeleteAllByApplicationID(ctx context.Context, tenant string, applicationID string) error {
	ret := _m.Called(ctx, tenant, applicationID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, tenant, applicationID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
