package formationassignment_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	emptyCtx      = context.TODO()
	externalTnt   = "externalTenant"
	ctxWithTenant = tenant.SaveToContext(emptyCtx, TestTenantID, externalTnt)

	testErr = errors.New("Test Error")

	faInput = fixFormationAssignmentModelInput(TestConfigValueRawJSON)

	first = 2
	after = "test"
)

func TestService_Create(t *testing.T) {
	testCases := []struct {
		Name                     string
		Context                  context.Context
		FormationAssignmentInput *model.FormationAssignmentInput
		FormationAssignmentRepo  func() *automock.FormationAssignmentRepository
		ExpectedOutput           string
		ExpectedErrorMsg         string
	}{
		{
			Name:                     "Success",
			Context:                  ctxWithTenant,
			FormationAssignmentInput: faInput,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Create", ctxWithTenant, faModel).Return(nil).Once()
				return repo
			},
			ExpectedOutput:   TestID,
			ExpectedErrorMsg: "",
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			ExpectedOutput:   "",
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:                     "Error when creating formation assignment",
			Context:                  ctxWithTenant,
			FormationAssignmentInput: faInput,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Create", ctxWithTenant, faModel).Return(testErr).Once()
				return repo
			},
			ExpectedOutput:   "",
			ExpectedErrorMsg: "while creating formation assignment for formation with ID:",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			uuidSvc := fixUUIDService()

			svc := formationassignment.NewService(nil, faRepo, uuidSvc, nil, nil, nil, nil)

			// WHEN
			r, err := svc.Create(testCase.Context, testCase.FormationAssignmentInput)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, testCase.ExpectedOutput, r)

			mock.AssertExpectationsForObjects(t, faRepo, uuidSvc)
		})
	}
}

func TestService_Get(t *testing.T) {
	testCases := []struct {
		Name                    string
		Context                 context.Context
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		ExpectedOutput          *model.FormationAssignment
		ExpectedErrorMsg        string
	}{
		{
			Name:    "Success",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(faModel, nil).Once()
				return repo
			},
			ExpectedOutput:   faModel,
			ExpectedErrorMsg: "",
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			ExpectedOutput:   nil,
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:    "Error when getting formation assignment",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: fmt.Sprintf("while getting formation assignment with ID: %q and tenant: %q", TestID, TestTenantID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(nil, faRepo, nil, nil, nil, nil, nil)

			// WHEN
			r, err := svc.Get(testCase.Context, TestID)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, testCase.ExpectedOutput, r)

			mock.AssertExpectationsForObjects(t, faRepo)
		})
	}
}

func TestService_GetForFormation(t *testing.T) {
	testCases := []struct {
		Name                    string
		Context                 context.Context
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		ExpectedOutput          *model.FormationAssignment
		ExpectedErrorMsg        string
	}{
		{
			Name:    "Success",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetForFormation", ctxWithTenant, TestTenantID, TestID, TestFormationID).Return(faModel, nil).Once()
				return repo
			},
			ExpectedOutput:   faModel,
			ExpectedErrorMsg: "",
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			ExpectedOutput:   nil,
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:    "Error when getting formation assignment for formation",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetForFormation", ctxWithTenant, TestTenantID, TestID, TestFormationID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: fmt.Sprintf("while getting formation assignment with ID: %q for formation with ID: %q", TestID, TestFormationID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(nil, faRepo, nil, nil, nil, nil, nil)

			// WHEN
			r, err := svc.GetForFormation(testCase.Context, TestID, TestFormationID)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, testCase.ExpectedOutput, r)

			mock.AssertExpectationsForObjects(t, faRepo)
		})
	}
}

func TestService_List(t *testing.T) {
	// GIVEN
	faModelPage := &model.FormationAssignmentPage{
		Data: []*model.FormationAssignment{faModel},
		PageInfo: &pagination.Page{
			StartCursor: "start",
			EndCursor:   "end",
			HasNextPage: false,
		},
		TotalCount: 1,
	}

	testCases := []struct {
		Name                    string
		Context                 context.Context
		InputPageSize           int
		InputCursor             string
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		ExpectedOutput          *model.FormationAssignmentPage
		ExpectedErrorMsg        string
	}{
		{
			Name:          "Success",
			Context:       ctxWithTenant,
			InputPageSize: first,
			InputCursor:   after,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("List", ctxWithTenant, first, after, TestTenantID).Return(faModelPage, nil).Once()
				return repo
			},
			ExpectedOutput:   faModelPage,
			ExpectedErrorMsg: "",
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			InputPageSize:    first,
			ExpectedOutput:   nil,
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:          "Error when listing formation assignment",
			Context:       ctxWithTenant,
			InputPageSize: first,
			InputCursor:   after,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("List", ctxWithTenant, first, after, TestTenantID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:             "Error when page size is invalid",
			Context:          ctxWithTenant,
			InputPageSize:    0,
			ExpectedOutput:   nil,
			ExpectedErrorMsg: "page size must be between 1 and 200",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(nil, faRepo, nil, nil, nil, nil, nil)

			// WHEN
			r, err := svc.List(testCase.Context, testCase.InputPageSize, testCase.InputCursor)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, testCase.ExpectedOutput, r)

			mock.AssertExpectationsForObjects(t, faRepo)
		})
	}
}

