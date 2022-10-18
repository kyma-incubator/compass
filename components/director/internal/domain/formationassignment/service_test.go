package formationassignment_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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

			svc := formationassignment.NewService(faRepo, uuidSvc, nil, nil, nil, nil, nil)

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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil)

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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil)

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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil)

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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil)

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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil)

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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil)

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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil)

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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil)

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
	applications := []*model.Application{{BaseEntity: &model.BaseEntity{ID: "app"}}}
	runtimes := []*model.Runtime{{ID: "runtime"}}
	runtimeContexts := []*model.RuntimeContext{{ID: "runtimeContext"}}

	formationAssignmentsForApplication := fixFormationAssignmentsWithObjectTypeAndID(graphql.FormationObjectTypeApplication, objectID, applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)
	formationAssignmentsForRuntime := fixFormationAssignmentsWithObjectTypeAndID(graphql.FormationObjectTypeRuntime, objectID, applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)
	formationAssignmentsForRuntimeContext := fixFormationAssignmentsWithObjectTypeAndID(graphql.FormationObjectTypeRuntimeContext, objectID, applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)
	formationAssignmentsForRuntimeContextWithParentInTheFormation := fixFormationAssignmentsForRtmCtxWithAppAndRtmCtx(graphql.FormationObjectTypeRuntimeContext, objectID, applications[0].ID, runtimeContexts[0].ID)

	formationAssignmentIDs := []string{"ID1", "ID2", "ID3", "ID4", "ID5", "ID6"}
	formationAssignmentIDsRtmCtxParentInFormation := []string{"ID1", "ID2", "ID3", "ID4"}

	formation := &model.Formation{
		Name: "testFormation",
		ID:   "ID",
	}
	testCases := []struct {
		Name                    string
		Context                 context.Context
		ObjectType              graphql.FormationObjectType
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		ApplicationRepo         func() *automock.ApplicationRepository
		RuntimeRepo             func() *automock.RuntimeRepository
		RuntimeContextRepo      func() *automock.RuntimeContextRepository
		UIDService              func() *automock.UIDService
		ExpectedOutput          []*model.FormationAssignment
		ExpectedErrorMsg        string
	}{
		{
			Name:       "Success",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeApplication,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				for i := range formationAssignmentsForApplication {
					repo.On("GetByTargetAndSource", ctxWithTenant, formationAssignmentsForApplication[i].Target, formationAssignmentsForApplication[i].Source, TestTenantID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.FormationAssignment)).Once()
					repo.On("Create", ctxWithTenant, formationAssignmentsForApplication[i]).Return(nil).Once()
				}
				repo.On("ListForIDs", ctxWithTenant, TestTenantID, formationAssignmentIDs).Return(formationAssignmentsForApplication, nil).Once()

				return repo
			},
			UIDService: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				for i := range formationAssignmentIDs {
					uidSvc.On("Generate").Return(formationAssignmentIDs[i]).Once()
				}
				return uidSvc
			},
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
			ExpectedOutput:   formationAssignmentsForApplication,
			ExpectedErrorMsg: "",
		},
		{
			Name:       "Success does not create formation assignment for application and itself",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeApplication,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				for i := range formationAssignmentsForApplication {
					repo.On("GetByTargetAndSource", ctxWithTenant, formationAssignmentsForApplication[i].Target, formationAssignmentsForApplication[i].Source, TestTenantID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.FormationAssignment)).Once()
					repo.On("Create", ctxWithTenant, formationAssignmentsForApplication[i]).Return(nil).Once()
				}
				repo.On("ListForIDs", ctxWithTenant, TestTenantID, formationAssignmentIDs).Return(formationAssignmentsForApplication, nil).Once()

				return repo
			},
			UIDService: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				for i := range formationAssignmentIDs {
					uidSvc.On("Generate").Return(formationAssignmentIDs[i]).Once()
				}
				return uidSvc
			},
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
			ExpectedOutput:   formationAssignmentsForApplication,
			ExpectedErrorMsg: "",
		},
		{
			Name:       "Success does not create formation assignment for runtime and itself",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeRuntime,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				for i := range formationAssignmentsForRuntime {
					repo.On("GetByTargetAndSource", ctxWithTenant, formationAssignmentsForRuntime[i].Target, formationAssignmentsForRuntime[i].Source, TestTenantID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.FormationAssignment)).Once()
					repo.On("Create", ctxWithTenant, formationAssignmentsForRuntime[i]).Return(nil).Once()
				}
				repo.On("ListForIDs", ctxWithTenant, TestTenantID, formationAssignmentIDs).Return(formationAssignmentsForRuntime, nil).Once()

				return repo
			},
			UIDService: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				for i := range formationAssignmentIDs {
					uidSvc.On("Generate").Return(formationAssignmentIDs[i]).Once()
				}
				return uidSvc
			},
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
			ExpectedOutput:   formationAssignmentsForRuntime,
			ExpectedErrorMsg: "",
		},
		{
			Name:       "Success does not create formation assignment for runtime context and itself",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeRuntimeContext,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				for i := range formationAssignmentsForRuntimeContext {
					repo.On("GetByTargetAndSource", ctxWithTenant, formationAssignmentsForRuntimeContext[i].Target, formationAssignmentsForRuntimeContext[i].Source, TestTenantID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.FormationAssignment)).Once()
					repo.On("Create", ctxWithTenant, formationAssignmentsForRuntimeContext[i]).Return(nil).Once()
				}
				repo.On("ListForIDs", ctxWithTenant, TestTenantID, formationAssignmentIDs).Return(formationAssignmentsForRuntimeContext, nil).Once()

				return repo
			},
			UIDService: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				for i := range formationAssignmentIDs {
					uidSvc.On("Generate").Return(formationAssignmentIDs[i]).Once()
				}
				return uidSvc
			},
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
				repo.On("GetByID", ctxWithTenant, TestTenantID, objectID).Return(&model.RuntimeContext{RuntimeID: "random"}, nil)
				return repo
			},
			ExpectedOutput:   formationAssignmentsForRuntimeContext,
			ExpectedErrorMsg: "",
		},
		{
			Name:       "Success does not create formation assignment for runtime context and it's parent runtime",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeRuntimeContext,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				for i := range formationAssignmentsForRuntimeContextWithParentInTheFormation {
					repo.On("GetByTargetAndSource", ctxWithTenant, formationAssignmentsForRuntimeContextWithParentInTheFormation[i].Target, formationAssignmentsForRuntimeContextWithParentInTheFormation[i].Source, TestTenantID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.FormationAssignment)).Once()
					repo.On("Create", ctxWithTenant, formationAssignmentsForRuntimeContextWithParentInTheFormation[i]).Return(nil).Once()
				}
				repo.On("ListForIDs", ctxWithTenant, TestTenantID, formationAssignmentIDsRtmCtxParentInFormation).Return(formationAssignmentsForRuntimeContextWithParentInTheFormation, nil).Once()

				return repo
			},
			UIDService: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				for i := range formationAssignmentIDsRtmCtxParentInFormation {
					uidSvc.On("Generate").Return(formationAssignmentIDsRtmCtxParentInFormation[i]).Once()
				}
				return uidSvc
			},
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
				repo.On("GetByID", ctxWithTenant, TestTenantID, objectID).Return(&model.RuntimeContext{RuntimeID: runtimes[0].ID}, nil)
				return repo
			},
			ExpectedOutput:   formationAssignmentsForRuntimeContextWithParentInTheFormation,
			ExpectedErrorMsg: "",
		},
		{
			Name:                    "Error while listing applications",
			Context:                 ctxWithTenant,
			ObjectType:              graphql.FormationObjectTypeApplication,
			FormationAssignmentRepo: unusedFormationAssignmentRepository,
			UIDService:              unusedUIDService,
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
			Name:                    "Error while listing runtimes",
			Context:                 ctxWithTenant,
			ObjectType:              graphql.FormationObjectTypeRuntime,
			FormationAssignmentRepo: unusedFormationAssignmentRepository,
			UIDService:              unusedUIDService,
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
			Name:                    "Error while listing runtime contexts",
			Context:                 ctxWithTenant,
			ObjectType:              graphql.FormationObjectTypeRuntimeContext,
			FormationAssignmentRepo: unusedFormationAssignmentRepository,
			UIDService:              unusedUIDService,
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
		{
			Name:                    "Error while getting runtime context by ID",
			Context:                 ctxWithTenant,
			ObjectType:              graphql.FormationObjectTypeRuntimeContext,
			FormationAssignmentRepo: unusedFormationAssignmentRepository,
			UIDService:              unusedUIDService,
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
				repo.On("GetByID", ctxWithTenant, TestTenantID, objectID).Return(nil, testErr)
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:       "Error while creating formation assignment",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeRuntimeContext,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetByTargetAndSource", ctxWithTenant, formationAssignmentsForRuntimeContext[0].Target, formationAssignmentsForRuntimeContext[0].Source, TestTenantID).Return(nil, testErr).Once()
				return repo
			},
			UIDService: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(formationAssignmentIDs[0]).Once()
				return uidSvc
			},
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
				repo.On("GetByID", ctxWithTenant, TestTenantID, objectID).Return(&model.RuntimeContext{RuntimeID: "random"}, nil)
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:       "Error while listing formation assignments",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeApplication,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				for i := range formationAssignmentsForApplication {
					repo.On("GetByTargetAndSource", ctxWithTenant, formationAssignmentsForApplication[i].Target, formationAssignmentsForApplication[i].Source, TestTenantID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.FormationAssignment)).Once()
					repo.On("Create", ctxWithTenant, formationAssignmentsForApplication[i]).Return(nil).Once()
				}
				repo.On("ListForIDs", ctxWithTenant, TestTenantID, formationAssignmentIDs).Return(nil, testErr).Once()

				return repo
			},
			UIDService: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				for i := range formationAssignmentIDs {
					uidSvc.On("Generate").Return(formationAssignmentIDs[i]).Once()
				}
				return uidSvc
			},
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
			ExpectedOutput:   nil,
			ExpectedErrorMsg: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationAssignmentRepo := testCase.FormationAssignmentRepo()
			appRepo := testCase.ApplicationRepo()
			runtimeRepo := testCase.RuntimeRepo()
			runtimeContextRepo := testCase.RuntimeContextRepo()
			uidSvc := testCase.UIDService()
			svc := formationassignment.NewService(formationAssignmentRepo, uidSvc, appRepo, runtimeRepo, runtimeContextRepo, nil, nil)

			// WHEN
			r, err := svc.GenerateAssignments(testCase.Context, TestTenantID, objectID, testCase.ObjectType, formation)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, testCase.ExpectedOutput, r)

			mock.AssertExpectationsForObjects(t, formationAssignmentRepo, appRepo, runtimeRepo, runtimeContextRepo)
		})
	}
}

