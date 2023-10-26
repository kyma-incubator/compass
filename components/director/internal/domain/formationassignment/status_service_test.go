package formationassignment_test

import (
	"context"
	"encoding/json"
	"strconv"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	initialStateAssignment       = fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, model.FormationAssignmentTypeApplication, model.FormationAssignmentTypeApplication, string(model.InitialAssignmentState), nil, nil)
	lastConfig                   = json.RawMessage(`{"foo": "bar"}`)
	assignmentWithStateAndConfig = fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, TestSourceType, TestTargetType, string(model.ConfigPendingAssignmentState), lastConfig, nil)
)

func TestStatusService_UpdateWithConstraints(t *testing.T) {
	preJoinPointDetails := fixNotificationStatusReturnedDetails(model.ApplicationResourceType, appSubtype, fa, reverseFa, formationconstraint.PreNotificationStatusReturned, initialState, "", TestTenantID)
	postJoinPointDetails := fixNotificationStatusReturnedDetails(model.ApplicationResourceType, appSubtype, fa, reverseFa, formationconstraint.PostNotificationStatusReturned, initialState, "", TestTenantID)

	// GIVEN
	testCases := []struct {
		Name                    string
		Context                 context.Context
		FormationAssignment     *model.FormationAssignment
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		NotificationSvc         func() *automock.FaNotificationService
		ConstraintEngine        func() *automock.ConstraintEngine
		ExpectedErrorMsg        string
	}{
		{
			Name:                "Success",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(initialStateAssignment, nil).Once()
				repo.On("Update", ctxWithTenant, fa).Return(nil).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PostNotificationStatusReturned, postJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation, initialStateAssignment.State, "").Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
		},
		{
			Name:                "Success with last formation assignment state",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(assignmentWithStateAndConfig, nil).Once()
				repo.On("Update", ctxWithTenant, fa).Return(nil).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PostNotificationStatusReturned, postJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation, assignmentWithStateAndConfig.State, strconv.Quote(string(lastConfig))).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
		},
		{
			Name:                "Error when can't prepare details",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(initialStateAssignment, nil).Once()
				return repo
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation, initialStateAssignment.State, "").Return(nil, testErr).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                "Error when enforcing PRE constraints",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(initialStateAssignment, nil).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(testErr).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation, initialStateAssignment.State, "").Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                "Error when checking for formation assignment existence",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                "Error when updating formation assignment",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(initialStateAssignment, nil).Once()
				repo.On("Update", ctxWithTenant, fa).Return(testErr).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation, initialStateAssignment.State, "").Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                "Not found error when updating formation assignment when update fails due to unauthorized",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(initialStateAssignment, nil).Once()
				repo.On("Update", ctxWithTenant, fa).Return(unauthorizedError).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation, initialStateAssignment.State, "").Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: notFoundError.Error(),
		},
		{
			Name:                "Error when enforcing POST constraints",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(initialStateAssignment, nil).Once()
				repo.On("Update", ctxWithTenant, fa).Return(nil).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PostNotificationStatusReturned, postJoinPointDetails, formation.FormationTemplateID).Return(testErr).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation, initialStateAssignment.State, "").Return(preJoinPointDetails, nil).Once()
				return notificationSvc
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
			constraintEngine := &automock.ConstraintEngine{}
			if testCase.ConstraintEngine != nil {
				constraintEngine = testCase.ConstraintEngine()
			}
			notificationSvc := &automock.FaNotificationService{}
			if testCase.NotificationSvc != nil {
				notificationSvc = testCase.NotificationSvc()
			}

			svc := formationassignment.NewFormationAssignmentStatusService(faRepo, constraintEngine, notificationSvc)

			// WHEN
			err := svc.UpdateWithConstraints(testCase.Context, testCase.FormationAssignment, assignOperation)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, faRepo, constraintEngine, notificationSvc)
		})
	}
}

