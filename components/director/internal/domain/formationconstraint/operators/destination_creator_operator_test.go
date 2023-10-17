package operators_test

import (
	"encoding/json"
	"fmt"
	"testing"

	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConstraintOperators_DestinationCreator(t *testing.T) {
	certData := fixCertificateData()

	designTimeDests := fixDesignTimeDestinations()
	basicDests := fixBasicDestinations()
	samlAssertionDests := fixSAMLAssertionDestinations()
	clientCertAuthDests := fixClientCertAuthDestinations()

	basicCreds := fixBasicCreds()
	samlAssertionCreds := fixSAMLCreds()
	clientCertAuthCreds := fixClientCertAuthCreds()

	testCases := []struct {
		Name                  string
		Input                 operators.OperatorInput
		DestinationSvc        func() *automock.DestinationService
		DestinationCreatorSvc func() *automock.DestinationCreatorService
		ExpectedResult        bool
		ExpectedErrorMsg      string
	}{
		{
			Name:  "Success when operation is 'unassign' and location is 'NotificationStatusReturned'",
			Input: inputForUnassignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("DeleteDestinations", ctx, fa, false).Return(nil)
				return destSvc
			},
			ExpectedResult: true,
		},
		{
			Name:           "Success when formation assignment state is in 'INITIAL' state",
			Input:          inputForAssignWithFormationAssignmentInitialState,
			ExpectedResult: true,
		},
		{
			Name:  "Success when operation is 'assign' and location is 'NotificationStatusReturned' with full destination config",
			Input: inputForAssignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateDesignTimeDestinations", ctx, designTimeDests, fa, false).Return(nil)
				return destSvc
			},
			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("CreateCertificate", ctx, samlAssertionDests, destinationcreatorpkg.AuthTypeSAMLAssertion, fa, uint8(0), false).Return(certData, nil)
				destCreatorSvc.On("EnrichAssignmentConfigWithSAMLCertificateData", fa.Value, destinationcreatorpkg.SAMLAssertionDestPath, certData).Return(destsConfigValueRawJSON, nil)
				destCreatorSvc.On("CreateCertificate", ctx, clientCertAuthDests, destinationcreatorpkg.AuthTypeClientCertificate, fa, uint8(0), false).Return(certData, nil)
				destCreatorSvc.On("EnrichAssignmentConfigWithCertificateData", fa.Value, destinationcreatorpkg.ClientCertAuthDestPath, certData).Return(destsConfigValueRawJSON, nil)
				return destCreatorSvc
			},
			ExpectedResult: true,
		},
		{
			Name:  "Success when operation is 'assign' and location is 'SendNotification'",
			Input: inputForAssignSendNotification,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateBasicCredentialDestinations", ctx, basicDests, basicCreds, fa, corrleationIDs, false).Return(nil)
				destSvc.On("CreateSAMLAssertionDestination", ctx, samlAssertionDests, samlAssertionCreds, fa, corrleationIDs, false).Return(nil)
				destSvc.On("CreateClientCertificateAuthenticationDestination", ctx, clientCertAuthDests, clientCertAuthCreds, fa, corrleationIDs, false).Return(nil)
				return destSvc
			},
			ExpectedResult: true,
		},
		{
			Name:           "Success when operation is 'Unassign' and location is 'SendNotification'",
			Input:          inputForUnassignSendNotification,
			ExpectedResult: true,
		},
		{
			Name:             "Error when parsing operator input",
			Input:            "wrong input",
			ExpectedErrorMsg: "Incompatible input for operator:",
		},
		{
			Name:             "Error when formation operation is invalid",
			Input:            inputWithInvalidOperation,
			ExpectedErrorMsg: "The formation operation is invalid:",
		},
		{
			Name:  "Error when operation is 'unassign' and location is 'NotificationStatusReturned' and the deletion of destinations fails",
			Input: inputForUnassignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("DeleteDestinations", ctx, fa, false).Return(testErr)
				return destSvc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:             "Error when operation is 'assign' and location is 'NotificationStatusReturned' and config unmarshalling fails",
			Input:            inputForAssignNotificationStatusReturnedWithInvalidFAConfig,
			ExpectedErrorMsg: "while unmarshalling tenant mapping response configuration from assignment with ID:",
		},
		{
			Name:             "Error when retrieving fa pointer fails",
			Input:            inputWithoutAssignmentMemoryAddress,
			ExpectedErrorMsg: "The join point details' assignment memory address cannot be 0",
		},
		{
			Name:  "Error when operation is 'assign' and location is 'NotificationStatusReturned' and the creation of design time dests fails",
			Input: inputForAssignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateDesignTimeDestinations", ctx, designTimeDests, fa, false).Return(testErr)
				return destSvc
			},
			ExpectedErrorMsg: fmt.Sprintf("while creating design time destinations: %s", testErr.Error()),
		},
		{
			Name:  "Success(no-op) when operation is 'assign', location is 'NotificationStatusReturned' and SAML assertion certificate data is already exists",
			Input: inputWithAssignmentWithSAMLCertData,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateDesignTimeDestinations", ctx, designTimeDests, faWithSAMLCertData, false).Return(nil)
				return destSvc
			},
			ExpectedResult: true,
		},
		{
			Name:  "Error when operation is 'assign', location is 'NotificationStatusReturned' and the creation of SAML assertion certificate fails",
			Input: inputForAssignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateDesignTimeDestinations", ctx, designTimeDests, fa, false).Return(nil)
				return destSvc
			},
			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("CreateCertificate", ctx, samlAssertionDests, destinationcreatorpkg.AuthTypeSAMLAssertion, fa, uint8(0), false).Return(nil, testErr)
				return destCreatorSvc
			},
			ExpectedErrorMsg: fmt.Sprintf("while creating SAML assertion certificate: %s", testErr.Error()),
		},
		{
			Name:  "Error when operation is 'assign' and location is 'NotificationStatusReturned' and the enrichment of config with SAML cert fails",
			Input: inputForAssignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateDesignTimeDestinations", ctx, designTimeDests, fa, false).Return(nil)
				return destSvc
			},
			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("CreateCertificate", ctx, samlAssertionDests, destinationcreatorpkg.AuthTypeSAMLAssertion, fa, uint8(0), false).Return(certData, nil)
				destCreatorSvc.On("EnrichAssignmentConfigWithSAMLCertificateData", fa.Value, destinationcreatorpkg.SAMLAssertionDestPath, certData).Return(json.RawMessage{}, testErr)
				return destCreatorSvc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:  "Success(no-op) when operation is 'assign', location is 'NotificationStatusReturned' and client cert auth certificate data is already exists",
			Input: inputWithAssignmentWithClientCertAuthCertData,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateDesignTimeDestinations", ctx, designTimeDests, faWithClientCertAuthCertData, false).Return(nil)
				return destSvc
			},
			ExpectedResult: true,
		},
		{
			Name:  "Error when operation is 'assign', location is 'NotificationStatusReturned' and the creation of client cert auth certificate fails",
			Input: inputForAssignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateDesignTimeDestinations", ctx, designTimeDests, fa, false).Return(nil)
				return destSvc
			},
			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("CreateCertificate", ctx, samlAssertionDests, destinationcreatorpkg.AuthTypeSAMLAssertion, fa, uint8(0), false).Return(certData, nil)
				destCreatorSvc.On("EnrichAssignmentConfigWithSAMLCertificateData", fa.Value, destinationcreatorpkg.SAMLAssertionDestPath, certData).Return(destsConfigValueRawJSON, nil)
				destCreatorSvc.On("CreateCertificate", ctx, clientCertAuthDests, destinationcreatorpkg.AuthTypeClientCertificate, fa, uint8(0), false).Return(nil, testErr)
				return destCreatorSvc
			},
			ExpectedErrorMsg: fmt.Sprintf("while creating client certificate authentication certificate: %s", testErr.Error()),
		},
		{
			Name:  "Error when operation is 'assign' and location is 'NotificationStatusReturned' and the enrichment of config with client cert auth cert fails",
			Input: inputForAssignNotificationStatusReturned,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateDesignTimeDestinations", ctx, designTimeDests, fa, false).Return(nil)
				return destSvc
			},
			DestinationCreatorSvc: func() *automock.DestinationCreatorService {
				destCreatorSvc := &automock.DestinationCreatorService{}
				destCreatorSvc.On("CreateCertificate", ctx, samlAssertionDests, destinationcreatorpkg.AuthTypeSAMLAssertion, fa, uint8(0), false).Return(certData, nil)
				destCreatorSvc.On("EnrichAssignmentConfigWithSAMLCertificateData", fa.Value, destinationcreatorpkg.SAMLAssertionDestPath, certData).Return(destsConfigValueRawJSON, nil)
				destCreatorSvc.On("CreateCertificate", ctx, clientCertAuthDests, destinationcreatorpkg.AuthTypeClientCertificate, fa, uint8(0), false).Return(certData, nil)
				destCreatorSvc.On("EnrichAssignmentConfigWithCertificateData", fa.Value, destinationcreatorpkg.ClientCertAuthDestPath, certData).Return(json.RawMessage{}, testErr)
				return destCreatorSvc
			},
			ExpectedErrorMsg: testErr.Error(),
		},
		{
			Name:             "Error when operation is 'assign' and location is 'SendNotification' and config unmarshalling fails",
			Input:            inputForAssignSendNotificationWithInvalidFAConfig,
			ExpectedErrorMsg: "while unmarshalling tenant mapping configuration response from assignment with ID:",
		},
		{
			Name:             "Error when operation is 'assign' and location is 'SendNotification' and reverse config unmarshalling fails",
			Input:            inputForAssignSendNotificationWithInvalidReverseFAConfig,
			ExpectedErrorMsg: "while unmarshalling tenant mapping configuration response from reverse assignment with ID:",
		},
		{
			Name:             "Success(no-op) when operation is 'assign' and location is 'SendNotification' and inbound details are nil",
			Input:            inputForAssignSendNotificationWhereFAConfigStructureIsDifferent,
			ExpectedResult:   true,
			ExpectedErrorMsg: "",
		},
		{
			Name:             "Success(no-op) when operation is 'assign' and location is 'SendNotification' and outbound details are nil",
			Input:            inputForAssignSendNotificationWhereReverseFAConfigStructureIsDifferent,
			ExpectedResult:   true,
			ExpectedErrorMsg: "",
		},
		{
			Name:             "Success(no-op) when operation is 'assign' and location is neither 'NotificationStatusReturned' or 'SendNotification'",
			Input:            inputForAssignGenerateFANotification,
			ExpectedResult:   true,
			ExpectedErrorMsg: "",
		},
		{
			Name:             "Error when operation is 'assign' and location is 'SendNotification' and retrieving reverse assignment pointer fails",
			Input:            inputForAssignSendNotificationWithoutReverseAssignmentMemoryAddress,
			ExpectedErrorMsg: "The join point details' assignment memory address cannot be 0",
		},
		{
			Name:  "Error when operation is 'assign' and location is 'SendNotification' and the creation of basic dests fails",
			Input: inputForAssignSendNotification,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateBasicCredentialDestinations", ctx, basicDests, basicCreds, fa, corrleationIDs, false).Return(testErr)
				return destSvc
			},
			ExpectedErrorMsg: fmt.Sprintf("while creating basic destinations: %s", testErr.Error()),
		},
		{
			Name:  "Error when operation is 'assign' and location is 'SendNotification' and the creation of SAML assertion dests fails",
			Input: inputForAssignSendNotification,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateBasicCredentialDestinations", ctx, basicDests, basicCreds, fa, corrleationIDs, false).Return(nil)
				destSvc.On("CreateSAMLAssertionDestination", ctx, samlAssertionDests, samlAssertionCreds, fa, corrleationIDs, false).Return(testErr)
				return destSvc
			},
			ExpectedErrorMsg: fmt.Sprintf("while creating SAML Assertion destinations: %s", testErr.Error()),
		},
		{
			Name:  "Error when operation is 'assign' and location is 'SendNotification' and the creation of client cert auth dests fails",
			Input: inputForAssignSendNotification,
			DestinationSvc: func() *automock.DestinationService {
				destSvc := &automock.DestinationService{}
				destSvc.On("CreateBasicCredentialDestinations", ctx, basicDests, basicCreds, fa, corrleationIDs, false).Return(nil)
				destSvc.On("CreateSAMLAssertionDestination", ctx, samlAssertionDests, samlAssertionCreds, fa, corrleationIDs, false).Return(nil)
				destSvc.On("CreateClientCertificateAuthenticationDestination", ctx, clientCertAuthDests, clientCertAuthCreds, fa, corrleationIDs, false).Return(testErr)
				return destSvc
			},
			ExpectedErrorMsg: fmt.Sprintf("while creating client certificate authentication destinations: %s", testErr.Error()),
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			destSvc := unusedDestinationService()
			if testCase.DestinationSvc != nil {
				destSvc = testCase.DestinationSvc()
			}

			destCreatorSvc := unusedDestinationCreatorService()
			if testCase.DestinationCreatorSvc != nil {
				destCreatorSvc = testCase.DestinationCreatorSvc()
			}
			defer mock.AssertExpectationsForObjects(t, destSvc, destCreatorSvc)

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
		})
	}
}
