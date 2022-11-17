// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	testing "testing"
)

// FormationRepository is an autogenerated mock type for the FormationRepository type
type FormationRepository struct {
	mock.Mock
}

// Get provides a mock function with given fields: ctx, id, tenantID
func (_m *FormationRepository) Get(ctx context.Context, id string, tenantID string) (*model.Formation, error) {
	ret := _m.Called(ctx, id, tenantID)

	var r0 *model.Formation
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Formation); ok {
		r0 = rf(ctx, id, tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Formation)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewFormationRepository creates a new instance of FormationRepository. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationRepository(t testing.TB) *FormationRepository {
	mock := &FormationRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
