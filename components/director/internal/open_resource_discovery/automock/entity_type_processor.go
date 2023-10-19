// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// EntityTypeProcessor is an autogenerated mock type for the EntityTypeProcessor type
type EntityTypeProcessor struct {
	mock.Mock
}

// Process provides a mock function with given fields: ctx, resourceType, resourceID, entityTypes, resourceHashes
func (_m *EntityTypeProcessor) Process(ctx context.Context, resourceType resource.Type, resourceID string, entityTypes []*model.EntityTypeInput, resourceHashes map[string]uint64) ([]*model.EntityType, error) {
	ret := _m.Called(ctx, resourceType, resourceID, entityTypes, resourceHashes)

	var r0 []*model.EntityType
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, []*model.EntityTypeInput, map[string]uint64) ([]*model.EntityType, error)); ok {
		return rf(ctx, resourceType, resourceID, entityTypes, resourceHashes)
	}
	if rf, ok := ret.Get(0).(func(context.Context, resource.Type, string, []*model.EntityTypeInput, map[string]uint64) []*model.EntityType); ok {
		r0 = rf(ctx, resourceType, resourceID, entityTypes, resourceHashes)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.EntityType)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, resource.Type, string, []*model.EntityTypeInput, map[string]uint64) error); ok {
		r1 = rf(ctx, resourceType, resourceID, entityTypes, resourceHashes)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewEntityTypeProcessor creates a new instance of EntityTypeProcessor. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewEntityTypeProcessor(t interface {
	mock.TestingT
	Cleanup(func())
}) *EntityTypeProcessor {
	mock := &EntityTypeProcessor{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