func TestService_ListByFormationIDs(t *testing.T) {
	// GIVEN
	faModelPages := []*model.FormationAssignmentPage{
		{
			Data: []*model.FormationAssignment{faModel},
			PageInfo: &pagination.Page{
				StartCursor: "start",
				EndCursor:   "end",
				HasNextPage: false,
			},
			TotalCount: 1,
		},
	}

	formationsIDs := []string{"formation-id-1", "formation-id-2"}

	testCases := []struct {
		Name                    string
		Context                 context.Context
		InputPageSize           int
		InputCursor             string
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		ExpectedOutput          []*model.FormationAssignmentPage
		ExpectedErrorMsg        string
	}{
		{
			Name:          "Success",
			Context:       ctxWithTenant,
			InputPageSize: first,
			InputCursor:   after,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListByFormationIDs", ctxWithTenant, TestTenantID, formationsIDs, first, after).Return(faModelPages, nil).Once()
				return repo
			},
			ExpectedOutput:   faModelPages,
			ExpectedErrorMsg: "",
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			InputPageSize:    first,
			ExpectedOutput:   nil,
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:          "Error when listing formation assignments by formations IDs",
			Context:       ctxWithTenant,
			InputPageSize: first,
			InputCursor:   after,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListByFormationIDs", ctxWithTenant, TestTenantID, formationsIDs, first, after).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:             "Error when page size is invalid",
			Context:          ctxWithTenant,
			InputPageSize:    0,
			ExpectedOutput:   nil,
			ExpectedErrorMsg: "page size must be between 1 and 200",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(nil, faRepo, nil, nil, nil, nil, nil)

			// WHEN
			r, err := svc.ListByFormationIDs(testCase.Context, formationsIDs, testCase.InputPageSize, testCase.InputCursor)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, testCase.ExpectedOutput, r)

			mock.AssertExpectationsForObjects(t, faRepo)
		})
	}
}

func TestService_ListFormationAssignmentsForObjectID(t *testing.T) {
	// GIVEN

	formationID := "formationID"
	objectID := "objectID"
	result := []*model.FormationAssignment{faModel}

	testCases := []struct {
		Name                    string
		Context                 context.Context
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		ExpectedOutput          []*model.FormationAssignment
		ExpectedErrorMsg        string
	}{
		{
			Name:    "Success",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListAllForObject", ctxWithTenant, TestTenantID, formationID, objectID).Return(result, nil).Once()
				return repo
			},
			ExpectedOutput:   result,
			ExpectedErrorMsg: "",
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			ExpectedOutput:   nil,
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:    "Error when listing formation assignment",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListAllForObject", ctxWithTenant, TestTenantID, formationID, objectID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(nil, faRepo, nil, nil, nil, nil, nil)

			// WHEN
			r, err := svc.ListFormationAssignmentsForObjectID(testCase.Context, formationID, objectID)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, testCase.ExpectedOutput, r)

			mock.AssertExpectationsForObjects(t, faRepo)
		})
	}
}

func TestService_Update(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name                     string
		Context                  context.Context
		FormationAssignmentInput *model.FormationAssignmentInput
		FormationAssignmentRepo  func() *automock.FormationAssignmentRepository
		ExpectedErrorMsg         string
	}{
		{
			Name:                     "Success",
			Context:                  ctxWithTenant,
			FormationAssignmentInput: faInput,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, faModel).Return(nil).Once()
				return repo
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:                     "Error when checking for formation assignment existence",
			Context:                  ctxWithTenant,
			FormationAssignmentInput: faInput,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(false, testErr).Once()
				return repo
			},
			ExpectedErrorMsg: fmt.Sprintf("while ensuring formation assignment with ID: %q exists", TestID),
		},
		{
			Name:                     "Error when formation assignment does not exists",
			Context:                  ctxWithTenant,
			FormationAssignmentInput: faInput,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(false, nil).Once()
				return repo
			},
			ExpectedErrorMsg: "Object not found",
		},
		{
			Name:                     "Error when updating formation assignment",
			Context:                  ctxWithTenant,
			FormationAssignmentInput: faInput,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, faModel).Return(testErr).Once()
				return repo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(nil, faRepo, nil, nil, nil, nil, nil)

			// WHEN
			err := svc.Update(testCase.Context, TestID, testCase.FormationAssignmentInput)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, faRepo)
		})
	}
}