func TestUpdater_SetAssignmentToErrorState(t *testing.T) {
	errorMsg := "Test Error Message"
	assignmentError := formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   errorMsg,
			ErrorCode: formationassignment.TechnicalError,
		},
	}
	marshaledAssignemntError, err := json.Marshal(assignmentError)
	require.NoError(t, err)

	fa := &model.FormationAssignment{
		ID:          TestID,
		FormationID: TestFormationID,
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      TestTarget,
		TargetType:  TestTargetType,
		State:       TestStateInitial,
		Value:       TestConfigValueRawJSON,
	}

	faErrorState := &model.FormationAssignment{
		ID:          TestID,
		FormationID: TestFormationID,
		TenantID:    TestTenantID,
		Source:      TestSource,
		SourceType:  TestSourceType,
		Target:      TestTarget,
		TargetType:  TestTargetType,
		State:       string(model.DeleteErrorFormationState),
		Value:       TestConfigValueRawJSON,
		Error:       marshaledAssignemntError,
	}

	reverseFaErrorState := fixReverseFormationAssignment(faErrorState)

	preJoinPointDetails := fixNotificationStatusReturnedDetails(model.ApplicationResourceType, appSubtype, faErrorState, reverseFaErrorState, formationconstraint.PreNotificationStatusReturned, initialState, "", TestTenantID)
	postJoinPointDetails := fixNotificationStatusReturnedDetails(model.ApplicationResourceType, appSubtype, faErrorState, reverseFaErrorState, formationconstraint.PostNotificationStatusReturned, initialState, "", TestTenantID)

	testCases := []struct {
		Name                    string
		Context                 context.Context
		FormationAssignment     *model.FormationAssignment
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		FormationRepo           func() *automock.FormationRepository
		ConstraintEngine        func() *automock.ConstraintEngine
		NotificationSvc         func() *automock.FaNotificationService
		FormationOperation      model.FormationOperation
		ExpectedErrorMsg        string
	}{
		{
			Name:                "Success",
			Context:             ctxWithTenant,
			FormationAssignment: fa.Clone(),
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(initialStateAssignment, nil).Once()
				repo.On("Update", ctxWithTenant, faErrorState).Return(nil).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PostNotificationStatusReturned, postJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, faErrorState, model.AssignFormation, initialStateAssignment.State, "").Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			FormationOperation: assignOperation,
		},
		{
			Name:                "Returns error when updating fails",
			Context:             ctxWithTenant,
			FormationAssignment: fa.Clone(),
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(initialStateAssignment, nil).Once()
				repo.On("Update", ctxWithTenant, faErrorState).Return(testErr).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, faErrorState, model.AssignFormation, initialStateAssignment.State, "").Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			FormationOperation: assignOperation,
			ExpectedErrorMsg:   testErr.Error(),
		},
		{
			Name:                "Returns not found error when updating fails with not found",
			Context:             ctxWithTenant,
			FormationAssignment: fa.Clone(),
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(initialStateAssignment, nil).Once()
				repo.On("Update", ctxWithTenant, faErrorState).Return(notFoundError).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, faErrorState, model.AssignFormation, initialStateAssignment.State, "").Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			FormationOperation: assignOperation,
			ExpectedErrorMsg:   notFoundError.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			faRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepo != nil {
				faRepo = testCase.FormationAssignmentRepo()
			}
			constraintEngine := &automock.ConstraintEngine{}
			if testCase.ConstraintEngine != nil {
				constraintEngine = testCase.ConstraintEngine()
			}
			formationRepo := &automock.FormationRepository{}
			if testCase.FormationRepo != nil {
				formationRepo = testCase.FormationRepo()
			}
			notificationSvc := &automock.FaNotificationService{}
			if testCase.NotificationSvc != nil {
				notificationSvc = testCase.NotificationSvc()
			}

			svc := formationassignment.NewFormationAssignmentStatusService(faRepo, constraintEngine, notificationSvc)

			// WHEN
			err := svc.SetAssignmentToErrorStateWithConstraints(testCase.Context, testCase.FormationAssignment, errorMsg, formationassignment.TechnicalError, model.DeleteErrorAssignmentState, assignOperation)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, faRepo, constraintEngine, formationRepo, notificationSvc)
		})
	}
}

