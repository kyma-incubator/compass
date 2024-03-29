// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	apperrors "github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/apperrors"

	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	mock "github.com/stretchr/testify/mock"
)

// Client is an autogenerated mock type for the Client type
type Client struct {
	mock.Mock
}

// GetApplication provides a mock function with given fields: ctx, systemAuthID
func (_m *Client) GetApplication(ctx context.Context, systemAuthID string) (graphql.ApplicationExt, apperrors.AppError) {
	ret := _m.Called(ctx, systemAuthID)

	var r0 graphql.ApplicationExt
	if rf, ok := ret.Get(0).(func(context.Context, string) graphql.ApplicationExt); ok {
		r0 = rf(ctx, systemAuthID)
	} else {
		r0 = ret.Get(0).(graphql.ApplicationExt)
	}

	var r1 apperrors.AppError
	if rf, ok := ret.Get(1).(func(context.Context, string) apperrors.AppError); ok {
		r1 = rf(ctx, systemAuthID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(apperrors.AppError)
		}
	}

	return r0, r1
}
