package formationassignment_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
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

	testErr       = errors.New("Test Error")
	notFoundError = apperrors.NewNotFoundError(resource.FormationAssignment, TestID)

	faInput = fixFormationAssignmentModelInput(TestConfigValueRawJSON)
	fa      = fixFormationAssignmentModelWithFormationID(TestFormationID)

	assignOperation = model.AssignFormation

	first = 2
	after = "test"

	readyState         = string(model.ReadyAssignmentState)
	configPendingState = string(model.ConfigPendingAssignmentState)
	initialState       = string(model.InitialAssignmentState)
	deleteErrorState   = string(model.DeleteErrorAssignmentState)
	invalidState       = "asd"

	formation = &model.Formation{
		ID:                  TestFormationID,
		TenantID:            TestTenantID,
		FormationTemplateID: TestFormationTemplateID,
		Name:                TestFormationName,
		State:               TestReadyState,
	}
	reverseFa = fixReverseFormationAssignment(fa)

	rtmTypeLabelKey = "rtmTypeLabelKey"
	appTypeLabelKey = "appTypeLabelKey"

	appLbl = &model.Label{Value: appSubtype}
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
			ExpectedOutput: TestID,
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

			svc := formationassignment.NewService(faRepo, uuidSvc, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

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

func TestService_CreateIfNotExists(t *testing.T) {
	testCases := []struct {
		Name                     string
		Context                  context.Context
		FormationAssignmentInput *model.FormationAssignmentInput
		FormationAssignmentRepo  func() *automock.FormationAssignmentRepository
		UUIDService              func() *automock.UIDService
		ExpectedOutput           string
		ExpectedErrorMsg         string
	}{
		{
			Name:                     "Success when formation assignment does not exist",
			Context:                  ctxWithTenant,
			FormationAssignmentInput: faInput,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetByTargetAndSource", ctxWithTenant, faModel.Target, faModel.Source, TestTenantID, faModel.FormationID).Return(nil, apperrors.NewNotFoundError(resource.FormationAssignment, faModel.Source)).Once()
				repo.On("Create", ctxWithTenant, faModel).Return(nil).Once()
				return repo
			},
			ExpectedOutput: TestID,
		},
		{
			Name:                     "error when fetching formation assignment returns error different from not found",
			Context:                  ctxWithTenant,
			FormationAssignmentInput: faInput,
			UUIDService:              unusedUIDService,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetByTargetAndSource", ctxWithTenant, faModel.Target, faModel.Source, TestTenantID, faModel.FormationID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                     "Success when formation assignment does not exist",
			Context:                  ctxWithTenant,
			FormationAssignmentInput: faInput,
			UUIDService:              unusedUIDService,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetByTargetAndSource", ctxWithTenant, faModel.Target, faModel.Source, TestTenantID, faModel.FormationID).Return(faModel, nil).Once()
				return repo
			},
			ExpectedOutput: TestID,
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			UUIDService:      unusedUIDService,
			ExpectedOutput:   "",
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			uuidSvc := fixUUIDService()
			if testCase.UUIDService != nil {
				uuidSvc = testCase.UUIDService()
			}

			svc := formationassignment.NewService(faRepo, uuidSvc, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			r, err := svc.CreateIfNotExists(testCase.Context, testCase.FormationAssignmentInput)

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
			ExpectedOutput: faModel,
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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

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

func TestService_GetGlobalByID(t *testing.T) {
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
				repo.On("GetGlobalByID", ctxWithTenant, TestID).Return(faModel, nil).Once()
				return repo
			},
			ExpectedOutput: faModel,
		},
		{
			Name:    "Error when getting formation assignment globally",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetGlobalByID", ctxWithTenant, TestID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: fmt.Sprintf("while getting formation assignment with ID: %q globally", TestID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			r, err := svc.GetGlobalByID(testCase.Context, TestID)

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

func TestService_GetGlobalByIDAndFormationID(t *testing.T) {
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
				repo.On("GetGlobalByIDAndFormationID", ctxWithTenant, TestID, TestFormationID).Return(faModel, nil).Once()
				return repo
			},
			ExpectedOutput: faModel,
		},
		{
			Name:    "Error when getting formation assignment globally",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetGlobalByIDAndFormationID", ctxWithTenant, TestID, TestFormationID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: fmt.Sprintf("while getting formation assignment with ID: %q and formation ID: %q globally", TestID, TestFormationID),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			r, err := svc.GetGlobalByIDAndFormationID(testCase.Context, TestID, TestFormationID)

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
			ExpectedOutput: faModel,
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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

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

func TestService_GetReverseBySourceAndTarget(t *testing.T) {
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
				repo.On("GetReverseBySourceAndTarget", ctxWithTenant, TestTenantID, TestFormationID, TestSource, TestTarget).Return(faModel, nil).Once()
				return repo
			},
			ExpectedOutput: faModel,
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			ExpectedOutput:   nil,
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:    "Error when getting reverse formation assignment by formation ID, source and target",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetReverseBySourceAndTarget", ctxWithTenant, TestTenantID, TestFormationID, TestSource, TestTarget).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: fmt.Sprintf("while getting reverse formation assignment for formation ID: %q and source: %q and target: %q", TestFormationID, TestSource, TestTarget),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			r, err := svc.GetReverseBySourceAndTarget(testCase.Context, TestFormationID, TestSource, TestTarget)

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
			ExpectedOutput: faModelPage,
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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

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
			ExpectedOutput: faModelPages,
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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

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

func TestService_ListByFormationIDsNoPaging(t *testing.T) {
	// GIVEN
	faModels := [][]*model.FormationAssignment{{faModel}}

	formationsIDs := []string{"formation-id-1", "formation-id-2"}

	testCases := []struct {
		Name                    string
		Context                 context.Context
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		ExpectedOutput          [][]*model.FormationAssignment
		ExpectedErrorMsg        string
	}{
		{
			Name:    "Success",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListByFormationIDsNoPaging", ctxWithTenant, TestTenantID, formationsIDs).Return(faModels, nil).Once()
				return repo
			},
			ExpectedOutput: faModels,
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			ExpectedOutput:   nil,
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:    "Error when listing formation assignments by formations IDs",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListByFormationIDsNoPaging", ctxWithTenant, TestTenantID, formationsIDs).Return(nil, testErr).Once()
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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			r, err := svc.ListByFormationIDsNoPaging(testCase.Context, formationsIDs)

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

func TestService_GetAssignmentsForFormationWithStates(t *testing.T) {
	// GIVEN
	faModels := []*model.FormationAssignment{faModel}

	testCases := []struct {
		Name                    string
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		ExpectedOutput          []*model.FormationAssignment
		ExpectedErrorMsg        string
	}{
		{
			Name: "Success",
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetAssignmentsForFormationWithStates", ctxWithTenant, TestTenantID, TestFormationID, []string{TestStateInitial}).Return(faModels, nil).Once()
				return repo
			},
			ExpectedOutput: faModels,
		},
		{
			Name: "Error when listing formation assignments by formations ID and states",
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetAssignmentsForFormationWithStates", ctxWithTenant, TestTenantID, TestFormationID, []string{TestStateInitial}).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: "while getting formation assignments with states for formation with ID",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			r, err := svc.GetAssignmentsForFormationWithStates(ctxWithTenant, TestTenantID, TestFormationID, []string{TestStateInitial})

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

func TestService_GetAssignmentsForFormation(t *testing.T) {
	// GIVEN
	faModels := []*model.FormationAssignment{faModel}

	testCases := []struct {
		Name                    string
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		ExpectedOutput          []*model.FormationAssignment
		ExpectedErrorMsg        string
	}{
		{
			Name: "Success",
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetAssignmentsForFormation", ctxWithTenant, TestTenantID, TestFormationID).Return(faModels, nil).Once()
				return repo
			},
			ExpectedOutput: faModels,
		},
		{
			Name: "Error when listing formation assignments by formations ID",
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetAssignmentsForFormation", ctxWithTenant, TestTenantID, TestFormationID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: "while getting formation assignments for formation with ID",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			r, err := svc.GetAssignmentsForFormation(ctxWithTenant, TestTenantID, TestFormationID)

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
			ExpectedOutput: result,
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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

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

func TestService_ListFormationAssignmentsForObjectIDs(t *testing.T) {
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
				repo.On("ListAllForObjectIDs", ctxWithTenant, TestTenantID, formationID, []string{objectID}).Return(result, nil).Once()
				return repo
			},
			ExpectedOutput: result,
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
				repo.On("ListAllForObjectIDs", ctxWithTenant, TestTenantID, formationID, []string{objectID}).Return(nil, testErr).Once()
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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			r, err := svc.ListFormationAssignmentsForObjectIDs(testCase.Context, formationID, []string{objectID})

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
		Name                    string
		Context                 context.Context
		FormationAssignment     *model.FormationAssignment
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		ExpectedErrorMsg        string
	}{
		{
			Name:                "Success",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, faModel).Return(nil).Once()
				return repo
			},
		},
		{
			Name:             "Error when loading tenant from context",
			Context:          emptyCtx,
			ExpectedErrorMsg: "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:                "Error when checking for formation assignment existence",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(false, testErr).Once()
				return repo
			},
			ExpectedErrorMsg: fmt.Sprintf("while ensuring formation assignment with ID: %q exists", TestID),
		},
		{
			Name:                "Error when formation assignment does not exists",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(false, nil).Once()
				return repo
			},
			ExpectedErrorMsg: "Object not found",
		},
		{
			Name:                "Error when updating formation assignment",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			err := svc.Update(testCase.Context, TestID, testCase.FormationAssignment)

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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

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

func TestService_DeleteAssignmentsForObjectID(t *testing.T) {
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
				repo.On("DeleteAssignmentsForObjectID", ctxWithTenant, TestTenantID, TestID, TestSource).Return(nil).Once()
				return repo
			},
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
				repo.On("DeleteAssignmentsForObjectID", ctxWithTenant, TestTenantID, TestID, TestSource).Return(testErr).Once()
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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			err := svc.DeleteAssignmentsForObjectID(testCase.Context, TestID, TestSource)

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

			svc := formationassignment.NewService(faRepo, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

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

	formationParticipantsIDs := []string{applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID}

	formationAssignmentsForApplication := fixFormationAssignmentsWithObjectTypeAndID(model.FormationAssignmentTypeApplication, objectID, applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)
	formationAssignmentsForRuntime := fixFormationAssignmentsWithObjectTypeAndID(model.FormationAssignmentTypeRuntime, objectID, applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)
	formationAssignmentsForRuntimeContext := fixFormationAssignmentsWithObjectTypeAndID(model.FormationAssignmentTypeRuntimeContext, objectID, applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)
	formationAssignmentsForSelf := fixFormationAssignmentsForSelf(applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)
	formationAssignmentsForRuntimeContextWithParentInTheFormation := fixFormationAssignmentsForRtmCtxWithAppAndRtmCtx(model.FormationAssignmentTypeRuntimeContext, objectID, applications[0].ID, runtimeContexts[0].ID)

	allAssignments := append(append(formationAssignmentsForApplication, append(formationAssignmentsForRuntime, append(formationAssignmentsForRuntimeContext, formationAssignmentsForRuntimeContextWithParentInTheFormation...)...)...), formationAssignmentsForSelf...)

	formationAssignmentInputsForApplication := fixFormationAssignmentInputsWithObjectTypeAndID(model.FormationAssignmentTypeApplication, objectID, applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)
	formationAssignmentInputsForRuntime := fixFormationAssignmentInputsWithObjectTypeAndID(model.FormationAssignmentTypeRuntime, objectID, applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)
	formationAssignmentInputsForRuntimeContext := fixFormationAssignmentInputsWithObjectTypeAndID(model.FormationAssignmentTypeRuntimeContext, objectID, applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)
	formationAssignmentInputsForRuntimeContextWithParentInTheFormation := fixFormationAssignmentInputsForRtmCtxWithAppAndRtmCtx(model.FormationAssignmentTypeRuntimeContext, objectID, applications[0].ID, runtimeContexts[0].ID)

	formationAssignmentIDs := []string{"ID1", "ID2", "ID3", "ID4", "ID5", "ID6", "ID7"}
	formationAssignmentIDsRtmCtxParentInFormation := []string{"ID1", "ID2", "ID3", "ID4", "ID5"}

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
		ExpectedOutput          []*model.FormationAssignmentInput
		ExpectedErrorMsg        string
	}{
		{
			Name:       "Success",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeApplication,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListAllForObjectIDs", ctxWithTenant, TestTenantID, formation.ID, formationParticipantsIDs).Return(allAssignments, nil).Once()
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
			ExpectedOutput: formationAssignmentInputsForApplication,
		},
		{
			Name:       "Success does not create formation assignment for entity that is being unassigned asynchronously",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeApplication,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				unassignAppFormationAssignments := fixFormationAssignmentsWithObjectTypeAndID(model.FormationAssignmentTypeApplication, objectID, applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)
				repo.On("ListAllForObjectIDs", ctxWithTenant, TestTenantID, formation.ID, []string{applications[0].ID, objectID, runtimes[0].ID, runtimeContexts[0].ID}).Return(append(allAssignments, unassignAppFormationAssignments...), nil).Once()
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
			ExpectedOutput: formationAssignmentInputsForApplication,
		},
		{
			Name:       "Success does not create formation assignment for application and itself",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeApplication,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListAllForObjectIDs", ctxWithTenant, TestTenantID, formation.ID, []string{applications[0].ID, objectID, runtimes[0].ID, runtimeContexts[0].ID}).Return(allAssignments, nil).Once()
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
			ExpectedOutput: formationAssignmentInputsForApplication,
		},
		{
			Name:       "Success does not create formation assignment for runtime and itself",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeRuntime,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListAllForObjectIDs", ctxWithTenant, TestTenantID, formation.ID, []string{applications[0].ID, runtimes[0].ID, objectID, runtimeContexts[0].ID}).Return(allAssignments, nil).Once()
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
			ExpectedOutput: formationAssignmentInputsForRuntime,
		},
		{
			Name:       "Success does not create formation assignment for runtime context and itself",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeRuntimeContext,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListAllForObjectIDs", ctxWithTenant, TestTenantID, formation.ID, append(formationParticipantsIDs, objectID)).Return(allAssignments, nil).Once()
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
			ExpectedOutput: formationAssignmentInputsForRuntimeContext,
		},
		{
			Name:       "Success does not create formation assignment for runtime context and it's parent runtime",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeRuntimeContext,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListAllForObjectIDs", ctxWithTenant, TestTenantID, formation.ID, append(formationParticipantsIDs, objectID)).Return(allAssignments, nil).Once()
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
			ExpectedOutput: formationAssignmentInputsForRuntimeContextWithParentInTheFormation,
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
			Name:       "Error while listing all formation assignments",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeRuntimeContext,
			UIDService: unusedUIDService,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListAllForObjectIDs", ctxWithTenant, TestTenantID, formation.ID, append(formationParticipantsIDs, objectID)).Return(nil, testErr).Once()

				return repo
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
				return repo
			},
			ExpectedOutput:   nil,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:       "Error while getting runtime context by ID",
			Context:    ctxWithTenant,
			ObjectType: graphql.FormationObjectTypeRuntimeContext,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("ListAllForObjectIDs", ctxWithTenant, TestTenantID, formation.ID, append(formationParticipantsIDs, objectID)).Return(allAssignments, nil).Once()
				return repo
			},
			UIDService: unusedUIDService,
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
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationAssignmentRepo := testCase.FormationAssignmentRepo()
			appRepo := testCase.ApplicationRepo()
			runtimeRepo := testCase.RuntimeRepo()
			runtimeContextRepo := testCase.RuntimeContextRepo()
			uidSvc := testCase.UIDService()
			svc := formationassignment.NewService(formationAssignmentRepo, uidSvc, appRepo, runtimeRepo, runtimeContextRepo, nil, nil, nil, nil, nil, "", "")

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

func TestService_PersistAssignments(t *testing.T) {
	// GIVEN
	objectID := "objectID"
	applications := []*model.Application{{BaseEntity: &model.BaseEntity{ID: "app"}}}
	runtimes := []*model.Runtime{{ID: "runtime"}}
	runtimeContexts := []*model.RuntimeContext{{ID: "runtimeContext"}}

	formationAssignments := fixFormationAssignmentsWithObjectTypeAndID(model.FormationAssignmentTypeApplication, objectID, applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)
	formationAssignmentInputs := fixFormationAssignmentInputsWithObjectTypeAndID(model.FormationAssignmentTypeApplication, objectID, applications[0].ID, runtimes[0].ID, runtimeContexts[0].ID)

	formationAssignmentIDs := []string{"ID1", "ID2", "ID3", "ID4", "ID5", "ID6", "ID7"}
	testCases := []struct {
		Name                    string
		Context                 context.Context
		Input                   []*model.FormationAssignmentInput
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		UIDService              func() *automock.UIDService
		ExpectedOutput          []*model.FormationAssignment
		ExpectedErrorMsg        string
	}{
		{
			Name:    "Success",
			Context: ctxWithTenant,
			Input:   formationAssignmentInputs,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				for i := range formationAssignments {
					repo.On("GetByTargetAndSource", ctxWithTenant, formationAssignments[i].Target, formationAssignments[i].Source, TestTenantID, formationAssignments[i].FormationID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.FormationAssignment)).Once()
					repo.On("Create", ctxWithTenant, formationAssignments[i]).Return(nil).Once()
				}
				repo.On("ListForIDs", ctxWithTenant, TestTenantID, formationAssignmentIDs).Return(formationAssignments, nil).Once()
				return repo
			},
			UIDService: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				for i := range formationAssignmentIDs {
					uidSvc.On("Generate").Return(formationAssignmentIDs[i]).Once()
				}
				return uidSvc
			},
			ExpectedOutput: formationAssignments,
		},
		{
			Name:    "error while listing for ids",
			Context: ctxWithTenant,
			Input:   formationAssignmentInputs,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				for i := range formationAssignments {
					repo.On("GetByTargetAndSource", ctxWithTenant, formationAssignments[i].Target, formationAssignments[i].Source, TestTenantID, formationAssignments[i].FormationID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.FormationAssignment)).Once()
					repo.On("Create", ctxWithTenant, formationAssignments[i]).Return(nil).Once()
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
			ExpectedErrorMsg: "while listing formationAssignments",
		},
		{
			Name:    "error while creating formation assignments",
			Context: ctxWithTenant,
			Input:   formationAssignmentInputs,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetByTargetAndSource", ctxWithTenant, formationAssignments[0].Target, formationAssignments[0].Source, TestTenantID, formationAssignments[0].FormationID).Return(nil, apperrors.NewNotFoundErrorWithType(resource.FormationAssignment)).Once()
				repo.On("Create", ctxWithTenant, formationAssignments[0]).Return(testErr).Once()
				return repo
			},
			UIDService: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(formationAssignmentIDs[0]).Once()
				return uidSvc
			},
			ExpectedErrorMsg: "while creating formationAssignment for formation",
		},
		{
			Name:    "error while getting formation assignments",
			Context: ctxWithTenant,
			Input:   formationAssignmentInputs,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetByTargetAndSource", ctxWithTenant, formationAssignments[0].Target, formationAssignments[0].Source, TestTenantID, formationAssignments[0].FormationID).Return(nil, testErr).Once()
				return repo
			},
			UIDService: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(formationAssignmentIDs[0]).Once()
				return uidSvc
			},
			ExpectedErrorMsg: "while getting formation assignment by target",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationAssignmentRepo := testCase.FormationAssignmentRepo()
			uidSvc := testCase.UIDService()
			svc := formationassignment.NewService(formationAssignmentRepo, uidSvc, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			// WHEN
			r, err := svc.PersistAssignments(testCase.Context, TestTenantID, formationAssignmentInputs)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, testCase.ExpectedOutput, r)

			mock.AssertExpectationsForObjects(t, formationAssignmentRepo)
		})
	}
}

