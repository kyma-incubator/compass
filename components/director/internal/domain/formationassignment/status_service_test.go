package formationassignment_test

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/statusreport"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	initialStateAssignment = fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, TestSourceType, TestTargetType, string(model.InitialAssignmentState), nil, nil)
	assignmentConfig       = json.RawMessage(`{"foo": "bar"}`)
	assignmentConfigOld    = json.RawMessage(`{"old": "config"}`)
	assignmentError        = json.RawMessage(`{"error":{"message":"error from report","errorCode":2}}`)

	assignmentWithStateAndConfig          = fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, TestSourceType, TestTargetType, string(model.ReadyAssignmentState), assignmentConfig, nil)
	assignmentWithoutConfig               = fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, TestSourceType, TestTargetType, string(model.ReadyAssignmentState), nil, nil)
	assignmentWithStateAndOldConfig       = fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, TestSourceType, TestTargetType, string(model.ConfigPendingAssignmentState), assignmentConfigOld, nil)
	assignmentWithConfigAndError          = fixFormationAssignmentModelWithParameters(TestID, TestFormationID, TestTenantID, TestSource, TestTarget, TestSourceType, TestTargetType, string(model.DeleteErrorAssignmentState), assignmentConfig, assignmentError)
	notificationStatusReport              = fixNotificationStatusReport()
	statusReportWithConfig                = fixNotificationStatusReportWithStateAndConfig(assignmentConfig, readyState)
	statusReportWithoutConfigAndError     = fixNotificationStatusReportWithStateAndConfig(nil, readyState)
	statusReportWithConfigConsideredEmpty = fixNotificationStatusReportWithStateAndConfig(json.RawMessage("{}"), readyState)
	statusReportWithError                 = fixNotificationStatusReportWithStateAndError(deleteErrorState, "error from report")
)

