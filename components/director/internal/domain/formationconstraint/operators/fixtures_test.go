package operators_test

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationassignment"
	"github.com/kyma-incubator/compass/components/director/internal/domain/statusreport"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"

	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/pkg/errors"
)

const (
	// IDs constants
	testID                   = "d1fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	testTenantID             = "d9fddec6-5456-4a1e-9ae0-74447f5d6ae9"
	testInternalTenantID     = "aaaddec6-5456-4a1e-9ae0-74447f5d6ae9"
	inputAppID               = "eb2d5110-ca3a-11ed-afa1-0242ac120002"
	appID                    = "b55131c4-ca3a-11ed-afa1-0242ac120002"
	runtimeID                = "c66341c4-ca3a-11ed-afa1-0242ac120564"
	runtimeCtxID             = "f7156h4-ca3a-11ed-afa1-0242ac121237"
	formationAssignmentID    = "c54341c4-ca3a-11ed-afa1-0242ac120564"
	assignmentOperationID    = "3a711086-3cac-4253-97e9-6c0417f5cc67"
	formationTemplateID      = "b87631c4-ca3a-11ed-afa1-0242ac120002"
	otherFormationTemplateID = "b05731c4-ca3a-11ed-afa1-0242ac120002"
	formationID              = "a16724q3-ba3a-13ef-a1c7-1247ac120123"
	webhookID                = "f4aac335-8afa-421f-a5ad-da9ce7a676bc"
	externalSubaccountID     = "04e408ff-7915-4642-b491-8a80960936b2"

	// Certificate constants
	testFileName   = "test-file-name"
	testCommonName = "test-common-name"
	testCertChain  = "test-cert-chain"

	// Destination constants
	designTimeDestName        = "design-time-name"
	basicDestName             = "name-basic"
	samlAssertionDestName     = "saml-assertion-name"
	clientCertAuthDestName    = "client-cert-auth-dest-name"
	oauth2ClientCredsDestName = "oauth2-client-creds-name"
	oauth2mTLSDestName        = "oauth2-mTLS-name"
	destinationURL            = "http://test-url"
	destinationType           = destinationcreatorpkg.TypeHTTP
	destinationProxyType      = destinationcreatorpkg.ProxyTypeInternet
	destinationNoAuthn        = destinationcreatorpkg.AuthTypeNoAuth

	// Creds constants
	basicDestUser                        = "user"
	basicDestPassword                    = "pwd"
	oauth2ClientCredsDestTokenServiceURL = "http://test-token-url"
	oauth2ClientCredsDestClientID        = "test-client-id"
	oauth2ClientCredsDestClientSecret    = "test-client-secret"
	oauth2mTLSDestTokenServiceURL        = "http://test-token-url-mTLS"
	oauth2mTLSDestClientID               = "test-client-id-mTLS"

	// Other
	formationConstraintName = "test constraint"
	operatorName            = operators.IsNotAssignedToAnyFormationOfTypeOperator
	resourceSubtype         = "test subtype"
	exceptResourceType      = "except subtype"
	inputTemplate           = `{"formation_template_id": "{{.FormationTemplateID}}","resource_type": "{{.ResourceType}}","resource_subtype": "{{.ResourceSubtype}}","resource_id": "{{.ResourceID}}","tenant": "{{.TenantID}}"}`
	scenario                = "test-scenario"
	runtimeType             = "runtimeType"
	applicationType         = "applicationType"
	exceptType              = "except-type"
	formationType           = "formationType"
	applicationTypeLabel    = "applicationType"
	runtimeTypeLabel        = "runtimeType"
	inputAppType            = "input-type"
	testScenarioGroup       = "scenarioGroup"
	testJSONConfig          = `{"key": "val"}`
)

// Common variables used across different operators' tests
var (
	ctx            = context.TODO()
	testErr        = errors.New("test error")
	corrleationIDs []string
	defaultTime    = time.Time{}

	preNotificationStatusReturnedLocation = fixJoinPointLocation(model.NotificationStatusReturned, model.PreOperation)
	preSendNotificationLocation           = fixJoinPointLocation(model.SendNotificationOperation, model.PreOperation)
	preGenerateFANotificationLocation     = fixJoinPointLocation(model.GenerateFormationAssignmentNotificationOperation, model.PreOperation)
	preAssignFormationLocation            = fixJoinPointLocation(model.AssignFormationOperation, model.PreOperation)

	details = formationconstraintpkg.AssignFormationOperationDetails{
		ResourceType:    "runtime",
		ResourceSubtype: "kyma",
	}

	// func TestConstraintEngine_EnforceConstraints
	formationConstraintUnsupportedOperatorModel = &model.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        "unsupported",
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
	}

	formationConstraintModel = &model.FormationConstraint{
		ID:              testID,
		Name:            formationConstraintName,
		ConstraintType:  model.PreOperation,
		TargetOperation: model.AssignFormationOperation,
		Operator:        operatorName,
		ResourceType:    model.ApplicationResourceType,
		ResourceSubtype: resourceSubtype,
		InputTemplate:   inputTemplate,
		ConstraintScope: model.FormationTypeFormationConstraintScope,
	}

	webhookModelSync          = graphql.WebhookModeSync
	webhookModelAsyncCallback = graphql.WebhookModeAsyncCallback
)

