package operators

import (
	"context"
	"encoding/json"
	"runtime/debug"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
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
	log.C(ctx).Infof("Starting executing operator: %q", DestinationCreatorOperator)

	defer func() {
		if err := recover(); err != nil {
			log.C(ctx).WithField(logrus.ErrorKey, err).Panic("recovered panic")
			debug.PrintStack()
		}
	}()

	di, ok := input.(*formationconstraint.DestinationCreatorInput)
	if !ok {
		return false, errors.Errorf("Incompatible input for operator: %q", DestinationCreatorOperator)
	}

	if di.Operation != model.AssignFormation && di.Operation != model.UnassignFormation {
		return false, errors.Errorf("The formation operation is invalid: %q. It should be either %q or %q", di.Operation, model.AssignFormation, model.UnassignFormation)
	}

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q and subtype: %q during %q operation", di.ResourceType, di.ResourceSubtype, di.Operation)

	if di.Operation == model.UnassignFormation && di.Location.OperationName == model.NotificationStatusReturned && di.FormationAssignment != nil && di.FormationAssignment.State == string(model.ReadyAssignmentState) {
		log.C(ctx).Infof("Handling %s operation for formation assignment with ID: %q", model.UnassignFormation, di.FormationAssignment.ID)
		if err := e.destinationSvc.DeleteDestinations(ctx, di.FormationAssignment); err != nil {
			return false, err
		}

		return true, nil
	}

	if di.FormationAssignment.State != string(model.ReadyAssignmentState) && di.FormationAssignment.State != string(model.ConfigPendingAssignmentState) {
		log.C(ctx).Warnf("The formation assignment with ID: %q has state: %q and no destination(s) will be created because of it", di.FormationAssignment.ID, di.FormationAssignment.State)
		return true, nil
	}

	if di.Operation == model.AssignFormation {
		log.C(ctx).Infof("Handling %s operation for formation assignment with ID: %q", model.AssignFormation, di.FormationAssignment.ID)

		if di.FormationAssignment != nil && string(di.FormationAssignment.Value) != "" && string(di.FormationAssignment.Value) != "\"\"" && di.Location.OperationName == model.NotificationStatusReturned {
			log.C(ctx).Infof("Location with constraint type: %q and operation name: %q is reached", di.Location.ConstraintType, di.Location.OperationName)

			var assignmentConfig Configuration
			if err := json.Unmarshal(di.FormationAssignment.Value, &assignmentConfig); err != nil {
				return false, errors.Wrapf(err, "while unmarshaling tenant mapping response configuration from assignment with ID: %q", di.FormationAssignment.ID)
			}

			formationAssignment, err := RetrieveFormationAssignmentPointer(ctx, di.JoinPointDetailsFAMemoryAddress)
			if err != nil {
				return false, err
			}

			log.C(ctx).Infof("There is/are %d design time destination(s) available in the configuration response", len(assignmentConfig.Destinations))
			for _, destDetails := range assignmentConfig.Destinations {
				if err := e.destinationSvc.CreateDesignTimeDestinations(ctx, destDetails, formationAssignment); err != nil {
					return false, errors.Wrapf(err, "while creating design time destination with name: %q", destDetails.Name)
				}
			}

			if assignmentConfig.Credentials.InboundCommunicationDetails != nil {
				samlAssertionDetails := assignmentConfig.Credentials.InboundCommunicationDetails.SAMLAssertionDetails
				if isSAMLDetailsExists := samlAssertionDetails; isSAMLDetailsExists != nil {
					log.C(ctx).Infof("There is/are %d SAML Assertion destination details in the configuration response", len(samlAssertionDetails.Destinations))
					for i, destDetails := range samlAssertionDetails.Destinations {
						certData, err := e.destinationCreatorSvc.CreateCertificate(ctx, destDetails, formationAssignment, 0)
						if err != nil {
							return false, errors.Wrapf(err, "while creating SAML assertion certificate with name: %q", destDetails.Name)
						}

						config, err := e.destinationCreatorSvc.EnrichAssignmentConfigWithCertificateData(formationAssignment.Value, certData, i)
						if err != nil {
							return false, err
						}

						formationAssignment.Value = config
					}
				}
			}

			return true, nil
		}

		if di.FormationAssignment != nil && string(di.FormationAssignment.Value) != "" && string(di.FormationAssignment.Value) != "\"\"" && di.ReverseFormationAssignment != nil && string(di.ReverseFormationAssignment.Value) != "" && string(di.ReverseFormationAssignment.Value) != "\"\"" && di.Location.OperationName == model.SendNotificationOperation {
			log.C(ctx).Infof("Location with constraint type: %q and operation name: %q is reached", di.Location.ConstraintType, di.Location.OperationName)

			var assignmentConfig Configuration
			if err := json.Unmarshal(di.FormationAssignment.Value, &assignmentConfig); err != nil {
				return false, errors.Wrapf(err, "while unmarshaling tenant mapping configuration response from assignment with ID: %q", di.FormationAssignment.ID)
			}

			var reverseAssignmentConfig Configuration
			if err := json.Unmarshal(di.ReverseFormationAssignment.Value, &reverseAssignmentConfig); err != nil {
				return false, errors.Wrapf(err, "while unmarshaling tenant mapping configuration response from reverse assignment with ID: %q", di.ReverseFormationAssignment.ID)
			}

			if assignmentConfig.Credentials.InboundCommunicationDetails == nil {
				return false, errors.New("The inbound communication destination details could not be empty")
			}

			if reverseAssignmentConfig.Credentials.OutboundCommunicationCredentials == nil {
				return false, errors.New("The outbound communication credentials could not be empty")
			}

			formationAssignment, err := RetrieveFormationAssignmentPointer(ctx, di.JoinPointDetailsFAMemoryAddress)
			if err != nil {
				return false, err
			}

			basicAuthDetails := assignmentConfig.Credentials.InboundCommunicationDetails.BasicAuthenticationDetails
			basicAuthCreds := reverseAssignmentConfig.Credentials.OutboundCommunicationCredentials.BasicAuthentication
			if basicAuthDetails != nil && basicAuthCreds != nil {
				log.C(ctx).Infof("There is/are %d inbound basic destination(s) details available in the configuration", len(basicAuthDetails.Destinations))
				for _, destDetails := range basicAuthDetails.Destinations {
					if err := e.destinationSvc.CreateBasicCredentialDestinations(ctx, destDetails, *basicAuthCreds, formationAssignment, basicAuthDetails.CorrelationIDs); err != nil {
						return false, errors.Wrapf(err, "while creating basic destination with name: %q", destDetails.Name)
					}
				}
			}

			samlAssertionDetails := assignmentConfig.Credentials.InboundCommunicationDetails.SAMLAssertionDetails
			samlAuthCreds := reverseAssignmentConfig.Credentials.OutboundCommunicationCredentials.SAMLAssertionAuthentication
			if samlAssertionDetails != nil && samlAuthCreds != nil {
				log.C(ctx).Infof("There is/are %d inbound SAML destination(s) available in the configuration", len(basicAuthDetails.Destinations))
				for _, destDetails := range samlAssertionDetails.Destinations {
					if err := e.destinationSvc.CreateSAMLAssertionDestination(ctx, destDetails, samlAuthCreds, formationAssignment, samlAssertionDetails.CorrelationIDs); err != nil {
						return false, errors.Wrapf(err, "while creating SAML assertion destination with name: %q", destDetails.Name)
					}
				}
			}

			return true, nil
		}
	}

	log.C(ctx).Infof("Finished executing operator: %q", DestinationCreatorOperator)
	return true, nil
}

