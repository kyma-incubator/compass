package operators

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/domain/destination/destinationcreator"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"net/http"
)

const (
	// DestinationCreatorOperator represents the destination creator operator
	DestinationCreatorOperator = "DestinationCreator"
)

// NewDestinationCreatorInput is input constructor for DestinationCreatorOperator. It returns empty OperatorInput
func NewDestinationCreatorInput() OperatorInput {
	return &formationconstraint.DestinationCreatorInput{}
}

// DestinationCreator is an operator that handles destination creations
func (e *ConstraintEngine) DestinationCreator(ctx context.Context, input OperatorInput) (bool, error) {
	log.C(ctx).Infof("Executing operator: %q", DestinationCreatorOperator)

	di, ok := input.(*formationconstraint.DestinationCreatorInput)
	if !ok {
		return false, errors.Errorf("Incompatible input for operator: %q", DestinationCreatorOperator)
	}

	if di.Operation != model.AssignFormation && di.Operation != model.UnassignFormation {
		return false, errors.Errorf("The formation operation is invalid: %q. It should be either %q or %q", di.Operation, model.AssignFormation, model.UnassignFormation)
	}

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q and subtype: %q during %q operation", di.ResourceType, di.ResourceSubtype, di.Operation)

	if di.Operation == model.UnassignFormation && di.Location.OperationName == model.SendNotificationOperation {
		if err := e.destinationSvc.DeleteDestinations(ctx, di.FormationAssignment); err != nil {
			return false, err
		}
		return true, nil
	}

	if di.FormationAssignment.State != string(model.ReadyAssignmentState) && di.FormationAssignment.State != string(model.ConfigPendingAssignmentState) {
		log.C(ctx).Warnf("The formation assignment with ID: %q has state: %q and no destination(s) will be created because of it", di.FormationAssignment.ID, di.FormationAssignment.State)
		return true, nil
	}

	if di.FormationAssignment != nil && di.FormationAssignment.Value != "" && di.ReverseFormationAssignment != nil && di.ReverseFormationAssignment.Value == "" && di.Location.OperationName == model.NotificationStatusReturned {
		log.C(ctx).Infof("Location with constraint type: %q and operation name: %q is reached", di.Location.ConstraintType, di.Location.OperationName)

		var confDetails Configuration
		if err := json.Unmarshal([]byte(di.FormationAssignment.Value), &confDetails); err != nil {
			return false, errors.Wrapf(err, "while unmarshaling tenant mapping response configuration from assignment with ID: %q", di.FormationAssignment.ID)
		}

		log.C(ctx).Infof("There is/are %d design time destination(s) available in the configuration response", len(confDetails.Destinations))
		for _, destDetails := range confDetails.Destinations {
			statusCode, err := e.destinationSvc.CreateDesignTimeDestinations(ctx, destDetails, di.FormationAssignment)
			if err != nil {
				return false, errors.Wrapf(err, "while creating destination with name: %q", destDetails.Name)
			}

			if statusCode == http.StatusConflict {
				log.C(ctx).Infof("The destination with name: %q already exists. Will be deleted and created again...", destDetails.Name)
				if err := e.destinationSvc.DeleteDestinations(ctx, di.FormationAssignment); err != nil {
					return false, errors.Wrapf(err, "while deleting destination with name: %q", destDetails.Name)
				}

				if _, err = e.destinationSvc.CreateDesignTimeDestinations(ctx, destDetails, di.FormationAssignment); err != nil {
					return false, errors.Wrapf(err, "while creating destination with name: %q", destDetails.Name)
				}
			}
		}

		samlAuthDetails := confDetails.Credentials.InboundCommunicationDetails.SAMLAssertionAuthenticationDetails
		if isSAMLDetailsExists := samlAuthDetails; isSAMLDetailsExists != nil {
			log.C(ctx).Infof("There is/are %d SAML Bearer destination details in the configuration response", len(samlAuthDetails.Destinations))
			for _, _ = range samlAuthDetails.Destinations {
				// todo:: (for phase 2) create KeyStore for every destination elements and enrich the destination before sending it to the Destination Creator component
			}
		}

		return true, nil
	}

	if di.FormationAssignment != nil && di.FormationAssignment.Value != "" && di.ReverseFormationAssignment != nil && di.ReverseFormationAssignment.Value != "" && di.Location.OperationName == model.SendNotificationOperation {
		log.C(ctx).Infof("Location with constraint type: %q and operation name: %q is reached", di.Location.ConstraintType, di.Location.OperationName)

		var confDetails Configuration
		if err := json.Unmarshal([]byte(di.FormationAssignment.Value), &confDetails); err != nil {
			return false, errors.Wrapf(err, "while unmarshaling tenant mapping configuration response from assignment with ID: %q", di.FormationAssignment.ID)
		}

		var confCredentials Configuration
		if err := json.Unmarshal([]byte(di.ReverseFormationAssignment.Value), &confCredentials); err != nil {
			return false, errors.Wrapf(err, "while unmarshaling tenant mapping configuration response from reverse assignment with ID: %q", di.ReverseFormationAssignment.ID)
		}

		if confDetails.Credentials.InboundCommunicationDetails == nil {
			return false, errors.New("The inbound communication destination details could not be empty")
		}

		if confCredentials.Credentials.InboundCommunicationCredentials == nil {
			return false, errors.New("The inbound communication credentials could not be empty")
		}

		basicAuthDetails := confDetails.Credentials.InboundCommunicationDetails.BasicAuthenticationDetails
		basicAuthCreds := confCredentials.Credentials.InboundCommunicationCredentials.BasicAuthentication
		if basicAuthDetails != nil && basicAuthCreds != nil {
			log.C(ctx).Infof("There is/are %d inbound basic destination(s) available in the configuration", len(basicAuthDetails.Destinations))
			for _, destDetails := range basicAuthDetails.Destinations {
				statusCode, err := e.destinationSvc.CreateBasicCredentialDestinations(ctx, destDetails, *basicAuthCreds, di.FormationAssignment)
				if err != nil {
					return false, errors.Wrapf(err, "while creating inbound basic destination with name: %q", destDetails.Name)
				}

				if statusCode == http.StatusConflict {
					log.C(ctx).Infof("The destination with name: %q already exists. Will be deleted and created again...", destDetails.Name)
					if err := e.destinationSvc.DeleteDestinationFromDestinationService(ctx, destDetails.Name, destDetails.SubaccountID, di.FormationAssignment); err != nil {
						return false, errors.Wrapf(err, "while deleting inbound basic destination with name: %q", destDetails.Name)
					}

					if _, err = e.destinationSvc.CreateBasicCredentialDestinations(ctx, destDetails, *basicAuthCreds, di.FormationAssignment); err != nil {
						return false, errors.Wrapf(err, "while creating inbound basic destination with name: %q", destDetails.Name)
					}
				}
			}
		}

		// todo:: will be implemented with the second phase of the destination operator
		samlAuthDetails := confDetails.Credentials.InboundCommunicationDetails.SAMLAssertionAuthenticationDetails
		samlAuthCreds := confCredentials.Credentials.InboundCommunicationCredentials.SAMLAssertionAuthentication
		if samlAuthDetails != nil && samlAuthCreds != nil {
			log.C(ctx).Infof("There is/are %d inbound SAML destination(s) available in the configuration", len(basicAuthDetails.Destinations))
			// todo:: (for phase 2) create SAML destination with the data from the "SAML details" and credentials response
		}

		return true, nil
	}

	return true, nil
}