// Destination Creator variables
var (
	emptyConfig                  = json.RawMessage("{}")
	invalidFAConfig              = json.RawMessage("invalid-Destination-config")
	configWithDifferentStructure = json.RawMessage(testJSONConfig)
	destsConfigValueRawJSON      = json.RawMessage(
		fmt.Sprintf(`{"credentials":{"inboundCommunication":{"samlAssertion":{"destinations":[{"url":"%s","name":"%s"}]},"clientCertificateAuthentication":{"destinations":[{"url":"%s","name":"%s"}]},"basicAuthentication":{"destinations":[{"url":"%s","name":"%s"}]},"oauth2ClientCredentials":{"destinations":[{"url":"%s","name":"%s"}]},"oauth2mtls":{"destinations":[{"url":"%s","name":"%s"}]}}},"destinations":[{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authenticationType":"%s"}]}`, destinationURL, samlAssertionDestName, destinationURL, clientCertAuthDestName, destinationURL, basicDestName, destinationURL, oauth2ClientCredsDestName, destinationURL, oauth2mTLSDestName, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)),
	)

	destsConfigValueRawJSONDeprecatedKey = json.RawMessage(
		fmt.Sprintf(`{"credentials":{"inboundCommunication":{"samlAssertion":{"destinations":[{"url":"%s","name":"%s"}]},"clientCertificateAuthentication":{"destinations":[{"url":"%s","name":"%s"}]},"basicAuthentication":{"destinations":[{"url":"%s","name":"%s"}]},"oauth2ClientCredentials":{"destinations":[{"url":"%s","name":"%s"}]}}},"destinations":[{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authentication":"%s"}]}`, destinationURL, samlAssertionDestName, destinationURL, clientCertAuthDestName, destinationURL, basicDestName, destinationURL, oauth2ClientCredsDestName, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)),
	)

	destsReverseConfigValueRawJSON = json.RawMessage(
		fmt.Sprintf(`{"credentials":{"inboundCommunication":{"samlAssertion":{"destinations":[{"url":"%s","name":"%s"}]},"basicAuthentication":{"destinations":[{"url":"%s","name":"%s"}]}},"outboundCommunication":{"basicAuthentication":{"url":"%s","username":"%s","password":"%s"},"samlAssertion":{"url":"%s"},"clientCertificateAuthentication":{"url":"%s"},"oauth2ClientCredentials":{"url":"%s","tokenServiceURL":"%s","clientId":"%s","clientSecret":"%s"},"oauth2mtls":{"url":"%s","tokenServiceURL":"%s","clientId":"%s"}}},"destinations":[{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authenticationType":"%s"}]}`, destinationURL, samlAssertionDestName, destinationURL, basicDestName, destinationURL, basicDestUser, basicDestPassword, destinationURL, destinationURL, destinationURL, oauth2ClientCredsDestTokenServiceURL, oauth2ClientCredsDestClientID, oauth2ClientCredsDestClientSecret, destinationURL, oauth2mTLSDestTokenServiceURL, oauth2mTLSDestClientID, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)),
	)

	destsConfigWithSAMLCertDataRawJSON = json.RawMessage(
		fmt.Sprintf(`{"credentials":{"inboundCommunication":{"samlAssertion":{"certificate":"cert-chain-data","assertionIssuer":"assertionIssuerValue","destinations":[{"url":"%s","name":"%s"}]},"clientCertificateAuthentication":{"destinations":[{"url":"%s","name":"%s"}]},"basicAuthentication":{"destinations":[{"url":"%s","name":"%s"}]}}},"destinations":[{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authenticationType":"%s"}]}`, destinationURL, samlAssertionDestName, destinationURL, clientCertAuthDestName, destinationURL, basicDestName, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)),
	)

	destsConfigWithClientCertauthCertDataRawJSON = json.RawMessage(
		fmt.Sprintf(`{"credentials":{"inboundCommunication":{"clientCertificateAuthentication":{"certificate":"cert-chain-data","destinations":[{"url":"%s","name":"%s"}]},"basicAuthentication":{"destinations":[{"url":"%s","name":"%s"}]}}},"destinations":[{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authenticationType":"%s"}]}`, destinationURL, clientCertAuthDestName, destinationURL, basicDestName, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)),
	)

	destsConfigWithOauth2mTLSCertDataRawJSON = json.RawMessage(
		fmt.Sprintf(`{"credentials":{"inboundCommunication":{"basicAuthentication":{"destinations":[{"url":"%s","name":"%s"}]},"oauth2ClientCredentials":{"destinations":[{"url":"%s","name":"%s"}]},"oauth2mtls":{"certificate":"cert-chain-data","destinations":[{"url":"%s","name":"%s"}]}}},"destinations":[{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authenticationType":"%s"}]}`, destinationURL, basicDestName, destinationURL, oauth2ClientCredsDestName, destinationURL, oauth2mTLSDestName, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)),
	)

	statusReportWithConfigAndReadyState                             = fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), destsConfigValueRawJSON)
	statusReportWithConfigAndReadyStateWithDeprecatedDestinationKey = fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), destsConfigValueRawJSONDeprecatedKey)
	statusReportWitInvalidConfig                                    = fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), invalidFAConfig)
	statusRportWitSAMLCertData                                      = fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), destsConfigWithSAMLCertDataRawJSON)
	statusRportWitClientCertAuthCertData                            = fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), destsConfigWithClientCertauthCertDataRawJSON)
	statusRportWitOauth2mTLSCertData                                = fixNotificationStatusReportWithStateAndConfig(string(model.ReadyAssignmentState), destsConfigWithOauth2mTLSCertDataRawJSON)

	fa                           = fixFormationAssignmentWithConfig(destsConfigValueRawJSON)
	faDeprecatedDestinationKey   = fixFormationAssignmentWithConfig(destsConfigValueRawJSONDeprecatedKey)
	reverseFa                    = fixFormationAssignmentWithConfig(destsReverseConfigValueRawJSON)
	faWithInitialState           = fixFormationAssignmentWithState(model.InitialAssignmentState)
	faWithInvalidConfig          = fixFormationAssignmentWithConfig(invalidFAConfig)
	faWithSAMLCertData           = fixFormationAssignmentWithConfig(destsConfigWithSAMLCertDataRawJSON)
	faWithClientCertAuthCertData = fixFormationAssignmentWithConfig(destsConfigWithClientCertauthCertDataRawJSON)
	faWithOauth2mTLSCertData     = fixFormationAssignmentWithConfig(destsConfigWithOauth2mTLSCertDataRawJSON)

	faConfigWithDifferentStructure = fixFormationAssignmentWithConfig(configWithDifferentStructure)

	destinationCreatorInputForUnassignNotificationStatusReturned = fixDestinationCreatorInputForUnassignWithLocationOperation(model.NotificationStatusReturned)
	destinationCreatorInputForUnassignSendNotification           = fixDestinationCreatorInputForUnassignWithLocationOperation(model.SendNotificationOperation)

	inputForAssignWithFormationAssignmentInitialState = &formationconstraintpkg.DestinationCreatorInput{
		Operation:       model.AssignFormation,
		FAMemoryAddress: faWithInitialState.GetAddress(),
	}

	inputWithAssignmentWithSAMLCertData                                  = fixDestinationCreatorInputWithAssignmentMemoryAddress(model.AssignFormation, faWithSAMLCertData, preNotificationStatusReturnedLocation, statusRportWitSAMLCertData)
	inputWithAssignmentWithClientCertAuthCertData                        = fixDestinationCreatorInputWithAssignmentMemoryAddress(model.AssignFormation, faWithClientCertAuthCertData, preNotificationStatusReturnedLocation, statusRportWitClientCertAuthCertData)
	inputWithAssignmentWithOauth2mTLSCertData                            = fixDestinationCreatorInputWithAssignmentMemoryAddress(model.AssignFormation, faWithOauth2mTLSCertData, preNotificationStatusReturnedLocation, statusRportWitOauth2mTLSCertData)
	inputForAssignNotificationStatusReturned                             = fixDestinationCreatorInputWithAssignmentMemoryAddress(model.AssignFormation, fa, preNotificationStatusReturnedLocation, statusReportWithConfigAndReadyState)
	inputForAssignNotificationStatusReturnedWithDeprecatedDestinationKey = fixDestinationCreatorInputWithAssignmentMemoryAddress(model.AssignFormation, faDeprecatedDestinationKey, preNotificationStatusReturnedLocation, statusReportWithConfigAndReadyStateWithDeprecatedDestinationKey)

	inputForAssignNotificationStatusReturnedWithCertSvcKeyStore         = fixDestinationCreatorInputWithAssignmentMemoryAddressAndCertSvcKeystore(model.AssignFormation, fa, preNotificationStatusReturnedLocation, true, statusReportWithConfigAndReadyState)
	inputForAssignNotificationStatusReturnedWithInvalidFAConfig         = fixDestinationCreatorInputWithAssignmentMemoryAddress(model.AssignFormation, faWithInvalidConfig, preNotificationStatusReturnedLocation, statusReportWitInvalidConfig)
	inputForAssignSendNotificationWithoutReverseAssignmentMemoryAddress = fixDestinationCreatorInputWithAssignmentMemoryAddress(model.AssignFormation, fa, preSendNotificationLocation, statusReportWithConfigAndReadyState)

	inputForAssignSendNotification                                         = fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, fa, reverseFa, preSendNotificationLocation, statusReportWithConfigAndReadyState)
	inputForAssignSendNotificationWithInvalidFAConfig                      = fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, faWithInvalidConfig, reverseFa, preSendNotificationLocation, statusReportWithConfigAndReadyState)
	inputForAssignSendNotificationWithInvalidReverseFAConfig               = fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, fa, faWithInvalidConfig, preSendNotificationLocation, statusReportWithConfigAndReadyState)
	inputForAssignSendNotificationWhereFAConfigStructureIsDifferent        = fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, faConfigWithDifferentStructure, faConfigWithDifferentStructure, preSendNotificationLocation, statusReportWithConfigAndReadyState)
	inputForAssignSendNotificationWhereReverseFAConfigStructureIsDifferent = fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, fa, faConfigWithDifferentStructure, preSendNotificationLocation, statusReportWithConfigAndReadyState)
	inputForAssignGenerateFANotification                                   = fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, fa, reverseFa, preGenerateFANotificationLocation, statusReportWithConfigAndReadyState)

	inputWithInvalidOperation = &formationconstraintpkg.DestinationCreatorInput{
		Operation: model.CreateFormation,
	}

	inputWithoutAssignmentMemoryAddress = &formationconstraintpkg.DestinationCreatorInput{
		Operation: model.AssignFormation,
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  model.NotificationStatusReturned,
			ConstraintType: model.PreOperation,
		},
	}
)