// Destination Creator Operator types

// Configuration represents a formation assignment(or reverse formation assignment) configuration
type Configuration struct {
	Destinations         []Destination        `json:"destinations"`
	Credentials          Credentials          `json:"credentials"`
	AdditionalProperties AdditionalProperties `json:"additionalProperties"`
}

// AdditionalProperties is an alias for slice of `json.RawMessage` elements
type AdditionalProperties []json.RawMessage

// Credentials represents different type of credentials configuration - inbound, outbound
type Credentials struct {
	OutboundCommunicationCredentials *OutboundCommunicationCredentials `json:"outboundCommunication,omitempty"`
	InboundCommunicationDetails      *InboundCommunicationDetails      `json:"inboundCommunication,omitempty"`
}

// OutboundCommunicationCredentials consists of different type of outbound authentications
type OutboundCommunicationCredentials struct {
	NoAuthentication                      *NoAuthentication                      `json:"noAuthentication,omitempty"`
	BasicAuthentication                   *BasicAuthentication                   `json:"basicAuthentication,omitempty"`
	SAMLAssertionAuthentication           *SAMLAssertionAuthentication           `json:"samlAssertion,omitempty"`
	OAuth2ClientCredentialsAuthentication *OAuth2ClientCredentialsAuthentication `json:"oauth2ClientCredentials,omitempty"`
	ClientCertAuthentication              *ClientCertAuthentication              `json:"clientCertificateAuthentication,omitempty"`
}