func TestService_ProcessFormationAssignments(t *testing.T) {
	// GIVEN
	operationContainer := &operationContainer{content: []*formationassignment.AssignmentMappingPairWithOperation{}, err: testErr}
	appID := "app"
	appID2 := "app2"
	appTemplateID := "appTemplate"
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
	appToAppRequests2, appToAppInputTemplate2, appToAppInputTemplateReverse2 := fixNotificationRequestAndReverseRequest(appID, appID2, []string{appID, appID2}, matchedApplicationAssignment, matchedApplicationAssignmentReverse, "application", "application", true)
	rtmCtxToAppRequests, rtmCtxToAppInputTemplate, rtmCtxToAppInputTemplateReverse := fixNotificationRequestAndReverseRequest(runtimeID, appID, []string{appID, runtimeCtxID}, matchedRuntimeContextAssignment, matchedRuntimeContextAssignmentReverse, "runtime", "application", true)

	appToAppRequestsWithAppTemplateWebhook, _, _ := fixNotificationRequestAndReverseRequest(appID, appID2, []string{appID, appID2}, matchedApplicationAssignment, matchedApplicationAssignmentReverse, "application", "application", true)
	appToAppRequestsWithAppTemplateWebhook[0].Webhook.ApplicationID = nil
	appToAppRequestsWithAppTemplateWebhook[0].Webhook.ApplicationTemplateID = str.Ptr(appTemplateID)

	sourceNotMatchTemplateInput := &automock.TemplateInput{}
	sourceNotMatchTemplateInput.Mock.On("GetParticipantsIDs").Return([]string{"random", "notMatch"}).Times(1)

	//TODO test two apps and one runtime to verify the mapping
	var testCases = []struct {
		Name                                      string
		Context                                   context.Context
		TemplateInput                             *automock.TemplateInput
		TemplateInputReverse                      *automock.TemplateInput
		FormationAssignments                      []*model.FormationAssignment
		Requests                                  []*webhookclient.FormationAssignmentNotificationRequest
		Operation                                 func(context.Context, *formationassignment.AssignmentMappingPairWithOperation) (bool, error)
		FormationOperation                        model.FormationOperation
		RuntimeContextToRuntimeMapping            map[string]string
		ApplicationsToApplicationTemplatesMapping map[string]string
		ExpectedMappings                          []*formationassignment.AssignmentMappingPairWithOperation
		ExpectedErrorMsg                          string
	}{
		{
			Name:                 "Success when match assignment for application",
			Context:              ctxWithTenant,
			TemplateInput:        appToAppInputTemplate,
			TemplateInputReverse: appToAppInputTemplateReverse,
			FormationAssignments: []*model.FormationAssignment{matchedApplicationAssignment, matchedApplicationAssignmentReverse},
			Requests:             appToAppRequests,
			Operation:            operationContainer.appendThatDoesNotProcessedReverse,
			FormationOperation:   assignOperation,
			ExpectedMappings: []*formationassignment.AssignmentMappingPairWithOperation{
				{
					AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
						AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             appToAppRequests[0],
							FormationAssignment: matchedApplicationAssignment,
						},
						ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             appToAppRequests[1],
							FormationAssignment: matchedApplicationAssignmentReverse,
						},
					},
					Operation: assignOperation,
				},
				{
					AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
						AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             appToAppRequests[1],
							FormationAssignment: matchedApplicationAssignmentReverse,
						},
						ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             appToAppRequests[0],
							FormationAssignment: matchedApplicationAssignment,
						},
					},
					Operation: assignOperation,
				},
			},
		},
		{
			Name:                 "Success when match assignment for application when webhook comes from applicationTemplate",
			Context:              ctxWithTenant,
			TemplateInput:        appToAppInputTemplate,
			TemplateInputReverse: appToAppInputTemplateReverse,
			FormationAssignments: []*model.FormationAssignment{matchedApplicationAssignment, matchedApplicationAssignmentReverse},
			Requests:             appToAppRequestsWithAppTemplateWebhook,
			Operation:            operationContainer.appendThatDoesNotProcessedReverse,
			ApplicationsToApplicationTemplatesMapping: map[string]string{appID: appTemplateID},
			FormationOperation:                        assignOperation,
			ExpectedMappings: []*formationassignment.AssignmentMappingPairWithOperation{
				{
					AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
						AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             appToAppRequestsWithAppTemplateWebhook[0],
							FormationAssignment: matchedApplicationAssignment,
						},
						ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             appToAppRequestsWithAppTemplateWebhook[1],
							FormationAssignment: matchedApplicationAssignmentReverse,
						},
					},
					Operation: assignOperation,
				},
				{
					AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
						AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             appToAppRequestsWithAppTemplateWebhook[1],
							FormationAssignment: matchedApplicationAssignmentReverse,
						},
						ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             appToAppRequestsWithAppTemplateWebhook[0],
							FormationAssignment: matchedApplicationAssignment,
						},
					},
					Operation: assignOperation,
				},
			},
		},
		{
			Name:                 "Does not process assignments multiple times",
			Context:              ctxWithTenant,
			TemplateInput:        appToAppInputTemplate2,
			TemplateInputReverse: appToAppInputTemplateReverse2,
			FormationAssignments: []*model.FormationAssignment{matchedApplicationAssignment, matchedApplicationAssignmentReverse},
			Requests:             appToAppRequests2,
			Operation:            operationContainer.appendThatProcessedReverse,
			FormationOperation:   assignOperation,
			ExpectedMappings: []*formationassignment.AssignmentMappingPairWithOperation{
				{
					AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
						AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             appToAppRequests2[0],
							FormationAssignment: matchedApplicationAssignment,
						},
						ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             appToAppRequests2[1],
							FormationAssignment: matchedApplicationAssignmentReverse,
						},
					},
					Operation: assignOperation,
				},
			},
		},
		{
			Name:                           "Success when match assignment for runtimeContext",
			Context:                        ctxWithTenant,
			TemplateInput:                  rtmCtxToAppInputTemplate,
			TemplateInputReverse:           rtmCtxToAppInputTemplateReverse,
			FormationAssignments:           []*model.FormationAssignment{matchedRuntimeContextAssignment, matchedRuntimeContextAssignmentReverse},
			Requests:                       rtmCtxToAppRequests,
			Operation:                      operationContainer.appendThatDoesNotProcessedReverse,
			RuntimeContextToRuntimeMapping: map[string]string{runtimeCtxID: runtimeID},
			FormationOperation:             assignOperation,
			ExpectedMappings: []*formationassignment.AssignmentMappingPairWithOperation{
				{
					AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
						AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             rtmCtxToAppRequests[0],
							FormationAssignment: matchedRuntimeContextAssignment,
						},
						ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             rtmCtxToAppRequests[1],
							FormationAssignment: matchedRuntimeContextAssignmentReverse,
						},
					},
					Operation: assignOperation,
				},
				{
					AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
						AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             rtmCtxToAppRequests[1],
							FormationAssignment: matchedRuntimeContextAssignmentReverse,
						},
						ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             rtmCtxToAppRequests[0],
							FormationAssignment: matchedRuntimeContextAssignment,
						},
					},
					Operation: assignOperation,
				},
			},
		},
		{
			Name:                 "Success when no matching assignment for source found",
			Context:              ctxWithTenant,
			TemplateInput:        sourceNotMatchTemplateInput,
			TemplateInputReverse: &automock.TemplateInput{},
			FormationAssignments: []*model.FormationAssignment{sourseNotMatchedAssignment, sourseNotMatchedAssignmentReverse},
			Requests: []*webhookclient.FormationAssignmentNotificationRequest{
				{
					Webhook: graphql.Webhook{
						ApplicationID: &appID,
					},
					Object: sourceNotMatchTemplateInput},
			},
			Operation:          operationContainer.appendThatDoesNotProcessedReverse,
			FormationOperation: assignOperation,
			ExpectedMappings: []*formationassignment.AssignmentMappingPairWithOperation{
				{
					AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
						AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             nil,
							FormationAssignment: sourseNotMatchedAssignment,
						},
						ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             nil,
							FormationAssignment: sourseNotMatchedAssignmentReverse,
						},
					},
					Operation: assignOperation,
				},
				{
					AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
						AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             nil,
							FormationAssignment: sourseNotMatchedAssignmentReverse,
						},
						ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             nil,
							FormationAssignment: sourseNotMatchedAssignment,
						},
					},
					Operation: assignOperation,
				},
			},
		},
		{
			Name:                 "Success when no match assignment for target found",
			Context:              ctxWithTenant,
			TemplateInput:        &automock.TemplateInput{},
			TemplateInputReverse: &automock.TemplateInput{},
			FormationAssignments: []*model.FormationAssignment{targetNotMatchedAssignment, targetNotMatchedAssignmentReverse},
			Requests:             appToAppRequests,
			Operation:            operationContainer.appendThatDoesNotProcessedReverse,
			FormationOperation:   assignOperation,
			ExpectedMappings: []*formationassignment.AssignmentMappingPairWithOperation{
				{
					AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
						AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             nil,
							FormationAssignment: targetNotMatchedAssignment,
						},
						ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             nil,
							FormationAssignment: targetNotMatchedAssignmentReverse,
						},
					},
					Operation: assignOperation,
				},
				{
					AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
						AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             nil,
							FormationAssignment: targetNotMatchedAssignmentReverse,
						},
						ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
							Request:             nil,
							FormationAssignment: targetNotMatchedAssignment,
						},
					},
					Operation: assignOperation,
				},
			},
		},
		{
			Name:                 "Fails on executing operation",
			Context:              ctxWithTenant,
			TemplateInput:        &automock.TemplateInput{},
			TemplateInputReverse: &automock.TemplateInput{},
			FormationAssignments: []*model.FormationAssignment{targetNotMatchedAssignment, targetNotMatchedAssignmentReverse},
			Requests:             appToAppRequests,
			Operation:            operationContainer.fail,
			FormationOperation:   assignOperation,
			ExpectedMappings:     []*formationassignment.AssignmentMappingPairWithOperation{},
			ExpectedErrorMsg:     testErr.Error(),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := formationassignment.NewService(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, "", "")

			//WHEN
			err := svc.ProcessFormationAssignments(testCase.Context, testCase.FormationAssignments, testCase.RuntimeContextToRuntimeMapping, testCase.ApplicationsToApplicationTemplatesMapping, testCase.Requests, testCase.Operation, testCase.FormationOperation)

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

func TestService_ProcessFormationAssignmentPair(t *testing.T) {
	// GIVEN
	config := "{\"key\":\"value\"}"
	ok := 200
	incomplete := 204

	deletingStateAssignment := &model.FormationAssignment{
		ID:          TestID,
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  model.FormationAssignmentTypeApplication,
		Target:      TestTarget,
		TargetType:  model.FormationAssignmentTypeApplication,
		FormationID: formation.ID,
		State:       string(model.DeletingAssignmentState),
	}
	marshaledErrTechnicalError, err := json.Marshal(formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   testErr.Error(),
			ErrorCode: 1,
		},
	})
	require.NoError(t, err)

	createErrorStateAssignment := &model.FormationAssignment{
		ID:          TestID,
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  model.FormationAssignmentTypeApplication,
		Target:      TestTarget,
		TargetType:  model.FormationAssignmentTypeApplication,
		FormationID: formation.ID,
		State:       string(model.CreateErrorAssignmentState),
		Value:       nil,
		Error:       marshaledErrTechnicalError,
	}
	initialStateSelfReferencingAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestSource, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.InitialAssignmentState), nil, nil)
	initialStateAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.InitialAssignmentState), nil, nil)
	reverseInitialStateAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestTarget, TestSource, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.InitialAssignmentState), nil, nil)
	readyStateAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), nil, nil)
	readyStateSelfReferencingAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestSource, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), nil, nil)
	configPendingStateWithConfigAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ConfigPendingAssignmentState), []byte(config), nil)
	configPendingStateAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ConfigPendingAssignmentState), nil, nil)
	configAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), []byte(config), nil)
	reverseConfigAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestTarget, TestSource, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), []byte(config), nil)
	reverseConfigPendingAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestTarget, TestSource, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ConfigPendingAssignmentState), []byte(config), nil)

	input := &webhook.FormationConfigurationChangeInput{
		Operation:   model.AssignFormation,
		FormationID: TestFormationID,
		Formation:   formation,
	}

	reqWebhook := &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: graphql.Webhook{
			ID: TestWebhookID,
		},
		Object:        input,
		CorrelationID: "",
	}

	whMode := graphql.WebhookModeAsyncCallback
	reqWebhookWithAsyncCallbackMode := &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: graphql.Webhook{
			ID:   TestWebhookID,
			Mode: &whMode,
			Type: graphql.WebhookTypeConfigurationChanged,
		},
		Object:        input,
		CorrelationID: "",
	}

	extendedFaNotificationInitialSelfReferencedReq := fixExtendedFormationAssignmentNotificationReq(reqWebhook, initialStateSelfReferencingAssignment)
	extendedFaNotificationInitialReq := fixExtendedFormationAssignmentNotificationReq(reqWebhook, initialStateAssignment)
	extendedFaNotificationInitialReqAsync := fixExtendedFormationAssignmentNotificationReq(reqWebhookWithAsyncCallbackMode, initialStateAssignment)

	testCases := []struct {
		Name                                 string
		Context                              context.Context
		FormationAssignmentRepo              func() *automock.FormationAssignmentRepository
		NotificationService                  func() *automock.NotificationService
		FormationAssignmentPairWithOperation *formationassignment.AssignmentMappingPairWithOperation
		FormationRepo                        func() *automock.FormationRepository
		FAStatusService                      func() *automock.StatusService
		FANotificationSvc                    func() *automock.FaNotificationService
		ExpectedIsReverseProcessed           bool
		ExpectedErrorMsg                     string
	}{
		{
			Name:    "Success: ready state assignment when assignment is already in ready state",
			Context: ctxWithTenant,
			FormationAssignmentPairWithOperation: &formationassignment.AssignmentMappingPairWithOperation{
				AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
					AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: fixFormationAssignmentModelWithIDAndTenantID(readyStateAssignment),
					},
					ReverseAssignmentReqMapping: nil,
				},
				Operation: model.AssignFormation,
			},
		},
		{
			Name:    "Success: ready state assignment with no request",
			Context: ctxWithTenant,
			FormationAssignmentPairWithOperation: &formationassignment.AssignmentMappingPairWithOperation{
				AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
					AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: initialStateAssignment.Clone(),
					},
					ReverseAssignmentReqMapping: nil,
				},
				Operation: model.AssignFormation,
			},
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, readyStateAssignment).Return(nil).Once()
				return repo
			},
		},
		{
			Name:    "Error when there is no request and update fails",
			Context: ctxWithTenant,
			FormationAssignmentPairWithOperation: &formationassignment.AssignmentMappingPairWithOperation{
				AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
					AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: initialStateAssignment.Clone(),
					},
					ReverseAssignmentReqMapping: nil,
				},
				Operation: model.AssignFormation,
			},
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, readyStateAssignment).Return(testErr).Once()
				return repo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "Success: state in response body",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &incomplete,
					State:                &configPendingState,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment.Clone(), reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("UpdateWithConstraints", ctxWithTenant, configPendingStateAssignment, assignOperation).Return(nil).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
		},
		{
			Name:    "Success: incomplete state assignment",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &incomplete,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("UpdateWithConstraints", ctxWithTenant, configPendingStateAssignment, assignOperation).Return(nil).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
		},
		{
			Name:    "Success: update self-referenced assignment to ready state without sending reverse notification",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialSelfReferencedReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &ok,
				}, nil)
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateSelfReferencingAssignment.Clone(), reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialSelfReferencedReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("UpdateWithConstraints", ctxWithTenant, readyStateSelfReferencingAssignment, assignOperation).Return(nil).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateSelfReferencingAssignment.Clone(), reqWebhook),
		},
		{
			Name:    "Error: update assignment to ready state if it is self-referenced formation assignment fails on update",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, readyStateSelfReferencingAssignment).Return(testErr).Once()
				return repo
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignment(initialStateSelfReferencingAssignment.Clone()),
			ExpectedErrorMsg:                     testErr.Error(),
		},
		{
			Name:    "Error: can't generate formation assignment extended notification",
			Context: ctxWithTenant,
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment.Clone(), reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(nil, testErr).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     testErr.Error(),
		},
		{
			Name:    "Error: state in body is not valid",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &incomplete,
					State:                &invalidState,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     fmt.Sprintf("The provided state in the response %q is not valid.", invalidState),
		},
		{
			Name:    "Error: state in body is INITIAL, but the previous assignment state is DELETING",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(reqWebhook, deletingStateAssignment)).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &incomplete,
					State:                &initialState,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(deletingStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(reqWebhook, deletingStateAssignment), nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(deletingStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     fmt.Sprintf("The provided state in the response %q is not valid.", initialState),
		},
		{
			Name:    "Error: state in body is DELETE_ERROR, but the previous assignment state is INITIAL",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &incomplete,
					State:                &deleteErrorState,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     fmt.Sprintf("The provided state in the response %q is not valid.", deleteErrorState),
		},
		{
			Name:    "Success: update assignment to ready state",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &ok,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("UpdateWithConstraints", ctxWithTenant, readyStateAssignment, assignOperation).Return(nil).Once()
				return updater
			},
		},
		{
			Name:    "Error: incomplete state assignment fails on update",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &incomplete,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("UpdateWithConstraints", ctxWithTenant, configPendingStateAssignment, assignOperation).Return(testErr).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     testErr.Error(),
		},
		{
			Name:    "Success with error from response",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					ActualStatusCode: &incomplete,
					Error:            str.Ptr(testErr.Error()),
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", ctxWithTenant, initialStateAssignment, testErr.Error(), formationassignment.AssignmentErrorCode(2), model.CreateErrorAssignmentState, assignOperation).Return(nil).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
		},
		{
			Name:    "Error with error from response while updating formation assignment",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					ActualStatusCode: &incomplete,
					Error:            str.Ptr(testErr.Error()),
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", ctxWithTenant, initialStateAssignment, testErr.Error(), formationassignment.AssignmentErrorCode(2), model.CreateErrorAssignmentState, assignOperation).Return(testErr).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     testErr.Error(),
		},
		{
			Name:    "Success while sending notification failing to update state to create error",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, createErrorStateAssignment).Return(nil).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(nil, testErr).Once()
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
		},
		{
			Name:    "Error while sending notification while updating state to create error",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, createErrorStateAssignment).Return(testErr).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(nil, testErr).Once()
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     testErr.Error(),
		},
		{
			Name:    "Success: webhook has mode ASYNC_CALLBACK",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, configPendingStateAssignment).Return(nil).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReqAsync).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &ok,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configPendingStateAssignment, reqWebhookWithAsyncCallbackMode)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReqAsync, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configPendingStateAssignment, reqWebhookWithAsyncCallbackMode),
		},
		{
			Name:    "ERROR: webhook has mode ASYNC_CALLBACK but fails on update",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, initialStateAssignment).Return(testErr).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReqAsync).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &ok,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhookWithAsyncCallbackMode)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReqAsync, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhookWithAsyncCallbackMode),
			ExpectedErrorMsg:                     testErr.Error(),
		},
		{
			Name:    "Success: assignment with config",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					Config:               &config,
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &ok,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("UpdateWithConstraints", ctxWithTenant, configAssignment, assignOperation).Return(nil).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				repo = testCase.FormationAssignmentRepo()
			}
			notificationSvc := &automock.NotificationService{}
			if testCase.NotificationService != nil {
				notificationSvc = testCase.NotificationService()
			}
			formationRepo := &automock.FormationRepository{}
			if testCase.FormationRepo != nil {
				formationRepo = testCase.FormationRepo()
			}
			faStatusService := &automock.StatusService{}
			if testCase.FAStatusService != nil {
				faStatusService = testCase.FAStatusService()
			}
			faNotificationSvc := &automock.FaNotificationService{}
			if testCase.FANotificationSvc != nil {
				faNotificationSvc = testCase.FANotificationSvc()
			}

			svc := formationassignment.NewService(repo, nil, nil, nil, nil, notificationSvc, faNotificationSvc, nil, formationRepo, faStatusService, rtmTypeLabelKey, appTypeLabelKey)

			// WHEN
			isReverseProcessed, err := svc.ProcessFormationAssignmentPair(testCase.Context, testCase.FormationAssignmentPairWithOperation)

			require.Equal(t, testCase.ExpectedIsReverseProcessed, isReverseProcessed)
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			mock.AssertExpectationsForObjects(t, repo, notificationSvc, formationRepo, faStatusService, faNotificationSvc)
		})
	}

	t.Run("success when propagating config to reverse assignment", func(t *testing.T) {
		mappingRequest := &webhookclient.FormationAssignmentNotificationRequest{
			Webhook: graphql.Webhook{
				ID: TestWebhookID,
			},
		}
		inputMock := &automock.TemplateInput{}
		inputMock.On("Clone").Return(inputMock)
		mappingRequest.Object = inputMock

		reverseMappingRequest := &webhookclient.FormationAssignmentNotificationRequest{
			Webhook: graphql.Webhook{
				ID: TestReverseWebhookID,
			},
		}
		reverseInputMock := &automock.TemplateInput{}
		reverseInputMock.On("Clone").Return(reverseInputMock)
		reverseMappingRequest.Object = reverseInputMock

		notificationSvc := &automock.NotificationService{}
		extendedReqWithReverseFA := fixExtendedFormationAssignmentNotificationReq(mappingRequest, initialStateAssignment)
		extendedReqWithReverseFA.ReverseFormationAssignment = reverseInitialStateAssignment
		notificationSvc.On("SendNotification", ctxWithTenant, extendedReqWithReverseFA).Return(&webhook.Response{
			Config:               &config,
			SuccessStatusCode:    &ok,
			IncompleteStatusCode: &incomplete,
			ActualStatusCode:     &ok,
		}, nil)

		extendedReqWithReverseFAForReverseNotification := fixExtendedFormationAssignmentNotificationReq(reverseMappingRequest, reverseInitialStateAssignment)
		extendedReqWithReverseFAForReverseNotification.ReverseFormationAssignment = configAssignment

		notificationSvc.On("SendNotification", ctxWithTenant, extendedReqWithReverseFAForReverseNotification).Return(&webhook.Response{
			Config:               &config,
			SuccessStatusCode:    &ok,
			IncompleteStatusCode: &incomplete,
			ActualStatusCode:     &ok,
		}, nil)

		repo := &automock.FormationAssignmentRepository{}

		assignmentPair := &formationassignment.AssignmentMappingPairWithOperation{
			AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
				AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
					Request:             mappingRequest,
					FormationAssignment: initialStateAssignment,
				},
				ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
					Request:             reverseMappingRequest,
					FormationAssignment: fixFormationAssignmentModelWithIDAndTenantID(reverseInitialStateAssignment),
				},
			},
			Operation: model.AssignFormation,
		}
		reverseInputMock.On("SetAssignment", assignmentPair.ReverseAssignmentReqMapping.FormationAssignment).Once()

		inputMock.On("SetReverseAssignment", assignmentPair.ReverseAssignmentReqMapping.FormationAssignment).Once()

		// once while processing the assignment and once when processing in the recursion call
		inputMock.On("SetAssignment", fixFormationAssignmentModelWithIDAndTenantID(configAssignment)).Twice()
		inputMock.On("SetReverseAssignment", fixFormationAssignmentModelWithIDAndTenantID(reverseConfigAssignment)).Once()

		reverseInputMock.On("SetAssignment", fixFormationAssignmentModelWithIDAndTenantID(reverseConfigAssignment)).Once()
		// once while processing the assignment and once when processing in the recursion call
		reverseInputMock.On("SetReverseAssignment", fixFormationAssignmentModelWithIDAndTenantID(configAssignment)).Twice()

		lblSvc := &automock.LabelService{}
		lblSvc.On("GetLabel", ctxWithTenant, TestTenantID, &model.LabelInput{
			Key:        appTypeLabelKey,
			ObjectID:   TestTarget,
			ObjectType: model.ApplicationLabelableObject,
		}).Return(appLbl, nil).Once()
		lblSvc.On("GetLabel", ctxWithTenant, TestTenantID, &model.LabelInput{
			Key:        appTypeLabelKey,
			ObjectID:   TestSource,
			ObjectType: model.ApplicationLabelableObject,
		}).Return(appLbl, nil).Once()

		formationRepo := &automock.FormationRepository{}
		formationRepo.On("Get", ctxWithTenant, TestFormationID, TestTenantID).Return(formation, nil).Times(2)

		faStatusService := &automock.StatusService{}
		faStatusService.On("UpdateWithConstraints", ctxWithTenant, configAssignment, assignOperation).Return(nil).Once()
		faStatusService.On("UpdateWithConstraints", ctxWithTenant, reverseConfigAssignment, assignOperation).Return(nil).Once()

		faNotificationSvc := &automock.FaNotificationService{}
		assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequestWithReverse(initialStateAssignment, reverseInitialStateAssignment, mappingRequest, reverseMappingRequest)
		faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedReqWithReverseFA, nil).Once()

		assignmentMapping = fixAssignmentMappingPairWithAssignmentAndRequestWithReverse(reverseInitialStateAssignment, configAssignment, reverseMappingRequest, mappingRequest)
		faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedReqWithReverseFAForReverseNotification, nil).Once()

		svc := formationassignment.NewService(repo, nil, nil, nil, nil, notificationSvc, faNotificationSvc, lblSvc, formationRepo, faStatusService, rtmTypeLabelKey, appTypeLabelKey)

		///WHEN
		isReverseProcessed, err := svc.ProcessFormationAssignmentPair(ctxWithTenant, assignmentPair)
		require.NoError(t, err)
		require.True(t, isReverseProcessed)

		mock.AssertExpectationsForObjects(t, inputMock, reverseInputMock, notificationSvc, repo, faStatusService, faNotificationSvc)
	})
	t.Run("error when updating to database in recursion call", func(t *testing.T) {
		mappingRequest := &webhookclient.FormationAssignmentNotificationRequest{
			Webhook: graphql.Webhook{
				ID: TestWebhookID,
			},
		}
		inputMock := &automock.TemplateInput{}
		inputMock.On("Clone").Return(inputMock)
		mappingRequest.Object = inputMock

		reverseMappingRequest := &webhookclient.FormationAssignmentNotificationRequest{
			Webhook: graphql.Webhook{
				ID: TestReverseWebhookID,
			},
		}
		reverseInputMock := &automock.TemplateInput{}
		reverseInputMock.On("Clone").Return(reverseInputMock)
		reverseMappingRequest.Object = reverseInputMock

		notificationSvc := &automock.NotificationService{}
		extendedReqWithReverseFA := fixExtendedFormationAssignmentNotificationReq(mappingRequest, initialStateAssignment)
		extendedReqWithReverseFA.ReverseFormationAssignment = reverseInitialStateAssignment
		notificationSvc.On("SendNotification", ctxWithTenant, extendedReqWithReverseFA).Return(&webhook.Response{
			Config:               &config,
			SuccessStatusCode:    &ok,
			IncompleteStatusCode: &incomplete,
			ActualStatusCode:     &ok,
		}, nil)

		extendedReqWithReverseFAForReverseNotification := fixExtendedFormationAssignmentNotificationReq(reverseMappingRequest, reverseInitialStateAssignment)
		extendedReqWithReverseFAForReverseNotification.ReverseFormationAssignment = configAssignment

		notificationSvc.On("SendNotification", ctxWithTenant, extendedReqWithReverseFAForReverseNotification).Return(&webhook.Response{
			Config:               &config,
			SuccessStatusCode:    &ok,
			IncompleteStatusCode: &incomplete,
			ActualStatusCode:     &ok,
		}, nil)

		repo := &automock.FormationAssignmentRepository{}

		assignmentPair := &formationassignment.AssignmentMappingPairWithOperation{
			AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
				AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
					Request:             mappingRequest,
					FormationAssignment: initialStateAssignment,
				},
				ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
					Request:             reverseMappingRequest,
					FormationAssignment: fixFormationAssignmentModelWithIDAndTenantID(reverseInitialStateAssignment),
				},
			},
			Operation: model.AssignFormation,
		}
		inputMock.On("SetAssignment", fixFormationAssignmentModelWithIDAndTenantID(configAssignment)).Once()
		inputMock.On("SetReverseAssignment", fixFormationAssignmentModelWithIDAndTenantID(reverseInitialStateAssignment)).Once()

		reverseInputMock.On("SetAssignment", assignmentPair.ReverseAssignmentReqMapping.FormationAssignment).Once()
		reverseInputMock.On("SetReverseAssignment", fixFormationAssignmentModelWithIDAndTenantID(configAssignment)).Once()

		lblSvc := &automock.LabelService{}
		lblSvc.On("GetLabel", ctxWithTenant, TestTenantID, &model.LabelInput{
			Key:        appTypeLabelKey,
			ObjectID:   TestTarget,
			ObjectType: model.ApplicationLabelableObject,
		}).Return(appLbl, nil).Once()
		lblSvc.On("GetLabel", ctxWithTenant, TestTenantID, &model.LabelInput{
			Key:        appTypeLabelKey,
			ObjectID:   TestSource,
			ObjectType: model.ApplicationLabelableObject,
		}).Return(appLbl, nil).Once()

		formationRepo := &automock.FormationRepository{}
		formationRepo.On("Get", ctxWithTenant, TestFormationID, TestTenantID).Return(formation, nil).Times(2)

		faStatusService := &automock.StatusService{}
		faStatusService.On("UpdateWithConstraints", ctxWithTenant, configAssignment, assignOperation).Return(nil).Once()
		faStatusService.On("UpdateWithConstraints", ctxWithTenant, reverseConfigAssignment, assignOperation).Return(testErr).Once()

		faNotificationSvc := &automock.FaNotificationService{}
		assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequestWithReverse(initialStateAssignment, reverseInitialStateAssignment, mappingRequest, reverseMappingRequest)
		faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedReqWithReverseFA, nil).Once()

		assignmentMapping = fixAssignmentMappingPairWithAssignmentAndRequestWithReverse(reverseInitialStateAssignment, configAssignment, reverseMappingRequest, mappingRequest)
		faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedReqWithReverseFAForReverseNotification, nil).Once()

		svc := formationassignment.NewService(repo, nil, nil, nil, nil, notificationSvc, faNotificationSvc, lblSvc, formationRepo, faStatusService, rtmTypeLabelKey, appTypeLabelKey)

		///WHEN
		isReverseProcessed, err := svc.ProcessFormationAssignmentPair(ctxWithTenant, assignmentPair)
		require.Error(t, err)
		require.Contains(t, err.Error(), testErr.Error())
		require.True(t, isReverseProcessed)

		mock.AssertExpectationsForObjects(t, inputMock, reverseInputMock, notificationSvc, repo, faStatusService, faNotificationSvc)
	})
	t.Run("success when reaching the maximum depth limit with two config pending assignments that return unfinished configurations", func(t *testing.T) {
		mappingRequest := &webhookclient.FormationAssignmentNotificationRequest{
			Webhook: graphql.Webhook{
				ID: TestWebhookID,
			},
		}
		inputMock := &automock.TemplateInput{}
		inputMock.On("Clone").Return(inputMock).Times(21)
		mappingRequest.Object = inputMock

		reverseMappingRequest := &webhookclient.FormationAssignmentNotificationRequest{
			Webhook: graphql.Webhook{
				ID: TestReverseWebhookID,
			},
		}
		reverseInputMock := &automock.TemplateInput{}
		reverseInputMock.On("Clone").Return(reverseInputMock).Times(21)
		reverseMappingRequest.Object = reverseInputMock

		notificationSvc := &automock.NotificationService{}
		extendedReqWithReverseFA := fixExtendedFormationAssignmentNotificationReq(mappingRequest, initialStateAssignment)
		extendedReqWithReverseFA.ReverseFormationAssignment = reverseInitialStateAssignment
		notificationSvc.On("SendNotification", ctxWithTenant, extendedReqWithReverseFA).Return(&webhook.Response{
			Config:               &config,
			SuccessStatusCode:    &ok,
			IncompleteStatusCode: &incomplete,
			ActualStatusCode:     &incomplete,
		}, nil)

		extendedReqWithReverseFAForReverseNotification := fixExtendedFormationAssignmentNotificationReq(reverseMappingRequest, reverseInitialStateAssignment)
		extendedReqWithReverseFAForReverseNotification.ReverseFormationAssignment = configPendingStateWithConfigAssignment

		notificationSvc.On("SendNotification", ctxWithTenant, extendedReqWithReverseFAForReverseNotification).Return(&webhook.Response{
			Config:               &config,
			SuccessStatusCode:    &ok,
			IncompleteStatusCode: &incomplete,
			ActualStatusCode:     &incomplete,
		}, nil)

		extendedReqWithReverseFAForReverseNotificationSecond := fixExtendedFormationAssignmentNotificationReq(mappingRequest, configPendingStateWithConfigAssignment)
		extendedReqWithReverseFAForReverseNotificationSecond.ReverseFormationAssignment = reverseConfigPendingAssignment

		notificationSvc.On("SendNotification", ctxWithTenant, extendedReqWithReverseFAForReverseNotificationSecond).Return(&webhook.Response{
			Config:               &config,
			SuccessStatusCode:    &ok,
			IncompleteStatusCode: &incomplete,
			ActualStatusCode:     &incomplete,
		}, nil)

		extendedReqWithReverseFAForReverseNotificationThird := fixExtendedFormationAssignmentNotificationReq(reverseMappingRequest, reverseConfigPendingAssignment)
		extendedReqWithReverseFAForReverseNotificationThird.ReverseFormationAssignment = configPendingStateWithConfigAssignment

		notificationSvc.On("SendNotification", ctxWithTenant, extendedReqWithReverseFAForReverseNotificationThird).Return(&webhook.Response{
			Config:               &config,
			SuccessStatusCode:    &ok,
			IncompleteStatusCode: &incomplete,
			ActualStatusCode:     &incomplete,
		}, nil)

		repo := &automock.FormationAssignmentRepository{}

		assignmentPair := &formationassignment.AssignmentMappingPairWithOperation{
			AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
				AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
					Request:             mappingRequest,
					FormationAssignment: initialStateAssignment.Clone(),
				},
				ReverseAssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
					Request:             reverseMappingRequest,
					FormationAssignment: fixFormationAssignmentModelWithIDAndTenantID(reverseInitialStateAssignment),
				},
			},
			Operation: model.AssignFormation,
		}
		reverseInputMock.On("SetAssignment", assignmentPair.ReverseAssignmentReqMapping.FormationAssignment)

		inputMock.On("SetReverseAssignment", assignmentPair.ReverseAssignmentReqMapping.FormationAssignment)

		inputMock.On("SetAssignment", fixFormationAssignmentModelWithIDAndTenantID(configPendingStateWithConfigAssignment))
		inputMock.On("SetReverseAssignment", fixFormationAssignmentModelWithIDAndTenantID(reverseConfigPendingAssignment))

		reverseInputMock.On("SetAssignment", fixFormationAssignmentModelWithIDAndTenantID(reverseConfigPendingAssignment))
		reverseInputMock.On("SetReverseAssignment", fixFormationAssignmentModelWithIDAndTenantID(configPendingStateWithConfigAssignment))

		lblSvc := &automock.LabelService{}
		lblSvc.On("GetLabel", ctxWithTenant, TestTenantID, &model.LabelInput{
			Key:        appTypeLabelKey,
			ObjectID:   TestTarget,
			ObjectType: model.ApplicationLabelableObject,
		}).Return(appLbl, nil)
		lblSvc.On("GetLabel", ctxWithTenant, TestTenantID, &model.LabelInput{
			Key:        appTypeLabelKey,
			ObjectID:   TestSource,
			ObjectType: model.ApplicationLabelableObject,
		}).Return(appLbl, nil)

		formationRepo := &automock.FormationRepository{}
		formationRepo.On("Get", ctxWithTenant, TestFormationID, TestTenantID).Return(formation, nil)

		faStatusService := &automock.StatusService{}
		faStatusService.On("UpdateWithConstraints", ctxWithTenant, configPendingStateWithConfigAssignment, assignOperation).Return(nil)
		faStatusService.On("UpdateWithConstraints", ctxWithTenant, reverseConfigPendingAssignment, assignOperation).Return(nil)

		faNotificationSvc := &automock.FaNotificationService{}
		assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequestWithReverse(initialStateAssignment, reverseInitialStateAssignment, mappingRequest, reverseMappingRequest)
		faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedReqWithReverseFA, nil)

		assignmentMapping = fixAssignmentMappingPairWithAssignmentAndRequestWithReverse(reverseInitialStateAssignment, configAssignment, reverseMappingRequest, mappingRequest)
		faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedReqWithReverseFAForReverseNotification, nil)

		assignmentMapping = fixAssignmentMappingPairWithAssignmentAndRequestWithReverse(reverseInitialStateAssignment, configPendingStateWithConfigAssignment, reverseMappingRequest, mappingRequest)
		faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedReqWithReverseFAForReverseNotification, nil)

		assignmentMapping = fixAssignmentMappingPairWithAssignmentAndRequestWithReverse(configPendingStateWithConfigAssignment, reverseConfigPendingAssignment, mappingRequest, reverseMappingRequest)
		faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedReqWithReverseFAForReverseNotificationSecond, nil)

		assignmentMapping = fixAssignmentMappingPairWithAssignmentAndRequestWithReverse(reverseConfigPendingAssignment, configPendingStateWithConfigAssignment, reverseMappingRequest, mappingRequest)
		faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedReqWithReverseFAForReverseNotificationThird, nil)

		svc := formationassignment.NewService(repo, nil, nil, nil, nil, notificationSvc, faNotificationSvc, lblSvc, formationRepo, faStatusService, rtmTypeLabelKey, appTypeLabelKey)

		///WHEN
		isReverseProcessed, err := svc.ProcessFormationAssignmentPair(ctxWithTenant, assignmentPair)
		require.NoError(t, err)
		require.True(t, isReverseProcessed)

		mock.AssertExpectationsForObjects(t, inputMock, reverseInputMock, notificationSvc, repo, faStatusService)
	})
}

