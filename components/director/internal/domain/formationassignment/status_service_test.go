package formationassignment_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestStatusService_UpdateWithConstraints(t *testing.T) {
	preJoinPointDetails := fixNotificationStatusReturnedDetails(fa, reverseFa, formationconstraint.PreNotificationStatusReturned)
	postJoinPointDetails := fixNotificationStatusReturnedDetails(fa, reverseFa, formationconstraint.PostNotificationStatusReturned)

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
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
		},
		{
			Name:                "Error when can't prepare details",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation).Return(nil, testErr).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                "Error when enforcing PRE constraints",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(testErr).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation).Return(preJoinPointDetails, nil).Once()
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
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(false, testErr).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: testErr.Error(),
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
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: apperrors.NewNotFoundError(resource.FormationAssignment, fa.ID).Error(),
		},
		{
			Name:                "Error when updating formation assignment",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                "Error when enforcing POST constraints",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.AssignFormation).Return(preJoinPointDetails, nil).Once()
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
	}
	assignmentError := formationassignment.AssignmentErrorWrapper{
		Error: formationassignment.AssignmentError{
			Message:   errorMsg,
			ErrorCode: formationassignment.TechnicalError,
		},
	}
	marshaled, err := json.Marshal(assignmentError)
	require.NoError(t, err)
	faErrorState.Value = marshaled

	reverseFaErrorState := fixReverseFormationAssignment(faErrorState)

	preJoinPointDetails := fixNotificationStatusReturnedDetails(faErrorState, reverseFaErrorState, formationconstraint.PreNotificationStatusReturned)
	postJoinPointDetails := fixNotificationStatusReturnedDetails(faErrorState, reverseFaErrorState, formationconstraint.PostNotificationStatusReturned)

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
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, faErrorState, model.AssignFormation).Return(preJoinPointDetails, nil).Once()
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
				repo.On("Exists", ctxWithTenant, TestID, TestTenantID).Return(true, nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, faErrorState, model.AssignFormation).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			FormationOperation: assignOperation,
			ExpectedErrorMsg:   testErr.Error(),
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
	preJoinPointDetails := fixNotificationStatusReturnedDetails(fa, reverseFa, formationconstraint.PreNotificationStatusReturned)
	postJoinPointDetails := fixNotificationStatusReturnedDetails(fa, reverseFa, formationconstraint.PostNotificationStatusReturned)

	faWithReadyState := fixFormationAssignmentModelWithFormationID(TestFormationID)
	faWithReadyState.State = string(model.ReadyAssignmentState)

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
			Name:    "Success",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa, nil).Once()
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyState).Return(nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.UnassignFormation).Return(preJoinPointDetails, nil).Once()
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
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa, nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyState).Return(testErr).Once()
				return repo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "Returns error when can't enforce post constraints",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa, nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyState).Return(nil).Once()
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.UnassignFormation).Return(preJoinPointDetails, nil).Once()
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
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa, nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyState).Return(nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.UnassignFormation).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:    "Returns error when can't enforce pre constraints",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa, nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyState).Return(nil).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(testErr).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.UnassignFormation).Return(preJoinPointDetails, nil).Once()
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
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(fa, nil).Once()
				repo.On("Update", ctxWithTenant, faWithReadyState).Return(nil).Once()
				return repo
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, fa, model.UnassignFormation).Return(nil, testErr).Once()
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
