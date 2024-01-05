package operators_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/statusreport"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConstraintOperators_AsynchronousFlowControlOperator(t *testing.T) {
	inputForNotificationStatusReturnedAssign := fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, fixWebhook(), preNotificationStatusReturnedLocation)
	inputForNotificationStatusReturnedUnassign := fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddress(model.UnassignFormation, fixWebhook(), preNotificationStatusReturnedLocation)
	inputForSendNotificationAssign := fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddressShouldRedirect(true, model.AssignFormation, fixWebhook(), preSendNotificationLocation)
	inputForSendNotificationUnassign := fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddress(model.UnassignFormation, fixWebhook(), preSendNotificationLocation)
	inputForPreAssign := fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddress(model.UnassignFormation, fixWebhook(), preAssignFormationLocation)

	testAssignmentPair := &formationassignment.AssignmentMappingPairWithOperation{
		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
			AssignmentReqMapping:        nil,
			ReverseAssignmentReqMapping: nil,
		},
		Operation: model.UnassignFormation,
	}

	testCases := []struct {
		Name                                   string
		Input                                  *formationconstraintpkg.AsynchronousFlowControlOperatorInput
		Assignment                             *model.FormationAssignment
		ReverseAssignment                      *model.FormationAssignment
		StatusReport                           *statusreport.NotificationStatusReport
		FormationAssignmentRepository          func() *automock.FormationAssignmentRepository
		FormationAssignmentService             func() *automock.FormationAssignmentService
		FormationAssignmentNotificationService func() *automock.FormationAssignmentNotificationService
		ExpectedResult                         bool
		ExpectedFormationAssignmentState       string
		ExpectedStatusReportState              string
		ExpectShouldRedirect                   bool
		ExpectedErrorMsg                       string
	}{
		// Assign during SendNotification
		{
			Name:                 "Success when sending notification and state is READY",
			Input:                inputForSendNotificationAssign,
			Assignment:           fixFormationAssignmentWithState(model.ReadyAssignmentState),
			ReverseAssignment:    fixFormationAssignmentWithState(model.ReadyAssignmentState),
			ExpectShouldRedirect: true,
			ExpectedResult:       true,
		},
		// Unassign during SendNotification
		{
			Name:                 "Success when sending notification and state is INSTANCE_CREATOR_DELETING state",
			Input:                inputForSendNotificationUnassign,
			Assignment:           fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState),
			ReverseAssignment:    fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectShouldRedirect: true,
			ExpectedResult:       true,
		},
		{
			Name:                 "Success when sending notification and state is INSTANCE_CREATOR_DELETE_ERROR state",
			Input:                inputForSendNotificationUnassign,
			Assignment:           fixFormationAssignmentWithState(model.InstanceCreatorDeleteErrorAssignmentState),
			ReverseAssignment:    fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectShouldRedirect: true,
			ExpectedResult:       true,
		},
		{
			Name:                 "Success when sending notification and state is READY state",
			Input:                inputForSendNotificationUnassign,
			Assignment:           fixFormationAssignmentWithState(model.ReadyAssignmentState),
			ReverseAssignment:    fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectShouldRedirect: false,
			ExpectedResult:       true,
		},
		{
			Name:                             "Error when retrieving formation assignment pointer fails during send notification",
			Input:                            inputForSendNotificationUnassign,
			ExpectedFormationAssignmentState: string(model.InitialAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), configWithDifferentStructure),
			ExpectedErrorMsg:                 "The join point details' assignment memory address cannot be 0",
		},
		// Assign during PreStatusReturned
		{
			Name:                             "Success when transitioning to READY state with inbound credentials",
			Input:                            inputForNotificationStatusReturnedAssign,
			Assignment:                       fixFormationAssignmentWithState(model.InitialAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.InitialAssignmentState),
			ExpectedFormationAssignmentState: string(model.InitialAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), invalidFAConfig),
			ExpectedStatusReportState:        string(model.ConfigPendingAssignmentState),
			ExpectedErrorMsg:                 "while unmarshalling tenant mapping response configuration for assignment with ID:",
		},
		{
			Name:                             "Success when transitioning to READY state with inbound credentials",
			Input:                            inputForNotificationStatusReturnedAssign,
			Assignment:                       fixFormationAssignmentWithState(model.InitialAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.InitialAssignmentState),
			ExpectedFormationAssignmentState: string(model.InitialAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), destsConfigValueRawJSON),
			ExpectedStatusReportState:        string(model.ConfigPendingAssignmentState),
			ExpectedResult:                   true,
		},
		{
			Name:                             "Success when transitioning to READY state without inbound credentials",
			Input:                            inputForNotificationStatusReturnedAssign,
			Assignment:                       fixFormationAssignmentWithState(model.InitialAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.InitialAssignmentState),
			ExpectedFormationAssignmentState: string(model.InitialAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), configWithDifferentStructure),
			ExpectedStatusReportState:        string(model.ReadyAssignmentState),
			ExpectedResult:                   true,
		},
		// Unassign during PreStatusReturned
		{
			Name:                             "Success when transitioning from DELETING to READY",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Assignment:                       fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.InstanceCreatorDeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.InstanceCreatorDeletingAssignmentState),
			FormationAssignmentRepository: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctx, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState)).Return(nil).Once()
				return repo
			},
			FormationAssignmentService: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("CleanupFormationAssignment", ctx, testAssignmentPair).Return(false, nil)
				return svc
			},
			FormationAssignmentNotificationService: func() *automock.FormationAssignmentNotificationService {
				notificationSvc := &automock.FormationAssignmentNotificationService{}
				notificationSvc.On("GenerateFormationAssignmentPair", ctx, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState), fixFormationAssignmentWithState(model.DeletingAssignmentState), model.UnassignFormation).Return(testAssignmentPair, nil).Once()
				return notificationSvc
			},
			ExpectedResult: true,
		},
		{
			Name:                             "Error when cleanup formation assignment fails",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Assignment:                       fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.InstanceCreatorDeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.InstanceCreatorDeletingAssignmentState),
			FormationAssignmentRepository: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctx, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState)).Return(nil).Once()
				return repo
			},
			FormationAssignmentService: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("CleanupFormationAssignment", ctx, testAssignmentPair).Return(false, testErr)
				return svc
			},
			FormationAssignmentNotificationService: func() *automock.FormationAssignmentNotificationService {
				notificationSvc := &automock.FormationAssignmentNotificationService{}
				notificationSvc.On("GenerateFormationAssignmentPair", ctx, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState), fixFormationAssignmentWithState(model.DeletingAssignmentState), model.UnassignFormation).Return(testAssignmentPair, nil).Once()
				return notificationSvc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                             "Error during generating formation assignment pair",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Assignment:                       fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.InstanceCreatorDeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.InstanceCreatorDeletingAssignmentState),
			FormationAssignmentRepository: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctx, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState)).Return(nil).Once()
				return repo
			},
			FormationAssignmentNotificationService: func() *automock.FormationAssignmentNotificationService {
				notificationSvc := &automock.FormationAssignmentNotificationService{}
				notificationSvc.On("GenerateFormationAssignmentPair", ctx, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState), fixFormationAssignmentWithState(model.DeletingAssignmentState), model.UnassignFormation).Return(nil, testErr).Once()
				return notificationSvc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                             "Error during formation assignment update to INSTANCE_CREATOR_DELETING state",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Assignment:                       fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.InstanceCreatorDeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.InstanceCreatorDeletingAssignmentState),
			FormationAssignmentRepository: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctx, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState)).Return(testErr).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                             "Success when transitioning from DELETING to DELETE_ERROR",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Assignment:                       fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.DeleteErrorAssignmentState),
			ExpectedStatusReportState:        string(model.DeleteErrorAssignmentState),
			ExpectedResult:                   true,
		},
		{
			Name:                             "Success when transitioning from INSTANCE_CREATOR_DELETING to DELETE_ERROR",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Assignment:                       fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.InstanceCreatorDeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.DeleteErrorAssignmentState),
			ExpectedStatusReportState:        string(model.InstanceCreatorDeleteErrorAssignmentState),
			ExpectedResult:                   true,
		},
		{
			Name:                             "Success when transitioning from INSTANCE_CREATOR_DELETING to READY",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Assignment:                       fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.InstanceCreatorDeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.ReadyAssignmentState),
			ExpectedResult:                   true,
		},
		{
			Name:                      "Success when transitioning from INSTANCE_CREATOR_DELETING to READY without reverse assignment",
			Input:                     inputForNotificationStatusReturnedUnassign,
			Assignment:                fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState),
			StatusReport:              fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState: string(model.ReadyAssignmentState),
			ExpectedResult:            true,
		},
		{
			Name:              "Error when retrieving status report pointer fails",
			Input:             inputForNotificationStatusReturnedUnassign,
			Assignment:        fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedErrorMsg:  "The join point details' notification status report memory address cannot be 0",
		},
		{
			Name:             "Error when retrieving formation assignment pointer fails",
			Input:            inputForNotificationStatusReturnedUnassign,
			ExpectedErrorMsg: "The join point details' assignment memory address cannot be 0",
		},
		{
			Name:           "Error when retrieving formation assignment pointer fails",
			Input:          inputForPreAssign,
			ExpectedResult: true,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			formationAssignmentRepo := &automock.FormationAssignmentRepository{}
			if testCase.FormationAssignmentRepository != nil {
				formationAssignmentRepo = testCase.FormationAssignmentRepository()
			}
			formationAssignmentService := &automock.FormationAssignmentService{}
			if testCase.FormationAssignmentService != nil {
				formationAssignmentService = testCase.FormationAssignmentService()
			}
			formationAssignmentNotificationService := &automock.FormationAssignmentNotificationService{}
			if testCase.FormationAssignmentNotificationService != nil {
				formationAssignmentNotificationService = testCase.FormationAssignmentNotificationService()
			}

			engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, formationAssignmentRepo, formationAssignmentService, formationAssignmentNotificationService, runtimeType, applicationType)

			inputClone := cloneAsynchronousFlowControlOperatorInput(testCase.Input)
			if testCase.Assignment != nil {
				setAssignmentToAsynchronousFlowControlInput(inputClone, testCase.Assignment)
			}
			if testCase.ReverseAssignment != nil {
				setReverseAssignmentToAsynchronousFlowControlInput(inputClone, testCase.ReverseAssignment)
			}
			if testCase.StatusReport != nil {
				setStatusReportToAsynchronousFlowControlInput(inputClone, testCase.StatusReport)
			}

			result, err := engine.AsynchronousFlowControlOperator(ctx, inputClone)

			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.Equal(t, testCase.ExpectedResult, result)
				if testCase.ExpectedFormationAssignmentState != "" {
					assert.Equal(t, testCase.ExpectedFormationAssignmentState, testCase.Assignment.State, result)
				}
				if testCase.ExpectedStatusReportState != "" {
					assert.Equal(t, testCase.ExpectedStatusReportState, testCase.StatusReport.State, result)
				}
				assert.Equal(t, testCase.ExpectShouldRedirect, inputClone.ShouldRedirect)
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, formationAssignmentRepo, formationAssignmentService, formationAssignmentNotificationService)
		})
	}

	t.Run("Error when incorrect input is provided", func(t *testing.T) {
		// GIVEN

		engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		input := "wrong input"
		result, err := engine.AsynchronousFlowControlOperator(ctx, input)

		// THEN
		assert.Equal(t, false, result)
		assert.Equal(t, "Incompatible input for operator: AsynchronousFlowControl", err.Error())
	})
}