func TestService_Delete(t *testing.T) {
	testCases := []struct {
		Name                    string
		Context                 context.Context
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		ExpectedErrorMsg        string
	}{
		{
			Name:    "Success",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(nil).Once()
				return repo
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:    "Error when deleting formation assignment",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(testErr).Once()
				return repo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(nil, faRepo, nil, nil, nil, nil, nil)

			// WHEN
			err := svc.Delete(testCase.Context, TestID)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, faRepo)
		})
	}
}

func TestService_Exists(t *testing.T) {
	testCases := []struct {
		Name                    string
		Context                 context.Context
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		ExpectedErrorMsg        string
	}{
		{
			Name:    "Success",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				return repo
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:    "Error when checking for formation assignment existence",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(false, testErr).Once()
				return repo
			},
			ExpectedErrorMsg: fmt.Sprintf("while checking formation assignment existence for ID: %q and tenant: %q", TestID, TestTenantID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(nil, faRepo, nil, nil, nil, nil, nil)

			// WHEN
			exists, err := svc.Exists(testCase.Context, TestID)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
				require.False(t, exists)
			} else {
				require.NoError(t, err)
				require.True(t, exists)
			}

			mock.AssertExpectationsForObjects(t, faRepo)
		})
	}
}

func TestService_GenerateAssignments(t *testing.T) {
	// GIVEN

	objectID := "objectID"
	objectType := graphql.FormationObjectTypeApplication
	applications := []*model.Application{&model.Application{BaseEntity: &model.BaseEntity{ID: "app"}}}
	runtimes := []*model.Runtime{&model.Runtime{ID: "runtime"}}
	runtimeContexts := []*model.RuntimeContext{&model.RuntimeContext{ID: "runtimeContext"}}
	result := []*model.FormationAssignment{
		&model.FormationAssignment{
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  string(objectType),
			Target:      applications[0].ID,
			TargetType:  string(graphql.FormationObjectTypeApplication),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      applications[0].ID,
			SourceType:  string(graphql.FormationObjectTypeApplication),
			Target:      objectID,
			TargetType:  string(objectType),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  string(objectType),
			Target:      runtimes[0].ID,
			TargetType:  string(graphql.FormationObjectTypeRuntime),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      runtimes[0].ID,
			SourceType:  string(graphql.FormationObjectTypeRuntime),
			Target:      objectID,
			TargetType:  string(objectType),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      objectID,
			SourceType:  string(objectType),
			Target:      runtimeContexts[0].ID,
			TargetType:  string(graphql.FormationObjectTypeRuntimeContext),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
		&model.FormationAssignment{
			FormationID: "ID",
			TenantID:    TestTenantID,
			Source:      runtimeContexts[0].ID,
			SourceType:  string(graphql.FormationObjectTypeRuntimeContext),
			Target:      objectID,
			TargetType:  string(objectType),
			State:       string(model.InitialAssignmentState),
			Value:       nil,
		},
	}
	formation := &model.Formation{
		Name: "testFormation",
		ID:   "ID",
	}
	testCases := []struct {
		Name               string
		Context            context.Context
		ApplicationRepo    func() *automock.ApplicationRepository
		RuntimeRepo        func() *automock.RuntimeRepository
		RuntimeContextRepo func() *automock.RuntimeContextRepository
		ExpectedOutput     []*model.FormationAssignment
		ExpectedErrorMsg   string
	}{
		{
			Name:    "Success",
			Context: ctxWithTenant,
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(applications, nil).Once()
				return repo
			},
			RuntimeRepo: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByScenarios", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(runtimes, nil).Once()
				return repo
			},
			RuntimeContextRepo: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(runtimeContexts, nil).Once()
				return repo
			},
			ExpectedOutput:   result,
			ExpectedErrorMsg: "",
		},
		{
			Name:    "Success does not create formation assignment for application and itself",
			Context: ctxWithTenant,
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(append(applications, &model.Application{BaseEntity: &model.BaseEntity{ID: objectID}}), nil).Once()
				return repo
			},
			RuntimeRepo: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByScenarios", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(runtimes, nil).Once()
				return repo
			},
			RuntimeContextRepo: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(runtimeContexts, nil).Once()
				return repo
			},
			ExpectedOutput:   result,
			ExpectedErrorMsg: "",
		},
		{
			Name:    "Success does not create formation assignment for runtime and itself",
			Context: ctxWithTenant,
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(applications, nil).Once()
				return repo
			},
			RuntimeRepo: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByScenarios", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(append(runtimes, &model.Runtime{ID: objectID}), nil).Once()
				return repo
			},
			RuntimeContextRepo: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(runtimeContexts, nil).Once()
				return repo
			},
			ExpectedOutput:   result,
			ExpectedErrorMsg: "",
		},
		{
			Name:    "Success does not create formation assignment for runtime context and itself",
			Context: ctxWithTenant,
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(applications, nil).Once()
				return repo
			},
			RuntimeRepo: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByScenarios", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(runtimes, nil).Once()
				return repo
			},
			RuntimeContextRepo: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(append(runtimeContexts, &model.RuntimeContext{ID: objectID}), nil).Once()
				return repo
			},
			ExpectedOutput:   result,
			ExpectedErrorMsg: "",
		},
		{
			Name:    "Error when listing applications",
			Context: ctxWithTenant,
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(nil, testErr).Once()
				return repo
			},
			RuntimeRepo:        unusedRuntimeRepository,
			RuntimeContextRepo: unusedRuntimeContextRepository,
			ExpectedOutput:     nil,
			ExpectedErrorMsg:   testErr.Error(),
		},
		{
			Name:    "Error when listing runtimes",
			Context: ctxWithTenant,
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(applications, nil).Once()
				return repo
			},
			RuntimeRepo: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByScenarios", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(nil, testErr).Once()
				return repo
			},
			RuntimeContextRepo: unusedRuntimeContextRepository,
			ExpectedOutput:     nil,
			ExpectedErrorMsg:   testErr.Error(),
		},
		{
			Name:    "Error when listing runtime contexts",
			Context: ctxWithTenant,
			ApplicationRepo: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("ListByScenariosNoPaging", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(applications, nil).Once()
				return repo
			},
			RuntimeRepo: func() *automock.RuntimeRepository {
				repo := &automock.RuntimeRepository{}
				repo.On("ListByScenarios", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(runtimes, nil).Once()
				return repo
			},
			RuntimeContextRepo: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("ListByScenarios", ctxWithTenant, TestTenantID, []string{formation.Name}).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := testCase.ApplicationRepo()
			runtimeRepo := testCase.RuntimeRepo()
			runtimeContextRepo := testCase.RuntimeContextRepo()
			svc := formationassignment.NewService(nil, nil, nil, appRepo, runtimeRepo, runtimeContextRepo, nil)

			// WHEN
			r, err := svc.GenerateAssignments(testCase.Context, TestTenantID, objectID, objectType, formation)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, testCase.ExpectedOutput, r)

			mock.AssertExpectationsForObjects(t, appRepo, runtimeRepo, runtimeContextRepo)
		})
	}
}

