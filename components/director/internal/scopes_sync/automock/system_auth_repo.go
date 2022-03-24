// Code generated by mockery v2.10.0. DO NOT EDIT.

package automock

import (
	context "context"

	repo "github.com/kyma-incubator/compass/components/director/internal/repo"
	mock "github.com/stretchr/testify/mock"

	pubModel "github.com/kyma-incubator/compass/components/director/pkg/model"
)

// SystemAuthRepo is an autogenerated mock type for the SystemAuthRepo type
type SystemAuthRepo struct {
	mock.Mock
}

// ListGlobalWithConditions provides a mock function with given fields: ctx, conditions
func (_m *SystemAuthRepo) ListGlobalWithConditions(ctx context.Context, conditions repo.Conditions) ([]pubModel.SystemAuth, error) {
	ret := _m.Called(ctx, conditions)

	var r0 []pubModel.SystemAuth
	if rf, ok := ret.Get(0).(func(context.Context, repo.Conditions) []pubModel.SystemAuth); ok {
		r0 = rf(ctx, conditions)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]pubModel.SystemAuth)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, repo.Conditions) error); ok {
		r1 = rf(ctx, conditions)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
