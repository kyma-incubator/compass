package operators_test

// todo::: adapt
//import (
//	"encoding/json"
//	"testing"
//
//	"github.com/stretchr/testify/mock"
//
//	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
//	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
//	"github.com/stretchr/testify/assert"
//	"github.com/stretchr/testify/require"
//)
//
//func TestConstraintOperators_DestinationCreator(t *testing.T) {
//	testCases := []struct {
//		Name                  string
//		Input                 operators.OperatorInput
//		DestinationSvc        func() *automock.DestinationService
//		DestinationCreatorSvc func() *automock.DestinationCreatorService
//		ExpectedResult        bool
//		ExpectedErrorMsg      string
//	}{
//		{
//			Name:  "Success when operation is 'unassign' and location is 'NotificationStatusReturned'",
//			Input: inputForUnassignNotificationStatusReturned,
//			DestinationSvc: func() *automock.DestinationService {
//				destSvc := &automock.DestinationService{}
//				destSvc.On("DeleteDestinations", ctx, fa).Return(nil)
//				return destSvc
//			},
//			ExpectedResult: true,
//		},
//		{
//			Name:           "Success when formation assignment state is not 'Ready' or 'Config pending'",
//			Input:          inputForAssignWithFormationAssignmentDeletingState,
//			ExpectedResult: true,
//		},
//		{
//			Name:  "Success when operation is 'assign' and location is 'NotificationStatusReturned'",
//			Input: inputForAssignNotificationStatusReturned,
//			DestinationSvc: func() *automock.DestinationService {
//				destSvc := &automock.DestinationService{}
//				destSvc.On("CreateDesignTimeDestinations", ctx, fixDesignTimeDestination(), fa).Return(nil)
//				return destSvc
//			},
//			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
//				destCreatorSvc := &automock.DestinationCreatorService{}
//				destCreatorSvc.On("CreateCertificate", ctx, fixSAMLAssertionDestination(), fa, uint8(0)).Return(fixCertificateData(), nil)
//				destCreatorSvc.On("EnrichAssignmentConfigWithCertificateData", fa.Value, fixCertificateData(), 0).Return(destsConfigValueRawJSON, nil)
//				return destCreatorSvc
//			},
//			ExpectedResult: true,
//		},
//		{
//			Name:  "Success when operation is 'assign' and location is 'SendNotification'",
//			Input: inputForAssignSendNotification,
//			DestinationSvc: func() *automock.DestinationService {
//				destSvc := &automock.DestinationService{}
//				destSvc.On("CreateBasicCredentialDestinations", ctx, fixBasicDestination(), fixBasicCreds(), fa, corrleationIDs).Return(nil)
//				destSvc.On("CreateSAMLAssertionDestination", ctx, fixSAMLAssertionDestination(), fixSAMLCreds(), fa, corrleationIDs).Return(nil)
//				return destSvc
//			},
//			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
//				return &automock.DestinationCreatorService{}
//			},
//			ExpectedResult: true,
//		},
//		{
//			Name:           "Success when operation is 'Unassign' and location is 'SendNotification'",
//			Input:          inputForUnassignSendNotification,
//			ExpectedResult: true,
//		},
//		{
//			Name:             "Error when parsing operator input",
//			Input:            "wrong input",
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "Incompatible input for operator:",
//		},
//		{
//			Name:             "Error when formation operation is invalid",
//			Input:            inputWithInvalidOperation,
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "The formation operation is invalid:",
//		},
//		{
//			Name:  "Error when operation is 'unassign' and location is 'NotificationStatusReturned' and the deletion of destinations fails",
//			Input: inputForUnassignNotificationStatusReturned,
//			DestinationSvc: func() *automock.DestinationService {
//				destSvc := &automock.DestinationService{}
//				destSvc.On("DeleteDestinations", ctx, fa).Return(testErr)
//				return destSvc
//			},
//			ExpectedResult:   false,
//			ExpectedErrorMsg: testErr.Error(),
//		},
//		{
//			Name:             "Error when operation is 'assign' and location is 'NotificationStatusReturned' and config unmarshalling fails",
//			Input:            inputForAssignNotificationStatusReturnedWithInvalidFAConfig,
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "while unmarshalling tenant mapping response configuration from assignment with ID:",
//		},
//		{
//			Name:             "Error when retrieving fa pointer fails",
//			Input:            inputWithoutAssignmentMemoryAddress,
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "The join point details' assignment memory address cannot be 0",
//		},
//		{
//			Name:  "Error when operation is 'assign' and location is 'NotificationStatusReturned' and the creation of design time dests fails",
//			Input: inputForAssignNotificationStatusReturned,
//			DestinationSvc: func() *automock.DestinationService {
//				destSvc := &automock.DestinationService{}
//				destSvc.On("CreateDesignTimeDestinations", ctx, fixDesignTimeDestination(), fa).Return(testErr)
//				return destSvc
//			},
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "while creating design time destination with name:",
//		},
//		{
//			Name:  "Error when operation is 'assign' and location is 'NotificationStatusReturned' and the creation of certificates fails",
//			Input: inputForAssignNotificationStatusReturned,
//			DestinationSvc: func() *automock.DestinationService {
//				destSvc := &automock.DestinationService{}
//				destSvc.On("CreateDesignTimeDestinations", ctx, fixDesignTimeDestination(), fa).Return(nil)
//				return destSvc
//			},
//			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
//				destCreatorSvc := &automock.DestinationCreatorService{}
//				destCreatorSvc.On("CreateCertificate", ctx, fixSAMLAssertionDestination(), fa, uint8(0)).Return(nil, testErr)
//				return destCreatorSvc
//			},
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "while creating SAML assertion certificate with name:",
//		},
//		{
//			Name:  "Error when operation is 'assign' and location is 'NotificationStatusReturned' and the enrichment of config fails",
//			Input: inputForAssignNotificationStatusReturned,
//			DestinationSvc: func() *automock.DestinationService {
//				destSvc := &automock.DestinationService{}
//				destSvc.On("CreateDesignTimeDestinations", ctx, fixDesignTimeDestination(), fa).Return(nil)
//				return destSvc
//			},
//			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
//				destCreatorSvc := &automock.DestinationCreatorService{}
//				destCreatorSvc.On("CreateCertificate", ctx, fixSAMLAssertionDestination(), fa, uint8(0)).Return(fixCertificateData(), nil)
//				destCreatorSvc.On("EnrichAssignmentConfigWithCertificateData", fa.Value, fixCertificateData(), 0).Return(json.RawMessage{}, testErr)
//				return destCreatorSvc
//			},
//			ExpectedResult:   false,
//			ExpectedErrorMsg: testErr.Error(),
//		},
//		{
//			Name:             "Error when operation is 'assign' and location is 'SendNotification' and config unmarshalling fails",
//			Input:            inputForAssignSendNotificationWithInvalidFAConfig,
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "while unmarshalling tenant mapping configuration response from assignment with ID:",
//		},
//		{
//			Name:             "Error when operation is 'assign' and location is 'SendNotification' and reverse config unmarshalling fails",
//			Input:            inputForAssignSendNotificationWithInvalidReverseFAConfig,
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "while unmarshalling tenant mapping configuration response from reverse assignment with ID:",
//		},
//		{
//			Name:             "Error when operation is 'assign' and location is 'SendNotification' and inbound details are nil",
//			Input:            inputForAssignSendNotificationWhereFAConfigStructureIsDifferent,
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "The inbound communication destination details could not be empty",
//		},
//		{
//			Name:             "Error when operation is 'assign' and location is 'SendNotification' and outbound details are nil",
//			Input:            inputForAssignSendNotificationWhereReverseFAConfigStructureIsDifferent,
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "The outbound communication credentials could not be empty",
//		},
//		{
//			Name:             "Error when operation is 'assign' and location is 'SendNotification' and retrieving reverse assignment pointer fails",
//			Input:            inputForAssignSendNotificationWithoutReverseAssignmentMemoryAddress,
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "The join point details' assignment memory address cannot be 0",
//		},
//		{
//			Name:  "Error when operation is 'assign' and location is 'SendNotification' and the creation of basic dests fails",
//			Input: inputForAssignSendNotification,
//			DestinationSvc: func() *automock.DestinationService {
//				destSvc := &automock.DestinationService{}
//				destSvc.On("CreateBasicCredentialDestinations", ctx, fixBasicDestination(), fixBasicCreds(), fa, corrleationIDs).Return(testErr)
//				return destSvc
//			},
//			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
//				return &automock.DestinationCreatorService{}
//			},
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "while creating basic destination with name:",
//		},
//		{
//			Name:  "Error when operation is 'assign' and location is 'SendNotification' and the creation of SAML assertion dests fails",
//			Input: inputForAssignSendNotification,
//			DestinationSvc: func() *automock.DestinationService {
//				destSvc := &automock.DestinationService{}
//				destSvc.On("CreateBasicCredentialDestinations", ctx, fixBasicDestination(), fixBasicCreds(), fa, corrleationIDs).Return(nil)
//				destSvc.On("CreateSAMLAssertionDestination", ctx, fixSAMLAssertionDestination(), fixSAMLCreds(), fa, corrleationIDs).Return(testErr)
//				return destSvc
//			},
//			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
//				destCreatorSvc := &automock.DestinationCreatorService{}
//				return destCreatorSvc
//			},
//			ExpectedResult:   false,
//			ExpectedErrorMsg: "while creating SAML assertion destination with name:",
//		},
//	}
//	for _, testCase := range testCases {
//		t.Run(testCase.Name, func(t *testing.T) {
//			// GIVEN
//			destSvc := UnusedDestinationService()
//			if testCase.DestinationSvc != nil {
//				destSvc = testCase.DestinationSvc()
//			}
//
//			destCreatorSvc := UnusedDestinationCreatorService()
//			if testCase.DestinationCreatorSvc != nil {
//				destCreatorSvc = testCase.DestinationCreatorSvc()
//			}
//			defer mock.AssertExpectationsForObjects(t, destSvc, destCreatorSvc)
//
//			engine := operators.NewConstraintEngine(nil, nil, nil, nil, destSvc, destCreatorSvc, nil, nil, nil, nil, nil, nil, nil, runtimeType, applicationType)
//
//			// WHEN
//			result, err := engine.DestinationCreator(ctx, testCase.Input)
//
//			// THEN
//			if testCase.ExpectedErrorMsg != "" {
//				require.Error(t, err)
//				assert.Contains(t, err.Error(), testCase.ExpectedErrorMsg)
//			} else {
//				assert.Equal(t, testCase.ExpectedResult, result)
//				assert.NoError(t, err)
//			}
//		})
//	}
//}