// NoAuthentication represents destination without authentication
type NoAuthentication struct {
	URL            string   `json:"url"`
	UIURL          string   `json:"uiUrl,omitempty"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

// BasicAuthentication represents destination with basic authentication
type BasicAuthentication struct {
	URL            string   `json:"url"`
	UIURL          string   `json:"uiUrl,omitempty"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

// SAMLAssertionAuthentication represents destination with SAML assertion authentication
type SAMLAssertionAuthentication struct {
	URL string `json:"url"`
}

// OAuth2ClientCredentialsAuthentication represents destination with OAuth 2 client credentials authentication
type OAuth2ClientCredentialsAuthentication struct {
	URL             string   `json:"url"`
	UIURL           string   `json:"uiUrl,omitempty"`
	TokenServiceURL string   `json:"tokenServiceUrl"`
	ClientID        string   `json:"clientId"`
	ClientSecret    string   `json:"clientSecret"`
	CorrelationIds  []string `json:"correlationIds,omitempty"`
}

// ClientCertAuthentication represents destination with client certificate authentication
type ClientCertAuthentication struct {
	URL            string   `json:"url"`
	UIURL          string   `json:"uiUrl,omitempty"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

// InboundCommunicationDetails consists of different type of inbound communication configuration details
type InboundCommunicationDetails struct {
	BasicAuthenticationDetails       *InboundBasicAuthenticationDetails       `json:"basicAuthentication,omitempty"`
	SAMLAssertionDetails             *InboundSAMLAssertionDetails             `json:"samlAssertion,omitempty"`
	OAuth2SAMLBearerAssertionDetails *InboundOAuth2SAMLBearerAssertionDetails `json:"oauth2SamlBearerAssertion,omitempty"`
}

// InboundBasicAuthenticationDetails represents destination configuration details for basic authentication
type InboundBasicAuthenticationDetails struct {
	CorrelationIDs []string      `json:"correlationIds"`
	Destinations   []Destination `json:"destinations"`
}

// InboundSAMLAssertionDetails represents destination configuration details for SAML assertion authentication
type InboundSAMLAssertionDetails struct {
	CorrelationIDs []string      `json:"correlationIds"`
	Destinations   []Destination `json:"destinations"`
}

// InboundOAuth2SAMLBearerAssertionDetails represents destination configuration details for SAML bearer assertion authentication
type InboundOAuth2SAMLBearerAssertionDetails struct {
	CorrelationIDs []string      `json:"correlationIds"`
	Destinations   []Destination `json:"destinations"`
}

// Destination holds different destination types properties
type Destination struct {
	Name                 string          `json:"name"`
	Type                 string          `json:"type,omitempty"`
	Description          string          `json:"description,omitempty"`
	ProxyType            string          `json:"proxyType,omitempty"`
	Authentication       string          `json:"authentication,omitempty"`
	URL                  string          `json:"url,omitempty"`
	SubaccountID         string          `json:"subaccountId,omitempty"`
	AdditionalProperties json.RawMessage `json:"additionalProperties,omitempty"`
	FileName             string          `json:"fileName,omitempty"`
	CommonName           string          `json:"commonName,omitempty"`
	CertificateChain     string          `json:"certificateChain,omitempty"`
}

// CertificateData contains the data for the certificate resource from the destination creator component
type CertificateData struct {
	FileName         string `json:"fileName"`
	CommonName       string `json:"commonName"`
	CertificateChain string `json:"certificateChain"`
}