func TestService_ProcessFormationAssignments(t *testing.T) {
	// GIVEN
	operationContainer := &operationContainer{content: []*formationassignment.AssignmentMappingPair{}, err: testErr}
	appID := "app"
	appID2 := "app2"
	runtimeID := "runtime"
	runtimeCtxID := "runtimeCtx"
	matchedApplicationAssignment := &model.FormationAssignment{
		Source:     appID2,
		SourceType: TestSourceType,
		Target:     appID,
		TargetType: "targetType",
	}
	matchedApplicationAssignmentReverse := &model.FormationAssignment{
		Source:     appID,
		SourceType: "targetType",
		Target:     appID2,
		TargetType: TestSourceType,
	}

	matchedRuntimeContextAssignment := &model.FormationAssignment{
		Source:     appID,
		SourceType: "APPLICATION",
		Target:     runtimeCtxID,
		TargetType: "RUNTIME_CONTEXT",
	}
	matchedRuntimeContextAssignmentReverse := &model.FormationAssignment{
		Source:     runtimeCtxID,
		SourceType: "RUNTIME_CONTEXT",
		Target:     appID,
		TargetType: "APPLICATION",
	}

	sourseNotMatchedAssignment := &model.FormationAssignment{
		Source:     "source3",
		SourceType: "sourceType",
		Target:     appID,
		TargetType: "targetType",
	}

	sourseNotMatchedAssignmentReverse := &model.FormationAssignment{
		Source:     appID,
		SourceType: "targetType",
		Target:     "source3",
		TargetType: "sourceType",
	}

	targetNotMatchedAssignment := &model.FormationAssignment{
		Source:     "source4",
		SourceType: "sourceType",
		Target:     "app3",
		TargetType: "targetType",
	}

	targetNotMatchedAssignmentReverse := &model.FormationAssignment{
		Source:     "app3",
		SourceType: "targetType",
		Target:     "source4",
		TargetType: "sourceType",
	}

	appToAppRequests, appToAppInputTemplate, appToAppInputTemplateReverse := fixNotificationRequestAndReverseRequest(appID, appID2, []string{appID, appID2}, matchedApplicationAssignment, matchedApplicationAssignmentReverse, "application", "application", true)
	rtmCtxToAppRequests, rtmCtxToAppInputTemplate, rtmCtxToAppInputTemplateReverse := fixNotificationRequestAndReverseRequest(runtimeID, appID, []string{appID, runtimeCtxID}, matchedRuntimeContextAssignment, matchedRuntimeContextAssignmentReverse, "runtime", "application", true)

	sourceNotMatchTemplateInput := &automock.TemplateInput{}
	sourceNotMatchTemplateInput.Mock.On("GetParticipantsIDs").Return([]string{"random", "notMatch"}).Times(1)

	//TODO test two apps and one runtime to verify the mapping
	testCases := []struct {
		Name                           string
		Context                        context.Context
		TemplateInput                  *automock.TemplateInput
		TemplateInputReverse           *automock.TemplateInput
		FormationAssignments           []*model.FormationAssignment
		Requests                       []*webhookclient.NotificationRequest
		Operation                      func(context.Context, *formationassignment.AssignmentMappingPair) error
		RuntimeContextToRuntimeMapping map[string]string
		ExpectedMappings               []*formationassignment.AssignmentMappingPair
		ExpectedErrorMsg               string
	}{
		{
			Name:                 "Success when match assignment for application",
			Context:              ctxWithTenant,
			TemplateInput:        appToAppInputTemplate,
			TemplateInputReverse: appToAppInputTemplateReverse,
			FormationAssignments: []*model.FormationAssignment{matchedApplicationAssignment, matchedApplicationAssignmentReverse},
			Requests:             appToAppRequests,
			Operation:            operationContainer.append,
			ExpectedMappings: []*formationassignment.AssignmentMappingPair{
				{
					Assignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             appToAppRequests[0],
						FormationAssignment: matchedApplicationAssignment,
					},
					ReverseAssignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             appToAppRequests[1],
						FormationAssignment: matchedApplicationAssignmentReverse,
					},
				},
				{
					Assignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             appToAppRequests[1],
						FormationAssignment: matchedApplicationAssignmentReverse,
					},
					ReverseAssignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             appToAppRequests[0],
						FormationAssignment: matchedApplicationAssignment,
					},
				},
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:                           "Success when match assignment for runtimeContext",
			Context:                        ctxWithTenant,
			TemplateInput:                  rtmCtxToAppInputTemplate,
			TemplateInputReverse:           rtmCtxToAppInputTemplateReverse,
			FormationAssignments:           []*model.FormationAssignment{matchedRuntimeContextAssignment, matchedRuntimeContextAssignmentReverse},
			Requests:                       rtmCtxToAppRequests,
			Operation:                      operationContainer.append,
			RuntimeContextToRuntimeMapping: map[string]string{runtimeCtxID: runtimeID},
			ExpectedMappings: []*formationassignment.AssignmentMappingPair{
				{
					Assignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             rtmCtxToAppRequests[0],
						FormationAssignment: matchedRuntimeContextAssignment,
					},
					ReverseAssignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             rtmCtxToAppRequests[1],
						FormationAssignment: matchedRuntimeContextAssignmentReverse,
					},
				},
				{
					Assignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             rtmCtxToAppRequests[1],
						FormationAssignment: matchedRuntimeContextAssignmentReverse,
					},
					ReverseAssignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             rtmCtxToAppRequests[0],
						FormationAssignment: matchedRuntimeContextAssignment,
					},
				},
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:                 "Success when no matching assignment for source found",
			Context:              ctxWithTenant,
			TemplateInput:        sourceNotMatchTemplateInput,
			TemplateInputReverse: &automock.TemplateInput{},
			FormationAssignments: []*model.FormationAssignment{sourseNotMatchedAssignment, sourseNotMatchedAssignmentReverse},
			Requests: []*webhookclient.NotificationRequest{
				{
					Webhook: graphql.Webhook{
						ApplicationID: &appID,
					},
					Object: sourceNotMatchTemplateInput},
			},
			Operation: operationContainer.append,
			ExpectedMappings: []*formationassignment.AssignmentMappingPair{
				{
					Assignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: sourseNotMatchedAssignment,
					},
					ReverseAssignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: sourseNotMatchedAssignmentReverse,
					},
				},
				{
					Assignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: sourseNotMatchedAssignmentReverse,
					},
					ReverseAssignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: sourseNotMatchedAssignment,
					},
				},
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:                 "Success when no match assignment for target found",
			Context:              ctxWithTenant,
			TemplateInput:        &automock.TemplateInput{},
			TemplateInputReverse: &automock.TemplateInput{},
			FormationAssignments: []*model.FormationAssignment{targetNotMatchedAssignment, targetNotMatchedAssignmentReverse},
			Requests:             appToAppRequests,
			Operation:            operationContainer.append,
			ExpectedMappings: []*formationassignment.AssignmentMappingPair{
				{
					Assignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: targetNotMatchedAssignment,
					},
					ReverseAssignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: targetNotMatchedAssignmentReverse,
					},
				},
				{
					Assignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: targetNotMatchedAssignmentReverse,
					},
					ReverseAssignment: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: targetNotMatchedAssignment,
					},
				},
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:                 "Fails on executing operation",
			Context:              ctxWithTenant,
			TemplateInput:        &automock.TemplateInput{},
			TemplateInputReverse: &automock.TemplateInput{},
			FormationAssignments: []*model.FormationAssignment{targetNotMatchedAssignment, targetNotMatchedAssignmentReverse},
			Requests:             appToAppRequests,
			Operation:            operationContainer.fail,
			ExpectedMappings:     []*formationassignment.AssignmentMappingPair{},
			ExpectedErrorMsg:     testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := formationassignment.NewService(nil, nil, nil, nil, nil, nil, nil)

			//WHEN
			err := svc.ProcessFormationAssignments(testCase.Context, testCase.FormationAssignments, testCase.RuntimeContextToRuntimeMapping, testCase.Requests, testCase.Operation)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			//THEN
			require.Equal(t, testCase.ExpectedMappings, operationContainer.content)

			mock.AssertExpectationsForObjects(t, testCase.TemplateInput, testCase.TemplateInputReverse)
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
			//repo := testCase.FormationAssignmentRepo()
			//conv := testCase.FormationAssignmentConverter()
			//uuidSvc := fixUUIDService()
			//svc := formationassignment.NewService(repo, uuidSvc, nil, nil, nil, conv, nil)
			//
			////WHEN
			//err := svc.UpdateFormationAssignment(testCase.Context, testCase.FormationAssignment, testCase.Response)
			//
			//if testCase.ExpectedErrorMsg != "" {
			//	require.Error(t, err)
			//	require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			//} else {
			//	require.NoError(t, err)
			//}
			//
			//// THEN
			//mock.AssertExpectationsForObjects(t, repo, conv, uuidSvc)
		})
	}
}

