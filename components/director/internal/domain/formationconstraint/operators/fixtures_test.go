package operators_test

import (
	"context"
	"encoding/json"
	"fmt"

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
	formationTemplateID      = "b87631c4-ca3a-11ed-afa1-0242ac120002"
	otherFormationTemplateID = "b05731c4-ca3a-11ed-afa1-0242ac120002"

	// Certificate constants
	testFileName   = "test-file-name"
	testCommonName = "test-common-name"
	testCertChain  = "test-cert-chain"

	// Destination constants
	designTimeDestName     = "design-time-name"
	basicDestName          = "name-basic"
	samlAssertionDestName  = "saml-assertion-name"
	clientCertAuthDestName = "client-cert-auth-dest-name"
	destinationURL         = "http://test-url"
	destinationType        = destinationcreatorpkg.TypeHTTP
	destinationProxyType   = destinationcreatorpkg.ProxyTypeInternet
	destinationNoAuthn     = destinationcreatorpkg.AuthTypeNoAuth

	// Creds constants
	basicDestUser     = "user"
	basicDestPassword = "pwd"

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
	testJSONConfig          = `{"key": "val"}`
)

var (
	ctx            = context.TODO()
	testErr        = errors.New("test error")
	corrleationIDs []string

	invalidFAConfig              = json.RawMessage("invalid-destination-config")
	configWithDifferentStructure = json.RawMessage(testJSONConfig)
	destsConfigValueRawJSON      = json.RawMessage(
		fmt.Sprintf(`{"credentials":{"inboundCommunication":{"samlAssertion":{"destinations":[{"url":"%s","name":"%s"}]},"clientCertificateAuthentication":{"destinations":[{"url":"%s","name":"%s"}]},"basicAuthentication":{"destinations":[{"url":"%s","name":"%s"}]}}},"destinations":[{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authentication":"%s"}]}`, destinationURL, samlAssertionDestName, destinationURL, clientCertAuthDestName, destinationURL, basicDestName, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)),
	)
	destsReverseConfigValueRawJSON = json.RawMessage(
		fmt.Sprintf(`{"credentials":{"inboundCommunication":{"samlAssertion":{"destinations":[{"url":"%s","name":"%s"}]},"basicAuthentication":{"destinations":[{"url":"%s","name":"%s"}]}},"outboundCommunication":{"basicAuthentication":{"url":"%s","username":"%s","password":"%s"},"samlAssertion":{"url":"%s"},"clientCertificateAuthentication":{"url":"%s"}}},"destinations":[{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authentication":"%s"}]}`, destinationURL, samlAssertionDestName, destinationURL, basicDestName, destinationURL, basicDestUser, basicDestPassword, destinationURL, destinationURL, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)),
	)

	destsConfigWithSAMLCertDataRawJSON = json.RawMessage(
		fmt.Sprintf(`{"credentials":{"inboundCommunication":{"samlAssertion":{"certificate":"cert-chain-data","assertionIssuer":"assertionIssuerValue","destinations":[{"url":"%s","name":"%s"}]},"clientCertificateAuthentication":{"destinations":[{"url":"%s","name":"%s"}]},"basicAuthentication":{"destinations":[{"url":"%s","name":"%s"}]}}},"destinations":[{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authentication":"%s"}]}`, destinationURL, samlAssertionDestName, destinationURL, clientCertAuthDestName, destinationURL, basicDestName, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)),
	)

	destsConfigWithClientCertauthCertDataRawJSON = json.RawMessage(
		fmt.Sprintf(`{"credentials":{"inboundCommunication":{"clientCertificateAuthentication":{"certificate":"cert-chain-data","destinations":[{"url":"%s","name":"%s"}]},"basicAuthentication":{"destinations":[{"url":"%s","name":"%s"}]}}},"destinations":[{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authentication":"%s"}]}`, destinationURL, clientCertAuthDestName, destinationURL, basicDestName, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)),
	)

	fa                           = fixFormationAssignmentWithConfig(destsConfigValueRawJSON)
	reverseFa                    = fixFormationAssignmentWithConfig(destsReverseConfigValueRawJSON)
	faWithInitialState           = fixFormationAssignmentWithState(model.InitialAssignmentState)
	faWithInvalidConfig          = fixFormationAssignmentWithConfig(invalidFAConfig)
	faWithSAMLCertData           = fixFormationAssignmentWithConfig(destsConfigWithSAMLCertDataRawJSON)
	faWithClientCertAuthCertData = fixFormationAssignmentWithConfig(destsConfigWithClientCertauthCertDataRawJSON)

	faConfigWithDifferentStructure = fixFormationAssignmentWithConfig(configWithDifferentStructure)

	inputForUnassignNotificationStatusReturned = fixDestinationCreatorInputForUnassignWithLocationOperation(model.NotificationStatusReturned)
	inputForUnassignSendNotification           = fixDestinationCreatorInputForUnassignWithLocationOperation(model.SendNotificationOperation)

	inputForAssignWithFormationAssignmentInitialState = &formationconstraintpkg.DestinationCreatorInput{
		Operation:                       model.AssignFormation,
		JoinPointDetailsFAMemoryAddress: faWithInitialState.GetAddress(),
	}

	preNotificationStatusReturnedLocation = fixJoinPointLocation(model.NotificationStatusReturned, model.PreOperation)
	preSendNotificationLocation           = fixJoinPointLocation(model.SendNotificationOperation, model.PreOperation)
	preGenerateFANotificationLocation     = fixJoinPointLocation(model.GenerateFormationAssignmentNotificationOperation, model.PreOperation)

	inputWithAssignmentWithSAMLCertData                                 = fixDestinationCreatorInputWithAssignmentMemoryAddress(model.AssignFormation, faWithSAMLCertData, preNotificationStatusReturnedLocation)
	inputWithAssignmentWithClientCertAuthCertData                       = fixDestinationCreatorInputWithAssignmentMemoryAddress(model.AssignFormation, faWithClientCertAuthCertData, preNotificationStatusReturnedLocation)
	inputForAssignNotificationStatusReturned                            = fixDestinationCreatorInputWithAssignmentMemoryAddress(model.AssignFormation, fa, preNotificationStatusReturnedLocation)
	inputForAssignNotificationStatusReturnedWithCertSvcKeyStore         = fixDestinationCreatorInputWithAssignmentMemoryAddressAndCertSvcKeystore(model.AssignFormation, fa, preNotificationStatusReturnedLocation, true)
	inputForAssignNotificationStatusReturnedWithInvalidFAConfig         = fixDestinationCreatorInputWithAssignmentMemoryAddress(model.AssignFormation, faWithInvalidConfig, preNotificationStatusReturnedLocation)
	inputForAssignSendNotificationWithoutReverseAssignmentMemoryAddress = fixDestinationCreatorInputWithAssignmentMemoryAddress(model.AssignFormation, fa, preSendNotificationLocation)

	inputForAssignSendNotification                                         = fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, fa, reverseFa, preSendNotificationLocation)
	inputForAssignSendNotificationWithInvalidFAConfig                      = fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, faWithInvalidConfig, reverseFa, preSendNotificationLocation)
	inputForAssignSendNotificationWithInvalidReverseFAConfig               = fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, fa, faWithInvalidConfig, preSendNotificationLocation)
	inputForAssignSendNotificationWhereFAConfigStructureIsDifferent        = fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, faConfigWithDifferentStructure, faConfigWithDifferentStructure, preSendNotificationLocation)
	inputForAssignSendNotificationWhereReverseFAConfigStructureIsDifferent = fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, fa, faConfigWithDifferentStructure, preSendNotificationLocation)
	inputForAssignGenerateFANotification                                   = fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(model.AssignFormation, fa, reverseFa, preGenerateFANotificationLocation)

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

	// func TestConstraintOperators_IsNotAssignedToAnyFormationOfType
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

	// func TestConstraintOperators_DoNotGenerateFormationAssignmentNotification
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

	scenariosLabel             = &model.Label{Value: []interface{}{scenario}}
	scenariosLabelInvalidValue = &model.Label{Value: "invalid"}

	formations = []*model.Formation{
		{
			FormationTemplateID: otherFormationTemplateID,
		},
	}

	formations2 = []*model.Formation{
		{
			FormationTemplateID: formationTemplateID,
		},
	}

	emptyAssignments = []*model.AutomaticScenarioAssignment{}

	assignments = []*model.AutomaticScenarioAssignment{
		{ScenarioName: scenario},
	}

	location = formationconstraintpkg.JoinPointLocation{
		OperationName:  "assign",
		ConstraintType: "pre",
	}

	details = formationconstraintpkg.AssignFormationOperationDetails{
		ResourceType:    "runtime",
		ResourceSubtype: "kyma",
	}
)

