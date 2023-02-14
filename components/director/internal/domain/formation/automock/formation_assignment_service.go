// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	formationassignment "github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"

	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
)

// FormationAssignmentService is an autogenerated mock type for the formationAssignmentService type
type FormationAssignmentService struct {
	mock.Mock
}

// CleanupFormationAssignment provides a mock function with given fields: ctx, mappingPair
func (_m *FormationAssignmentService) CleanupFormationAssignment(ctx context.Context, mappingPair *formationassignment.AssignmentMappingPair) (bool, error) {
	ret := _m.Called(ctx, mappingPair)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, *formationassignment.AssignmentMappingPair) bool); ok {
		r0 = rf(ctx, mappingPair)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *formationassignment.AssignmentMappingPair) error); ok {
		r1 = rf(ctx, mappingPair)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Delete provides a mock function with given fields: ctx, id
func (_m *FormationAssignmentService) Delete(ctx context.Context, id string) error {
	ret := _m.Called(ctx, id)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string) error); ok {
		r0 = rf(ctx, id)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GenerateAssignments provides a mock function with given fields: ctx, tnt, objectID, objectType, _a4
func (_m *FormationAssignmentService) GenerateAssignments(ctx context.Context, tnt string, objectID string, objectType graphql.FormationObjectType, _a4 *model.Formation) ([]*model.FormationAssignment, error) {
	ret := _m.Called(ctx, tnt, objectID, objectType, _a4)

	var r0 []*model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, string, graphql.FormationObjectType, *model.Formation) []*model.FormationAssignment); ok {
		r0 = rf(ctx, tnt, objectID, objectType, _a4)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string, graphql.FormationObjectType, *model.Formation) error); ok {
		r1 = rf(ctx, tnt, objectID, objectType, _a4)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// GetForFormation provides a mock function with given fields: ctx, id, formationID
func (_m *FormationAssignmentService) GetForFormation(ctx context.Context, id string, formationID string) (*model.FormationAssignment, error) {
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

// ListByFormationIDs provides a mock function with given fields: ctx, formationIDs, pageSize, cursor
func (_m *FormationAssignmentService) ListByFormationIDs(ctx context.Context, formationIDs []string, pageSize int, cursor string) ([]*model.FormationAssignmentPage, error) {
	ret := _m.Called(ctx, formationIDs, pageSize, cursor)

	var r0 []*model.FormationAssignmentPage
	if rf, ok := ret.Get(0).(func(context.Context, []string, int, string) []*model.FormationAssignmentPage); ok {
		r0 = rf(ctx, formationIDs, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationAssignmentPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string, int, string) error); ok {
		r1 = rf(ctx, formationIDs, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListByFormationIDsNoPaging provides a mock function with given fields: ctx, formationIDs
func (_m *FormationAssignmentService) ListByFormationIDsNoPaging(ctx context.Context, formationIDs []string) ([][]*model.FormationAssignment, error) {
	ret := _m.Called(ctx, formationIDs)

	var r0 [][]*model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, []string) [][]*model.FormationAssignment); ok {
		r0 = rf(ctx, formationIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([][]*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, []string) error); ok {
		r1 = rf(ctx, formationIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListFormationAssignmentsForObjectID provides a mock function with given fields: ctx, formationID, objectID
func (_m *FormationAssignmentService) ListFormationAssignmentsForObjectID(ctx context.Context, formationID string, objectID string) ([]*model.FormationAssignment, error) {
	ret := _m.Called(ctx, formationID, objectID)

	var r0 []*model.FormationAssignment
	if rf, ok := ret.Get(0).(func(context.Context, string, string) []*model.FormationAssignment); ok {
		r0 = rf(ctx, formationID, objectID)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.FormationAssignment)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, string) error); ok {
		r1 = rf(ctx, formationID, objectID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProcessFormationAssignmentPair provides a mock function with given fields: ctx, mappingPair
func (_m *FormationAssignmentService) ProcessFormationAssignmentPair(ctx context.Context, mappingPair *formationassignment.AssignmentMappingPair) (bool, error) {
	ret := _m.Called(ctx, mappingPair)

	var r0 bool
	if rf, ok := ret.Get(0).(func(context.Context, *formationassignment.AssignmentMappingPair) bool); ok {
		r0 = rf(ctx, mappingPair)
	} else {
		r0 = ret.Get(0).(bool)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *formationassignment.AssignmentMappingPair) error); ok {
		r1 = rf(ctx, mappingPair)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ProcessFormationAssignments provides a mock function with given fields: ctx, formationAssignmentsForObject, runtimeContextIDToRuntimeIDMapping, applicationIDToApplicationTemplateIDMapping, requests, operation
func (_m *FormationAssignmentService) ProcessFormationAssignments(ctx context.Context, formationAssignmentsForObject []*model.FormationAssignment, runtimeContextIDToRuntimeIDMapping map[string]string, applicationIDToApplicationTemplateIDMapping map[string]string, requests []*webhookclient.NotificationRequest, operation func(context.Context, *formationassignment.AssignmentMappingPair) (bool, error)) error {
	ret := _m.Called(ctx, formationAssignmentsForObject, runtimeContextIDToRuntimeIDMapping, applicationIDToApplicationTemplateIDMapping, requests, operation)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []*model.FormationAssignment, map[string]string, map[string]string, []*webhookclient.NotificationRequest, func(context.Context, *formationassignment.AssignmentMappingPair) (bool, error)) error); ok {
		r0 = rf(ctx, formationAssignmentsForObject, runtimeContextIDToRuntimeIDMapping, applicationIDToApplicationTemplateIDMapping, requests, operation)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type NewFormationAssignmentServiceT interface {
	mock.TestingT
	Cleanup(func())
}

// NewFormationAssignmentService creates a new instance of FormationAssignmentService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewFormationAssignmentService(t NewFormationAssignmentServiceT) *FormationAssignmentService {
	mock := &FormationAssignmentService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
