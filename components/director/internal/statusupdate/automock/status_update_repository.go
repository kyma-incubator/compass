// Code generated by mockery v2.9.4. DO NOT EDIT.

package automock

import (
	context "context"

	statusupdate "github.com/kyma-incubator/compass/components/director/internal/statusupdate"
	mock "github.com/stretchr/testify/mock"
)

// StatusUpdateRepository is an autogenerated mock type for the StatusUpdateRepository type
type StatusUpdateRepository struct {
	mock.Mock
}

// IsConnected provides a mock function with given fields: ctx, id, object
func (_m *StatusUpdateRepository) IsConnected(ctx context.Context, id string, object statusupdate.WithStatusObject) (bool, error) {
	ret := _m.Called(ctx, id, object)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, statusupdate.WithStatusObject) bool); ok {
		r0 = rf(ctx, id, object)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, statusupdate.WithStatusObject) error); ok {
		r1 = rf(ctx, id, object)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateStatus provides a mock function with given fields: ctx, id, object
func (_m *StatusUpdateRepository) UpdateStatus(ctx context.Context, id string, object statusupdate.WithStatusObject) error {
	ret := _m.Called(ctx, id, object)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, statusupdate.WithStatusObject) error); ok {
		r0 = rf(ctx, id, object)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