// Destination Creator Operator types

// Configuration todo::: add godoc
type Configuration struct {
	Destinations         []Destination        `json:"destinations"`
	Credentials          Credentials          `json:"credentials"`
	AdditionalProperties AdditionalProperties `json:"additionalProperties"`
}

// AdditionalProperties todo::: add godoc
type AdditionalProperties []json.RawMessage

// Credentials todo::: add godoc
type Credentials struct {
	OutboundCommunicationCredentials *OutboundCommunicationCredentials `json:"outboundCommunication,omitempty"`
	InboundCommunicationDetails      *InboundCommunicationDetails      `json:"inboundCommunication,omitempty"`
	InboundCommunicationCredentials  *InboundCommunicationCredentials  `json:"inboundCommunicationCredentials,omitempty"`
}

// InboundCommunicationCredentials todo::: add godoc
type InboundCommunicationCredentials struct {
	BasicAuthentication         *BasicAuthentication         `json:"basicAuthentication,omitempty"`
	SAMLAssertionAuthentication *SAMLAssertionAuthentication `json:"samlAssertionAuthentication,omitempty"` // todo::: double check the json tag/key
}

// OutboundCommunicationCredentials todo::: add godoc
type OutboundCommunicationCredentials struct {
	NoAuthentication                *NoAuthentication                `json:"noAuthentication,omitempty"`
	BasicAuthentication             *BasicAuthentication             `json:"basicAuthentication,omitempty"`
	ClientCredentialsAuthentication *ClientCredentialsAuthentication `json:"clientCredentialsAuthentication,omitempty"`
	ClientCertAuthentication        *ClientCertAuthentication        `json:"clientCertAuthentication,omitempty"`
}

