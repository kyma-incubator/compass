package operators

import (
	"context"
	"encoding/json"
	"runtime/debug"

	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

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

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q and subtype: %q for location with constraint type: %q and operation name: %q during %q operation", di.ResourceType, di.ResourceSubtype, di.Location.ConstraintType, di.Location.OperationName, di.Operation)

	formationAssignment, err := RetrieveFormationAssignmentPointer(ctx, di.JoinPointDetailsFAMemoryAddress)
	if err != nil {
		return false, err
	}

	if di.Operation == model.UnassignFormation {
		if di.Location.OperationName == model.NotificationStatusReturned && formationAssignment != nil && formationAssignment.State == string(model.ReadyAssignmentState) {
			log.C(ctx).Infof("Handling %s operation for formation assignment with ID: %q", model.UnassignFormation, formationAssignment.ID)
			if err := e.destinationSvc.DeleteDestinations(ctx, formationAssignment, di.SkipSubaccountValidation); err != nil {
				return false, err
			}
		}
		log.C(ctx).Infof("Finished executing operator: %q for %q operation", DestinationCreatorOperator, model.UnassignFormation)
		return true, nil
	}

	isAssignmentCfgEmpty := isFormationAssignmentConfigEmpty(formationAssignment)
	if formationAssignment != nil && (formationAssignment.State == string(model.InitialAssignmentState) || isAssignmentCfgEmpty) {
		log.C(ctx).Infof("The formation assignment with ID: %s has either %s state or empty configuration. Returning without executing the destination creator operator.", formationAssignment.ID, model.InitialAssignmentState)
		return true, nil
	}

	if formationAssignment != nil && !isAssignmentCfgEmpty && di.Location.OperationName == model.NotificationStatusReturned {
		log.C(ctx).Infof("Location with constraint type: %q and operation name: %q is reached", di.Location.ConstraintType, di.Location.OperationName)

		var assignmentConfig Configuration
		if err := json.Unmarshal(formationAssignment.Value, &assignmentConfig); err != nil {
			return false, errors.Wrapf(err, "while unmarshalling tenant mapping response configuration from assignment with ID: %q", formationAssignment.ID)
		}

		if len(assignmentConfig.Destinations) > 0 {
			log.C(ctx).Infof("There is/are %d design time destination(s) available in the configuration response", len(assignmentConfig.Destinations))
			if err := e.destinationSvc.CreateDesignTimeDestinations(ctx, assignmentConfig.Destinations, formationAssignment, di.SkipSubaccountValidation); err != nil {
				return false, errors.Wrap(err, "while creating design time destinations")
			}
		}

		if assignmentConfig.Credentials.InboundCommunicationDetails != nil {
			if samlAssertionDetails := assignmentConfig.Credentials.InboundCommunicationDetails.SAMLAssertionDetails; samlAssertionDetails != nil && len(samlAssertionDetails.Destinations) > 0 {
				log.C(ctx).Infof("There is/are %d SAML Assertion destination details in the configuration response", len(samlAssertionDetails.Destinations))

				if samlAssertionDetails.Certificate != nil && *samlAssertionDetails.Certificate != "" && samlAssertionDetails.AssertionIssuer != nil && *samlAssertionDetails.AssertionIssuer != "" {
					log.C(ctx).Infof("The certificate and assertion issuer for SAML Assertion destination already exist. No new certificate will be generated.")
					return true, nil
				}

				certData, err := e.destinationCreatorSvc.CreateCertificate(ctx, samlAssertionDetails.Destinations, destinationcreatorpkg.AuthTypeSAMLAssertion, formationAssignment, 0, di.SkipSubaccountValidation)
				if err != nil {
					return false, errors.Wrap(err, "while creating SAML assertion certificate")
				}

				config, err := e.destinationCreatorSvc.EnrichAssignmentConfigWithSAMLCertificateData(formationAssignment.Value, destinationcreatorpkg.SAMLAssertionDestPath, certData)
				if err != nil {
					return false, err
				}
				formationAssignment.Value = config
			}

			if clientCertDetails := assignmentConfig.Credentials.InboundCommunicationDetails.ClientCertificateAuthenticationDetails; clientCertDetails != nil && len(clientCertDetails.Destinations) > 0 {
				log.C(ctx).Infof("There is/are %d client certificate destination details in the configuration response", len(clientCertDetails.Destinations))

				if clientCertDetails.Certificate != nil && *clientCertDetails.Certificate != "" {
					log.C(ctx).Infof("The certificate for client certificate authentication destination already exists. No new certificate will be generated.")
					return true, nil
				}

				certData, err := e.destinationCreatorSvc.CreateCertificate(ctx, clientCertDetails.Destinations, destinationcreatorpkg.AuthTypeClientCertificate, formationAssignment, 0, di.SkipSubaccountValidation)
				if err != nil {
					return false, errors.Wrap(err, "while creating client certificate authentication certificate")
				}

				config, err := e.destinationCreatorSvc.EnrichAssignmentConfigWithCertificateData(formationAssignment.Value, destinationcreatorpkg.ClientCertAuthDestPath, certData)
				if err != nil {
					return false, err
				}
				formationAssignment.Value = config
			}
		}

		log.C(ctx).Infof("Finished executing operator: %q for location with constraint type: %q and operation name: %q during %q operation", DestinationCreatorOperator, di.Location.ConstraintType, di.Location.OperationName, model.AssignFormation)
		return true, nil
	}

	reverseFormationAssignment, err := RetrieveFormationAssignmentPointer(ctx, di.JoinPointDetailsReverseFAMemoryAddress)
	if err != nil {
		return false, err
	}
	isReverseAssignmentCfgEmpty := isFormationAssignmentConfigEmpty(reverseFormationAssignment)

	if formationAssignment != nil && !isAssignmentCfgEmpty && reverseFormationAssignment != nil && !isReverseAssignmentCfgEmpty && di.Location.OperationName == model.SendNotificationOperation {
		log.C(ctx).Infof("Location with constraint type: %q and operation name: %q is reached", di.Location.ConstraintType, di.Location.OperationName)

		var assignmentConfig Configuration
		if err := json.Unmarshal(formationAssignment.Value, &assignmentConfig); err != nil {
			return false, errors.Wrapf(err, "while unmarshalling tenant mapping configuration response from assignment with ID: %q", formationAssignment.ID)
		}

		var reverseAssignmentConfig Configuration
		if err := json.Unmarshal(reverseFormationAssignment.Value, &reverseAssignmentConfig); err != nil {
			return false, errors.Wrapf(err, "while unmarshalling tenant mapping configuration response from reverse assignment with ID: %q", reverseFormationAssignment.ID)
		}

		if assignmentConfig.Credentials.InboundCommunicationDetails == nil {
			log.C(ctx).Warnf("No destination details are provided in the inbound communication property in the assignment configuration. No destination will be created.")
			return true, nil
		}

		if reverseAssignmentConfig.Credentials.OutboundCommunicationCredentials == nil {
			log.C(ctx).Warnf("No outbound communication credentials are provided in the reverse assignment configuration. No destination will be created.")
			return true, nil
		}

		basicAuthDetails := assignmentConfig.Credentials.InboundCommunicationDetails.BasicAuthenticationDetails
		basicAuthCreds := reverseAssignmentConfig.Credentials.OutboundCommunicationCredentials.BasicAuthentication
		if basicAuthDetails != nil && basicAuthCreds != nil && len(basicAuthDetails.Destinations) > 0 {
			log.C(ctx).Infof("There is/are %d inbound basic destination(s) details available in the configuration", len(basicAuthDetails.Destinations))
			if err := e.destinationSvc.CreateBasicCredentialDestinations(ctx, basicAuthDetails.Destinations, *basicAuthCreds, formationAssignment, basicAuthDetails.CorrelationIDs, di.SkipSubaccountValidation); err != nil {
				return false, errors.Wrap(err, "while creating basic destinations")
			}
		}

		samlAssertionDetails := assignmentConfig.Credentials.InboundCommunicationDetails.SAMLAssertionDetails
		samlAuthCreds := reverseAssignmentConfig.Credentials.OutboundCommunicationCredentials.SAMLAssertionAuthentication
		if samlAssertionDetails != nil && samlAuthCreds != nil && len(samlAssertionDetails.Destinations) > 0 {
			log.C(ctx).Infof("There is/are %d inbound SAML Assertion destination(s) available in the configuration", len(samlAssertionDetails.Destinations))
			if err := e.destinationSvc.CreateSAMLAssertionDestination(ctx, samlAssertionDetails.Destinations, samlAuthCreds, formationAssignment, samlAssertionDetails.CorrelationIDs, di.SkipSubaccountValidation); err != nil {
				return false, errors.Wrap(err, "while creating SAML Assertion destinations")
			}
		}

		clientCertDetails := assignmentConfig.Credentials.InboundCommunicationDetails.ClientCertificateAuthenticationDetails
		clientCertCreds := reverseAssignmentConfig.Credentials.OutboundCommunicationCredentials.ClientCertAuthentication
		if clientCertDetails != nil && clientCertCreds != nil && len(clientCertDetails.Destinations) > 0 {
			log.C(ctx).Infof("There is/are %d inbound client certificate authentication destination(s) available in the configuration", len(clientCertDetails.Destinations))
			if err := e.destinationSvc.CreateClientCertificateAuthenticationDestination(ctx, clientCertDetails.Destinations, clientCertCreds, formationAssignment, clientCertDetails.CorrelationIDs, di.SkipSubaccountValidation); err != nil {
				return false, errors.Wrapf(err, "while creating client certificate authentication destinations")
			}
		}

		log.C(ctx).Infof("Finished executing operator: %q for location with constraint type: %q and operation name: %q during %q operation", DestinationCreatorOperator, di.Location.ConstraintType, di.Location.OperationName, model.AssignFormation)
		return true, nil
	}

	log.C(ctx).Infof("Finished executing operator: %q during %q operation", DestinationCreatorOperator, model.AssignFormation)
	return true, nil
}

