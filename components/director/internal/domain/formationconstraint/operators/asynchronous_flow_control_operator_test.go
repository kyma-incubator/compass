package operators_test

import (
	"context"
	"testing"

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

	inputForNotificationStatusReturnedAssign := fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, fixWebhookWithAsyncCallbackMode(), preNotificationStatusReturnedLocation)
	inputForNotificationStatusReturnedUnassign := fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddress(model.UnassignFormation, fixWebhookWithAsyncCallbackMode(), preNotificationStatusReturnedLocation)
	inputForSendNotificationAssign := fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddressShouldRedirect(true, model.AssignFormation, fixWebhookWithAsyncCallbackMode(), preSendNotificationLocation, false, false)
	inputForSendNotificationUnassign := fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddress(model.UnassignFormation, fixWebhookWithAsyncCallbackMode(), preSendNotificationLocation)

	assignmentOperationWithUnassignType := fixAssignmentOperationModelWithTypeAndTrigger(model.Unassign, model.UnassignObject)
	assignmentOperationWithInstanceCreatorUnassignType := fixAssignmentOperationModelWithTypeAndTrigger(model.InstanceCreatorUnassign, model.UnassignObject)

	testCases := []struct {
		Name                                   string
		Context                                context.Context
		Input                                  *formationconstraintpkg.AsynchronousFlowControlOperatorInput
		Assignment                             *model.FormationAssignment
		ReverseAssignment                      *model.FormationAssignment
		StatusReport                           *statusreport.NotificationStatusReport
		FormationAssignmentRepository          func() *automock.FormationAssignmentRepository
		LabelRepository                        func() *automock.LabelRepository
		FormationAssignmentService             func() *automock.FormationAssignmentService
		FormationAssignmentNotificationService func() *automock.FormationAssignmentNotificationService
		AssignmentOperationService             func() *automock.AssignmentOperationService
		ExpectedResult                         bool
		ExpectedFormationAssignmentState       string
		ExpectedStatusReportState              string
		ExpectShouldRedirect                   bool
		ExpectedErrorMsg                       string
	}{
		// Assign during SendNotification
		{
			Name:              "Success when sending notification and state is READY",
			Input:             inputForSendNotificationAssign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.ReadyAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.ReadyAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			ExpectShouldRedirect: true,
			ExpectedResult:       true,
		},
		// Unassign during SendNotification
		{
			Name:              "Success when sending notification and latest assignment operation is INSTANCE_CREATOR_UNASSIGN",
			Input:             inputForSendNotificationUnassign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			AssignmentOperationService: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", ctxWithCertConsumer, formationAssignmentID, formationID).Return(assignmentOperationWithInstanceCreatorUnassignType, nil).Once()
				return svc
			},
			ExpectShouldRedirect: true,
			ExpectedResult:       true,
		},
		{
			Name:              "Success when sending notification and state is READY state",
			Input:             inputForSendNotificationUnassign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.ReadyAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			AssignmentOperationService: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", ctxWithCertConsumer, formationAssignmentID, formationID).Return(assignmentOperationWithUnassignType, nil).Once()
				return svc
			},
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
		{
			Name:              "Error when sending notification and getting latest assignment operation fails",
			Input:             inputForSendNotificationUnassign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			AssignmentOperationService: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", ctxWithCertConsumer, formationAssignmentID, formationID).Return(nil, testErr).Once()
				return svc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		// Assign during PreStatusReturned
		{
			Name:              "Error when formation assignment config is invalid",
			Input:             inputForNotificationStatusReturnedAssign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.InitialAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.InitialAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			ExpectedFormationAssignmentState: string(model.InitialAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), invalidFAConfig),
			ExpectedStatusReportState:        string(model.ConfigPendingAssignmentState),
			ExpectedErrorMsg:                 "while unmarshalling tenant mapping response configuration for assignment with ID:",
		},
		{
			Name:              "Success when transitioning to READY state with no configuration",
			Input:             inputForNotificationStatusReturnedAssign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.InitialAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.InitialAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			ExpectedFormationAssignmentState: string(model.InitialAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), emptyConfig),
			ExpectedStatusReportState:        string(model.ReadyAssignmentState),
			ExpectedResult:                   true,
		},
		{
			Name:              "Success when transitioning to READY state with inbound credentials",
			Input:             inputForNotificationStatusReturnedAssign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.InitialAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.InitialAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			ExpectedFormationAssignmentState: string(model.InitialAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), destsConfigValueRawJSON),
			ExpectedStatusReportState:        string(model.ConfigPendingAssignmentState),
			ExpectedResult:                   true,
		},
		{
			Name:              "Success when transitioning to READY state without inbound credentials",
			Input:             inputForNotificationStatusReturnedAssign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.InitialAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.InitialAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			ExpectedFormationAssignmentState: string(model.InitialAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), configWithDifferentStructure),
			ExpectedStatusReportState:        string(model.ReadyAssignmentState),
			ExpectedResult:                   true,
		},
		// Unassign during PreStatusReturned
		{
			Name:              "Success when transitioning from DELETING to READY",
			Input:             inputForNotificationStatusReturnedUnassign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.DeletingAssignmentState),
			FormationAssignmentService: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("CleanupFormationAssignment", ctxWithCertConsumer, fixAssignmentPairWithAsyncWebhook()).Return(false, nil)
				return svc
			},
			FormationAssignmentNotificationService: func() *automock.FormationAssignmentNotificationService {
				notificationSvc := &automock.FormationAssignmentNotificationService{}
				notificationSvc.On("GenerateFormationAssignmentPair", ctxWithCertConsumer, fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState), fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState), model.UnassignFormation).Return(fixAssignmentPairWithAsyncWebhook(), nil).Once()
				return notificationSvc
			},
			AssignmentOperationService: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", ctxWithCertConsumer, fixAssignmentOperationInputWithTypeAndTrigger(model.InstanceCreatorUnassign, model.UnassignObject)).Return(assignmentOperationID, nil).Once()
				return svc
			},
			ExpectedResult: true,
		},
		{
			Name:              "Success when transitioning from DELETING to READY but consumer type is instance creator",
			Input:             inputForNotificationStatusReturnedUnassign,
			Context:           ctxWithInstanceCreatorConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithInstanceCreatorConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.ReadyAssignmentState),
			ExpectedResult:                   true,
		},
		{
			Name:              "Success when input has fail on synchronous application set to false on notification status returned",
			Input:             inputForNotificationStatusReturnedUnassign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.DeletingAssignmentState),
			FormationAssignmentNotificationService: func() *automock.FormationAssignmentNotificationService {
				notificationSvc := &automock.FormationAssignmentNotificationService{}
				notificationSvc.On("GenerateFormationAssignmentPair", ctxWithCertConsumer, fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState), fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState), model.UnassignFormation).Return(fixAssignmentPairWithSyncWebhook(), nil).Once()
				return notificationSvc
			},
			ExpectedResult: true,
		},
		{
			Name:              "Error when input has fail on synchronous application set to true on notification status returned",
			Input:             fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddressAndShouldFail(model.UnassignFormation, fixWebhookWithAsyncCallbackMode(), preNotificationStatusReturnedLocation, false, true),
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.DeletingAssignmentState),
			FormationAssignmentNotificationService: func() *automock.FormationAssignmentNotificationService {
				notificationSvc := &automock.FormationAssignmentNotificationService{}
				notificationSvc.On("GenerateFormationAssignmentPair", ctxWithCertConsumer, fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState), fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState), model.UnassignFormation).Return(fixAssignmentPairWithSyncWebhook(), nil).Once()
				return notificationSvc
			},
			ExpectedErrorMsg: "Instance creator is not supported on synchronous participants",
		},
		{
			Name:              "Error when transitioning from DELETING to READY but there is no consumer in context",
			Input:             inputForNotificationStatusReturnedUnassign,
			Context:           ctx,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctx, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.ReadyAssignmentState),
			ExpectedResult:                   false,
			ExpectedErrorMsg:                 "while fetching consumer info from context",
		},
		{
			Name:              "Error when creating assignment operation fails",
			Input:             inputForNotificationStatusReturnedUnassign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			FormationAssignmentNotificationService: func() *automock.FormationAssignmentNotificationService {
				notificationSvc := &automock.FormationAssignmentNotificationService{}
				notificationSvc.On("GenerateFormationAssignmentPair", ctxWithCertConsumer, fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState), fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState), model.UnassignFormation).Return(fixAssignmentPairWithAsyncWebhook(), nil).Once()
				return notificationSvc
			},
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.DeletingAssignmentState),
			AssignmentOperationService: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", ctxWithCertConsumer, fixAssignmentOperationInputWithTypeAndTrigger(model.InstanceCreatorUnassign, model.UnassignObject)).Return("", testErr).Once()
				return svc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:              "Error when cleanup formation assignment fails",
			Input:             inputForNotificationStatusReturnedUnassign,
			Context:           ctxWithCertConsumer,
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.DeletingAssignmentState),
			FormationAssignmentService: func() *automock.FormationAssignmentService {
				svc := &automock.FormationAssignmentService{}
				svc.On("CleanupFormationAssignment", ctxWithCertConsumer, fixAssignmentPairWithAsyncWebhook()).Return(false, testErr)
				return svc
			},
			FormationAssignmentNotificationService: func() *automock.FormationAssignmentNotificationService {
				notificationSvc := &automock.FormationAssignmentNotificationService{}
				notificationSvc.On("GenerateFormationAssignmentPair", ctxWithCertConsumer, fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState), fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState), model.UnassignFormation).Return(fixAssignmentPairWithAsyncWebhook(), nil).Once()
				return notificationSvc
			},
			AssignmentOperationService: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("Create", ctxWithCertConsumer, fixAssignmentOperationInputWithTypeAndTrigger(model.InstanceCreatorUnassign, model.UnassignObject)).Return(assignmentOperationID, nil).Once()
				return svc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                             "Error during generating formation assignment pair",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Context:                          ctxWithCertConsumer,
			Assignment:                       fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.ReadyAssignmentState),
			ExpectedStatusReportState:        string(model.DeletingAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			FormationAssignmentNotificationService: func() *automock.FormationAssignmentNotificationService {
				notificationSvc := &automock.FormationAssignmentNotificationService{}
				notificationSvc.On("GenerateFormationAssignmentPair", ctxWithCertConsumer, fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState), fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState), model.UnassignFormation).Return(nil, testErr).Once()
				return notificationSvc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                             "Error when getting latest assignment operation fails",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Context:                          ctxWithCertConsumer,
			Assignment:                       fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.DeleteErrorAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			AssignmentOperationService: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", ctxWithCertConsumer, formationAssignmentID, formationID).Return(nil, testErr).Once()
				return svc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                             "Success when transitioning from INSTANCE_CREATOR_UNASSIGN latest operation to DELETE_ERROR",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Context:                          ctxWithCertConsumer,
			Assignment:                       fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.DeleteErrorAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			AssignmentOperationService: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", ctxWithCertConsumer, formationAssignmentID, formationID).Return(assignmentOperationWithInstanceCreatorUnassignType, nil).Once()
				return svc
			},
			ExpectedStatusReportState: string(model.DeleteErrorAssignmentState),
			ExpectedResult:            true,
		},
		{
			Name:                             "Success when transitioning from DELETING to DELETE_ERROR",
			Input:                            inputForNotificationStatusReturnedUnassign,
			Context:                          ctxWithCertConsumer,
			Assignment:                       fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment:                fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedFormationAssignmentState: string(model.DeletingAssignmentState),
			StatusReport:                     fixNotificationStatusReportWithState(model.DeleteErrorAssignmentState),
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			AssignmentOperationService: func() *automock.AssignmentOperationService {
				svc := &automock.AssignmentOperationService{}
				svc.On("GetLatestOperation", ctxWithCertConsumer, formationAssignmentID, formationID).Return(assignmentOperationWithUnassignType, nil).Once()
				return svc
			},
			ExpectedStatusReportState: string(model.DeleteErrorAssignmentState),
			ExpectedResult:            true,
		},
		{
			Name:    "Error when retrieving status report pointer fails",
			Input:   inputForNotificationStatusReturnedUnassign,
			Context: ctxWithCertConsumer,
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedErrorMsg:  "The join point details' notification status report memory address cannot be 0",
		},
		{
			Name:  "Error when retrieving formation assignment pointer fails",
			Input: inputForNotificationStatusReturnedUnassign,
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			Context:          ctxWithCertConsumer,
			ExpectedErrorMsg: "The join point details' assignment memory address cannot be 0",
		},
		{
			Name:    "Success when input has fail on non-BTP application set to false",
			Input:   inputForNotificationStatusReturnedUnassign,
			Context: ctxWithCertConsumer,
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(map[string]*model.Label{}, nil)
				return labelRepo
			},
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedResult:    true,
		},
		{
			Name:    "Error when input has fail on non-BTP application set to true",
			Input:   fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddressAndShouldFail(model.UnassignFormation, fixWebhookWithAsyncCallbackMode(), preNotificationStatusReturnedLocation, true, false),
			Context: ctxWithCertConsumer,
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(map[string]*model.Label{}, nil)
				return labelRepo
			},
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedResult:    false,
			ExpectedErrorMsg:  "Instance creator is not supported on non-BTP participants",
		},
		{
			Name:    "Error when input has fail on non-BTP application set to true and label is invalid type",
			Input:   fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddressAndShouldFail(model.UnassignFormation, fixWebhookWithAsyncCallbackMode(), preNotificationStatusReturnedLocation, true, false),
			Context: ctxWithCertConsumer,
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(map[string]*model.Label{operators.GlobalSubaccountLabelKey: {Value: []string{"invalid"}}}, nil)
				return labelRepo
			},
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedResult:    false,
			ExpectedErrorMsg:  "Instance creator is not supported on non-BTP participants",
		},
		{
			Name:    "Error when input has fail on non-BTP application set to true and label does not have value",
			Input:   fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddressAndShouldFail(model.UnassignFormation, fixWebhookWithAsyncCallbackMode(), preNotificationStatusReturnedLocation, true, false),
			Context: ctxWithCertConsumer,
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(map[string]*model.Label{operators.GlobalSubaccountLabelKey: {}}, nil)
				return labelRepo
			},
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedResult:    false,
			ExpectedErrorMsg:  "Instance creator is not supported on non-BTP participants",
		},
		{
			Name:    "Error when input has fail on non-BTP application set to true and getting label fails",
			Input:   fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddressAndShouldFail(model.UnassignFormation, fixWebhookWithAsyncCallbackMode(), preNotificationStatusReturnedLocation, true, false),
			Context: ctxWithCertConsumer,
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(nil, testErr)
				return labelRepo
			},
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedResult:    false,
			ExpectedErrorMsg:  "Instance creator is not supported on non-BTP participants",
		},
		{
			Name:    "Success when input has fail on synchronous application set to false",
			Input:   fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddress(model.UnassignFormation, fixWebhookWithSyncMode(), preSendNotificationLocation),
			Context: ctxWithCertConsumer,
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			StatusReport:      fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), emptyConfig),
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedResult:    true,
		},
		{
			Name:    "Success when input has fail on synchronous application set to false",
			Input:   fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddressAndShouldFail(model.UnassignFormation, fixWebhookWithSyncMode(), preSendNotificationLocation, false, true),
			Context: ctxWithCertConsumer,
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			StatusReport:      fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), emptyConfig),
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedErrorMsg:  "Instance creator is not supported on synchronous participants",
		},
		{
			Name:    "Error when retrieving webhook pointer fails during send notification",
			Input:   fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddress(model.UnassignFormation, nil, preSendNotificationLocation),
			Context: ctxWithCertConsumer,
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedResult:    false,
			ExpectedErrorMsg:  "The webhook memory address cannot be 0",
		},
		{
			Name:  "Success when join point is not supported",
			Input: fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddress(model.UnassignFormation, fixWebhookWithAsyncCallbackMode(), preAssignFormationLocation),

			Context: ctxWithCertConsumer,
			LabelRepository: func() *automock.LabelRepository {
				labelRepo := &automock.LabelRepository{}
				labelRepo.On("ListForObject", ctxWithCertConsumer, testTenantID, model.ApplicationLabelableObject, appID).Return(subaccountnLbl, nil)
				return labelRepo
			},
			Assignment:        fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ReverseAssignment: fixFormationAssignmentWithStateAndTarget(model.DeletingAssignmentState),
			ExpectedResult:    true,
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
			assignmentOperationService := &automock.AssignmentOperationService{}
			if testCase.AssignmentOperationService != nil {
				assignmentOperationService = testCase.AssignmentOperationService()
			}
			labelRepo := &automock.LabelRepository{}
			if testCase.LabelRepository != nil {
				labelRepo = testCase.LabelRepository()
			}

			engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, labelRepo, nil, nil, nil, nil, formationAssignmentRepo, formationAssignmentService, formationAssignmentNotificationService, assignmentOperationService, runtimeType, applicationType)

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

			mock.AssertExpectationsForObjects(t, formationAssignmentRepo, formationAssignmentService, formationAssignmentNotificationService, assignmentOperationService)
		})
	}

	t.Run("Error when incorrect input is provided", func(t *testing.T) {
		// GIVEN

		engine := operators.NewConstraintEngine(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

		// WHEN
		input := "wrong input"
		result, err := engine.AsynchronousFlowControlOperator(ctxWithCertConsumer, input)

		// THEN
		assert.Equal(t, false, result)
		assert.Equal(t, "Incompatible input for operator: AsynchronousFlowControl", err.Error())
	})
}
