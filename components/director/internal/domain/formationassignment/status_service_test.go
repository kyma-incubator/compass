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

func TestUpdater_Update(t *testing.T) {
	preJoinPointDetails := fixNotificationStatusReturnedDetails(webhookFa, reverseWebhookFa, formationconstraint.PreNotificationStatusReturned)
	postJoinPointDetails := fixNotificationStatusReturnedDetails(webhookFa, reverseWebhookFa, formationconstraint.PostNotificationStatusReturned)

	// GIVEN
	testCases := []struct {
		Name                    string
		Context                 context.Context
		FormationAssignment     *model.FormationAssignment
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		FormationRepo           func() *automock.FormationRepository
		FormationTemplateRepo   func() *automock.FormationTemplateRepository
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
				repo.On("GetReverseBySourceAndTarget", ctxWithTenant, TestTenantID, formation.ID, fa.Source, fa.Target).Return(reverseFa, nil).Once()
				return repo
			},
			FormationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", ctxWithTenant, fa.FormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepo: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctxWithTenant, formation.FormationTemplateID).Return(formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PostNotificationStatusReturned, postJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
			},
			ExpectedErrorMsg: "",
		},
		{
			Name:                "Error when loading tenant from context",
			Context:             emptyCtx,
			FormationAssignment: fa,
			ExpectedErrorMsg:    "while loading tenant from context: cannot read tenant from context",
		},
		{
			Name:                "Error when can't get formation while preparing details",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", ctxWithTenant, fa.FormationID, TestTenantID).Return(nil, testErr).Once()
				return formationRepo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                "Error when can't get formation template while preparing details",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", ctxWithTenant, fa.FormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepo: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctxWithTenant, formation.FormationTemplateID).Return(nil, testErr).Once()
				return formationTemplateRepo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                "Error when can't get reverse formation assignment while preparing details",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetReverseBySourceAndTarget", ctxWithTenant, TestTenantID, formation.ID, fa.Source, fa.Target).Return(nil, testErr).Once()
				return repo
			},
			FormationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", ctxWithTenant, fa.FormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepo: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctxWithTenant, formation.FormationTemplateID).Return(formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                "Error when enforcing PRE constraints",
			Context:             ctxWithTenant,
			FormationAssignment: fa,
			FormationAssignmentRepo: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("GetReverseBySourceAndTarget", ctxWithTenant, TestTenantID, formation.ID, fa.Source, fa.Target).Return(reverseFa, nil).Once()
				return repo
			},
			FormationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", ctxWithTenant, fa.FormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepo: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctxWithTenant, formation.FormationTemplateID).Return(formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(testErr).Once()
				return constraintEngine
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
				repo.On("GetReverseBySourceAndTarget", ctxWithTenant, TestTenantID, formation.ID, fa.Source, fa.Target).Return(reverseFa, nil).Once()
				return repo
			},
			FormationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", ctxWithTenant, fa.FormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepo: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctxWithTenant, formation.FormationTemplateID).Return(formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
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
				repo.On("GetReverseBySourceAndTarget", ctxWithTenant, TestTenantID, formation.ID, fa.Source, fa.Target).Return(reverseFa, nil).Once()
				return repo
			},
			FormationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", ctxWithTenant, fa.FormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepo: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctxWithTenant, formation.FormationTemplateID).Return(formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
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
				repo.On("GetReverseBySourceAndTarget", ctxWithTenant, TestTenantID, formation.ID, fa.Source, fa.Target).Return(reverseFa, nil).Once()
				return repo
			},
			FormationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", ctxWithTenant, fa.FormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepo: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctxWithTenant, formation.FormationTemplateID).Return(formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
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
				repo.On("GetReverseBySourceAndTarget", ctxWithTenant, TestTenantID, formation.ID, fa.Source, fa.Target).Return(reverseFa, nil).Once()
				return repo
			},
			FormationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", ctxWithTenant, fa.FormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepo: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctxWithTenant, formation.FormationTemplateID).Return(formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PostNotificationStatusReturned, postJoinPointDetails, formation.FormationTemplateID).Return(testErr).Once()
				return constraintEngine
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
			formationRepo := &automock.FormationRepository{}
			if testCase.FormationRepo != nil {
				formationRepo = testCase.FormationRepo()
			}
			formationTemplateRepo := &automock.FormationTemplateRepository{}
			if testCase.FormationTemplateRepo != nil {
				formationTemplateRepo = testCase.FormationTemplateRepo()
			}

			svc := formationassignment.NewFormationAssignmentStatusService(faRepo, constraintEngine, formationRepo, formationTemplateRepo)

			// WHEN
			err := svc.UpdateWithConstraints(testCase.Context, testCase.FormationAssignment, assignOperation)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, faRepo, constraintEngine, formationRepo, formationTemplateRepo)
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

	preJoinPointDetails := fixNotificationStatusReturnedDetails(convertFormationAssignmentFromModel(faErrorState), convertFormationAssignmentFromModel(reverseFaErrorState), formationconstraint.PreNotificationStatusReturned)
	postJoinPointDetails := fixNotificationStatusReturnedDetails(convertFormationAssignmentFromModel(faErrorState), convertFormationAssignmentFromModel(reverseFaErrorState), formationconstraint.PostNotificationStatusReturned)

	testCases := []struct {
		Name                    string
		Context                 context.Context
		FormationAssignment     *model.FormationAssignment
		FormationAssignmentRepo func() *automock.FormationAssignmentRepository
		FormationRepo           func() *automock.FormationRepository
		FormationTemplateRepo   func() *automock.FormationTemplateRepository
		ConstraintEngine        func() *automock.ConstraintEngine
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
				repo.On("GetReverseBySourceAndTarget", ctxWithTenant, TestTenantID, formation.ID, fa.Source, fa.Target).Return(reverseFaErrorState, nil).Once()
				return repo
			},
			FormationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", ctxWithTenant, fa.FormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepo: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctxWithTenant, formation.FormationTemplateID).Return(formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PostNotificationStatusReturned, postJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
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
				repo.On("GetReverseBySourceAndTarget", ctxWithTenant, TestTenantID, formation.ID, fa.Source, fa.Target).Return(reverseFaErrorState, nil).Once()
				return repo
			},
			FormationRepo: func() *automock.FormationRepository {
				formationRepo := &automock.FormationRepository{}
				formationRepo.On("Get", ctxWithTenant, fa.FormationID, TestTenantID).Return(formation, nil).Once()
				return formationRepo
			},
			FormationTemplateRepo: func() *automock.FormationTemplateRepository {
				formationTemplateRepo := &automock.FormationTemplateRepository{}
				formationTemplateRepo.On("Get", ctxWithTenant, formation.FormationTemplateID).Return(formationTemplate, nil).Once()
				return formationTemplateRepo
			},
			ConstraintEngine: func() *automock.ConstraintEngine {
				constraintEngine := &automock.ConstraintEngine{}
				constraintEngine.On("EnforceConstraints", ctxWithTenant, formationconstraint.PreNotificationStatusReturned, preJoinPointDetails, formation.FormationTemplateID).Return(nil).Once()
				return constraintEngine
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
			formationTemplateRepo := &automock.FormationTemplateRepository{}
			if testCase.FormationTemplateRepo != nil {
				formationTemplateRepo = testCase.FormationTemplateRepo()
			}

			svc := formationassignment.NewFormationAssignmentStatusService(faRepo, constraintEngine, formationRepo, formationTemplateRepo)

			// WHEN
			err := svc.SetAssignmentToErrorStateWithConstraints(testCase.Context, testCase.FormationAssignment, errorMsg, formationassignment.TechnicalError, model.DeleteErrorAssignmentState, assignOperation)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				require.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, faRepo, constraintEngine, formationRepo, formationTemplateRepo)
		})
	}
}