func TestService_ProcessFormationAssignmentPairWithReset(t *testing.T) {
	// GIVEN
	config := "{\"key\":\"value\"}"
	ok := 200
	incomplete := 204

	deletingStateAssignment := &model.FormationAssignment{
		ID:          TestID,
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  model.FormationAssignmentTypeApplication,
		Target:      TestTarget,
		TargetType:  model.FormationAssignmentTypeApplication,
		FormationID: formation.ID,
		State:       string(model.DeletingAssignmentState),
	}
	marshaledErrTechnicalError, err := json.Marshal(formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   testErr.Error(),
			ErrorCode: 1,
		},
	})
	require.NoError(t, err)

	createErrorStateAssignment := &model.FormationAssignment{
		ID:          TestID,
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  model.FormationAssignmentTypeApplication,
		Target:      TestTarget,
		TargetType:  model.FormationAssignmentTypeApplication,
		FormationID: formation.ID,
		State:       string(model.CreateErrorAssignmentState),
		Value:       nil,
		Error:       marshaledErrTechnicalError,
	}
	initialStateSelfReferencingAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestSource, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.InitialAssignmentState), nil, nil)
	initialStateAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.InitialAssignmentState), nil, nil)
	readyStateAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), nil, nil)
	readyStateSelfReferencingAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestSource, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), nil, nil)
	configPendingStateAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ConfigPendingAssignmentState), nil, nil)
	configAssignment := fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), []byte(config), nil)

	input := &webhook.FormationConfigurationChangeInput{
		Operation:   model.AssignFormation,
		FormationID: TestFormationID,
		Formation:   formation,
	}

	reqWebhook := &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: graphql.Webhook{
			ID: TestWebhookID,
		},
		Object:        input,
		CorrelationID: "",
	}

	whMode := graphql.WebhookModeAsyncCallback
	reqWebhookWithAsyncCallbackMode := &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: graphql.Webhook{
			ID:   TestWebhookID,
			Mode: &whMode,
			Type: graphql.WebhookTypeConfigurationChanged,
		},
		Object:        input,
		CorrelationID: "",
	}

	extendedFaNotificationInitialSelfReferencedReq := fixExtendedFormationAssignmentNotificationReq(reqWebhook, initialStateSelfReferencingAssignment)
	extendedFaNotificationInitialReq := fixExtendedFormationAssignmentNotificationReq(reqWebhook, initialStateAssignment)
	extendedFaNotificationInitialReqAsync := fixExtendedFormationAssignmentNotificationReq(reqWebhookWithAsyncCallbackMode, initialStateAssignment)

	testCases := []struct {
		Name                                 string
		Context                              context.Context
		FormationAssignmentRepo              func() *automock.FormationAssignmentRepository
		NotificationService                  func() *automock.NotificationService
		FormationAssignmentPairWithOperation *formationassignment.AssignmentMappingPairWithOperation
		FormationRepo                        func() *automock.FormationRepository
		FAStatusService                      func() *automock.StatusService
		FANotificationSvc                    func() *automock.FaNotificationService
		ExpectedIsReverseProcessed           bool
		ExpectedErrorMsg                     string
	}{
		{
			Name:    "Success: ready state assignment when assignment is already in ready state",
			Context: ctxWithTenant,
			FormationAssignmentPairWithOperation: &formationassignment.AssignmentMappingPairWithOperation{
				AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
					AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: fixFormationAssignmentModelWithIDAndTenantID(readyStateAssignment),
					},
					ReverseAssignmentReqMapping: nil,
				},
				Operation: model.AssignFormation,
			},
		},
		{
			Name:    "Success: ready state assignment with no request",
			Context: ctxWithTenant,
			FormationAssignmentPairWithOperation: &formationassignment.AssignmentMappingPairWithOperation{
				AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
					AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: initialStateAssignment.Clone(),
					},
					ReverseAssignmentReqMapping: nil,
				},
				Operation: model.AssignFormation,
			},
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, readyStateAssignment).Return(nil).Once()
				return repo
			},
		},
		{
			Name:    "Error when there is no request and update fails",
			Context: ctxWithTenant,
			FormationAssignmentPairWithOperation: &formationassignment.AssignmentMappingPairWithOperation{
				AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
					AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
						Request:             nil,
						FormationAssignment: initialStateAssignment.Clone(),
					},
					ReverseAssignmentReqMapping: nil,
				},
				Operation: model.AssignFormation,
			},
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, readyStateAssignment).Return(testErr).Once()
				return repo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "Success: state in response body",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &incomplete,
					State:                &configPendingState,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment.Clone(), reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("UpdateWithConstraints", ctxWithTenant, configPendingStateAssignment, assignOperation).Return(nil).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
		},
		{
			Name:    "Success: incomplete state assignment",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &incomplete,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("UpdateWithConstraints", ctxWithTenant, configPendingStateAssignment, assignOperation).Return(nil).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
		},
		{
			Name:    "Success: update self-referenced assignment to ready state without sending reverse notification",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialSelfReferencedReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &ok,
				}, nil)
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateSelfReferencingAssignment.Clone(), reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialSelfReferencedReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("UpdateWithConstraints", ctxWithTenant, readyStateSelfReferencingAssignment, assignOperation).Return(nil).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateSelfReferencingAssignment.Clone(), reqWebhook),
		},
		{
			Name:    "Error: update assignment to ready state if it is self-referenced formation assignment fails on update",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, readyStateSelfReferencingAssignment).Return(testErr).Once()
				return repo
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignment(initialStateSelfReferencingAssignment.Clone()),
			ExpectedErrorMsg:                     testErr.Error(),
		},
		{
			Name:    "Error: can't generate formation assignment extended notification",
			Context: ctxWithTenant,
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment.Clone(), reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(nil, testErr).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     testErr.Error(),
		},
		{
			Name:    "Error: state in body is not valid",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &incomplete,
					State:                &invalidState,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     fmt.Sprintf("The provided state in the response %q is not valid.", invalidState),
		},
		{
			Name:    "Error: state in body is INITIAL, but the previous assignment state is DELETING",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(reqWebhook, deletingStateAssignment)).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &incomplete,
					State:                &initialState,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(deletingStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(reqWebhook, deletingStateAssignment), nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(deletingStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     fmt.Sprintf("The provided state in the response %q is not valid.", initialState),
		},
		{
			Name:    "Error: state in body is DELETE_ERROR, but the previous assignment state is INITIAL",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &incomplete,
					State:                &deleteErrorState,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     fmt.Sprintf("The provided state in the response %q is not valid.", deleteErrorState),
		},
		{
			Name:    "Success: update assignment to ready state",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &ok,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("UpdateWithConstraints", ctxWithTenant, readyStateAssignment, assignOperation).Return(nil).Once()
				return updater
			},
		},
		{
			Name:    "Error: incomplete state assignment fails on update",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &incomplete,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("UpdateWithConstraints", ctxWithTenant, configPendingStateAssignment, assignOperation).Return(testErr).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     testErr.Error(),
		},
		{
			Name:    "Success with error from response",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					ActualStatusCode: &incomplete,
					Error:            str.Ptr(testErr.Error()),
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", ctxWithTenant, initialStateAssignment, testErr.Error(), formationassignment.AssignmentErrorCode(2), model.CreateErrorAssignmentState, assignOperation).Return(nil).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
		},
		{
			Name:    "Error with error from response while updating formation assignment",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					ActualStatusCode: &incomplete,
					Error:            str.Ptr(testErr.Error()),
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", ctxWithTenant, initialStateAssignment, testErr.Error(), formationassignment.AssignmentErrorCode(2), model.CreateErrorAssignmentState, assignOperation).Return(testErr).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     testErr.Error(),
		},
		{
			Name:    "Success while sending notification failing to update state to create error",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, createErrorStateAssignment).Return(nil).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(nil, testErr).Once()
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
		},
		{
			Name:    "Error while sending notification while updating state to create error",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, createErrorStateAssignment).Return(testErr).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(nil, testErr).Once()
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
			ExpectedErrorMsg:                     testErr.Error(),
		},
		{
			Name:    "Success: webhook has mode ASYNC_CALLBACK: set assignment state to INITIAL",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, initialStateAssignment).Return(nil).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReqAsync).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &ok,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configPendingStateAssignment, reqWebhookWithAsyncCallbackMode)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReqAsync, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configPendingStateAssignment, reqWebhookWithAsyncCallbackMode),
		},
		{
			Name:    "ERROR: webhook has mode ASYNC_CALLBACK but fails on update",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, initialStateAssignment).Return(testErr).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReqAsync).Return(&webhook.Response{
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &ok,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhookWithAsyncCallbackMode)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReqAsync, nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhookWithAsyncCallbackMode),
			ExpectedErrorMsg:                     testErr.Error(),
		},
		{
			Name:    "Success: assignment with config",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				notificationSvc := &automock.NotificationService{}
				notificationSvc.On("SendNotification", ctxWithTenant, extendedFaNotificationInitialReq).Return(&webhook.Response{
					Config:               &config,
					SuccessStatusCode:    &ok,
					IncompleteStatusCode: &incomplete,
					ActualStatusCode:     &ok,
				}, nil)
				return notificationSvc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(extendedFaNotificationInitialReq, nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("UpdateWithConstraints", ctxWithTenant, configAssignment, assignOperation).Return(nil).Once()
				return updater
			},
			FormationAssignmentPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(initialStateAssignment, reqWebhook),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				repo = testCase.FormationAssignmentRepo()
			}
			notificationSvc := &automock.NotificationService{}
			if testCase.NotificationService != nil {
				notificationSvc = testCase.NotificationService()
			}
			formationRepo := &automock.FormationRepository{}
			if testCase.FormationRepo != nil {
				formationRepo = testCase.FormationRepo()
			}
			faStatusService := &automock.StatusService{}
			if testCase.FAStatusService != nil {
				faStatusService = testCase.FAStatusService()
			}
			faNotificationSvc := &automock.FaNotificationService{}
			if testCase.FANotificationSvc != nil {
				faNotificationSvc = testCase.FANotificationSvc()
			}

			svc := formationassignment.NewService(repo, nil, nil, nil, nil, notificationSvc, faNotificationSvc, nil, formationRepo, faStatusService, rtmTypeLabelKey, appTypeLabelKey)

			// WHEN
			isReverseProcessed, err := svc.ProcessFormationAssignmentPairWithReset(testCase.Context, testCase.FormationAssignmentPairWithOperation, true)

			require.Equal(t, testCase.ExpectedIsReverseProcessed, isReverseProcessed)
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			mock.AssertExpectationsForObjects(t, repo, notificationSvc, formationRepo, faStatusService, faNotificationSvc)
		})
	}
}

