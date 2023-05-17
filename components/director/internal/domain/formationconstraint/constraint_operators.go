package formationconstraint

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	formationconstraintpkg "github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	// IsNotAssignedToAnyFormationOfTypeOperator represents the IsNotAssignedToAnyFormationOfType operator
	IsNotAssignedToAnyFormationOfTypeOperator = "IsNotAssignedToAnyFormationOfType"
	// DoesNotContainResourceOfSubtypeOperator represents the DoesNotContainResourceOfSubtype operator
	DoesNotContainResourceOfSubtypeOperator = "DoesNotContainResourceOfSubtype"
	// DestinationCreatorOperator represents the destination creator operator
	DestinationCreatorOperator = "DestinationCreatorOperator"
)

// OperatorName represents the constraint operator name
type OperatorName string

// OperatorInput represents the input needed by the constraint operator
type OperatorInput interface{}

// OperatorFunc provides an interface for functions implementing constraint operators
type OperatorFunc func(ctx context.Context, input OperatorInput) (bool, error)

// OperatorInputConstructor returns empty OperatorInput for a certain constraint operator
type OperatorInputConstructor func() OperatorInput

// NewIsNotAssignedToAnyFormationOfTypeInput is input constructor for IsNotAssignedToAnyFormationOfType operator. It returns empty OperatorInput.
func NewIsNotAssignedToAnyFormationOfTypeInput() OperatorInput {
	return &formationconstraint.IsNotAssignedToAnyFormationOfTypeInput{}
}

// NewDoesNotContainResourceOfSubtypeInput is input constructor for DoesNotContainResourceOfSubtypeOperator operator. It returns empty OperatorInput
func NewDoesNotContainResourceOfSubtypeInput() OperatorInput {
	return &formationconstraint.DoesNotContainResourceOfSubtypeInput{}
}

// NewDestinationCreatorInput is input constructor for DestinationCreatorOperator. It returns empty OperatorInput
func NewDestinationCreatorInput() OperatorInput {
	return &formationconstraint.DestinationCreatorInput{}
}

// IsNotAssignedToAnyFormationOfType is a constraint operator. It checks if the resource from the OperatorInput is already part of formation of the type that the operator is associated with
func (e *ConstraintEngine) IsNotAssignedToAnyFormationOfType(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: %s", IsNotAssignedToAnyFormationOfTypeOperator)

	i, ok := input.(*formationconstraint.IsNotAssignedToAnyFormationOfTypeInput)
	if !ok {
		return false, errors.New("Incompatible input")
	}

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q, subtype: %q and ID: %q", i.ResourceType, i.ResourceSubtype, i.ResourceID)

	var assignedFormations []string
	switch i.ResourceType {
	case model.TenantResourceType:
		tenantInternalID, err := e.tenantSvc.GetInternalTenant(ctx, i.ResourceID)
		if err != nil {
			return false, err
		}

		assignments, err := e.asaSvc.ListForTargetTenant(ctx, tenantInternalID)
		if err != nil {
			return false, err
		}

		assignedFormations = make([]string, 0, len(assignments))
		for _, a := range assignments {
			assignedFormations = append(assignedFormations, a.ScenarioName)
		}
	case model.ApplicationResourceType:
		scenariosLabel, err := e.labelRepo.GetByKey(ctx, i.Tenant, model.ApplicationLabelableObject, i.ResourceID, model.ScenariosKey)
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				return true, nil
			}
			return false, err
		}
		assignedFormations, err = label.ValueToStringsSlice(scenariosLabel.Value)
		if err != nil {
			return false, err
		}
	default:
		return false, errors.Errorf("Unsupported resource type %q", i.ResourceType)
	}

	isAllowedToParticipateInFormationsOfType, err := e.isAllowedToParticipateInFormationsOfType(ctx, assignedFormations, i.Tenant, i.FormationTemplateID, i.ResourceSubtype, i.ExceptSystemTypes)
	if err != nil {
		return false, err
	}

	return isAllowedToParticipateInFormationsOfType, nil
}

