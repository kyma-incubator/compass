package operators

import (
	"context"
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	// DestinationCreatorOperator represents the destination creator operator
	DestinationCreatorOperator = "DestinationCreatorOperator"
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

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q, subtype: %q and ID: %q during %q operation", di.ResourceType, di.ResourceSubtype, di.ResourceID, di.Operation)

	if di.Operation == model.AssignFormation {
		if di.Assignment.Value != "" && di.JoinPointLocation == formationconstraint.PostNotificationStatusReturned {
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

		if di.Assignment.Value != "" && di.ReverseAssignment.Value != "" && di.JoinPointLocation == formationconstraint.PreSendNotification {
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

// Destination Creator types

// ConfigurationResponse todo::: add godoc
type ConfigurationResponse struct {
	State         string        `json:"state"`
	Configuration Configuration `json:"configuration"`
}

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
	OutboundCommunicationCredentials OutboundCommunicationCredentials `json:"outboundCommunication"`
	InboundCommunicationDetails      InboundCommunicationDetails      `json:"inboundCommunication"`
	InboundCommunicationCredentials  InboundCommunicationCredentials  `json:"inboundCommunicationCredentials"`
}

// InboundCommunicationCredentials todo::: add godoc
type InboundCommunicationCredentials struct {
	BasicAuthentication      *BasicAuthentication `json:"basicAuthentication"`
	SAMLBearerAuthentication *SAMLAuthentication  `json:"samlBearerAuthentication"`
}

// OutboundCommunicationCredentials todo::: add godoc
type OutboundCommunicationCredentials struct {
	NoAuthentication                NoAuthentication                `json:"noAuthentication"`
	BasicAuthentication             BasicAuthentication             `json:"basicAuthentication"`
	ClientCredentialsAuthentication ClientCredentialsAuthentication `json:"clientCredentialsAuthentication"`
	ClientCertAuthentication        ClientCertAuthentication        `json:"clientCertAuthentication"`
}

// NoAuthentication todo::: add godoc
type NoAuthentication struct {
	Url            string   `json:"url"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

// BasicAuthentication todo::: add godoc
type BasicAuthentication struct {
	Authentication string   `json:"authentication"`
	Url            string   `json:"url"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

// ClientCredentialsAuthentication todo::: add godoc
type ClientCredentialsAuthentication struct {
	Url             string   `json:"url"`
	TokenServiceUrl string   `json:"tokenServiceUrl"`
	ClientId        string   `json:"clientId"`
	ClientSecret    string   `json:"clientSecret"`
	CorrelationIds  []string `json:"correlationIds,omitempty"`
}

// ClientCertAuthentication todo::: add godoc
type ClientCertAuthentication struct {
	Url            string   `json:"url"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

// SAMLAuthentication todo::: add godoc
type SAMLAuthentication struct {
	Authentication      string `json:"authentication"`
	URL                 string `json:"url"`
	Audience            string `json:"audience"`
	TokenServiceURL     string `json:"tokenServiceURL"`
	TokenServiceURLType string `json:"tokenServiceURLType"`
}

// InboundCommunicationDetails todo::: add godoc
type InboundCommunicationDetails struct {
	BasicAuthenticationDetails      *InboundBasicAuthenticationDetails      `json:"basicAuthentication"`
	SAMLBearerAuthenticationDetails *InboundSamlBearerAuthenticationDetails `json:"samlBearerAuthentication"`
}

// InboundBasicAuthenticationDetails todo::: add godoc
type InboundBasicAuthenticationDetails struct {
	Destinations []Destination `json:"destinations"`
}

// InboundSamlBearerAuthenticationDetails todo::: add godoc
type InboundSamlBearerAuthenticationDetails struct {
	Destinations []Destination `json:"destinations"`
}

// Destination todo::: add godoc
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

// AdditionalAttributes todo::: add godoc
type AdditionalAttributes json.RawMessage