// IsNotAssignedToAnyFormationOfTypeOperator variables
var (
	inputTenantResourceType = &formationconstraintpkg.IsNotAssignedToAnyFormationOfTypeInput{
		FormationTemplateID: formationTemplateID,
		ResourceType:        model.TenantResourceType,
		ResourceSubtype:     "account",
		ResourceID:          testID,
		Tenant:              testTenantID,
	}

	inputApplicationResourceType = &formationconstraintpkg.IsNotAssignedToAnyFormationOfTypeInput{
		FormationTemplateID: formationTemplateID,
		ResourceType:        model.ApplicationResourceType,
		ResourceSubtype:     "app",
		ResourceID:          testID,
		Tenant:              testTenantID,
		ExceptSystemTypes:   []string{exceptResourceType},
	}

	inputApplicationResourceTypeWithSubtypeThatIsException = &formationconstraintpkg.IsNotAssignedToAnyFormationOfTypeInput{
		FormationTemplateID: otherFormationTemplateID,
		ResourceType:        model.ApplicationResourceType,
		ResourceSubtype:     exceptResourceType,
		ResourceID:          testID,
		Tenant:              testTenantID,
		ExceptSystemTypes:   []string{exceptResourceType},
	}

	inputRuntimeResourceType = &formationconstraintpkg.IsNotAssignedToAnyFormationOfTypeInput{
		FormationTemplateID: formationTemplateID,
		ResourceType:        model.RuntimeResourceType,
		ResourceSubtype:     "account",
		ResourceID:          testID,
		Tenant:              testTenantID,
	}

	emptyAssignments = []*model.AutomaticScenarioAssignment{}

	assignments = []*model.AutomaticScenarioAssignment{
		{ScenarioName: scenario},
	}

	formations = []*model.Formation{
		{
			Name:                scenario,
			FormationTemplateID: otherFormationTemplateID,
		},
	}
	formations2 = []*model.Formation{
		{
			Name:                scenario,
			FormationTemplateID: formationTemplateID,
		},
	}

	formationAssignments = []*model.FormationAssignment{
		{
			FormationID: formationID,
		},
	}
)

