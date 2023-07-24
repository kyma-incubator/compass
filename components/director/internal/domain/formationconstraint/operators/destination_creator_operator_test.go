package operators_test

import (
	"encoding/json"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestConstraintOperators_DestinationCreator(t *testing.T) {
	testCases := []struct {
		Name                            string
		Input                           operators.OperatorInput
		DestinationSvc                  func() *automock.DestinationService
		DestinationCreatorSvc           func() *automock.DestinationCreatorService
		ExpectedResult                  bool
		ExpectedErrorMsg                string
		InputFormationAssignment        *model.FormationAssignment // since we work with the actual fa memory address in the operator, we need to have a mechanism to 'reset'/pass a new 'clean' fa after it's been modified in the flow so that the next test can use a correct fa
		InputReverseFormationAssignment *model.FormationAssignment // since we work with the actual fa memory address in the operator, we need to have a mechanism to 'reset'/pass a new 'clean' fa after it's been modified in the flow so that the next test can use a correct fa
	}{
		{
			Name:  "Success when operation is 'unassign' and location is 'NotificationStatusReturned'",
			Input: inputForUnassignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("DeleteDestinations", ctx, inputForUnassignNotificationStatusReturned.FormationAssignment).Return(nil)
				return destSvc
			},
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        true,
		},
		{
			Name:                  "Success when formation assignment state is not 'Ready' or 'Config pending'",
			Input:                 inputForAssignWithFormationAssignmentDeletingState,
			DestinationSvc:        UnusedDestinationService,
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        true,
		},
		{
			Name:  "Success when operation is 'assign' and location is 'NotificationStatusReturned'",
			Input: inputForAssignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateDesignTimeDestinations", ctx, fixDesignTimeDestination(), inputForAssignNotificationStatusReturned.FormationAssignment).Return(nil)
				return destSvc
			},
			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("CreateCertificate", ctx, fixSAMLAssertionDestination(), inputForAssignNotificationStatusReturned.FormationAssignment, uint8(0)).Return(fixCertificateData(), nil)
				destCreatorSvc.On("EnrichAssignmentConfigWithCertificateData", inputForAssignNotificationStatusReturned.FormationAssignment.Value, fixCertificateData(), 0).Return(destsConfigValueRawJSON, nil)
				return destCreatorSvc
			},
			InputFormationAssignment: fixFormationAssignmentWithConfig(destsConfigValueRawJSON),
			ExpectedResult:           true,
		},
		{
			Name:  "Success when operation is 'assign' and location is 'SendNotification'",
			Input: inputForAssignSendNotification,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateBasicCredentialDestinations", ctx, fixBasicDestination(), fixBasicCreds(), inputForAssignSendNotification.FormationAssignment, corrleationIDs).Return(nil)
				destSvc.On("CreateSAMLAssertionDestination", ctx, fixSAMLAssertionDestination(), fixSAMLCreds(), inputForAssignSendNotification.FormationAssignment, corrleationIDs).Return(nil)
				return destSvc
			},
			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
				return &automock.DestinationCreatorService{}
			},
			InputFormationAssignment:        fixFormationAssignmentWithConfig(destsConfigValueRawJSON),
			InputReverseFormationAssignment: fixFormationAssignmentWithConfig(destsReverseConfigValueRawJSON),
			ExpectedResult:                  true,
		},
		{
			Name:                  "Success when operation is 'Unassign' and location is 'SendNotification'",
			Input:                 inputForUnassignSendNotification,
			DestinationSvc:        UnusedDestinationService,
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        true,
		},
		{
			Name:                  "Error when parsing operator input",
			Input:                 "wrong input",
			DestinationSvc:        UnusedDestinationService,
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        false,
			ExpectedErrorMsg:      "Incompatible input for operator:",
		},
		{
			Name:                  "Error when formation operation is invalid",
			Input:                 inputWithInvalidOperation,
			DestinationSvc:        UnusedDestinationService,
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        false,
			ExpectedErrorMsg:      "The formation operation is invalid:",
		},
		{
			Name:  "Error when operation is 'unassign' and location is 'NotificationStatusReturned' and the deletion of destinations fails",
			Input: inputForUnassignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("DeleteDestinations", ctx, inputForUnassignNotificationStatusReturned.FormationAssignment).Return(testErr)
				return destSvc
			},
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        false,
			ExpectedErrorMsg:      testErr.Error(),
		},
		{
			Name:                  "Error when operation is 'assign' and location is 'NotificationStatusReturned' and config unmarshalling fails",
			Input:                 inputForAssignNotificationStatusReturnedWithInvalidFAConfig,
			DestinationSvc:        UnusedDestinationService,
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        false,
			ExpectedErrorMsg:      "while unmarshaling tenant mapping response configuration from assignment with ID:",
		},
		{
			Name:                  "Error when operation is 'assign' and location is 'NotificationStatusReturned' and retrieving fa pointer fails",
			Input:                 inputForAssignNotificationStatusReturnedWithoutMemoryAddress,
			DestinationSvc:        UnusedDestinationService,
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        false,
			ExpectedErrorMsg:      "The join point details' assignment memory address cannot be 0",
		},
		{
			Name:  "Error when operation is 'assign' and location is 'NotificationStatusReturned' and the creation of design time dests fails",
			Input: inputForAssignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateDesignTimeDestinations", ctx, fixDesignTimeDestination(), inputForAssignNotificationStatusReturned.FormationAssignment).Return(testErr)
				return destSvc
			},
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        false,
			ExpectedErrorMsg:      "while creating design time destination with name:",
		},
		{
			Name:  "Error when operation is 'assign' and location is 'NotificationStatusReturned' and the creation of certificates fails",
			Input: inputForAssignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateDesignTimeDestinations", ctx, fixDesignTimeDestination(), inputForAssignNotificationStatusReturned.FormationAssignment).Return(nil)
				return destSvc
			},
			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("CreateCertificate", ctx, fixSAMLAssertionDestination(), inputForAssignNotificationStatusReturned.FormationAssignment, uint8(0)).Return(nil, testErr)
				return destCreatorSvc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: "while creating SAML assertion certificate with name:",
		},
		{
			Name:  "Error when operation is 'assign' and location is 'NotificationStatusReturned' and the enrichment of config fails",
			Input: inputForAssignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateDesignTimeDestinations", ctx, fixDesignTimeDestination(), inputForAssignNotificationStatusReturned.FormationAssignment).Return(nil)
				return destSvc
			},
			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("CreateCertificate", ctx, fixSAMLAssertionDestination(), inputForAssignNotificationStatusReturned.FormationAssignment, uint8(0)).Return(fixCertificateData(), nil)
				destCreatorSvc.On("EnrichAssignmentConfigWithCertificateData", inputForAssignNotificationStatusReturned.FormationAssignment.Value, fixCertificateData(), 0).Return(json.RawMessage{}, testErr)
				return destCreatorSvc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:                  "Error when operation is 'assign' and location is 'SendNotification' and config unmarshalling fails",
			Input:                 inputForAssignSendNotificationWithInvalidFAConfig,
			DestinationSvc:        UnusedDestinationService,
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        false,
			ExpectedErrorMsg:      "while unmarshaling tenant mapping configuration response from assignment with ID:",
		},
		{
			Name:                  "Error when operation is 'assign' and location is 'SendNotification' and reverse config unmarshalling fails",
			Input:                 inputForAssignSendNotificationWithInvalidReverseFAConfig,
			DestinationSvc:        UnusedDestinationService,
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        false,
			ExpectedErrorMsg:      "while unmarshaling tenant mapping configuration response from reverse assignment with ID:",
		},
		{
			Name:                  "Error when operation is 'assign' and location is 'SendNotification' and inbound details are nil",
			Input:                 inputForAssignSendNotificationWhereFAConfigStructureIsDifferent,
			DestinationSvc:        UnusedDestinationService,
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        false,
			ExpectedErrorMsg:      "The inbound communication destination details could not be empty",
		},
		{
			Name:                  "Error when operation is 'assign' and location is 'SendNotification' and outbound details are nil",
			Input:                 inputForAssignSendNotificationWhereReverseFAConfigStructureIsDifferent,
			DestinationSvc:        UnusedDestinationService,
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        false,
			ExpectedErrorMsg:      "The outbound communication credentials could not be empty",
		},
		{
			Name:                  "Error when operation is 'assign' and location is 'SendNotification' and retrieving fa pointer fails",
			Input:                 inputForAssignSendNotificationWithoutMemoryAddress,
			DestinationSvc:        UnusedDestinationService,
			DestinationCreatorSvc: UnusedDestinationCreatorService,
			ExpectedResult:        false,
			ExpectedErrorMsg:      "The join point details' assignment memory address cannot be 0",
		},
		{
			Name:  "Error when operation is 'assign' and location is 'SendNotification' and the creation of basic dests fails",
			Input: inputForAssignSendNotification,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateBasicCredentialDestinations", ctx, fixBasicDestination(), fixBasicCreds(), inputForAssignSendNotification.FormationAssignment, corrleationIDs).Return(testErr)
				return destSvc
			},
			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
				return &automock.DestinationCreatorService{}
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: "while creating basic destination with name:",
		},
		{
			Name:  "Error when operation is 'assign' and location is 'SendNotification' and the creation of SAML assertion dests fails",
			Input: inputForAssignSendNotification,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateBasicCredentialDestinations", ctx, fixBasicDestination(), fixBasicCreds(), inputForAssignSendNotification.FormationAssignment, corrleationIDs).Return(nil)
				destSvc.On("CreateSAMLAssertionDestination", ctx, fixSAMLAssertionDestination(), fixSAMLCreds(), inputForAssignSendNotification.FormationAssignment, corrleationIDs).Return(testErr)
				return destSvc
			},
			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				return destCreatorSvc
			},
			ExpectedResult:   false,
			ExpectedErrorMsg: "while creating SAML assertion destination with name:",
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			destSvc := UnusedDestinationService()
			if testCase.DestinationSvc != nil {
				destSvc = testCase.DestinationSvc()
			}

			destCreatorSvc := UnusedDestinationCreatorService()
			if testCase.DestinationSvc != nil {
				destCreatorSvc = testCase.DestinationCreatorSvc()
			}

			if testCase.InputFormationAssignment != nil {
				in, ok := testCase.Input.(*formationconstraint.DestinationCreatorInput)
				require.True(t, ok)
				in.FormationAssignment = testCase.InputFormationAssignment
				in.JoinPointDetailsFAMemoryAddress = testCase.InputFormationAssignment.GetAddress()
			}

			if testCase.InputReverseFormationAssignment != nil {
				in, ok := testCase.Input.(*formationconstraint.DestinationCreatorInput)
				require.True(t, ok)
				in.ReverseFormationAssignment = testCase.InputReverseFormationAssignment
				in.JoinPointDetailsReverseFAMemoryAddress = testCase.InputReverseFormationAssignment.GetAddress()
			}

			engine := operators.NewConstraintEngine(nil, nil, nil, nil, destSvc, destCreatorSvc, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)

			// WHEN
			result, err := engine.DestinationCreator(ctx, testCase.Input)

			// THEN
			if testCase.ExpectedErrorMsg != "" {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
			} else {
				assert.Equal(t, testCase.ExpectedResult, result)
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, destSvc, destCreatorSvc)
		})
	}
}