// DoesNotContainResourceOfSubtype is a constraint operator. It checks if the formation contains resource with the same subtype as the resource subtype from the OperatorInput
func (e *ConstraintEngine) DoesNotContainResourceOfSubtype(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: %q", DoesNotContainResourceOfSubtypeOperator)

	i, ok := input.(*formationconstraint.DoesNotContainResourceOfSubtypeInput)
	if !ok {
		return false, errors.New(fmt.Sprintf("Incompatible input for operator %q", DoesNotContainResourceOfSubtypeOperator))
	}

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q, subtype: %q and ID: %q", i.ResourceType, i.ResourceSubtype, i.ResourceID)

	switch i.ResourceType {
	case model.ApplicationResourceType:
		applications, err := e.applicationRepository.ListByScenariosNoPaging(ctx, i.Tenant, []string{i.FormationName})
		if err != nil {
			return false, errors.Wrap(err, fmt.Sprintf("while listing applications in scenario %q", i.FormationName))
		}

		for _, application := range applications {
			appTypeLbl, err := e.labelService.GetByKey(ctx, i.Tenant, model.ApplicationLabelableObject, application.ID, i.ResourceTypeLabelKey)
			if err != nil {
				return false, errors.Wrap(err, fmt.Sprintf("while getting label with key %q of application with ID %q in tenant %q", i.ResourceTypeLabelKey, application.ID, i.Tenant))
			}

			if i.ResourceSubtype == appTypeLbl.Value.(string) {
				return false, nil
			}
		}
	default:
		return false, errors.Errorf("Unsupported resource type %q", i.ResourceType)
	}

	return true, nil
}

// DestinationCreator is an operator that handles destination creations
func (e *ConstraintEngine) DestinationCreator(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: %q", DestinationCreatorOperator)

	di, ok := input.(*formationconstraint.DestinationCreatorInput)
	if !ok {
		return false, errors.Errorf("Incompatible input for operator: %q", DestinationCreatorOperator)
	}

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q, subtype: %q and ID: %q during %q operation", di.ResourceType, di.ResourceSubtype, di.ResourceID, di.Operation)

	if di.Operation == model.AssignFormation {
		if di.Assignment.Value != "" && di.JoinPointLocation == formationconstraintpkg.PostNotificationStatusReturned {
			log.C(ctx).Infof("Location with constraint type: %q and operation name: %q is reached", model.NotificationStatusReturned, model.PostOperation)

			var confDetailsResp ConfigurationResponse
			if err := json.Unmarshal([]byte(di.Assignment.Value), &confDetailsResp); err != nil {
				return false, errors.Wrapf(err, "while unmarshaling tenant mapping configuration response from assignment with ID: %q", di.Assignment.ID)
			}

			designTimeDestLength := len(confDetailsResp.Configuration.Destinations)
			if designTimeDestLength > 0 {
				log.C(ctx).Infof("There is/are %d design time destination(s) available in the configuration response", designTimeDestLength)
				// create design time destination
				// mtlsClient containing our client cert --> k8s client, get secret from our cluster
			}

			// todo:: will be implemented with the second phase of the destination operator
			samlAuthDetails := confDetailsResp.Configuration.Credentials.InboundCommunicationDetails.SAMLBearerAuthenticationDetails
			if isSAMLDetailsExists := samlAuthDetails; isSAMLDetailsExists != nil {
				log.C(ctx).Infof("There is/are %d SAML Bearer destination details in the configuration response", designTimeDestLength)
				for _, _ = range samlAuthDetails.Destinations {
					// todo:: (for phase 2) create KeyStore for every destination elements and enrich the destination before sending it to the Destination Creator component
				}
			}

			return true, nil
		}

		if di.Assignment.Value != "" && di.ReverseAssignment.Value != "" && di.JoinPointLocation == formationconstraintpkg.PreSendNotification {
			log.C(ctx).Infof("Location with constraint type: %q and operation name: %q is reached", model.SendNotificationOperation, model.PreOperation)
			var confDetailsResp ConfigurationResponse
			if err := json.Unmarshal([]byte(di.Assignment.Value), &confDetailsResp); err != nil {
				return false, errors.Wrapf(err, "while unmarshaling tenant mapping configuration response from assignment with ID: %q", di.Assignment.ID)
			}

			var confCredentialsResp ConfigurationResponse
			if err := json.Unmarshal([]byte(di.ReverseAssignment.Value), &confCredentialsResp); err != nil {
				return false, errors.Wrapf(err, "while unmarshaling tenant mapping configuration response from reverse assignment with ID: %q", di.ReverseAssignment.ID)
			}

			basicAuthDetails := confDetailsResp.Configuration.Credentials.InboundCommunicationDetails.BasicAuthenticationDetails
			basicAuthCreds := confCredentialsResp.Configuration.Credentials.InboundCommunicationCredentials.BasicAuthentication
			if basicAuthDetails != nil && basicAuthCreds != nil {
				// loop through dest details and build a req based on: dest details && returned basic inboundCommCreds from ext svc

				for _, destination := range basicAuthDetails.Destinations {
					// build request using 'destination details' --> send the request to destination creator -> create destination with the credentials from the reverseAssignment
					log.C(ctx).Info(destination) // todo::: delete
				}
			}

			// todo:: will be implemented with the second phase of the destination operator
			samlAuthDetails := confDetailsResp.Configuration.Credentials.InboundCommunicationDetails.SAMLBearerAuthenticationDetails
			samlAuthCreds := confCredentialsResp.Configuration.Credentials.InboundCommunicationCredentials.SAMLBearerAuthentication
			if samlAuthDetails != nil && samlAuthCreds != nil {
				// todo:: (for phase 2) create SAML destination with the data from the "SAML details" and credentials response
			}

			return true, nil
		}
	} else if di.Operation == model.UnassignFormation {

	} else {
		return false, nil // todo::: consider validate func
	}

	return true, nil
}