// DoNotGenerateFormationAssignmentNotificationOperator and DoNotGenerateFormationAssignmentNotificationForLoopsOperator variables
var (
	in = &formationconstraintpkg.DoNotGenerateFormationAssignmentNotificationInput{
		ResourceType:       model.ApplicationResourceType,
		ResourceSubtype:    inputAppType,
		ResourceID:         inputAppID,
		SourceResourceType: model.ApplicationResourceType,
		SourceResourceID:   appID,
		Tenant:             testTenantID,
		ExceptSubtypes:     []string{exceptType},
	}

	inLoop = &formationconstraintpkg.DoNotGenerateFormationAssignmentNotificationInput{
		ResourceType:       model.ApplicationResourceType,
		ResourceSubtype:    inputAppType,
		ResourceID:         inputAppID,
		SourceResourceType: model.ApplicationResourceType,
		SourceResourceID:   inputAppID,
		Tenant:             testTenantID,
		ExceptSubtypes:     []string{exceptType},
	}

	inWithFormationTypeException = &formationconstraintpkg.DoNotGenerateFormationAssignmentNotificationInput{
		ResourceType:         model.ApplicationResourceType,
		FormationTemplateID:  formationTemplateID,
		ResourceSubtype:      inputAppType,
		ResourceID:           inputAppID,
		SourceResourceType:   model.ApplicationResourceType,
		SourceResourceID:     appID,
		Tenant:               testTenantID,
		ExceptSubtypes:       []string{exceptType},
		ExceptFormationTypes: []string{formationType},
	}

	inWithFormationTypeExceptionLoop = &formationconstraintpkg.DoNotGenerateFormationAssignmentNotificationInput{
		ResourceType:         model.ApplicationResourceType,
		FormationTemplateID:  formationTemplateID,
		ResourceSubtype:      inputAppType,
		ResourceID:           inputAppID,
		SourceResourceType:   model.ApplicationResourceType,
		SourceResourceID:     inputAppID,
		Tenant:               testTenantID,
		ExceptSubtypes:       []string{exceptType},
		ExceptFormationTypes: []string{formationType},
	}

	runtimeIn = &formationconstraintpkg.DoNotGenerateFormationAssignmentNotificationInput{
		ResourceType:       model.ApplicationResourceType,
		ResourceSubtype:    inputAppType,
		ResourceID:         inputAppID,
		SourceResourceType: model.RuntimeResourceType,
		SourceResourceID:   runtimeID,
		Tenant:             testTenantID,
		ExceptSubtypes:     []string{exceptType},
	}

	runtimeContextIn = &formationconstraintpkg.DoNotGenerateFormationAssignmentNotificationInput{
		ResourceType:       model.ApplicationResourceType,
		ResourceSubtype:    inputAppType,
		ResourceID:         inputAppID,
		SourceResourceType: model.RuntimeContextResourceType,
		SourceResourceID:   runtimeCtxID,
		Tenant:             testTenantID,
		ExceptSubtypes:     []string{exceptType},
	}

	subaccountnLbl = map[string]*model.Label{
		operators.GlobalSubaccountLabelKey: {
			Value: externalSubaccountID,
		},
	}
)