func TestService_CleanupFormationAssignment(t *testing.T) {
	// GIVEN
	ok := http.StatusOK
	accepted := http.StatusNoContent
	notFound := http.StatusNotFound
	mode := graphql.WebhookModeAsyncCallback

	req := &webhookclient.FormationAssignmentNotificationRequest{
		Webhook:       graphql.Webhook{},
		Object:        nil,
		CorrelationID: "",
	}

	callbackReq := &webhookclient.FormationAssignmentNotificationRequest{
		Webhook: graphql.Webhook{
			Mode: &mode,
		},
		Object:        nil,
		CorrelationID: "",
	}

	config := "{\"key\":\"value\"}"
	errMsg := "Test Error"

	configAssignmentWithoutConfig := fixFormationAssignmentModelWithParameters(TestID, formation.ID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), nil, nil)

	configAssignmentWithTenantAndID := fixFormationAssignmentModelWithParameters(TestID, formation.ID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.ReadyAssignmentState), []byte(config), nil)

	marshaledErrTechnicalError, err := json.Marshal(formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   "Error while deleting assignment: config propagation is not supported on unassign notifications",
			ErrorCode: 2,
		},
	})
	require.NoError(t, err)

	assignmentWithTenantAndIDInDeleteError := fixFormationAssignmentModelWithParameters(TestID, formation.ID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.DeleteErrorFormationState), nil, marshaledErrTechnicalError)

	marshaledDeleteError, err := json.Marshal(formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   "while deleting formation assignment with ID: \"c861c3db-1265-4143-a05c-1ced1291d816\": Test Error",
			ErrorCode: 1,
		},
	})
	require.NoError(t, err)

	deleteErrorStateAssignmentDeleteErr := fixFormationAssignmentModelWithParameters(TestID, formation.ID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.DeleteErrorFormationState), nil, marshaledDeleteError)

	marshaledFailedRequestTechnicalErr, err := json.Marshal(formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   errMsg,
			ErrorCode: 1,
		},
	})
	require.NoError(t, err)

	deleteErrorStateAssignmentTechnicalErr := fixFormationAssignmentModelWithParameters(TestID, formation.ID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.DeleteErrorFormationState), []byte(config), marshaledFailedRequestTechnicalErr)

	marshaledWhileDeletingError, err := json.Marshal(formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   "error while deleting assignment",
			ErrorCode: 1,
		},
	})
	require.NoError(t, err)
	deleteErrorStateAssignmentWhileDeletingErr := fixFormationAssignmentModelWithParameters(TestID, formation.ID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.DeleteErrorFormationState), []byte(config), marshaledWhileDeletingError)

	successResponse := &webhook.Response{ActualStatusCode: &ok, SuccessStatusCode: &ok, IncompleteStatusCode: &accepted}
	incompleteResponse := &webhook.Response{ActualStatusCode: &accepted, SuccessStatusCode: &ok, IncompleteStatusCode: &accepted}
	errorResponse := &webhook.Response{ActualStatusCode: &notFound, SuccessStatusCode: &ok, IncompleteStatusCode: &accepted, Error: &errMsg}

	successResponseWithStateInBody := &webhook.Response{ActualStatusCode: &ok, SuccessStatusCode: &ok, IncompleteStatusCode: &accepted, State: &readyState}
	deleteErrorResponseWithStateInBody := &webhook.Response{ActualStatusCode: &ok, SuccessStatusCode: &ok, IncompleteStatusCode: &accepted, State: &deleteErrorState}
	responseWithInvalidStateInBody := &webhook.Response{ActualStatusCode: &ok, SuccessStatusCode: &ok, IncompleteStatusCode: &accepted, State: &invalidState}

	testCases := []struct {
		Name                                        string
		Context                                     context.Context
		FormationAssignmentRepo                     func() *automock.FormationAssignmentRepository
		NotificationService                         func() *automock.NotificationService
		LabelService                                func() *automock.LabelService
		FormationRepo                               func() *automock.FormationRepository
		RuntimeContextRepo                          func() *automock.RuntimeContextRepository
		FANotificationSvc                           func() *automock.FaNotificationService
		FAStatusService                             func() *automock.StatusService
		FormationAssignmentMappingPairWithOperation *formationassignment.AssignmentMappingPairWithOperation
		ExpectedErrorMsg                            string
	}{
		{
			Name:    "success delete assignment when there is no request",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(nil).Once()
				return repo
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithID(TestID),
		},
		{
			Name:    "success delete assignment when response code matches success status code",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(successResponse, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("DeleteWithConstraints", ctxWithTenant, TestID).Return(nil).Once()
				return svc
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID, req),
		},
		{
			Name:    "success delete assignment when there is a READY state in response",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(successResponseWithStateInBody, nil).Once()
				return svc
			},
			FAStatusService: func() *automock.StatusService {
				svc := &automock.StatusService{}
				svc.On("DeleteWithConstraints", ctxWithTenant, TestID).Return(nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
		},
		{
			Name:    "sets assignment in deleting state when webhook is async callback",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(callbackReq, configAssignmentWithTenantAndID)).Return(successResponse, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), callbackReq)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(callbackReq, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), callbackReq),
		},
		{
			Name:    "incomplete response code matches actual response code",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, assignmentWithTenantAndIDInDeleteError).Return(nil).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithoutConfig)).Return(incompleteResponse, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithoutConfig.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithoutConfig), nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithoutConfig.Clone(), req),
			ExpectedErrorMsg: "Error while deleting assignment: config propagation is not supported on unassign notifications",
		},
		{
			Name:    "error when can't generate extended formation assignment notification",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, deleteErrorStateAssignmentTechnicalErr).Return(nil).Once()
				return repo
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(nil, testErr).Once()
				return faNotificationSvc
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "error when can't generate extended formation assignment notification and updating assignment to error state fails",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(false, testErr).Once()
				return repo
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(nil, testErr).Once()
				return faNotificationSvc
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "response contains error",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(errorResponse, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", ctxWithTenant, configAssignmentWithTenantAndID.Clone(), testErr.Error(), formationassignment.AssignmentErrorCode(2), model.DeleteErrorAssignmentState, assignOperation).Return(nil).Once()
				return updater
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
			ExpectedErrorMsg: "Received error from response: Test Error",
		},
		{
			Name:    "error while delete assignment when there is no request succeed in updating",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, deleteErrorStateAssignmentDeleteErr).Return(nil).Once()
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(testErr).Once()
				return repo
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithoutConfig.Clone(), nil),
			ExpectedErrorMsg: "while deleting formation assignment with id",
		},
		{
			Name:    "error while delete assignment when there is no request then error while updating",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, deleteErrorStateAssignmentDeleteErr).Return(testErr).Once()
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(testErr).Once()
				return repo
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithoutConfig.Clone(), nil),
			ExpectedErrorMsg: "while updating error state:",
		},
		{
			Name:    "error while delete assignment when there is no request then error while updating",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(notFoundError).Once()
				return repo
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), nil),
		},
		{
			Name:    "error while delete assignment when there is no request succeed in updating",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, deleteErrorStateAssignmentTechnicalErr).Return(nil).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(nil, testErr).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
			ExpectedErrorMsg: "while sending notification for formation assignment with ID",
		},
		{
			Name:    "error while delete assignment when there is no request then error while updating",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, deleteErrorStateAssignmentTechnicalErr).Return(testErr).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(nil, testErr).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
			ExpectedErrorMsg: "while sending notification for formation assignment with ID",
		},
		{
			Name:    "error incomplete response code matches actual response code fails on update",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, assignmentWithTenantAndIDInDeleteError).Return(testErr).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithoutConfig)).Return(incompleteResponse, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithoutConfig.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithoutConfig), nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithoutConfig.Clone(), req),
			ExpectedErrorMsg: "while updating error state for formation with ID",
		},
		{
			Name:    "response contains error fails on update",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(errorResponse, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", ctxWithTenant, configAssignmentWithTenantAndID.Clone(), testErr.Error(), formationassignment.AssignmentErrorCode(2), model.DeleteErrorAssignmentState, assignOperation).Return(testErr).Once()
				return updater
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "error when fails on delete when success response",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, deleteErrorStateAssignmentWhileDeletingErr).Return(nil).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(successResponse, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("DeleteWithConstraints", ctxWithTenant, TestID).Return(testErr).Once()
				return updater
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "error when fails on delete when success response",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(successResponse, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("DeleteWithConstraints", ctxWithTenant, TestID).Return(notFoundError).Once()
				return updater
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
		},
		{
			Name:    "error when fails on delete when success response then fail on update",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, deleteErrorStateAssignmentWhileDeletingErr).Return(testErr).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(successResponse, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("DeleteWithConstraints", ctxWithTenant, TestID).Return(testErr).Once()
				return updater
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "error when fails on delete when success response then fail on update with not found",
			Context: ctxWithTenant,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
				repo.On("Update", ctxWithTenant, deleteErrorStateAssignmentWhileDeletingErr).Return(notFoundError).Once()
				return repo
			},
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(successResponse, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("DeleteWithConstraints", ctxWithTenant, TestID).Return(testErr).Once()
				return updater
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
		},
		{
			Name:    "error when state in body is invalid",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(responseWithInvalidStateInBody, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
			ExpectedErrorMsg: fmt.Sprintf("The provided state in the response %q is not valid.", invalidState),
		},
		{
			Name:    "error state in body is DELETE_ERROR and fails on update with not found",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(deleteErrorResponseWithStateInBody, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", ctxWithTenant, configAssignmentWithTenantAndID.Clone(), "", formationassignment.AssignmentErrorCode(2), model.DeleteErrorAssignmentState, assignOperation).Return(notFoundError).Once()
				return updater
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
		},
		{
			Name:    "error state in body is DELETE_ERROR and fails on update",
			Context: ctxWithTenant,
			NotificationService: func() *automock.NotificationService {
				svc := &automock.NotificationService{}
				svc.On("SendNotification", ctxWithTenant, fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID)).Return(deleteErrorResponseWithStateInBody, nil).Once()
				return svc
			},
			FANotificationSvc: func() *automock.FaNotificationService {
				faNotificationSvc := &automock.FaNotificationService{}
				assignmentMapping := fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req)
				faNotificationSvc.On("GenerateFormationAssignmentNotificationExt", ctxWithTenant, assignmentMapping.AssignmentReqMapping, assignmentMapping.ReverseAssignmentReqMapping, model.AssignFormation).Return(fixExtendedFormationAssignmentNotificationReq(req, configAssignmentWithTenantAndID), nil).Once()
				return faNotificationSvc
			},
			FAStatusService: func() *automock.StatusService {
				updater := &automock.StatusService{}
				updater.On("SetAssignmentToErrorStateWithConstraints", ctxWithTenant, configAssignmentWithTenantAndID.Clone(), "", formationassignment.AssignmentErrorCode(2), model.DeleteErrorAssignmentState, assignOperation).Return(testErr).Once()
				return updater
			},
			FormationAssignmentMappingPairWithOperation: fixAssignmentMappingPairWithAssignmentAndRequest(configAssignmentWithTenantAndID.Clone(), req),
			ExpectedErrorMsg: "while updating error state for formation with ID",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				repo = testCase.FormationAssignmentRepo()
			}
			notificationSvc := &automock.NotificationService{}
			if testCase.NotificationService != nil {
				notificationSvc = testCase.NotificationService()
			}
			lblSvc := &automock.LabelService{}
			if testCase.LabelService != nil {
				lblSvc = testCase.LabelService()
			}
			formationRepo := &automock.FormationRepository{}
			if testCase.FormationRepo != nil {
				formationRepo = testCase.FormationRepo()
			}
			rtmCtxRepo := &automock.RuntimeContextRepository{}
			if testCase.RuntimeContextRepo != nil {
				rtmCtxRepo = testCase.RuntimeContextRepo()
			}
			updater := &automock.StatusService{}
			if testCase.FAStatusService != nil {
				updater = testCase.FAStatusService()
			}
			faNotificationSvc := &automock.FaNotificationService{}
			if testCase.FANotificationSvc != nil {
				faNotificationSvc = testCase.FANotificationSvc()
			}

			svc := formationassignment.NewService(repo, nil, nil, nil, rtmCtxRepo, notificationSvc, faNotificationSvc, lblSvc, formationRepo, updater, rtmTypeLabelKey, appTypeLabelKey)

			// WHEN
			isReverseProcessed, err := svc.CleanupFormationAssignment(testCase.Context, testCase.FormationAssignmentMappingPairWithOperation)

			require.False(t, isReverseProcessed)
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			mock.AssertExpectationsForObjects(t, repo, notificationSvc, lblSvc, formationRepo, rtmCtxRepo, updater, faNotificationSvc)
		})
	}
}

type operationContainer struct {
	content []*formationassignment.AssignmentMappingPairWithOperation
	err     error
}

func (o *operationContainer) appendThatProcessedReverse(_ context.Context, a *formationassignment.AssignmentMappingPairWithOperation) (bool, error) {
	o.content = append(o.content, a)
	return true, nil
}

func (o *operationContainer) appendThatDoesNotProcessedReverse(_ context.Context, a *formationassignment.AssignmentMappingPairWithOperation) (bool, error) {
	o.content = append(o.content, a)
	return false, nil
}

func (o *operationContainer) fail(context.Context, *formationassignment.AssignmentMappingPairWithOperation) (bool, error) {
	return false, o.err
}

func (o *operationContainer) clear() {
	o.content = []*formationassignment.AssignmentMappingPairWithOperation{}
}
