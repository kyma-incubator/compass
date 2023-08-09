package operators_test

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/destinationcreator"

	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/pkg/errors"
)

const (
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

	testFileName   = "test-file-name"
	testCommonName = "test-common-name"
	testCertChain  = "test-cert-chain"

	designTimeDestName    = "design-time-name"
	destinationURL        = "http://test-url"
	destinationType       = destinationcreator.TypeHTTP
	destinationProxyType  = destinationcreator.ProxyTypeInternet
	destinationNoAuthn    = destinationcreator.AuthTypeNoAuth
	basicDestName         = "name-basic"
	basicUser             = "user"
	basicPassword         = "pwd"
	samlAssertionDestName = "saml-assertion-name"
	testJSONConfig        = `{"key": "val"}`
)

var (
	ctx            = context.TODO()
	testErr        = errors.New("test error")
	corrleationIDs []string

	invalidFAConfig              = json.RawMessage("invalid-config")
	configWithDifferentStructure = json.RawMessage(testJSONConfig)
	destsConfigValueRawJSON      = json.RawMessage(
		fmt.Sprintf(`{"credentials":{"inboundCommunication":{"samlAssertion":{"destinations":[{"url":"%s","name":"%s"}]},"basicAuthentication":{"destinations":[{"url":"%s","name":"%s"}]}}},"destinations":[{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authentication":"%s"}]}`, destinationURL, samlAssertionDestName, destinationURL, basicDestName, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)),
	)
	destsReverseConfigValueRawJSON = json.RawMessage(
		fmt.Sprintf(`{"credentials":{"inboundCommunication":{"samlAssertion":{"destinations":[{"url":"%s","name":"%s"}]},"basicAuthentication":{"destinations":[{"url":"%s","name":"%s"}]}},"outboundCommunication":{"basicAuthentication":{"url":"%s","username":"%s","password":"%s"},"samlAssertion":{"url":"%s"}}},"destinations":[{"url":"%s","name":"%s","type":"%s","proxyType":"%s","authentication":"%s"}]}`, destinationURL, samlAssertionDestName, destinationURL, basicDestName, destinationURL, basicUser, basicPassword, destinationURL, destinationURL, designTimeDestName, string(destinationType), string(destinationProxyType), string(destinationNoAuthn)),
	)
	fa                  = fixFormationAssignmentWithConfig(destsConfigValueRawJSON)
	reverseFa           = fixFormationAssignmentWithConfig(destsReverseConfigValueRawJSON)
	faWithDeletingState = fixFormationAssignmentWithState(model.DeletingAssignmentState)
	faWithInvalidConfig = fixFormationAssignmentWithConfig(invalidFAConfig)

	faConfigWithDifferentStructure = fixFormationAssignmentWithConfig(configWithDifferentStructure)

	inputForUnassignNotificationStatusReturned = fixDestinationCreatorInputForUnassignWithLocationOperation(model.NotificationStatusReturned)
	inputForUnassignSendNotification           = fixDestinationCreatorInputForUnassignWithLocationOperation(model.SendNotificationOperation)

	inputForAssignWithFormationAssignmentDeletingState = &formationconstraintpkg.DestinationCreatorInput{
		Operation:                       model.AssignFormation,
		JoinPointDetailsFAMemoryAddress: faWithDeletingState.GetAddress(),
	}

	inputForAssignNotificationStatusReturned = &formationconstraintpkg.DestinationCreatorInput{
		Operation: model.AssignFormation,
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  model.NotificationStatusReturned,
			ConstraintType: model.PreOperation,
		},
		JoinPointDetailsFAMemoryAddress: fa.GetAddress(),
	}

	inputForAssignSendNotification = &formationconstraintpkg.DestinationCreatorInput{
		Operation:                              model.AssignFormation,
		JoinPointDetailsFAMemoryAddress:        fa.GetAddress(),
		JoinPointDetailsReverseFAMemoryAddress: reverseFa.GetAddress(),
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  model.SendNotificationOperation,
			ConstraintType: model.PreOperation,
		},
	}

	inputWithInvalidOperation = &formationconstraintpkg.DestinationCreatorInput{
		Operation: model.CreateFormation,
	}

	inputForAssignNotificationStatusReturnedWithInvalidFAConfig = &formationconstraintpkg.DestinationCreatorInput{
		Operation:                       model.AssignFormation,
		JoinPointDetailsFAMemoryAddress: faWithInvalidConfig.GetAddress(),
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  model.NotificationStatusReturned,
			ConstraintType: model.PreOperation,
		},
	}

	inputForAssignSendNotificationWithInvalidFAConfig = &formationconstraintpkg.DestinationCreatorInput{
		Operation:                              model.AssignFormation,
		JoinPointDetailsFAMemoryAddress:        faWithInvalidConfig.GetAddress(),
		JoinPointDetailsReverseFAMemoryAddress: reverseFa.GetAddress(),
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  model.SendNotificationOperation,
			ConstraintType: model.PreOperation,
		},
	}

	inputForAssignSendNotificationWithInvalidReverseFAConfig = &formationconstraintpkg.DestinationCreatorInput{
		Operation:                              model.AssignFormation,
		JoinPointDetailsFAMemoryAddress:        fa.GetAddress(),
		JoinPointDetailsReverseFAMemoryAddress: faWithInvalidConfig.GetAddress(),
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  model.SendNotificationOperation,
			ConstraintType: model.PreOperation,
		},
	}

	inputForAssignSendNotificationWhereFAConfigStructureIsDifferent = &formationconstraintpkg.DestinationCreatorInput{
		Operation:                              model.AssignFormation,
		JoinPointDetailsFAMemoryAddress:        faConfigWithDifferentStructure.GetAddress(),
		JoinPointDetailsReverseFAMemoryAddress: faConfigWithDifferentStructure.GetAddress(),
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  model.SendNotificationOperation,
			ConstraintType: model.PreOperation,
		},
	}

	inputForAssignSendNotificationWhereReverseFAConfigStructureIsDifferent = &formationconstraintpkg.DestinationCreatorInput{
		Operation:                              model.AssignFormation,
		JoinPointDetailsFAMemoryAddress:        fa.GetAddress(),
		JoinPointDetailsReverseFAMemoryAddress: faConfigWithDifferentStructure.GetAddress(),
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  model.SendNotificationOperation,
			ConstraintType: model.PreOperation,
		},
	}

	inputWithoutAssignmentMemoryAddress = &formationconstraintpkg.DestinationCreatorInput{
		Operation: model.AssignFormation,
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  model.NotificationStatusReturned,
			ConstraintType: model.PreOperation,
		},
	}

	inputForAssignSendNotificationWithoutReverseAssignmentMemoryAddress = &formationconstraintpkg.DestinationCreatorInput{
		Operation:                       model.AssignFormation,
		JoinPointDetailsFAMemoryAddress: fa.GetAddress(),
		Location: formationconstraintpkg.JoinPointLocation{
			OperationName:  model.SendNotificationOperation,
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

func fixBasicDestination() operators.Destination {
	return operators.Destination{
		Name: basicDestName,
		URL:  destinationURL,
	}
}

func fixSAMLAssertionDestination() operators.Destination {
	return operators.Destination{
		Name: samlAssertionDestName,
		URL:  destinationURL,
	}
}

func fixBasicCreds() operators.BasicAuthentication {
	return operators.BasicAuthentication{
		URL:      destinationURL,
		Username: basicUser,
		Password: basicPassword,
	}
}

func fixSAMLCreds() *operators.SAMLAssertionAuthentication {
	return &operators.SAMLAssertionAuthentication{
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

func UnusedTenantService() *automock.TenantService {
	return &automock.TenantService{}
}

func UnusedASAService() *automock.AutomaticScenarioAssignmentService {
	return &automock.AutomaticScenarioAssignmentService{}
}

func UnusedLabelRepo() *automock.LabelRepository {
	return &automock.LabelRepository{}
}

func UnusedFormationRepo() *automock.FormationRepository {
	return &automock.FormationRepository{}
}

func UnusedApplicationRepo() *automock.ApplicationRepository {
	return &automock.ApplicationRepository{}
}

func UnusedLabelService() *automock.LabelService {
	return &automock.LabelService{}
}

func UnusedDestinationService() *automock.DestinationService {
	return &automock.DestinationService{}
}

func UnusedDestinationCreatorService() *automock.DestinationCreatorService {
	return &automock.DestinationCreatorService{}
}