// RedirectNotificationOperator variables
var (
	graphqlWebhook                   = fixWebhookWithAsyncCallbackMode()
	inputWithoutWebhookMemoryAddress = &formationconstraintpkg.RedirectNotificationInput{}
	webhookURL                       = "testWebhookURL"
	webhookURLTemplate               = "testWebhookURLTemplate"
)

// AsynchronousFlowControlOperator fixtures

func fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddressAndShouldFail(operation model.FormationOperation, webhook *graphql.Webhook, location formationconstraintpkg.JoinPointLocation, shouldFailOnGlobaSubaccountLabel, shouldFailOnSync bool) *formationconstraintpkg.AsynchronousFlowControlOperatorInput {
	return fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddressShouldRedirect(false, operation, webhook, location, shouldFailOnGlobaSubaccountLabel, shouldFailOnSync)
}

func fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddress(operation model.FormationOperation, webhook *graphql.Webhook, location formationconstraintpkg.JoinPointLocation) *formationconstraintpkg.AsynchronousFlowControlOperatorInput {
	return fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddressShouldRedirect(false, operation, webhook, location, false, false)
}

func fixAsynchronousFlowControlOperatorInputWithAssignmentAndReverseFAMemoryAddressShouldRedirect(shouldRedirect bool, operation model.FormationOperation, webhook *graphql.Webhook, location formationconstraintpkg.JoinPointLocation, shouldFailOnGlobaSubaccountLabel, shouldFailOnSync bool) *formationconstraintpkg.AsynchronousFlowControlOperatorInput {
	return &formationconstraintpkg.AsynchronousFlowControlOperatorInput{
		RedirectNotificationInput: formationconstraintpkg.RedirectNotificationInput{
			ShouldRedirect:       shouldRedirect,
			WebhookMemoryAddress: webhook.GetAddress(),
			Operation:            operation,
			Location:             location,
		},
		FailOnNonBTPParticipants: shouldFailOnGlobaSubaccountLabel,
		FailOnSyncParticipants:   shouldFailOnSync,
	}
}

