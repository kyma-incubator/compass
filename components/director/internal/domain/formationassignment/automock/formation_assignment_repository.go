// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// FormationAssignmentRepository is an autogenerated mock type for the FormationAssignmentRepository type
type FormationAssignmentRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: ctx, item
func (_m *FormationAssignmentRepository) Create(ctx context.Context, item *model.FormationAssignment) error {
	ret := _m.Called(ctx, item)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.FormationAssignment) error); ok {
		r0 = rf(ctx, item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: ctx, id, tenantID
func (_m *FormationAssignmentRepository) Delete(ctx context.Context, id string, tenantID string) error {
	ret := _m.Called(ctx, id, tenantID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, id, tenantID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteAssignmentsForObjectID provides a mock function with given fields: ctx, tnt, formationID, objectID
func (_m *FormationAssignmentRepository) DeleteAssignmentsForObjectID(ctx context.Context, tnt string, formationID string, objectID string) error {
	ret := _m.Called(ctx, tnt, formationID, objectID)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) error); ok {
		r0 = rf(ctx, tnt, formationID, objectID)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Exists provides a mock function with given fields: ctx, id, tenantID
func (_m *FormationAssignmentRepository) Exists(ctx context.Context, id string, tenantID string) (bool, error) {
	ret := _m.Called(ctx, id, tenantID)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, string, string) bool); ok {
		r0 = rf(ctx, id, tenantID)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Get provides a mock function with given fields: ctx, id, tenantID
func (_m *FormationAssignmentRepository) Get(ctx context.Context, id string, tenantID string) (*model.FormationAssignment, error) {
	ret := _m.Called(ctx, id, tenantID)

	var r0 *model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.FormationAssignment); ok {
		r0 = rf(ctx, id, tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationAssignment)
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

// GetAssignmentsForFormation provides a mock function with given fields: ctx, tenantID, formationID
func (_m *FormationAssignmentRepository) GetAssignmentsForFormation(ctx context.Context, tenantID string, formationID string) ([]*model.FormationAssignment, error) {
	ret := _m.Called(ctx, tenantID, formationID)

	var r0 []*model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []*model.FormationAssignment); ok {
		r0 = rf(ctx, tenantID, formationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, tenantID, formationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetAssignmentsForFormationWithStates provides a mock function with given fields: ctx, tenantID, formationID, states
func (_m *FormationAssignmentRepository) GetAssignmentsForFormationWithStates(ctx context.Context, tenantID string, formationID string, states []string) ([]*model.FormationAssignment, error) {
	ret := _m.Called(ctx, tenantID, formationID, states)

	var r0 []*model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, string, []string) []*model.FormationAssignment); ok {
		r0 = rf(ctx, tenantID, formationID, states)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, []string) error); ok {
		r1 = rf(ctx, tenantID, formationID, states)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetByTargetAndSource provides a mock function with given fields: ctx, target, source, tenantID, formationID
func (_m *FormationAssignmentRepository) GetByTargetAndSource(ctx context.Context, target string, source string, tenantID string, formationID string) (*model.FormationAssignment, error) {
	ret := _m.Called(ctx, target, source, tenantID, formationID)

	var r0 *model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string) *model.FormationAssignment); ok {
		r0 = rf(ctx, target, source, tenantID, formationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, string) error); ok {
		r1 = rf(ctx, target, source, tenantID, formationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetForFormation provides a mock function with given fields: ctx, tenantID, id, formationID
func (_m *FormationAssignmentRepository) GetForFormation(ctx context.Context, tenantID string, id string, formationID string) (*model.FormationAssignment, error) {
	ret := _m.Called(ctx, tenantID, id, formationID)

	var r0 *model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) *model.FormationAssignment); ok {
		r0 = rf(ctx, tenantID, id, formationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, tenantID, id, formationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetGlobalByID provides a mock function with given fields: ctx, id
func (_m *FormationAssignmentRepository) GetGlobalByID(ctx context.Context, id string) (*model.FormationAssignment, error) {
	ret := _m.Called(ctx, id)

	var r0 *model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string) *model.FormationAssignment); ok {
		r0 = rf(ctx, id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationAssignment)
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

// GetGlobalByIDAndFormationID provides a mock function with given fields: ctx, id, formationID
func (_m *FormationAssignmentRepository) GetGlobalByIDAndFormationID(ctx context.Context, id string, formationID string) (*model.FormationAssignment, error) {
	ret := _m.Called(ctx, id, formationID)

	var r0 *model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, string) *model.FormationAssignment); ok {
		r0 = rf(ctx, id, formationID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, id, formationID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetReverseBySourceAndTarget provides a mock function with given fields: ctx, tenantID, formationID, sourceID, targetID
func (_m *FormationAssignmentRepository) GetReverseBySourceAndTarget(ctx context.Context, tenantID string, formationID string, sourceID string, targetID string) (*model.FormationAssignment, error) {
	ret := _m.Called(ctx, tenantID, formationID, sourceID, targetID)

	var r0 *model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string, string) *model.FormationAssignment); ok {
		r0 = rf(ctx, tenantID, formationID, sourceID, targetID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string, string) error); ok {
		r1 = rf(ctx, tenantID, formationID, sourceID, targetID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: ctx, pageSize, cursor, tenantID
func (_m *FormationAssignmentRepository) List(ctx context.Context, pageSize int, cursor string, tenantID string) (*model.FormationAssignmentPage, error) {
	ret := _m.Called(ctx, pageSize, cursor, tenantID)

	var r0 *model.FormationAssignmentPage
	if rf, ok := ret.Get(0).(func(context.Context, int, string, string) *model.FormationAssignmentPage); ok {
		r0 = rf(ctx, pageSize, cursor, tenantID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.FormationAssignmentPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, int, string, string) error); ok {
		r1 = rf(ctx, pageSize, cursor, tenantID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllForObject provides a mock function with given fields: ctx, tenant, formationID, objectID
func (_m *FormationAssignmentRepository) ListAllForObject(ctx context.Context, tenant string, formationID string, objectID string) ([]*model.FormationAssignment, error) {
	ret := _m.Called(ctx, tenant, formationID, objectID)

	var r0 []*model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, string, string) []*model.FormationAssignment); ok {
		r0 = rf(ctx, tenant, formationID, objectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, string) error); ok {
		r1 = rf(ctx, tenant, formationID, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListAllForObjectIDs provides a mock function with given fields: ctx, tenant, formationID, objectIDs
func (_m *FormationAssignmentRepository) ListAllForObjectIDs(ctx context.Context, tenant string, formationID string, objectIDs []string) ([]*model.FormationAssignment, error) {
	ret := _m.Called(ctx, tenant, formationID, objectIDs)

	var r0 []*model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, string, []string) []*model.FormationAssignment); ok {
		r0 = rf(ctx, tenant, formationID, objectIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, []string) error); ok {
		r1 = rf(ctx, tenant, formationID, objectIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByFormationIDs provides a mock function with given fields: ctx, tenantID, formationIDs, pageSize, cursor
func (_m *FormationAssignmentRepository) ListByFormationIDs(ctx context.Context, tenantID string, formationIDs []string, pageSize int, cursor string) ([]*model.FormationAssignmentPage, error) {
	ret := _m.Called(ctx, tenantID, formationIDs, pageSize, cursor)

	var r0 []*model.FormationAssignmentPage
	if rf, ok := ret.Get(0).(func(context.Context, string, []string, int, string) []*model.FormationAssignmentPage); ok {
		r0 = rf(ctx, tenantID, formationIDs, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationAssignmentPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string, int, string) error); ok {
		r1 = rf(ctx, tenantID, formationIDs, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByFormationIDsNoPaging provides a mock function with given fields: ctx, tenantID, formationIDs
func (_m *FormationAssignmentRepository) ListByFormationIDsNoPaging(ctx context.Context, tenantID string, formationIDs []string) ([][]*model.FormationAssignment, error) {
	ret := _m.Called(ctx, tenantID, formationIDs)

	var r0 [][]*model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) [][]*model.FormationAssignment); ok {
		r0 = rf(ctx, tenantID, formationIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string) error); ok {
		r1 = rf(ctx, tenantID, formationIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListForIDs provides a mock function with given fields: ctx, tenant, ids
func (_m *FormationAssignmentRepository) ListForIDs(ctx context.Context, tenant string, ids []string) ([]*model.FormationAssignment, error) {
	ret := _m.Called(ctx, tenant, ids)

	var r0 []*model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, []string) []*model.FormationAssignment); ok {
		r0 = rf(ctx, tenant, ids)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, []string) error); ok {
		r1 = rf(ctx, tenant, ids)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: ctx, _a1
func (_m *FormationAssignmentRepository) Update(ctx context.Context, _a1 *model.FormationAssignment) error {
	ret := _m.Called(ctx, _a1)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.FormationAssignment) error); ok {
		r0 = rf(ctx, _a1)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewFormationAssignmentRepository interface {
	mock.TestingT
	Cleanup(func())
}

// NewFormationAssignmentRepository creates a new instance of FormationAssignmentRepository. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationAssignmentRepository(t mockConstructorTestingTNewFormationAssignmentRepository) *FormationAssignmentRepository {
	mock := &FormationAssignmentRepository{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