func TestService_ProcessFormationAssignments(t *testing.T) {
	// GIVEN
	txGen := txtest.NewTransactionContextGenerator(testErr)
	operationContainer := &operationContainer{content: []*assignmentResponsePair{}, err: testErr}
	appID := "app"
	runtimeID := "runtime"
	runtimCtxID := "runtimeCtx"
	runtimeContext := &model.RuntimeContext{RuntimeID: runtimeID}
	templateInput := &automock.TemplateInput{}
	templateInput.Mock.On("GetParticipantsIDs").Return([]string{"source1", "source2"})

	matchedApplicationAssignment := &model.FormationAssignment{
		Source:     "source1",
		Target:     appID,
		TargetType: "targetType",
	}

	matchedRuntimeContextAssignment := &model.FormationAssignment{
		Source:     "source2",
		Target:     runtimCtxID,
		TargetType: "RUNTIME_CONTEXT",
	}

	sourseNotMatchedAssignment := &model.FormationAssignment{
		Source:     "source3",
		Target:     appID,
		TargetType: "targetType",
	}

	targetNotMatchedAssignment := &model.FormationAssignment{
		Source:     "source4",
		Target:     "app2",
		TargetType: "targetType",
	}

	applicationRequest := &webhookclient.Request{Webhook: graphql.Webhook{ApplicationID: &appID}, Object: templateInput}
	runtimeContextRequest := &webhookclient.Request{Webhook: graphql.Webhook{RuntimeID: &runtimeID}, Object: templateInput}

	applicationResponse := &webhook.Response{Location: str.Ptr("location1")}
	runtimeContextResponse := &webhook.Response{Location: str.Ptr("location2")}

	testCases := []struct {
		Name                 string
		Context              context.Context
		TxFn                 func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		RuntimeContextRepo   func() *automock.RuntimeContextRepository
		FormationAssignments []*model.FormationAssignment
		Requests             []*webhookclient.Request
		Responses            []*webhook.Response
		Operation            func(context.Context, *model.FormationAssignment, *webhook.Response) error
		ExpectedMappings     []*assignmentResponsePair
		ExpectedErrorMsg     string
	}{
		{
			Name:                 "Success when match assignment for application",
			Context:              ctxWithTenant,
			TxFn:                 txGen.ThatSucceeds,
			RuntimeContextRepo:   unusedRuntimeContextRepository,
			FormationAssignments: []*model.FormationAssignment{matchedApplicationAssignment},
			Requests:             []*webhookclient.Request{applicationRequest, runtimeContextRequest},
			Responses:            []*webhook.Response{applicationResponse, runtimeContextResponse},
			Operation:            operationContainer.append,
			ExpectedMappings:     []*assignmentResponsePair{&assignmentResponsePair{assignment: matchedApplicationAssignment, response: applicationResponse}},
			ExpectedErrorMsg:     "",
		},
		{
			Name:    "Success when match assignment for runtimeContext",
			Context: ctxWithTenant,
			TxFn:    txGen.ThatSucceeds,
			RuntimeContextRepo: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenant, TestTenantID, runtimCtxID).Return(runtimeContext, nil).Once()
				return repo
			},
			FormationAssignments: []*model.FormationAssignment{matchedRuntimeContextAssignment},
			Requests:             []*webhookclient.Request{applicationRequest, runtimeContextRequest},
			Responses:            []*webhook.Response{applicationResponse, runtimeContextResponse},
			Operation:            operationContainer.append,
			ExpectedMappings:     []*assignmentResponsePair{&assignmentResponsePair{assignment: matchedRuntimeContextAssignment, response: runtimeContextResponse}},
			ExpectedErrorMsg:     "",
		},
		{
			Name:                 "Success when no matching assignment for source found",
			Context:              ctxWithTenant,
			TxFn:                 txGen.ThatSucceeds,
			RuntimeContextRepo:   unusedRuntimeContextRepository,
			FormationAssignments: []*model.FormationAssignment{sourseNotMatchedAssignment},
			Requests:             []*webhookclient.Request{applicationRequest, runtimeContextRequest},
			Responses:            []*webhook.Response{applicationResponse, runtimeContextResponse},
			Operation:            operationContainer.append,
			ExpectedMappings:     []*assignmentResponsePair{&assignmentResponsePair{assignment: sourseNotMatchedAssignment, response: nil}},
			ExpectedErrorMsg:     "",
		},
		{
			Name:                 "Success when no match assignment for target found",
			Context:              ctxWithTenant,
			TxFn:                 txGen.ThatSucceeds,
			RuntimeContextRepo:   unusedRuntimeContextRepository,
			FormationAssignments: []*model.FormationAssignment{targetNotMatchedAssignment},
			Requests:             []*webhookclient.Request{applicationRequest, runtimeContextRequest},
			Responses:            []*webhook.Response{applicationResponse, runtimeContextResponse},
			Operation:            operationContainer.append,
			ExpectedMappings:     []*assignmentResponsePair{&assignmentResponsePair{assignment: targetNotMatchedAssignment, response: nil}},
			ExpectedErrorMsg:     "",
		},
		{
			Name:                 "Fails on transaction begin",
			Context:              ctxWithTenant,
			TxFn:                 txGen.ThatFailsOnBegin,
			RuntimeContextRepo:   unusedRuntimeContextRepository,
			FormationAssignments: []*model.FormationAssignment{},
			Requests:             []*webhookclient.Request{},
			Responses:            []*webhook.Response{},
			Operation:            operationContainer.append,
			ExpectedMappings:     []*assignmentResponsePair{},
			ExpectedErrorMsg:     testErr.Error(),
		},
		{
			Name:    "Fails on fetching runtimeContext",
			Context: ctxWithTenant,
			TxFn:    txGen.ThatDoesntExpectCommit,
			RuntimeContextRepo: func() *automock.RuntimeContextRepository {
				repo := &automock.RuntimeContextRepository{}
				repo.On("GetByID", ctxWithTenant, TestTenantID, runtimCtxID).Return(nil, testErr).Once()
				return repo
			},
			FormationAssignments: []*model.FormationAssignment{matchedRuntimeContextAssignment},
			Requests:             []*webhookclient.Request{applicationRequest, runtimeContextRequest},
			Responses:            []*webhook.Response{applicationResponse, runtimeContextResponse},
			Operation:            operationContainer.append,
			ExpectedMappings:     []*assignmentResponsePair{},
			ExpectedErrorMsg:     testErr.Error(),
		},
		{
			Name:                 "Fails on executing operation",
			Context:              ctxWithTenant,
			TxFn:                 txGen.ThatDoesntExpectCommit,
			RuntimeContextRepo:   unusedRuntimeContextRepository,
			FormationAssignments: []*model.FormationAssignment{matchedApplicationAssignment},
			Requests:             []*webhookclient.Request{applicationRequest, runtimeContextRequest},
			Responses:            []*webhook.Response{applicationResponse, runtimeContextResponse},
			Operation:            operationContainer.fail,
			ExpectedMappings:     []*assignmentResponsePair{},
			ExpectedErrorMsg:     testErr.Error(),
		},
		{
			Name:                 "Fails on commit",
			Context:              ctxWithTenant,
			TxFn:                 txGen.ThatFailsOnCommit,
			RuntimeContextRepo:   unusedRuntimeContextRepository,
			FormationAssignments: []*model.FormationAssignment{matchedApplicationAssignment},
			Requests:             []*webhookclient.Request{applicationRequest, runtimeContextRequest},
			Responses:            []*webhook.Response{applicationResponse, runtimeContextResponse},
			Operation:            operationContainer.append,
			ExpectedMappings:     []*assignmentResponsePair{&assignmentResponsePair{assignment: matchedApplicationAssignment, response: applicationResponse}},
			ExpectedErrorMsg:     testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TxFn()
			runtimeContextRepo := testCase.RuntimeContextRepo()
			svc := formationassignment.NewService(transact, nil, nil, nil, nil, runtimeContextRepo, nil)

			// WHEN
			err := svc.ProcessFormationAssignments(testCase.Context, TestTenantID, testCase.FormationAssignments, testCase.Requests, testCase.Responses, testCase.Operation)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, testCase.ExpectedMappings, operationContainer.content)

			mock.AssertExpectationsForObjects(t, persist, runtimeContextRepo)
			operationContainer.clear()
		})
	}
}