func cloneAsynchronousFlowControlOperatorInput(input *formationconstraintpkg.AsynchronousFlowControlOperatorInput) *formationconstraintpkg.AsynchronousFlowControlOperatorInput {
	return &formationconstraintpkg.AsynchronousFlowControlOperatorInput{
		RedirectNotificationInput: formationconstraintpkg.RedirectNotificationInput{
			ShouldRedirect:       input.ShouldRedirect,
			WebhookMemoryAddress: input.WebhookMemoryAddress,
			Operation:            input.Operation,
			Location:             input.Location,
		},
		FailOnNonBTPParticipants: input.FailOnNonBTPParticipants,
		FailOnSyncParticipants:   input.FailOnSyncParticipants,
	}
}

func setAssignmentToAsynchronousFlowControlInput(input *formationconstraintpkg.AsynchronousFlowControlOperatorInput, assignment *model.FormationAssignment) {
	input.FAMemoryAddress = assignment.GetAddress()
}

func setReverseAssignmentToAsynchronousFlowControlInput(input *formationconstraintpkg.AsynchronousFlowControlOperatorInput, assignment *model.FormationAssignment) {
	input.ReverseFAMemoryAddress = assignment.GetAddress()
}

func setStatusReportToAsynchronousFlowControlInput(input *formationconstraintpkg.AsynchronousFlowControlOperatorInput, report *statusreport.NotificationStatusReport) {
	input.NotificationStatusReportMemoryAddress = report.GetAddress()
}

func fixAssignmentOperationModelWithTypeAndTrigger(opType model.AssignmentOperationType, opTrigger model.OperationTrigger) *model.AssignmentOperation {
	return &model.AssignmentOperation{
		ID:                    assignmentOperationID,
		Type:                  opType,
		FormationAssignmentID: formationAssignmentID,
		FormationID:           formationID,
		TriggeredBy:           opTrigger,
		StartedAtTimestamp:    &defaultTime,
		FinishedAtTimestamp:   &defaultTime,
	}
}

func fixAssignmentOperationInputWithTypeAndTrigger(opType model.AssignmentOperationType, opTrigger model.OperationTrigger) *model.AssignmentOperationInput {
	return &model.AssignmentOperationInput{
		Type:                  opType,
		FormationAssignmentID: formationAssignmentID,
		FormationID:           formationID,
		TriggeredBy:           opTrigger,
	}
}

// Destination Creator operator fixtures

func fixDestinationCreatorInputWithAssignmentMemoryAddress(operation model.FormationOperation, formationAssignment *model.FormationAssignment, location formationconstraintpkg.JoinPointLocation, report *statusreport.NotificationStatusReport) *formationconstraintpkg.DestinationCreatorInput {
	return &formationconstraintpkg.DestinationCreatorInput{
		Operation:                             operation,
		FAMemoryAddress:                       formationAssignment.GetAddress(),
		NotificationStatusReportMemoryAddress: report.GetAddress(),
		Location:                              location,
	}
}

func fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(operation model.FormationOperation, assignment, reverseAssignment *model.FormationAssignment, location formationconstraintpkg.JoinPointLocation, report *statusreport.NotificationStatusReport) *formationconstraintpkg.DestinationCreatorInput {
	return &formationconstraintpkg.DestinationCreatorInput{
		Operation:                             operation,
		FAMemoryAddress:                       assignment.GetAddress(),
		ReverseFAMemoryAddress:                reverseAssignment.GetAddress(),
		NotificationStatusReportMemoryAddress: report.GetAddress(),
		Location:                              location,
	}
}

func fixDestinationCreatorInputForUnassignWithLocationOperation(operationName model.TargetOperation) *formationconstraintpkg.DestinationCreatorInput {
	return &formationconstraintpkg.DestinationCreatorInput{
		Operation:                             model.UnassignFormation,
		FAMemoryAddress:                       fa.GetAddress(),
		NotificationStatusReportMemoryAddress: statusReportWithConfigAndReadyState.GetAddress(),
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  operationName,
			ConstraintType: model.PreOperation,
		},
	}
}