func fixDestinationCreatorInputWithAssignmentMemoryAddress(operation model.FormationOperation, formationAssignment *model.FormationAssignment, location formationconstraintpkg.JoinPointLocation) *formationconstraintpkg.DestinationCreatorInput {
	return &formationconstraintpkg.DestinationCreatorInput{
		Operation:                       operation,
		JoinPointDetailsFAMemoryAddress: formationAssignment.GetAddress(),
		Location:                        location,
	}
}

func fixDestinationCreatorInputWithAssignmentAndReverseFAMemoryAddress(operation model.FormationOperation, assignment, reverseAssignment *model.FormationAssignment, location formationconstraintpkg.JoinPointLocation) *formationconstraintpkg.DestinationCreatorInput {
	return &formationconstraintpkg.DestinationCreatorInput{
		Operation:                              operation,
		JoinPointDetailsFAMemoryAddress:        assignment.GetAddress(),
		JoinPointDetailsReverseFAMemoryAddress: reverseAssignment.GetAddress(),
		Location:                               location,
	}
}

func fixDestinationCreatorInputForUnassignWithLocationOperation(operationName model.TargetOperation) *formationconstraintpkg.DestinationCreatorInput {
	return &formationconstraintpkg.DestinationCreatorInput{
		Operation:                       model.UnassignFormation,
		JoinPointDetailsFAMemoryAddress: fa.GetAddress(),
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  operationName,
			ConstraintType: model.PreOperation,
		},
	}
}