func isFormationAssignmentConfigEmpty(assignment *model.FormationAssignment) bool {
	if assignment != nil && (string(assignment.Value) == "" || string(assignment.Value) == "{}" || string(assignment.Value) == "\"\"" || string(assignment.Value) == "null") {
		return true
	}

	return false
}

// Destination Creator Operator types

// Configuration represents a formation assignment (or reverse formation assignment) configuration
type Configuration struct {
	Destinations         []Destination        `json:"destinations"`
	Credentials          Credentials          `json:"credentials"`
	AdditionalProperties AdditionalProperties `json:"additionalProperties"`
}

// AdditionalProperties is an alias for slice of `json.RawMessage` elements
type AdditionalProperties []json.RawMessage

// Credentials represent a different type of credentials configuration - inbound, outbound
type Credentials struct {
	OutboundCommunicationCredentials *OutboundCommunicationCredentials `json:"outboundCommunication,omitempty"`
	InboundCommunicationDetails      *InboundCommunicationDetails      `json:"inboundCommunication,omitempty"`
}

// OutboundCommunicationCredentials consists of a different type of outbound authentications
type OutboundCommunicationCredentials struct {
	NoAuthentication                        *NoAuthentication                        `json:"noAuthentication,omitempty"`
	BasicAuthentication                     *BasicAuthentication                     `json:"basicAuthentication,omitempty"`
	SAMLAssertionAuthentication             *SAMLAssertionAuthentication             `json:"samlAssertion,omitempty"`
	OAuth2SAMLBearerAssertionAuthentication *OAuth2SAMLBearerAssertionAuthentication `json:"oauth2SamlBearerAssertion,omitempty"`
	ClientCertAuthentication                *ClientCertAuthentication                `json:"clientCertificateAuthentication,omitempty"`
	OAuth2ClientCredentialsAuthentication   *OAuth2ClientCredentialsAuthentication   `json:"oauth2ClientCredentials,omitempty"`
}