func TestStatusService_DeleteWithConstraints(t *testing.T) {
	preJoinPointDetails := fixNotificationStatusReturnedDetails(model.ApplicationResourceType, appSubtype, fa, reverseFa, formationconstraint.PreNotificationStatusReturned, initialState, "", TestTenantID)
	postJoinPointDetails := fixNotificationStatusReturnedDetails(model.ApplicationResourceType, appSubtype, fa, reverseFa, formationconstraint.PostNotificationStatusReturned, initialState, "", TestTenantID)

	fa := fixFormationAssignmentModelWithFormationID(TestFormationID)
	faWithInitialStateAndNoConfig := fixFormationAssignmentModel(nil)
	faWithReadyStateAndNoConfig := fixFormationAssignmentModel(nil)
	faWithReadyStateAndNoConfig.State = string(model.ReadyAssignmentState)

	// GIVEN
	testCases := []struct {
		Name                    string
		Context                 context.Context
		InputID                 string
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		NotificationSvc         func() *automock.FaNotificationService
		ConstraintEngine        func() *automock.ConstraintEngine
		ExpectedErrorMsg        string
	}{
		{
			Name:    "Success with last formation assignment state",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa.Clone(), nil).Once()
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyStateAndNoConfig).Return(nil).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PostNotificationStatusReturned, postJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, faWithReadyStateAndNoConfig, model.UnassignFormation, fa.State, strconv.Quote(string(TestConfigValueRawJSON))).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
		},
		{
			Name:    "Success without last formation assignment state",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(faWithInitialStateAndNoConfig.Clone(), nil).Once()
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyStateAndNoConfig).Return(nil).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PostNotificationStatusReturned, postJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, faWithReadyStateAndNoConfig, model.UnassignFormation, faWithInitialStateAndNoConfig.State, "").Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
		},
		{
			Name:             "Returns error when there is no tenant in the context",
			Context:          emptyCtx,
			InputID:          TestID,
			ExpectedErrorMsg: "while loading tenant from context",
		},
		{
			Name:    "Returns error when can't get the formation assignment",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "Returns error when can't update the formation assignment",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa.Clone(), nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyStateAndNoConfig).Return(testErr).Once()
				return repo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "Returns not found error when can't update the formation assignment due to conflict",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa.Clone(), nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyStateAndNoConfig).Return(unauthorizedError).Once()
				return repo
			},
			ExpectedErrorMsg: notFoundError.Error(),
		},
		{
			Name:    "Returns error when can't enforce post constraints",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa.Clone(), nil).Once()
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyStateAndNoConfig).Return(nil).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PostNotificationStatusReturned, postJoinPointDetails, formation.FormationTemplateID).Return(testErr).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, faWithReadyStateAndNoConfig, model.UnassignFormation, fa.State, strconv.Quote(string(TestConfigValueRawJSON))).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "Returns error when delete fails",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa.Clone(), nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyStateAndNoConfig).Return(nil).Once()
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(testErr).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, faWithReadyStateAndNoConfig, model.UnassignFormation, fa.State, strconv.Quote(string(TestConfigValueRawJSON))).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "Returns not found error when delete fails due to unauthorized",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa.Clone(), nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyStateAndNoConfig).Return(nil).Once()
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(unauthorizedError).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, faWithReadyStateAndNoConfig, model.UnassignFormation, fa.State, strconv.Quote(string(TestConfigValueRawJSON))).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: notFoundError.Error(),
		},
		{
			Name:    "Returns error when can't enforce pre constraints",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa.Clone(), nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyStateAndNoConfig).Return(nil).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(testErr).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, faWithReadyStateAndNoConfig, model.UnassignFormation, fa.State, strconv.Quote(string(TestConfigValueRawJSON))).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "Returns error when can't prepare details",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa.Clone(), nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyStateAndNoConfig).Return(nil).Once()
				return repo
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, faWithReadyStateAndNoConfig, model.UnassignFormation, fa.State, strconv.Quote(string(TestConfigValueRawJSON))).Return(nil, testErr).Once()
				return notificationSvc
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
			constraintEngine := &automock.ConstraintEngine{}
			if testCase.ConstraintEngine != nil {
				constraintEngine = testCase.ConstraintEngine()
			}
			notificationSvc := &automock.FaNotificationService{}
			if testCase.NotificationSvc != nil {
				notificationSvc = testCase.NotificationSvc()
			}

			svc := formationassignment.NewFormationAssignmentStatusService(faRepo, constraintEngine, notificationSvc)

			// WHEN
			err := svc.DeleteWithConstraints(testCase.Context, testCase.InputID)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, faRepo, constraintEngine, notificationSvc)
		})
	}
}
