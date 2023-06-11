// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	resource "github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// SpecService is an autogenerated mock type for the SpecService type
type SpecService struct {
	mock.Mock
}

// CreateByReferenceObjectID provides a mock function with given fields: ctx, in, resourceType, objectType, objectID
func (_m *SpecService) CreateByReferenceObjectID(ctx context.Context, in model.SpecInput, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) (string, error) {
	ret := _m.Called(ctx, in, resourceType, objectType, objectID)

	var r0 string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.SpecInput, resource.Type, model.SpecReferenceObjectType, string) (string, error)); ok {
		return rf(ctx, in, resourceType, objectType, objectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.SpecInput, resource.Type, model.SpecReferenceObjectType, string) string); ok {
		r0 = rf(ctx, in, resourceType, objectType, objectID)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.SpecInput, resource.Type, model.SpecReferenceObjectType, string) error); ok {
		r1 = rf(ctx, in, resourceType, objectType, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// CreateByReferenceObjectIDWithDelayedFetchRequest provides a mock function with given fields: ctx, in, resourceType, objectType, objectID
func (_m *SpecService) CreateByReferenceObjectIDWithDelayedFetchRequest(ctx context.Context, in model.SpecInput, resourceType resource.Type, objectType model.SpecReferenceObjectType, objectID string) (string, *model.FetchRequest, error) {
	ret := _m.Called(ctx, in, resourceType, objectType, objectID)

	var r0 string
	var r1 *model.FetchRequest
	var r2 error
	if rf, ok := ret.Get(0).(func(context.Context, model.SpecInput, resource.Type, model.SpecReferenceObjectType, string) (string, *model.FetchRequest, error)); ok {
		return rf(ctx, in, resourceType, objectType, objectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.SpecInput, resource.Type, model.SpecReferenceObjectType, string) string); ok {
		r0 = rf(ctx, in, resourceType, objectType, objectID)
	} else {
		r0 = ret.Get(0).(string)
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.SpecInput, resource.Type, model.SpecReferenceObjectType, string) *model.FetchRequest); ok {
		r1 = rf(ctx, in, resourceType, objectType, objectID)
	} else {
		if ret.Get(1) != nil {
			r1 = ret.Get(1).(*model.FetchRequest)
		}
	}

	if rf, ok := ret.Get(2).(func(context.Context, model.SpecInput, resource.Type, model.SpecReferenceObjectType, string) error); ok {
		r2 = rf(ctx, in, resourceType, objectType, objectID)
	} else {
		r2 = ret.Error(2)
	}

	return r0, r1, r2
}

// DeleteByReferenceObjectID provides a mock function with given fields: ctx, objectType, objectID
func (_m *SpecService) DeleteByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) error {
	ret := _m.Called(ctx, objectType, objectID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.SpecReferenceObjectType, string) error); ok {
		r0 = rf(ctx, objectType, objectID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByID provides a mock function with given fields: ctx, id, objectType
func (_m *SpecService) GetByID(ctx context.Context, id string, objectType model.SpecReferenceObjectType) (*model.Spec, error) {
	ret := _m.Called(ctx, id, objectType)

	var r0 *model.Spec
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.SpecReferenceObjectType) (*model.Spec, error)); ok {
		return rf(ctx, id, objectType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, model.SpecReferenceObjectType) *model.Spec); ok {
		r0 = rf(ctx, id, objectType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Spec)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, model.SpecReferenceObjectType) error); ok {
		r1 = rf(ctx, id, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByIDGlobal provides a mock function with given fields: ctx, id
func (_m *SpecService) GetByIDGlobal(ctx context.Context, id string) (*model.Spec, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Spec
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string) (*model.Spec, error)); ok {
		return rf(ctx, id)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Spec); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Spec)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListFetchRequestsByReferenceObjectIDs provides a mock function with given fields: ctx, tenant, objectIDs, objectType
func (_m *SpecService) ListFetchRequestsByReferenceObjectIDs(ctx context.Context, tenant string, objectIDs []string, objectType model.SpecReferenceObjectType) ([]*model.FetchRequest, error) {
	ret := _m.Called(ctx, tenant, objectIDs, objectType)

	var r0 []*model.FetchRequest
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, model.SpecReferenceObjectType) ([]*model.FetchRequest, error)); ok {
		return rf(ctx, tenant, objectIDs, objectType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, model.SpecReferenceObjectType) []*model.FetchRequest); ok {
		r0 = rf(ctx, tenant, objectIDs, objectType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FetchRequest)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, []string, model.SpecReferenceObjectType) error); ok {
		r1 = rf(ctx, tenant, objectIDs, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListIDByReferenceObjectID provides a mock function with given fields: ctx, objectType, objectID
func (_m *SpecService) ListIDByReferenceObjectID(ctx context.Context, objectType model.SpecReferenceObjectType, objectID string) ([]string, error) {
	ret := _m.Called(ctx, objectType, objectID)

	var r0 []string
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, model.SpecReferenceObjectType, string) ([]string, error)); ok {
		return rf(ctx, objectType, objectID)
	}
	if rf, ok := ret.Get(0).(func(context.Context, model.SpecReferenceObjectType, string) []string); ok {
		r0 = rf(ctx, objectType, objectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]string)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, model.SpecReferenceObjectType, string) error); ok {
		r1 = rf(ctx, objectType, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// RefetchSpec provides a mock function with given fields: ctx, id, objectType
func (_m *SpecService) RefetchSpec(ctx context.Context, id string, objectType model.SpecReferenceObjectType) (*model.Spec, error) {
	ret := _m.Called(ctx, id, objectType)

	var r0 *model.Spec
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, string, model.SpecReferenceObjectType) (*model.Spec, error)); ok {
		return rf(ctx, id, objectType)
	}
	if rf, ok := ret.Get(0).(func(context.Context, string, model.SpecReferenceObjectType) *model.Spec); ok {
		r0 = rf(ctx, id, objectType)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Spec)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, string, model.SpecReferenceObjectType) error); ok {
		r1 = rf(ctx, id, objectType)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// UpdateSpecOnly provides a mock function with given fields: ctx, spec
func (_m *SpecService) UpdateSpecOnly(ctx context.Context, spec model.Spec) error {
	ret := _m.Called(ctx, spec)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.Spec) error); ok {
		r0 = rf(ctx, spec)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// UpdateSpecOnlyGlobal provides a mock function with given fields: ctx, spec
func (_m *SpecService) UpdateSpecOnlyGlobal(ctx context.Context, spec model.Spec) error {
	ret := _m.Called(ctx, spec)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, model.Spec) error); ok {
		r0 = rf(ctx, spec)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewSpecService interface {
	mock.TestingT
	Cleanup(func())
}

// NewSpecService creates a new instance of SpecService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewSpecService(t mockConstructorTestingTNewSpecService) *SpecService {
	mock := &SpecService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
