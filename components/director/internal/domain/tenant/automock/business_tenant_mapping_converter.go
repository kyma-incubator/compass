// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	repo "github.com/kyma-incubator/compass/components/director/internal/repo"
)

// BusinessTenantMappingConverter is an autogenerated mock type for the BusinessTenantMappingConverter type
type BusinessTenantMappingConverter struct {
	mock.Mock
}

// InputFromGraphQL provides a mock function with given fields: ctx, tnt, externalTenantToType, retrieveTenantTypeFn
func (_m *BusinessTenantMappingConverter) InputFromGraphQL(ctx context.Context, tnt graphql.BusinessTenantMappingInput, externalTenantToType map[string]string, retrieveTenantTypeFn func(context.Context, string) (string, error)) (model.BusinessTenantMappingInput, error) {
	ret := _m.Called(ctx, tnt, externalTenantToType, retrieveTenantTypeFn)

	var r0 model.BusinessTenantMappingInput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, graphql.BusinessTenantMappingInput, map[string]string, func(context.Context, string) (string, error)) (model.BusinessTenantMappingInput, error)); ok {
		return rf(ctx, tnt, externalTenantToType, retrieveTenantTypeFn)
	}
	if rf, ok := ret.Get(0).(func(context.Context, graphql.BusinessTenantMappingInput, map[string]string, func(context.Context, string) (string, error)) model.BusinessTenantMappingInput); ok {
		r0 = rf(ctx, tnt, externalTenantToType, retrieveTenantTypeFn)
	} else {
		r0 = ret.Get(0).(model.BusinessTenantMappingInput)
	}

	if rf, ok := ret.Get(1).(func(context.Context, graphql.BusinessTenantMappingInput, map[string]string, func(context.Context, string) (string, error)) error); ok {
		r1 = rf(ctx, tnt, externalTenantToType, retrieveTenantTypeFn)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MultipleInputFromGraphQL provides a mock function with given fields: ctx, in, retrieveTenantTypeFn
func (_m *BusinessTenantMappingConverter) MultipleInputFromGraphQL(ctx context.Context, in []*graphql.BusinessTenantMappingInput, retrieveTenantTypeFn func(context.Context, string) (string, error)) ([]model.BusinessTenantMappingInput, error) {
	ret := _m.Called(ctx, in, retrieveTenantTypeFn)

	var r0 []model.BusinessTenantMappingInput
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []*graphql.BusinessTenantMappingInput, func(context.Context, string) (string, error)) ([]model.BusinessTenantMappingInput, error)); ok {
		return rf(ctx, in, retrieveTenantTypeFn)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []*graphql.BusinessTenantMappingInput, func(context.Context, string) (string, error)) []model.BusinessTenantMappingInput); ok {
		r0 = rf(ctx, in, retrieveTenantTypeFn)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]model.BusinessTenantMappingInput)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []*graphql.BusinessTenantMappingInput, func(context.Context, string) (string, error)) error); ok {
		r1 = rf(ctx, in, retrieveTenantTypeFn)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *BusinessTenantMappingConverter) MultipleToGraphQL(in []*model.BusinessTenantMapping) []*graphql.Tenant {
	ret := _m.Called(in)

	var r0 []*graphql.Tenant
	if rf, ok := ret.Get(0).(func([]*model.BusinessTenantMapping) []*graphql.Tenant); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.Tenant)
		}
	}

	return r0
}

// TenantAccessFromEntity provides a mock function with given fields: in
func (_m *BusinessTenantMappingConverter) TenantAccessFromEntity(in *repo.TenantAccess) *model.TenantAccess {
	ret := _m.Called(in)

	var r0 *model.TenantAccess
	if rf, ok := ret.Get(0).(func(*repo.TenantAccess) *model.TenantAccess); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.TenantAccess)
		}
	}

	return r0
}

// TenantAccessInputFromGraphQL provides a mock function with given fields: in
func (_m *BusinessTenantMappingConverter) TenantAccessInputFromGraphQL(in graphql.TenantAccessInput) (*model.TenantAccess, error) {
	ret := _m.Called(in)

	var r0 *model.TenantAccess
	var r1 error
	if rf, ok := ret.Get(0).(func(graphql.TenantAccessInput) (*model.TenantAccess, error)); ok {
		return rf(in)
	}
	if rf, ok := ret.Get(0).(func(graphql.TenantAccessInput) *model.TenantAccess); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.TenantAccess)
		}
	}

	if rf, ok := ret.Get(1).(func(graphql.TenantAccessInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TenantAccessToEntity provides a mock function with given fields: in
func (_m *BusinessTenantMappingConverter) TenantAccessToEntity(in *model.TenantAccess) *repo.TenantAccess {
	ret := _m.Called(in)

	var r0 *repo.TenantAccess
	if rf, ok := ret.Get(0).(func(*model.TenantAccess) *repo.TenantAccess); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*repo.TenantAccess)
		}
	}

	return r0
}

// TenantAccessToGraphQL provides a mock function with given fields: in
func (_m *BusinessTenantMappingConverter) TenantAccessToGraphQL(in *model.TenantAccess) (*graphql.TenantAccess, error) {
	ret := _m.Called(in)

	var r0 *graphql.TenantAccess
	var r1 error
	if rf, ok := ret.Get(0).(func(*model.TenantAccess) (*graphql.TenantAccess, error)); ok {
		return rf(in)
	}
	if rf, ok := ret.Get(0).(func(*model.TenantAccess) *graphql.TenantAccess); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.TenantAccess)
		}
	}

	if rf, ok := ret.Get(1).(func(*model.TenantAccess) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGraphQL provides a mock function with given fields: in
func (_m *BusinessTenantMappingConverter) ToGraphQL(in *model.BusinessTenantMapping) *graphql.Tenant {
	ret := _m.Called(in)

	var r0 *graphql.Tenant
	if rf, ok := ret.Get(0).(func(*model.BusinessTenantMapping) *graphql.Tenant); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.Tenant)
		}
	}

	return r0
}

// NewBusinessTenantMappingConverter creates a new instance of BusinessTenantMappingConverter. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewBusinessTenantMappingConverter(t interface {
	mock.TestingT
	Cleanup(func())
}) *BusinessTenantMappingConverter {
	mock := &BusinessTenantMappingConverter{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