func TestService_CreateOrUpdateFormationAssignment(t *testing.T) {
	// GIVEN
	source := "source"
	config := "{\"key\":\"value\"}"
	ok := 200
	incomplete := 204
	errMsg := "Test Error"
	marshaledErr, err := json.Marshal(struct{ Error string }{Error: errMsg})
	require.NoError(t, err)

	readystateAssignmentInput := &model.FormationAssignmentInput{
		Source: source,
		State:  string(model.ReadyAssignmentState),
	}
	configPendingStateAssignmentInput := &model.FormationAssignmentInput{
		Source: source,
		State:  string(model.ConfigPendingAssignmentState),
	}
	configAssignmentInput := &model.FormationAssignmentInput{
		Source: source,
		State:  string(model.ReadyAssignmentState),
		Value:  []byte(config),
	}
	errorStateAssignmentInput := &model.FormationAssignmentInput{
		Source: source,
		State:  string(model.CreateErrorAssignmentState),
		Value:  marshaledErr,
	}

	readyStateAssignment := &model.FormationAssignment{
		Source: source,
		State:  string(model.ReadyAssignmentState),
	}

	configPendingStateAssignment := &model.FormationAssignment{
		Source: source,
		State:  string(model.ConfigPendingAssignmentState),
	}

	configAssignment := &model.FormationAssignment{
		Source: source,
		State:  string(model.ReadyAssignmentState),
		Value:  []byte(config),
	}

	errorStateAssignment := &model.FormationAssignment{
		Source: source,
		State:  string(model.CreateErrorAssignmentState),
		Value:  marshaledErr,
	}

	successResponse := &webhook.Response{ActualStatusCode: &ok, SuccessStatusCode: &ok, IncompleteStatusCode: &incomplete}
	incompleteResponse := &webhook.Response{ActualStatusCode: &incomplete, SuccessStatusCode: &ok, IncompleteStatusCode: &incomplete}
	configResponse := &webhook.Response{ActualStatusCode: &ok, SuccessStatusCode: &ok, IncompleteStatusCode: &incomplete, Config: &config}
	errorResponse := &webhook.Response{ActualStatusCode: &ok, SuccessStatusCode: &ok, IncompleteStatusCode: &incomplete, Error: &errMsg}

	testCases := []struct {
		Name                         string
		Context                      context.Context
		FormationAssignmentRepo      func() *automock.FormationAssignmentRepository
		FormationAssignmentConverter func() *automock.FormationAssignmentConverter
		FormationAssignment          *model.FormationAssignment
		Response                     *webhook.Response
		ExpectedErrorMsg             string
	}{
		{
			Name:    "Success: ready state assignment",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Create", ctxWithTenant, fixFormationAssignmentModelWithIDAndTenantID(readyStateAssignment)).Return(nil).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", readyStateAssignment).Return(readystateAssignmentInput).Once()
				return conv
			},
			FormationAssignment: fixFormationAssignment(),
			Response:            successResponse,
			ExpectedErrorMsg:    "",
		},
		{
			Name:    "Success: incomplete state assignment",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Create", ctxWithTenant, fixFormationAssignmentModelWithIDAndTenantID(configPendingStateAssignment)).Return(nil).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", configPendingStateAssignment).Return(configPendingStateAssignmentInput).Once()
				return conv
			},
			FormationAssignment: fixFormationAssignment(),
			Response:            incompleteResponse,
			ExpectedErrorMsg:    "",
		},
		{
			Name:    "Success: assignment with config",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Create", ctxWithTenant, fixFormationAssignmentModelWithIDAndTenantID(configAssignment)).Return(nil).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", configAssignment).Return(configAssignmentInput).Once()
				return conv
			},
			FormationAssignment: fixFormationAssignment(),
			Response:            configResponse,
			ExpectedErrorMsg:    "",
		},
		{
			Name:    "Success: error state assignment",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Create", ctxWithTenant, fixFormationAssignmentModelWithIDAndTenantID(errorStateAssignment)).Return(nil).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", errorStateAssignment).Return(errorStateAssignmentInput).Once()
				return conv
			},
			FormationAssignment: fixFormationAssignment(),
			Response:            errorResponse,
			ExpectedErrorMsg:    "",
		},
		{
			Name:    "Fails on create",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Create", ctxWithTenant, fixFormationAssignmentModelWithIDAndTenantID(readyStateAssignment)).Return(testErr).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", readyStateAssignment).Return(readystateAssignmentInput).Once()
				return conv
			},
			FormationAssignment: fixFormationAssignment(),
			Response:            successResponse,
			ExpectedErrorMsg:    testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.FormationAssignmentRepo()
			conv := testCase.FormationAssignmentConverter()
			uuidSvc := fixUUIDService()
			svc := formationassignment.NewService(nil, repo, uuidSvc, nil, nil, nil, conv)

			// WHEN
			err := svc.CreateOrUpdateFormationAssignment(testCase.Context, testCase.FormationAssignment, testCase.Response)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			mock.AssertExpectationsForObjects(t, repo, conv, uuidSvc)
		})
	}
}

