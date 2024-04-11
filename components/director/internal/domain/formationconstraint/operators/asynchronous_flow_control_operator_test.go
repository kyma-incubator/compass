package operators_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/statusreport"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConstraintOperators_AsynchronousFlowControlOperator(t *testing.T) {
	certConsumer := consumer.Consumer{
		Type: consumer.ExternalCertificate,
	}
	instanceCreatorConsumer := consumer.Consumer{
		Type: consumer.InstanceCreator,
	}
	ctxWithCertConsumer := consumer.SaveToContext(context.TODO(), certConsumer)
	ctxWithInstanceCreatorConsumer := consumer.SaveToContext(context.TODO(), instanceCreatorConsumer)

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
		Context                                context.Context
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
			Context:              ctxWithCertConsumer,
			Assignment:           fixFormationAssignmentWithState(model.ReadyAssignmentState),
			ReverseAssignment:    fixFormationAssignmentWithState(model.ReadyAssignmentState),
			ExpectShouldRedirect: true,
			ExpectedResult:       true,
		},
		// Unassign during SendNotification
		{
			Name:                 "Success when sending notification and state is INSTANCE_CREATOR_DELETING state",
			Input:                inputForSendNotificationUnassign,
			Context:              ctxWithCertConsumer,
			Assignment:           fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState),
			ReverseAssignment:    fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectShouldRedirect: true,
			ExpectedResult:       true,
		},
		{
			Name:                 "Success when sending notification and state is INSTANCE_CREATOR_DELETE_ERROR state",
			Input:                inputForSendNotificationUnassign,
			Context:              ctxWithCertConsumer,
			Assignment:           fixFormationAssignmentWithState(model.InstanceCreatorDeleteErrorAssignmentState),
			ReverseAssignment:    fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectShouldRedirect: true,
			ExpectedResult:       true,
		},
		{
			Name:                 "Success when sending notification and state is READY state",
			Input:                inputForSendNotificationUnassign,
			Context:              ctxWithCertConsumer,
			Assignment:           fixFormationAssignmentWithState(model.ReadyAssignmentState),
			ReverseAssignment:    fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectShouldRedirect: false,
			ExpectedResult:       true,
		},
		{
			Name:                             "Error when retrieving formation assignment pointer fails during send notification",
			Input:                            inputForSendNotificationUnassign,
			Context:                          ctxWithCertConsumer,
			ExpectedFormationAssignmentState: string(model.InitialAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), configWithDifferentStructure),
			ExpectedErrorMsg:                 "The join point details' assignment memory address cannot be 0",
		},
		// Assign during PreStatusReturned
		{
			Name:                             "Error when formation assignment config is invalid",
			Input:                            inputForNotificationStatusReturnedAssign,
			Context:                          ctxWithCertConsumer,
			Assignment:                       fixFormationAssignmentWithState(model.InitialAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.InitialAssignmentState),
			ExpectedFormationAssignmentState: string(model.InitialAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), invalidFAConfig),
			ExpectedStatusReportState:        string(model.ConfigPendingAssignmentState),
			ExpectedErrorMsg:                 "while unmarshalling tenant mapping response configuration for assignment with ID:",
		},
		{
			Name:                             "Success when transitioning to READY state with no configuration",
			Input:                            inputForNotificationStatusReturnedAssign,
			Context:                          ctxWithCertConsumer,
			Assignment:                       fixFormationAssignmentWithState(model.InitialAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.InitialAssignmentState),
			ExpectedFormationAssignmentState: string(model.InitialAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), emptyConfig),
			ExpectedStatusReportState:        string(model.ReadyAssignmentState),
			ExpectedResult:                   true,
		},
		{
			Name:                             "Success when transitioning to READY state with inbound credentials",
			Input:                            inputForNotificationStatusReturnedAssign,
			Context:                          ctxWithCertConsumer,
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
			Context:                          ctxWithCertConsumer,
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
			Context:                          ctxWithCertConsumer,
			Assignment:                       fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.InstanceCreatorDeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.InstanceCreatorDeletingAssignmentState),
			FormationAssignmentRepository: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctxWithCertConsumer, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState)).Return(nil).Once()
				return repo
			},
			FormationAssignmentService: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("CleanupFormationAssignment", ctxWithCertConsumer, testAssignmentPair).Return(false, nil)
				return svc
			},
			FormationAssignmentNotificationService: func() *automock.FormationAssignmentNotificationService {
				notificationSvc := &automock.FormationAssignmentNotificationService{}
				notificationSvc.On("GenerateFormationAssignmentPair", ctxWithCertConsumer, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState), fixFormationAssignmentWithState(model.DeletingAssignmentState), model.UnassignFormation).Return(testAssignmentPair, nil).Once()
				return notificationSvc
			},
			ExpectedResult: true,
		},
		{
			Name:                             "Success when transitioning from DELETING to READY but consumer type is instance creator",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Context:                          ctxWithInstanceCreatorConsumer,
			Assignment:                       fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.ReadyAssignmentState),
			ExpectedResult:                   true,
		},
		{
			Name:                             "Error when transitioning from DELETING to READY but there is no consumer in context",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Context:                          ctx,
			Assignment:                       fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.ReadyAssignmentState),
			ExpectedResult:                   false,
			ExpectedErrorMsg:                 "while fetching consumer info from context",
		},
		{
			Name:                             "Error when cleanup formation assignment fails",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Context:                          ctxWithCertConsumer,
			Assignment:                       fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.InstanceCreatorDeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.InstanceCreatorDeletingAssignmentState),
			FormationAssignmentRepository: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctxWithCertConsumer, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState)).Return(nil).Once()
				return repo
			},
			FormationAssignmentService: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("CleanupFormationAssignment", ctxWithCertConsumer, testAssignmentPair).Return(false, testErr)
				return svc
			},
			FormationAssignmentNotificationService: func() *automock.FormationAssignmentNotificationService {
				notificationSvc := &automock.FormationAssignmentNotificationService{}
				notificationSvc.On("GenerateFormationAssignmentPair", ctxWithCertConsumer, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState), fixFormationAssignmentWithState(model.DeletingAssignmentState), model.UnassignFormation).Return(testAssignmentPair, nil).Once()
				return notificationSvc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                             "Error during generating formation assignment pair",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Context:                          ctxWithCertConsumer,
			Assignment:                       fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.InstanceCreatorDeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.InstanceCreatorDeletingAssignmentState),
			FormationAssignmentRepository: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctxWithCertConsumer, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState)).Return(nil).Once()
				return repo
			},
			FormationAssignmentNotificationService: func() *automock.FormationAssignmentNotificationService {
				notificationSvc := &automock.FormationAssignmentNotificationService{}
				notificationSvc.On("GenerateFormationAssignmentPair", ctxWithCertConsumer, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState), fixFormationAssignmentWithState(model.DeletingAssignmentState), model.UnassignFormation).Return(nil, testErr).Once()
				return notificationSvc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                             "Error during formation assignment update to INSTANCE_CREATOR_DELETING state",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Context:                          ctxWithCertConsumer,
			Assignment:                       fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.InstanceCreatorDeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.InstanceCreatorDeletingAssignmentState),
			FormationAssignmentRepository: func() *automock.FormationAssignmentRepository {
				repo := &automock.FormationAssignmentRepository{}
				repo.On("Update", ctxWithCertConsumer, fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState)).Return(testErr).Once()
				return repo
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                             "Success when transitioning from DELETING to DELETE_ERROR",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Context:                          ctxWithCertConsumer,
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
			Context:                          ctxWithCertConsumer,
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
			Context:                          ctxWithCertConsumer,
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
			Context:                   ctxWithCertConsumer,
			Assignment:                fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState),
			StatusReport:              fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState: string(model.ReadyAssignmentState),
			ExpectedResult:            true,
		},
		{
			Name:              "Error when retrieving status report pointer fails",
			Input:             inputForNotificationStatusReturnedUnassign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithState(model.InstanceCreatorDeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithState(model.DeletingAssignmentState),
			ExpectedErrorMsg:  "The join point details' notification status report memory address cannot be 0",
		},
		{
			Name:             "Error when retrieving formation assignment pointer fails",
			Input:            inputForNotificationStatusReturnedUnassign,
			Context:          ctxWithCertConsumer,
			ExpectedErrorMsg: "The join point details' assignment memory address cannot be 0",
		},
		{
			Name:           "Error when retrieving formation assignment pointer fails",
			Input:          inputForPreAssign,
			Context:        ctxWithCertConsumer,
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

			result, err := engine.AsynchronousFlowControlOperator(testCase.Context, inputClone)

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
		result, err := engine.AsynchronousFlowControlOperator(ctxWithCertConsumer, input)

		// THEN
		assert.Equal(t, false, result)
		assert.Equal(t, "Incompatible input for operator: AsynchronousFlowControl", err.Error())
	})
}