func TestService_CleanupFormationAssignment(t *testing.T) {
	// GIVEN
	source := "source"
	ok := http.StatusOK
	accepted := http.StatusNoContent
	notFound := http.StatusNotFound

	req := &webhookclient.NotificationRequest{
		Webhook:       graphql.Webhook{},
		Object:        nil,
		CorrelationID: "",
	}

	config := "{\"key\":\"value\"}"
	errMsg := "Test Error"
	marshaled, err := json.Marshal(formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   "Error while deleting assignment: config propagation is not supported on unassign notifications",
			ErrorCode: 2,
		},
	})
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

	marshaledErrClientError, err := json.Marshal(formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   errMsg,
			ErrorCode: 2,
		},
	})
	require.NoError(t, err)
	errorStateAssignment := &model.FormationAssignmentInput{
		Source: source,
		State:  string(model.CreateErrorAssignmentState),
		Value:  marshaledErrClientError,
	}
	errorStateAssignmentWithTenantAndID := &model.FormationAssignment{
		ID:       TestID,
		TenantID: TestTenantID,
		Source:   source,
		State:    string(model.CreateErrorAssignmentState),
		Value:    marshaledErrClientError,
	}

	marshaledTechnicalErr, err := json.Marshal(formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   "while deleting formation assignment with ID: \"c861c3db-1265-4143-a05c-1ced1291d816\": Test Error",
			ErrorCode: 1,
		},
	})
	require.NoError(t, err)
	errorStateAssignmentInputFailedToDelete := &model.FormationAssignmentInput{
		Source: source,
		State:  string(model.CreateErrorAssignmentState),
		Value:  marshaledTechnicalErr,
	}
	errorStateAssignmentFailedToDelete := &model.FormationAssignment{
		ID:       TestID,
		TenantID: TestTenantID,
		Source:   source,
		State:    string(model.CreateErrorAssignmentState),
		Value:    marshaledTechnicalErr,
	}

	marshaledFailedRequestTechnicalErr, err := json.Marshal(formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   errMsg,
			ErrorCode: 1,
		},
	})
	require.NoError(t, err)
	errorStateAssignmentInputFailedRequest := &model.FormationAssignmentInput{
		Source: source,
		State:  string(model.CreateErrorAssignmentState),
		Value:  marshaledFailedRequestTechnicalErr,
	}
	errorStateAssignmentFailedRequest := &model.FormationAssignment{
		ID:       TestID,
		TenantID: TestTenantID,
		Source:   source,
		State:    string(model.CreateErrorAssignmentState),
		Value:    marshaledFailedRequestTechnicalErr,
	}

	marshaledWhileDeletingError, err := json.Marshal(formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   "error while deleting assignment",
			ErrorCode: 1,
		},
	})
	require.NoError(t, err)
	errorStateAssignmentInputWhileDeletingError := &model.FormationAssignmentInput{
		Source: source,
		State:  string(model.CreateErrorAssignmentState),
		Value:  marshaledWhileDeletingError,
	}
	errorStateAssignmentWhileDeletingError := &model.FormationAssignment{
		ID:       TestID,
		TenantID: TestTenantID,
		Source:   source,
		State:    string(model.CreateErrorAssignmentState),
		Value:    marshaledWhileDeletingError,
	}

	successResponse := &webhook.Response{ActualStatusCode: &ok, SuccessStatusCode: &ok, IncompleteStatusCode: &accepted}
	incompleteResponse := &webhook.Response{ActualStatusCode: &accepted, SuccessStatusCode: &ok, IncompleteStatusCode: &accepted}
	errorResponse := &webhook.Response{ActualStatusCode: &notFound, SuccessStatusCode: &ok, IncompleteStatusCode: &accepted, Error: &errMsg}

	testCases := []struct {
		Name                           string
		Context                        context.Context
		FormationAssignmentRepo        func() *automock.FormationAssignmentRepository
		FormationAssignmentConverter   func() *automock.FormationAssignmentConverter
		NotificationService            func() *automock.NotificationService
		FormationAssignmentMappingPair *formationassignment.AssignmentMappingPair
		ExpectedErrorMsg               string
	}{
		{
			Name:    "success delete assignment when there is no request",
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
			NotificationService:            unusedNotificationService,
			FormationAssignmentMappingPair: fixAssignmentMappingPairWithID(TestID),
			ExpectedErrorMsg:               "",
		},
		{
			Name:    "success delete assignment when response code matches success status code",
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
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, req).Return(successResponse, nil).Once()
				return svc
			},
			FormationAssignmentMappingPair: fixAssignmentMappingPairWithIDAndRequest(TestID, req),
			ExpectedErrorMsg:               "",
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
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, req).Return(incompleteResponse, nil).Once()
				return svc
			},
			FormationAssignmentMappingPair: fixAssignmentMappingPairWithIDAndRequest(TestID, req),
			ExpectedErrorMsg:               "Error while deleting assignment: config propagation is not supported on unassign notifications",
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
					Value:  marshaledErrClientError,
				}).Return(errorStateAssignment).Once()
				return conv
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, req).Return(errorResponse, nil).Once()
				return svc
			},
			FormationAssignmentMappingPair: fixAssignmentMappingPairWithIDAndRequest(TestID, req),
			ExpectedErrorMsg:               "Received error from response: Test Error",
		},
		{
			Name:    "error while delete assignment when there is no request succeed in updating",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(testErr).Once()
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, errorStateAssignmentFailedToDelete).Return(nil).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", &model.FormationAssignment{
					ID:     TestID,
					Source: source,
					State:  string(model.DeleteErrorAssignmentState),
					Value:  marshaledTechnicalErr,
				}).Return(errorStateAssignmentInputFailedToDelete).Once()

				return conv
			},
			NotificationService:            unusedNotificationService,
			FormationAssignmentMappingPair: fixAssignmentMappingPairWithID(TestID),
			ExpectedErrorMsg:               "while deleting formation assignment with id",
		},
		{
			Name:    "error while delete assignment when there is no request then error while updating",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(testErr).Once()
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, errorStateAssignmentFailedToDelete).Return(testErr).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", &model.FormationAssignment{
					ID:     TestID,
					Source: source,
					State:  string(model.DeleteErrorAssignmentState),
					Value:  marshaledTechnicalErr,
				}).Return(errorStateAssignmentInputFailedToDelete).Once()

				return conv
			},
			NotificationService:            unusedNotificationService,
			FormationAssignmentMappingPair: fixAssignmentMappingPairWithID(TestID),
			ExpectedErrorMsg:               "while updating error state:",
		},
		{
			Name:    "error while delete assignment when there is no request succeed in updating",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, errorStateAssignmentFailedRequest).Return(nil).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", &model.FormationAssignment{
					ID:     TestID,
					Source: source,
					State:  string(model.DeleteErrorAssignmentState),
					Value:  marshaledFailedRequestTechnicalErr,
				}).Return(errorStateAssignmentInputFailedRequest).Once()

				return conv
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, req).Return(nil, testErr).Once()
				return svc
			},
			FormationAssignmentMappingPair: fixAssignmentMappingPairWithIDAndRequest(TestID, req),
			ExpectedErrorMsg:               "while sending notification for formation assignment with ID \"c861c3db-1265-4143-a05c-1ced1291d816\": Test Error",
		},
		{
			Name:    "error while delete assignment when there is no request then error while updating",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, errorStateAssignmentFailedToDelete).Return(testErr).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", &model.FormationAssignment{
					ID:     TestID,
					Source: source,
					State:  string(model.DeleteErrorAssignmentState),
					Value:  marshaledFailedRequestTechnicalErr,
				}).Return(errorStateAssignmentInputFailedToDelete).Once()

				return conv
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, req).Return(nil, testErr).Once()
				return svc
			},
			FormationAssignmentMappingPair: fixAssignmentMappingPairWithIDAndRequest(TestID, req),
			ExpectedErrorMsg:               "while updating error state:",
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
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, req).Return(incompleteResponse, nil).Once()
				return svc
			},
			FormationAssignmentMappingPair: fixAssignmentMappingPairWithIDAndRequest(TestID, req),
			ExpectedErrorMsg:               "while updating error state for formation with ID",
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
					Value:  marshaledErrClientError,
				}).Return(errorStateAssignment).Once()
				return conv
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, req).Return(errorResponse, nil).Once()
				return svc
			},
			FormationAssignmentMappingPair: fixAssignmentMappingPairWithIDAndRequest(TestID, req),
			ExpectedErrorMsg:               "while updating error state for formation with ID",
		},
		{
			Name:    "error when fails on delete when success response",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(testErr).Once()
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, errorStateAssignmentWhileDeletingError).Return(nil).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", &model.FormationAssignment{
					ID:     TestID,
					Source: source,
					State:  string(model.DeleteErrorAssignmentState),
					Value:  marshaledWhileDeletingError,
				}).Return(errorStateAssignmentInputWhileDeletingError).Once()
				return conv
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, req).Return(successResponse, nil).Once()
				return svc
			},
			FormationAssignmentMappingPair: fixAssignmentMappingPairWithIDAndRequest(TestID, req),
			ExpectedErrorMsg:               "while deleting formation assignment with id",
		},
		{
			Name:    "error when fails on delete when success response then fail on update",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(testErr).Once()
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, errorStateAssignmentWhileDeletingError).Return(testErr).Once()
				return repo
			},
			FormationAssignmentConverter: func() *automock.FormationAssignmentConverter {
				conv := &automock.FormationAssignmentConverter{}
				conv.On("ToInput", &model.FormationAssignment{
					ID:     TestID,
					Source: source,
					State:  string(model.DeleteErrorAssignmentState),
					Value:  marshaledWhileDeletingError,
				}).Return(errorStateAssignmentInputWhileDeletingError).Once()
				return conv
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, req).Return(successResponse, nil).Once()
				return svc
			},
			FormationAssignmentMappingPair: fixAssignmentMappingPairWithIDAndRequest(TestID, req),
			ExpectedErrorMsg:               "while updating error state: while deleting formation assignment with id",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.FormationAssignmentRepo()
			conv := testCase.FormationAssignmentConverter()
			notificationSvc := testCase.NotificationService()
			svc := formationassignment.NewService(repo, nil, nil, nil, nil, conv, notificationSvc)

			// WHEN
			err := svc.CleanupFormationAssignment(testCase.Context, testCase.FormationAssignmentMappingPair)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			mock.AssertExpectationsForObjects(t, repo, conv, notificationSvc)
		})
	}
}

//TODO create if not exist test
//TODO update formation assignment test
//TODO cleanup formation assignment
func unusedFormationAssignmentRepository() *automock.FormationAssignmentRepository {
	repo := &automock.FormationAssignmentRepository{}
	return repo
}

func unusedUIDService() *automock.UIDService {
	svc := &automock.UIDService{}
	return svc
}

func unusedNotificationService() *automock.NotificationService {
	svc := &automock.NotificationService{}
	return svc
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
	content []*formationassignment.AssignmentMappingPair
	err     error
}

func (o *operationContainer) append(_ context.Context, a *formationassignment.AssignmentMappingPair) error {
	o.content = append(o.content, a)
	return nil
}

func (o *operationContainer) fail(context.Context, *formationassignment.AssignmentMappingPair) error {
	return o.err
}

func (o *operationContainer) clear() {
	o.content = []*formationassignment.AssignmentMappingPair{}
}