// NoAuthentication todo::: add godoc
type NoAuthentication struct {
	URL            string   `json:"url"`
	UIURL          string   `json:"uiUrl,omitempty"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

// BasicAuthentication todo::: add godoc
type BasicAuthentication struct {
	URL            string   `json:"url"`
	UIURL          string   `json:"uiUrl,omitempty"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

// ClientCredentialsAuthentication todo::: add godoc
type ClientCredentialsAuthentication struct {
	URL             string   `json:"url"`
	UIURL           string   `json:"uiUrl,omitempty"`
	TokenServiceURL string   `json:"tokenServiceUrl"`
	ClientID        string   `json:"clientId"`
	ClientSecret    string   `json:"clientSecret"`
	CorrelationIds  []string `json:"correlationIds,omitempty"`
}

// ClientCertAuthentication todo::: add godoc
type ClientCertAuthentication struct {
	URL            string   `json:"url"`
	UIURL          string   `json:"uiUrl,omitempty"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

// SAMLAssertionAuthentication todo::: add godoc
type SAMLAssertionAuthentication struct { // todo::: check fields
	URL                  string `json:"url"`
	Audience             string `json:"audience"`
	TokenServiceURL      string `json:"tokenServiceURL"`
	TokenServiceURLType  string `json:"tokenServiceURLType"`
	AuthnContextClassRef string `json:"authnContextClassRef"`
}

// InboundCommunicationDetails todo::: add godoc
type InboundCommunicationDetails struct {
	BasicAuthenticationDetails         *InboundBasicAuthenticationDetails         `json:"basicAuthentication,omitempty"`
	SAMLAssertionAuthenticationDetails *InboundSAMLAssertionAuthenticationDetails `json:"samlAssertionAuthentication,omitempty"` // todo::: double check the json tag/key
}

// InboundBasicAuthenticationDetails todo::: add godoc
type InboundBasicAuthenticationDetails struct {
	Destinations []Destination `json:"destinations"`
}

// InboundSAMLAssertionAuthenticationDetails todo::: add godoc
type InboundSAMLAssertionAuthenticationDetails struct {
	Destinations []Destination `json:"destinations"`
}

// Destination todo::: add godoc
type Destination struct {
	Name                 string                       `json:"name"`
	Type                 destinationcreator.Type      `json:"type"`
	Description          string                       `json:"description,omitempty"`
	ProxyType            destinationcreator.ProxyType `json:"proxyType"`
	Authentication       destinationcreator.AuthType  `json:"authentication"`
	URL                  string                       `json:"url"`
	SubaccountID         string                       `json:"subaccountId,omitempty"`
	AdditionalAttributes json.RawMessage              `json:"additionalAttributes,omitempty"`
	// todo:: additional fields for KeyStore(phase 2)
}
