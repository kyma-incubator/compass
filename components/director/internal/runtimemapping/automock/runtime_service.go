// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// RuntimeService is an autogenerated mock type for the RuntimeService type
type RuntimeService struct {
	mock.Mock
}

// GetByTokenIssuer provides a mock function with given fields: ctx, issuer
func (_m *RuntimeService) GetByTokenIssuer(ctx context.Context, issuer string) (*model.Runtime, error) {
	ret := _m.Called(ctx, issuer)

	var r0 *model.Runtime
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Runtime); ok {
		r0 = rf(ctx, issuer)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Runtime)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, issuer)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