func TestService_CleanupFormationAssignment(t *testing.T) {
	// GIVEN
	source := "source"
	ok := http.StatusOK
	accepted := http.StatusNoContent
	notFound := http.StatusNotFound

	config := "{\"key\":\"value\"}"
	errMsg := "Test Error"
	marshaledErr, err := json.Marshal(struct{ Error string }{Error: errMsg})
	require.NoError(t, err)

	configAssignment := &model.FormationAssignmentInput{
		Source: source,
		State:  string(model.ReadyAssignmentState),
		Value:  []byte(config),
	}
	configAssignmentWithTenantAndID := &model.FormationAssignment{
		ID:       TestID,
		TenantID: TestTenantID,
		Source:   source,
		State:    string(model.ReadyAssignmentState),
		Value:    []byte(config),
	}

	errorStateAssignment := &model.FormationAssignmentInput{
		Source: source,
		State:  string(model.CreateErrorAssignmentState),
		Value:  marshaledErr,
	}
	errorStateAssignmentWithTenantAndID := &model.FormationAssignment{
		ID:       TestID,
		TenantID: TestTenantID,
		Source:   source,
		State:    string(model.CreateErrorAssignmentState),
		Value:    marshaledErr,
	}
	marshaled, err := json.Marshal(struct{ Error string }{Error: "Error while deleting assignment: config propagation is not supported on unassign notifications"})
	require.NoError(t, err)

	successResponse := &webhook.Response{ActualStatusCode: &ok, SuccessStatusCode: &ok, IncompleteStatusCode: &accepted}
	incompleteResponse := &webhook.Response{ActualStatusCode: &accepted, SuccessStatusCode: &ok, IncompleteStatusCode: &accepted}
	errorResponse := &webhook.Response{ActualStatusCode: &notFound, SuccessStatusCode: &ok, IncompleteStatusCode: &accepted, Error: &errMsg}

	testCases := []struct {
		Name                         string
		Context                      context.Context
		FormationAssignmentRepo      func() *automock.FormationAssignmentRepository
		FormationAssignmentConverter func() *automock.FormationAssignmentConverter
		FormationAssignment          *model.FormationAssignment
		Response                     *webhook.Response
		ExpectedErrorMsg             string
	}{
		{
			Name:    "success response code matches actual response code",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(nil).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				return conv
			},
			FormationAssignment: fixFormationAssignmentWithID(TestID),
			Response:            successResponse,
			ExpectedErrorMsg:    "",
		},
		{
			Name:    "incomplete response code matches actual response code",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, configAssignmentWithTenantAndID).Return(nil).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", &model.FormationAssignment{
					ID:     TestID,
					Source: source,
					State:  string(model.DeleteErrorAssignmentState),
					Value:  marshaled,
				}).Return(configAssignment).Once()
				return conv
			},
			FormationAssignment: fixFormationAssignmentWithID(TestID),
			Response:            incompleteResponse,
			ExpectedErrorMsg:    "",
		},
		{
			Name:    "error incomplete response code matches actual response code fails on update",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, configAssignmentWithTenantAndID).Return(testErr).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", &model.FormationAssignment{
					ID:     TestID,
					Source: source,
					State:  string(model.DeleteErrorAssignmentState),
					Value:  marshaled,
				}).Return(configAssignment).Once()
				return conv
			},
			FormationAssignment: fixFormationAssignmentWithID(TestID),
			Response:            incompleteResponse,
			ExpectedErrorMsg:    testErr.Error(),
		},
		{
			Name:    "response contains error",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, errorStateAssignmentWithTenantAndID).Return(nil).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", &model.FormationAssignment{
					ID:     TestID,
					Source: source,
					State:  string(model.DeleteErrorAssignmentState),
					Value:  marshaledErr,
				}).Return(errorStateAssignment).Once()
				return conv
			},
			FormationAssignment: fixFormationAssignmentWithID(TestID),
			Response:            errorResponse,
			ExpectedErrorMsg:    "",
		},
		{
			Name:    "response contains error fails on update",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, errorStateAssignmentWithTenantAndID).Return(testErr).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", &model.FormationAssignment{
					ID:     TestID,
					Source: source,
					State:  string(model.DeleteErrorAssignmentState),
					Value:  marshaledErr,
				}).Return(errorStateAssignment).Once()
				return conv
			},
			FormationAssignment: fixFormationAssignmentWithID(TestID),
			Response:            errorResponse,
			ExpectedErrorMsg:    testErr.Error(),
		},
		{
			Name:    "error when fails on delete",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(testErr).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				return conv
			},
			FormationAssignment: fixFormationAssignmentWithID(TestID),
			ExpectedErrorMsg:    testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.FormationAssignmentRepo()
			conv := testCase.FormationAssignmentConverter()
			svc := formationassignment.NewService(nil, repo, nil, nil, nil, nil, conv)

			// WHEN
			err := svc.CleanupFormationAssignment(testCase.Context, testCase.FormationAssignment, testCase.Response)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			mock.AssertExpectationsForObjects(t, repo, conv)
		})
	}
}

func unusedRuntimeRepository() *automock.RuntimeRepository {
	repo := &automock.RuntimeRepository{}
	return repo
}

func unusedRuntimeContextRepository() *automock.RuntimeContextRepository {
	repo := &automock.RuntimeContextRepository{}
	return repo
}

type assignmentResponsePair struct {
	assignment *model.FormationAssignment
	response   *webhook.Response
}

type operationContainer struct {
	content []*assignmentResponsePair
	err     error
}

func (o *operationContainer) append(_ context.Context, a *model.FormationAssignment, r *webhook.Response) error {
	o.content = append(o.content, &assignmentResponsePair{assignment: a, response: r})
	return nil
}

func (o *operationContainer) fail(context.Context, *model.FormationAssignment, *webhook.Response) error {
	return o.err
}

func (o *operationContainer) clear() {
	o.content = []*assignmentResponsePair{}
}