// NoAuthentication represents outbound communication without any authentication
type NoAuthentication struct {
	URL            string   `json:"url"`
	UIURL          string   `json:"uiUrl,omitempty"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

// BasicAuthentication represents outbound communication with basic authentication
type BasicAuthentication struct {
	URL            string   `json:"url"`
	UIURL          string   `json:"uiUrl,omitempty"`
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

// SAMLAssertionAuthentication represents outbound communication with SAML Assertion authentication
type SAMLAssertionAuthentication struct {
	URL string `json:"url"`
}

// OAuth2SAMLBearerAssertionAuthentication represents outbound communication with OAuth 2 SAML Bearer Assertion authentication
type OAuth2SAMLBearerAssertionAuthentication struct {
	URL             string `json:"url"`
	TokenServiceURL string `json:"tokenServiceUrl"`
	ClientID        string `json:"clientId"`
	ClientSecret    string `json:"clientSecret"`
}

// OAuth2ClientCredentialsAuthentication represents outbound communication with OAuth 2 client credentials authentication
type OAuth2ClientCredentialsAuthentication struct {
	URL             string   `json:"url"`
	UIURL           string   `json:"uiUrl,omitempty"`
	TokenServiceURL string   `json:"tokenServiceUrl"`
	ClientID        string   `json:"clientId"`
	ClientSecret    string   `json:"clientSecret"`
	CorrelationIds  []string `json:"correlationIds,omitempty"`
}

// ClientCertAuthentication represents outbound communication with client certificate authentication
type ClientCertAuthentication struct {
	URL            string   `json:"url"`
	UIURL          string   `json:"uiUrl,omitempty"`
	CorrelationIds []string `json:"correlationIds,omitempty"`
}

// InboundCommunicationDetails consists of a different type of inbound communication configuration details
type InboundCommunicationDetails struct {
	BasicAuthenticationDetails             *InboundBasicAuthenticationDetails       `json:"basicAuthentication,omitempty"`
	SAMLAssertionDetails                   *InboundSAMLAssertionDetails             `json:"samlAssertion,omitempty"`
	OAuth2SAMLBearerAssertionDetails       *InboundOAuth2SAMLBearerAssertionDetails `json:"oauth2SamlBearerAssertion,omitempty"`
	ClientCertificateAuthenticationDetails *InboundClientCertAuthenticationDetails  `json:"clientCertificateAuthentication,omitempty"`
}

// InboundBasicAuthenticationDetails represents inbound communication configuration details for basic authentication
type InboundBasicAuthenticationDetails struct {
	CorrelationIDs []string      `json:"correlationIds"`
	Destinations   []Destination `json:"destinations"`
}

// InboundSAMLAssertionDetails represents inbound communication configuration details for SAML assertion authentication
type InboundSAMLAssertionDetails struct {
	CorrelationIDs  []string      `json:"correlationIds"`
	Destinations    []Destination `json:"destinations"`
	Certificate     *string       `json:"certificate,omitempty"`
	AssertionIssuer *string       `json:"assertionIssuer,omitempty"`
}

// InboundOAuth2SAMLBearerAssertionDetails represents inbound communication configuration details for SAML bearer assertion authentication
type InboundOAuth2SAMLBearerAssertionDetails struct {
	CorrelationIDs  []string      `json:"correlationIds"`
	Destinations    []Destination `json:"destinations"`
	Certificate     *string       `json:"certificate,omitempty"`
	AssertionIssuer *string       `json:"assertionIssuer,omitempty"`
}

// InboundClientCertAuthenticationDetails represents inbound communication configuration details for client certificate authentication
type InboundClientCertAuthenticationDetails struct {
	CorrelationIDs []string      `json:"correlationIds"`
	Destinations   []Destination `json:"destinations"`
	Certificate    *string       `json:"certificate,omitempty"`
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
	InstanceID           string          `json:"instanceId,omitempty"`
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