func (e *ConstraintEngine) isAllowedToParticipateInFormationsOfType(ctx context.Context, assignedFormationNames []string, tenant, formationTemplateID, resourceSubtype string, exceptSystemTypes []string) (bool, error) {
	if len(assignedFormationNames) == 0 {
		return true, nil
	}

	if len(exceptSystemTypes) > 0 {
		for _, exceptType := range exceptSystemTypes {
			if resourceSubtype == exceptType {
				return true, nil
			}
		}
	}

	assignedFormations, err := e.formationRepo.ListByFormationNames(ctx, assignedFormationNames, tenant)
	if err != nil {
		return false, err
	}

	for _, formation := range assignedFormations {
		if formation.FormationTemplateID == formationTemplateID {
			return false, nil
		}
	}

	return true, nil
}

type ConfigurationResponse struct {
	State         string        `json:"state"`
	Configuration Configuration `json:"configuration"`
}

type Configuration struct {
	Destinations         []Destination        `json:"destinations"`
	Credentials          Credentials          `json:"credentials"`
	AdditionalProperties AdditionalProperties `json:"additionalProperties"`
}

type AdditionalProperties []json.RawMessage

type Credentials struct {
	OutboundCommunicationCredentials OutboundCommunicationCredentials `json:"outboundCommunication"`
	InboundCommunicationDetails      InboundCommunicationDetails      `json:"inboundCommunication"`
	InboundCommunicationCredentials  InboundCommunicationCredentials  `json:"inboundCommunicationCredentials"`
}

type InboundCommunicationCredentials struct {
	BasicAuthentication      *BasicAuthentication `json:"basicAuthentication"`
	SAMLBearerAuthentication *SAMLAuthentication  `json:"samlBearerAuthentication"`
}

type OutboundCommunicationCredentials struct {
	NoAuthentication                NoAuthentication                `json:"noAuthentication"`
	BasicAuthentication             BasicAuthentication             `json:"basicAuthentication"`
	ClientCredentialsAuthentication ClientCredentialsAuthentication `json:"clientCredentialsAuthentication"`
	ClientCertAuthentication        ClientCertAuthentication        `json:"clientCertAuthentication"`
}

type NoAuthentication struct {
	Url            string   `json:"url"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

type BasicAuthentication struct {
	Authentication string   `json:"authentication"`
	Url            string   `json:"url"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

type ClientCredentialsAuthentication struct {
	Url             string   `json:"url"`
	TokenServiceUrl string   `json:"tokenServiceUrl"`
	ClientId        string   `json:"clientId"`
	ClientSecret    string   `json:"clientSecret"`
	CorrelationIds  []string `json:"correlationIds,omitempty"`
}

type ClientCertAuthentication struct {
	Url            string   `json:"url"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

type SAMLAuthentication struct {
	Authentication      string `json:"authentication"`
	URL                 string `json:"url"`
	Audience            string `json:"audience"`
	TokenServiceURL     string `json:"tokenServiceURL"`
	TokenServiceURLType string `json:"tokenServiceURLType"`
}

type InboundCommunicationDetails struct {
	BasicAuthenticationDetails      *InboundBasicAuthenticationDetails      `json:"basicAuthentication"`
	SAMLBearerAuthenticationDetails *InboundSamlBearerAuthenticationDetails `json:"samlBearerAuthentication"`
}

type InboundBasicAuthenticationDetails struct {
	Destinations []Destination `json:"destinations"`
}

type InboundSamlBearerAuthenticationDetails struct {
	Destinations []Destination `json:"destinations"`
}

type Destination struct {
	Name                 string               `json:"name"`
	Type                 string               `json:"type"`
	Description          string               `json:"description,omitempty"`
	ProxyType            string               `json:"proxyType"`
	Authentication       string               `json:"authentication"`
	Url                  string               `json:"url"`
	SubaccountId         string               `json:"subaccountId,omitempty"`
	AdditionalAttributes AdditionalAttributes `json:"additionalAttributes,omitempty"`
	// todo::: additional fields for KeyStore(phase 2)
}

type AdditionalAttributes json.RawMessage