func fixDestinationCreatorInputWithAssignmentMemoryAddressAndCertSvcKeystore(operation model.FormationOperation, formationAssignment *model.FormationAssignment, location formationconstraintpkg.JoinPointLocation, useCertSvcKeystoreForSAML bool, report *statusreport.NotificationStatusReport) *formationconstraintpkg.DestinationCreatorInput {
	return &formationconstraintpkg.DestinationCreatorInput{
		Operation:                             operation,
		FAMemoryAddress:                       formationAssignment.GetAddress(),
		NotificationStatusReportMemoryAddress: report.GetAddress(),
		Location:                              location,
		UseCertSvcKeystoreForSAML:             useCertSvcKeystoreForSAML,
	}
}

func fixDesignTimeDestination() operators.DestinationRaw {
	return operators.DestinationRaw{
		Destination: json.RawMessage(fmt.Sprintf(`{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authenticationType":"%s"}`, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)))}
}

func fixDesignTimeDestinations() []operators.DestinationRaw {
	return []operators.DestinationRaw{
		fixDesignTimeDestination(),
	}
}

func fixBasicDestinations() []operators.Destination {
	return []operators.Destination{
		fixDestination(basicDestName, destinationURL),
	}
}

func fixSAMLAssertionDestinations() []operators.Destination {
	return []operators.Destination{
		fixDestination(samlAssertionDestName, destinationURL),
	}
}

func fixClientCertAuthDestinations() []operators.Destination {
	return []operators.Destination{
		fixDestination(clientCertAuthDestName, destinationURL),
	}
}

func fixOAuth2ClientCredsDestinations() []operators.Destination {
	return []operators.Destination{
		fixDestination(oauth2ClientCredsDestName, destinationURL),
	}
}

func fixOAuth2mTLSDestinations() []operators.Destination {
	return []operators.Destination{
		fixDestination(oauth2mTLSDestName, destinationURL),
	}
}

func fixDestination(name, url string) operators.Destination {
	return operators.Destination{
		Name: name,
		URL:  url,
	}
}

func fixBasicCreds() operators.BasicAuthentication {
	return operators.BasicAuthentication{
		URL:      destinationURL,
		Username: basicDestUser,
		Password: basicDestPassword,
	}
}

func fixSAMLCreds() *operators.SAMLAssertionAuthentication {
	return &operators.SAMLAssertionAuthentication{
		URL: destinationURL,
	}
}

func fixClientCertAuthCreds() *operators.ClientCertAuthentication {
	return &operators.ClientCertAuthentication{
		URL: destinationURL,
	}
}

func fixOAuth2ClientCreds() *operators.OAuth2ClientCredentialsAuthentication {
	return &operators.OAuth2ClientCredentialsAuthentication{
		URL:             destinationURL,
		TokenServiceURL: oauth2ClientCredsDestTokenServiceURL,
		ClientID:        oauth2ClientCredsDestClientID,
		ClientSecret:    oauth2ClientCredsDestClientSecret,
	}
}

func fixOAuth2mTLSAuthn() *operators.OAuth2mTLSAuthentication {
	return &operators.OAuth2mTLSAuthentication{
		URL:             destinationURL,
		TokenServiceURL: oauth2mTLSDestTokenServiceURL,
		ClientID:        oauth2mTLSDestClientID,
	}
}

func fixCertificateData() *operators.CertificateData {
	return &operators.CertificateData{
		FileName:         testFileName,
		CommonName:       testCommonName,
		CertificateChain: testCertChain,
	}
}

// Config Mutator operator fixtures

func fixConfigMutatorInput(fa *model.FormationAssignment, notificationStatusReport *statusreport.NotificationStatusReport, state, config *string, onlyForSourceSubtypes []string) *formationconstraintpkg.ConfigMutatorInput {
	return &formationconstraintpkg.ConfigMutatorInput{
		Operation:                             model.UnassignFormation,
		NotificationStatusReportMemoryAddress: notificationStatusReport.GetAddress(),
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  model.NotificationStatusReturned,
			ConstraintType: model.PreOperation,
		},
		SourceResourceType:    model.ResourceType(fa.SourceType),
		SourceResourceID:      fa.Source,
		ModifiedConfiguration: config,
		State:                 state,
		Tenant:                testTenantID,
		OnlyForSourceSubtypes: onlyForSourceSubtypes,
	}
}

// Redirect Notification operator fixtures

func fixRedirectNotificationOperatorInput(shouldRedirect bool) *formationconstraintpkg.RedirectNotificationInput {
	return &formationconstraintpkg.RedirectNotificationInput{
		ShouldRedirect:       shouldRedirect,
		URLTemplate:          "redirectNotificationOperatorInputURLTemplate",
		URL:                  "redirectNotificationOperatorInputURL",
		WebhookMemoryAddress: graphqlWebhook.GetAddress(),
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  model.SendNotificationOperation,
			ConstraintType: model.PreOperation,
		},
	}
}

