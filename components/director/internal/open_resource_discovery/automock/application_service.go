// Code generated by mockery v2.9.4. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// ApplicationService is an autogenerated mock type for the ApplicationService type
type ApplicationService struct {
	mock.Mock
}

// ListGlobal provides a mock function with given fields: ctx, pageSize, cursor
func (_m *ApplicationService) ListGlobal(ctx context.Context, pageSize int, cursor string) (*model.ApplicationPage, error) {
	ret := _m.Called(ctx, pageSize, cursor)

	var r0 *model.ApplicationPage
	if rf, ok := ret.Get(0).(func(context.Context, int, string) *model.ApplicationPage); ok {
		r0 = rf(ctx, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int, string) error); ok {
		r1 = rf(ctx, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