func fixDestinationCreatorInputWithAssignmentMemoryAddressAndCertSvcKeystore(operation model.FormationOperation, formationAssignment *model.FormationAssignment, location formationconstraintpkg.JoinPointLocation, useCertSvcKeystoreForSAML bool) *formationconstraintpkg.DestinationCreatorInput {
	return &formationconstraintpkg.DestinationCreatorInput{
		Operation:                       operation,
		JoinPointDetailsFAMemoryAddress: formationAssignment.GetAddress(),
		Location:                        location,
		UseCertSvcKeystoreForSAML:       useCertSvcKeystoreForSAML,
	}
}

func fixJoinPointLocation(operationName model.TargetOperation, constraintType model.FormationConstraintType) formationconstraintpkg.JoinPointLocation {
	return formationconstraintpkg.JoinPointLocation{
		OperationName:  operationName,
		ConstraintType: constraintType,
	}
}

func fixFormationAssignmentWithConfig(config json.RawMessage) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:    formationAssignmentID,
		State: string(model.ReadyAssignmentState),
		Value: config,
	}
}

func fixFormationAssignmentWithState(state model.FormationAssignmentState) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:    formationAssignmentID,
		State: string(state),
	}
}

func fixDesignTimeDestination() operators.Destination {
	return operators.Destination{
		Name:           designTimeDestName,
		Type:           string(destinationType),
		ProxyType:      string(destinationProxyType),
		Authentication: string(destinationNoAuthn),
		URL:            destinationURL,
	}
}

func fixDesignTimeDestinations() []operators.Destination {
	return []operators.Destination{
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

func fixCertificateData() *operators.CertificateData {
	return &operators.CertificateData{
		FileName:         testFileName,
		CommonName:       testCommonName,
		CertificateChain: testCertChain,
	}
}

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

func unusedDestinationCreatorService() *automock.DestinationCreatorService {
	return &automock.DestinationCreatorService{}
}