func fixNotificationStatusReport() *statusreport.NotificationStatusReport {
	return &statusreport.NotificationStatusReport{}
}

func fixNotificationStatusReportWithStateAndConfig(state string, config json.RawMessage) *statusreport.NotificationStatusReport {
	return &statusreport.NotificationStatusReport{
		State:         state,
		Configuration: config,
	}
}

func fixNotificationStatusReportWithState(state model.FormationAssignmentState) *statusreport.NotificationStatusReport {
	return &statusreport.NotificationStatusReport{
		State: string(state),
	}
}

func fixWebhookWithSyncMode() *graphql.Webhook {
	return &graphql.Webhook{
		ID:          webhookID,
		URL:         &webhookURL,
		URLTemplate: &webhookURLTemplate,
		Mode:        &webhookModelSync,
	}
}

func fixWebhookWithAsyncCallbackMode() *graphql.Webhook {
	return &graphql.Webhook{
		ID:          webhookID,
		URL:         &webhookURL,
		URLTemplate: &webhookURLTemplate,
		Mode:        &webhookModelAsyncCallback,
	}
}

func fixAssignmentPairWithAsyncWebhook() *formationassignment.AssignmentMappingPairWithOperation {
	return &formationassignment.AssignmentMappingPairWithOperation{
		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
			AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
				Request: &webhookclient.FormationAssignmentNotificationRequest{
					Webhook: fixWebhookWithAsyncCallbackMode(),
				},
			},
			ReverseAssignmentReqMapping: nil,
		},
		Operation: model.UnassignFormation,
	}
}

func fixAssignmentPairWithSyncWebhook() *formationassignment.AssignmentMappingPairWithOperation {
	return &formationassignment.AssignmentMappingPairWithOperation{
		AssignmentMappingPair: &formationassignment.AssignmentMappingPair{
			AssignmentReqMapping: &formationassignment.FormationAssignmentRequestMapping{
				Request: &webhookclient.FormationAssignmentNotificationRequest{
					Webhook: fixWebhookWithSyncMode(),
				},
			},
			ReverseAssignmentReqMapping: nil,
		},
		Operation: model.UnassignFormation,
	}
}

// Common fixtures for all operators

func fixFormationAssignmentWithConfig(config json.RawMessage) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:    formationAssignmentID,
		State: string(model.ReadyAssignmentState),
		Value: config,
	}
}

func fixFormationAssignmentWithState(state model.FormationAssignmentState) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          formationAssignmentID,
		FormationID: formationID,
		State:       string(state),
	}
}

func fixFormationAssignmentWithStateAndTarget(state model.FormationAssignmentState) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          formationAssignmentID,
		FormationID: formationID,
		TenantID:    testTenantID,
		TargetType:  model.FormationAssignmentTypeApplication,
		Target:      appID,
		State:       string(state),
	}
}

func fixJoinPointLocation(operationName model.TargetOperation, constraintType model.FormationConstraintType) formationconstraintpkg.JoinPointLocation {
	return formationconstraintpkg.JoinPointLocation{
		OperationName:  operationName,
		ConstraintType: constraintType,
	}
}

// Unused service mocks

func unusedTenantService() *automock.TenantService {
	return &automock.TenantService{}
}

func unusedASAService() *automock.AutomaticScenarioAssignmentService {
	return &automock.AutomaticScenarioAssignmentService{}
}

func unusedLabelRepo() *automock.LabelRepository {
	return &automock.LabelRepository{}
}

func unusedFormationRepo() *automock.FormationRepository {
	return &automock.FormationRepository{}
}

func unusedFormationTemplateRepo() *automock.FormationTemplateRepo {
	return &automock.FormationTemplateRepo{}
}

func unusedRuntimeContextRepo() *automock.RuntimeContextRepo {
	return &automock.RuntimeContextRepo{}
}

func unusedApplicationRepo() *automock.ApplicationRepository {
	return &automock.ApplicationRepository{}
}

func unusedLabelService() *automock.LabelService {
	return &automock.LabelService{}
}

func unusedDestinationService() *automock.DestinationService {
	return &automock.DestinationService{}
}

func unusedSystemAuthService() *automock.SystemAuthService {
	return &automock.SystemAuthService{}
}

func unusedDestinationCreatorService() *automock.DestinationCreatorService {
	return &automock.DestinationCreatorService{}
}
