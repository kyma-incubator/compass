// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	labelfilter "github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	uuid "github.com/google/uuid"
)

// ApplicationRepository is an autogenerated mock type for the ApplicationRepository type
type ApplicationRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, tenant, item
func (_m *ApplicationRepository) Create(ctx context.Context, tenant string, item *model.Application) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Application) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, tenant, id
func (_m *ApplicationRepository) Delete(ctx context.Context, tenant string, id string) error {
	ret := _m.Called(ctx, tenant, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteGlobal provides a mock function with given fields: ctx, id
func (_m *ApplicationRepository) DeleteGlobal(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: ctx, tenant, id
func (_m *ApplicationRepository) Exists(ctx context.Context, tenant string, id string) (bool, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByFilter provides a mock function with given fields: ctx, tenant, filter
func (_m *ApplicationRepository) GetByFilter(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) (*model.Application, error) {
	ret := _m.Called(ctx, tenant, filter)

	var r0 *model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string, []*labelfilter.LabelFilter) *model.Application); ok {
		r0 = rf(ctx, tenant, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []*labelfilter.LabelFilter) error); ok {
		r1 = rf(ctx, tenant, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByID provides a mock function with given fields: ctx, tenant, id
func (_m *ApplicationRepository) GetByID(ctx context.Context, tenant string, id string) (*model.Application, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 *model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Application); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByIDForUpdate provides a mock function with given fields: ctx, tenant, id
func (_m *ApplicationRepository) GetByIDForUpdate(ctx context.Context, tenant string, id string) (*model.Application, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 *model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Application); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetBySystemNumber provides a mock function with given fields: ctx, tenant, systemNumber
func (_m *ApplicationRepository) GetBySystemNumber(ctx context.Context, tenant string, systemNumber string) (*model.Application, error) {
	ret := _m.Called(ctx, tenant, systemNumber)

	var r0 *model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.Application); ok {
		r0 = rf(ctx, tenant, systemNumber)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, systemNumber)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGlobalByID provides a mock function with given fields: ctx, id
func (_m *ApplicationRepository) GetGlobalByID(ctx context.Context, id string) (*model.Application, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.Application); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, tenant, filter, pageSize, cursor
func (_m *ApplicationRepository) List(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (*model.ApplicationPage, error) {
	ret := _m.Called(ctx, tenant, filter, pageSize, cursor)

	var r0 *model.ApplicationPage
	if rf, ok := ret.Get(0).(func(context.Context, string, []*labelfilter.LabelFilter, int, string) *model.ApplicationPage); ok {
		r0 = rf(ctx, tenant, filter, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []*labelfilter.LabelFilter, int, string) error); ok {
		r1 = rf(ctx, tenant, filter, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAll provides a mock function with given fields: ctx, tenant
func (_m *ApplicationRepository) ListAll(ctx context.Context, tenant string) ([]*model.Application, error) {
	ret := _m.Called(ctx, tenant)

	var r0 []*model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.Application); ok {
		r0 = rf(ctx, tenant)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, tenant)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllByApplicationTemplateID provides a mock function with given fields: ctx, applicationTemplateID
func (_m *ApplicationRepository) ListAllByApplicationTemplateID(ctx context.Context, applicationTemplateID string) ([]*model.Application, error) {
	ret := _m.Called(ctx, applicationTemplateID)

	var r0 []*model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string) []*model.Application); ok {
		r0 = rf(ctx, applicationTemplateID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string) error); ok {
		r1 = rf(ctx, applicationTemplateID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllByFilter provides a mock function with given fields: ctx, tenant, filter
func (_m *ApplicationRepository) ListAllByFilter(ctx context.Context, tenant string, filter []*labelfilter.LabelFilter) ([]*model.Application, error) {
	ret := _m.Called(ctx, tenant, filter)

	var r0 []*model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string, []*labelfilter.LabelFilter) []*model.Application); ok {
		r0 = rf(ctx, tenant, filter)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []*labelfilter.LabelFilter) error); ok {
		r1 = rf(ctx, tenant, filter)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByScenarios provides a mock function with given fields: ctx, tenantID, scenarios, pageSize, cursor, hidingSelectors
func (_m *ApplicationRepository) ListByScenarios(ctx context.Context, tenantID uuid.UUID, scenarios []string, pageSize int, cursor string, hidingSelectors map[string][]string) (*model.ApplicationPage, error) {
	ret := _m.Called(ctx, tenantID, scenarios, pageSize, cursor, hidingSelectors)

	var r0 *model.ApplicationPage
	if rf, ok := ret.Get(0).(func(context.Context, uuid.UUID, []string, int, string, map[string][]string) *model.ApplicationPage); ok {
		r0 = rf(ctx, tenantID, scenarios, pageSize, cursor, hidingSelectors)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, uuid.UUID, []string, int, string, map[string][]string) error); ok {
		r1 = rf(ctx, tenantID, scenarios, pageSize, cursor, hidingSelectors)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByScenariosAndIDs provides a mock function with given fields: ctx, tenant, scenarios, ids
func (_m *ApplicationRepository) ListByScenariosAndIDs(ctx context.Context, tenant string, scenarios []string, ids []string) ([]*model.Application, error) {
	ret := _m.Called(ctx, tenant, scenarios, ids)

	var r0 []*model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, []string) []*model.Application); ok {
		r0 = rf(ctx, tenant, scenarios, ids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string, []string) error); ok {
		r1 = rf(ctx, tenant, scenarios, ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByScenariosNoPaging provides a mock function with given fields: ctx, tenant, scenarios
func (_m *ApplicationRepository) ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error) {
	ret := _m.Called(ctx, tenant, scenarios)

	var r0 []*model.Application
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) []*model.Application); ok {
		r0 = rf(ctx, tenant, scenarios)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string) error); ok {
		r1 = rf(ctx, tenant, scenarios)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListGlobal provides a mock function with given fields: ctx, pageSize, cursor
func (_m *ApplicationRepository) ListGlobal(ctx context.Context, pageSize int, cursor string) (*model.ApplicationPage, error) {
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

// OwnerExists provides a mock function with given fields: ctx, tenant, id
func (_m *ApplicationRepository) OwnerExists(ctx context.Context, tenant string, id string) (bool, error) {
	ret := _m.Called(ctx, tenant, id)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, tenant, id)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenant, id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// TechnicalUpdate provides a mock function with given fields: ctx, item
func (_m *ApplicationRepository) TechnicalUpdate(ctx context.Context, item *model.Application) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.Application) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// TrustedUpsert provides a mock function with given fields: ctx, tenant, _a2
func (_m *ApplicationRepository) TrustedUpsert(ctx context.Context, tenant string, _a2 *model.Application) (string, error) {
	ret := _m.Called(ctx, tenant, _a2)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Application) string); ok {
		r0 = rf(ctx, tenant, _a2)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *model.Application) error); ok {
		r1 = rf(ctx, tenant, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, tenant, item
func (_m *ApplicationRepository) Update(ctx context.Context, tenant string, item *model.Application) error {
	ret := _m.Called(ctx, tenant, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Application) error); ok {
		r0 = rf(ctx, tenant, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Upsert provides a mock function with given fields: ctx, tenant, _a2
func (_m *ApplicationRepository) Upsert(ctx context.Context, tenant string, _a2 *model.Application) (string, error) {
	ret := _m.Called(ctx, tenant, _a2)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.Application) string); ok {
		r0 = rf(ctx, tenant, _a2)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *model.Application) error); ok {
		r1 = rf(ctx, tenant, _a2)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewApplicationRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewApplicationRepository creates a new instance of ApplicationRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewApplicationRepository(t mockConstructorTestingTNewApplicationRepository) *ApplicationRepository {
	mock := &ApplicationRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