func TestStatusService_UpdateWithConstraints(t *testing.T) {
	preJoinPointDetails := fixNotificationStatusReturnedDetails(model.ApplicationResourceType, appSubtype, initialStateAssignment, reverseFa, formationconstraint.PreNotificationStatusReturned, TestTenantID, notificationStatusReport)
	postJoinPointDetails := fixNotificationStatusReturnedDetails(model.ApplicationResourceType, appSubtype, initialStateAssignment, reverseFa, formationconstraint.PostNotificationStatusReturned, TestTenantID, notificationStatusReport)

	// GIVEN
	testCases := []struct {
		Name                     string
		Context                  context.Context
		FormationAssignment      *model.FormationAssignment
		FormationAssignmentRepo  func() *automock.FormationAssignmentRepository
		NotificationSvc          func() *automock.FaNotificationService
		ConstraintEngine         func() *automock.ConstraintEngine
		NotificationStatusReport *statusreport.NotificationStatusReport
		ExpectedErrorMsg         string
	}{
		{
			Name:                "Success with config in notification status report",
			Context:             ctxWithTenant,
			FormationAssignment: initialStateAssignment,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctxWithTenant, assignmentWithStateAndConfig).Return(nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, initialStateAssignment, model.AssignFormation, statusReportWithConfig).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			NotificationStatusReport: statusReportWithConfig,
		},
		{
			Name:                "Success with config in notification status report - replace previous config if config in report",
			Context:             ctxWithTenant,
			FormationAssignment: assignmentWithStateAndOldConfig,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctxWithTenant, assignmentWithStateAndConfig).Return(nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndOldConfig, model.AssignFormation, statusReportWithConfig).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			NotificationStatusReport: statusReportWithConfig,
		},
		{
			Name:                "Success with config in notification status report - clear previous config if no config in report",
			Context:             ctxWithTenant,
			FormationAssignment: assignmentWithStateAndOldConfig,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctxWithTenant, assignmentWithoutConfig).Return(nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndOldConfig, model.AssignFormation, statusReportWithoutConfigAndError).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			NotificationStatusReport: statusReportWithoutConfigAndError,
		},
		{
			Name:                "Success with config in notification status report - do not set config if config from report is considered empty - \\\"\\\" or {}",
			Context:             ctxWithTenant,
			FormationAssignment: assignmentWithStateAndOldConfig,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctxWithTenant, assignmentWithoutConfig).Return(nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndOldConfig, model.AssignFormation, statusReportWithConfigConsideredEmpty).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			NotificationStatusReport: statusReportWithConfigConsideredEmpty,
		},
		{
			Name:                "Success with error in notification status report - do not clear config",
			Context:             ctxWithTenant,
			FormationAssignment: assignmentWithStateAndConfig,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctxWithTenant, assignmentWithConfigAndError).Return(nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndConfig, model.AssignFormation, statusReportWithError).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			NotificationStatusReport: statusReportWithError,
		},
		{
			Name:                "Error while enforcing constraints POST",
			Context:             ctxWithTenant,
			FormationAssignment: assignmentWithStateAndConfig,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctxWithTenant, assignmentWithConfigAndError).Return(nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndConfig, model.AssignFormation, statusReportWithError).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			NotificationStatusReport: statusReportWithError,
			ExpectedErrorMsg:         fmt.Sprintf("while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PostOperation),
		},
		{
			Name:                "Error while updating formation assignment",
			Context:             ctxWithTenant,
			FormationAssignment: assignmentWithStateAndConfig,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctxWithTenant, assignmentWithConfigAndError).Return(testErr).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndConfig, model.AssignFormation, statusReportWithError).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			NotificationStatusReport: statusReportWithError,
			ExpectedErrorMsg:         "while updating formation assignment with ID:",
		},
		{
			Name:                "Error while updating formation assignment - unauthorized",
			Context:             ctxWithTenant,
			FormationAssignment: assignmentWithStateAndConfig,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctxWithTenant, assignmentWithConfigAndError).Return(apperrors.NewUnauthorizedError(testErr.Error())).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndConfig, model.AssignFormation, statusReportWithError).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			NotificationStatusReport: statusReportWithError,
			ExpectedErrorMsg:         notFoundError.Error(),
		},
		{
			Name:                "Error while enforcing constraints PRE",
			Context:             ctxWithTenant,
			FormationAssignment: assignmentWithStateAndConfig,
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(testErr).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndConfig, model.AssignFormation, statusReportWithError).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			NotificationStatusReport: statusReportWithError,
			ExpectedErrorMsg:         fmt.Sprintf("while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PreOperation),
		},
		{
			Name:                "Error while preparing details",
			Context:             ctxWithTenant,
			FormationAssignment: assignmentWithStateAndConfig,
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndConfig, model.AssignFormation, statusReportWithError).Return(nil, testErr).Once()
				return notificationSvc
			},
			NotificationStatusReport: statusReportWithError,
			ExpectedErrorMsg:         "while preparing details for NotificationStatusReturned",
		},
		{
			Name:                     "Error while loading tenant from context",
			Context:                  emptyCtx,
			FormationAssignment:      assignmentWithStateAndConfig,
			NotificationStatusReport: statusReportWithError,
			ExpectedErrorMsg:         "while loading tenant from context",
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
			err := svc.UpdateWithConstraints(testCase.Context, testCase.NotificationStatusReport, testCase.FormationAssignment, assignOperation)

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

func TestStatusService_DeleteWithConstraints(t *testing.T) {
	preJoinPointDetails := fixNotificationStatusReturnedDetails(model.ApplicationResourceType, appSubtype, fa, reverseFa, formationconstraint.PreNotificationStatusReturned, TestTenantID, notificationStatusReport)
	postJoinPointDetails := fixNotificationStatusReturnedDetails(model.ApplicationResourceType, appSubtype, fa, reverseFa, formationconstraint.PostNotificationStatusReturned, TestTenantID, notificationStatusReport)

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
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(assignmentWithStateAndConfig, nil).Once()
				repo.On("Delete", ctxWithTenant, TestID, TestTenantID).Return(nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndConfig, model.UnassignFormation, notificationStatusReport).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
		},
		{
			Name:    "Returns error when can't enforce post constraints",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(assignmentWithStateAndConfig, nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndConfig, model.UnassignFormation, notificationStatusReport).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: fmt.Sprintf("while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PostOperation),
		},
		{
			Name:    "Returns error when delete fails",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(assignmentWithStateAndConfig, nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndConfig, model.UnassignFormation, notificationStatusReport).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: "while deleting formation assignment",
		},
		{
			Name:    "Returns not found error when delete fails due to unauthorized",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(assignmentWithStateAndConfig, nil).Once()
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
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndConfig, model.UnassignFormation, notificationStatusReport).Return(preJoinPointDetails, nil).Once()
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
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(assignmentWithStateAndConfig, nil).Once()
				return repo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(testErr).Once()
				return constraintEngine
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndConfig, model.UnassignFormation, notificationStatusReport).Return(preJoinPointDetails, nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: fmt.Sprintf("while enforcing constraints for target operation %q and constraint type %q", model.NotificationStatusReturned, model.PreOperation),
		},
		{
			Name:    "Returns error when can't prepare details",
			Context: ctxWithTenant,
			InputID: TestID,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Get", ctxWithTenant, TestID, TestTenantID).Return(assignmentWithStateAndConfig, nil).Once()
				return repo
			},
			NotificationSvc: func() *automock.FaNotificationService {
				notificationSvc := &automock.FaNotificationService{}
				notificationSvc.On("PrepareDetailsForNotificationStatusReturned", ctxWithTenant, TestTenantID, assignmentWithStateAndConfig, model.UnassignFormation, notificationStatusReport).Return(nil, testErr).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: "while preparing details for NotificationStatusReturned",
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
			ExpectedErrorMsg: "while getting formation assignment with id",
		},
		{
			Name:             "Returns error when there is no tenant in the context",
			Context:          emptyCtx,
			InputID:          TestID,
			ExpectedErrorMsg: "while loading tenant from context",
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
			err := svc.DeleteWithConstraints(testCase.Context, testCase.InputID, notificationStatusReport)

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
