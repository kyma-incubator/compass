package operators

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/formationconstraint"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	// DestinationCreatorOperator represents the destination creator operator
	DestinationCreatorOperator = "DestinationCreatorOperator"
	ClientUserHeaderKey        = "CLIENT_USER"
	GlobalSubaccountLabelKey   = "global_subaccount_id"
	regionLabelKey             = "region"

	DestinationTypeHTTP DestinationType = "HTTP"
	DestinationTypeRFC  DestinationType = "RFC"
	DestinationTypeLDAP DestinationType = "LDAP"
	DestinationTypeMAIL DestinationType = "MAIL"

	DestinationAuthTypeNoAuth     DestinationAuthType = "NoAuthentication"
	DestinationAuthTypeBasic      DestinationAuthType = "BasicAuthentication"
	DestinationAuthTypeSAMLBearer DestinationAuthType = "OAuth2SAMLBearerAssertion"

	DestinationProxyTypeInternet    DestinationProxyType = "Internet"
	DestinationProxyTypeOnPremise   DestinationProxyType = "OnPremise"
	DestinationProxyTypePrivateLink DestinationProxyType = "PrivateLink"
)

type DestinationType string
type DestinationAuthType string
type DestinationProxyType string

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

	log.C(ctx).Infof("Enforcing constraint on resource of type: %q, subtype: %q and ID: %q during %q operation", di.ResourceType, di.ResourceSubtype, di.ResourceID, di.Operation)

	if di.Operation == model.UnassignFormation {
		// todo::: specific logic for unassign
		return true, nil
	}

	if di.Assignment.Value != "" && di.JoinPointLocation == formationconstraint.PostNotificationStatusReturned {
		log.C(ctx).Infof("Location with constraint type: %q and operation name: %q is reached", model.NotificationStatusReturned, model.PostOperation)

		var confDetailsResp ConfigurationResponse
		if err := json.Unmarshal([]byte(di.Assignment.Value), &confDetailsResp); err != nil {
			return false, errors.Wrapf(err, "while unmarshaling tenant mapping configuration response from assignment with ID: %q", di.Assignment.ID)
		}

		// todo:: will be implemented with the second phase of the destination operator
		//log.C(ctx).Infof("There is/are %d design time destination(s) available in the configuration response", len(confDetailsResp.Configuration.Destinations))
		//for _, destination := range confDetailsResp.Configuration.Destinations {
		//	statusCode, err := e.createDesignTimeDestinations(ctx, destination, di.Assignment)
		//	if err != nil {
		//		return false, errors.Wrapf(err, "while creating destination with name: %q", destination.Name)
		//	}
		//
		//	if statusCode == http.StatusConflict {
		//		log.C(ctx).Infof("The destination with name: %q already exists. Will be deleted and created again...")
		//		if err := e.deleteDestinations(ctx); err != nil {
		//			return false, errors.Wrapf(err, "while deleting destination with name: %q", destination.Name)
		//		}
		//
		//		if _, err = e.createDesignTimeDestinations(ctx, destination, di.Assignment); err != nil {
		//			return false, errors.Wrapf(err, "while creating destination with name: %q", destination.Name)
		//		}
		//	}
		//}
		//
		//samlAuthDetails := confDetailsResp.Configuration.Credentials.InboundCommunicationDetails.SAMLBearerAuthenticationDetails
		//if isSAMLDetailsExists := samlAuthDetails; isSAMLDetailsExists != nil {
		//	log.C(ctx).Infof("There is/are %d SAML Bearer destination details in the configuration response", len(samlAuthDetails.Destinations))
		//	for _, _ = range samlAuthDetails.Destinations {
		//		// todo:: (for phase 2) create KeyStore for every destination elements and enrich the destination before sending it to the Destination Creator component
		//	}
		//}

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
			for _, destination := range basicAuthDetails.Destinations {
				// todo::: build request using 'destination details' --> send the request to destination creator -> create destination with the credentials from the reverseAssignment
				if err := e.createCredentialDestinations(ctx, destination); err != nil {
					return false, err
				}
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

	return true, nil
}

func (e *ConstraintEngine) createCredentialDestinations(ctx context.Context, destination Destination,) error {
	return nil
}

func (e *ConstraintEngine) createDesignTimeDestinations(ctx context.Context, destination Destination, formationAssignment *webhook.FormationAssignment) (statusCode int, err error) {
	subaccountID, err := e.validateDestinationSubaccount(ctx, destination, formationAssignment)
	if err != nil {
		return statusCode, err
	}

	region, err := e.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return statusCode, err
	}

	strURL, err := buildURL(e.destinationCfg, region, subaccountID, "", false)
	if err != nil {
		return statusCode, errors.Wrapf(err, "while building destination URL")
	}

	destReqBody := &DestinationReqBody{
		Name:               destination.Name,
		Url:                destination.Url,
		Type:               destination.Type,
		ProxyType:          destination.ProxyType,
		AuthenticationType: destination.Authentication,
		User:               "dummyValue?", // todo::: TBD in case of generic auth? since it's a required field
	}

	if err := validateDestinationReqBody(destReqBody); err != nil {
		return statusCode, err
	}

	destReqbodyBytes, err := json.Marshal(destReqBody)
	if err != nil {
		return statusCode, errors.Wrapf(err, "while marshalling destination request body")
	}

	req, err := http.NewRequest(http.MethodPost, strURL, bytes.NewBuffer(destReqbodyBytes))
	req.Header.Set(ClientUserHeaderKey, "dummyValue?") // todo::: double check what should be the value of the header??

	log.C(ctx).Infof("Creating destination with name: %q", destination.Name)
	resp, err := e.mtlsHTTPClient.Do(req)
	if err != nil {
		return statusCode, err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return statusCode, errors.Errorf("Failed to read destination response body: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return statusCode, errors.Errorf("Failed to create destination, status: %d, body: %s", resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusConflict {
		log.C(ctx).Infof("The destination with name: %q already exists, continue with the next one.", destination.Name)
		return http.StatusConflict, nil
	}
	log.C(ctx).Infof("Successfully create destination with name: %q", destination.Name)

	return statusCode, nil
}

func (e *ConstraintEngine) validateDestinationSubaccount(ctx context.Context, destination Destination, formationAssignment *webhook.FormationAssignment) (string, error) {
	consumerSubaccountID, err := e.getConsumerTenant(ctx, formationAssignment)
	if err != nil {
		return "", err
	}

	var subaccountID string
	subaccountID = consumerSubaccountID

	if destination.SubaccountID != "" && destination.SubaccountID != consumerSubaccountID {
		switch formationAssignment.TargetType {
		case model.FormationAssignmentTypeApplication:
			if err := e.validateAppTemplateProviderSubaccount(ctx, formationAssignment, destination.SubaccountID); err != nil {
				return "", err
			}
		case model.FormationAssignmentTypeRuntime:
			if err := e.validateRuntimeProviderSubaccount(ctx, formationAssignment.Target, destination.SubaccountID); err != nil {
				return "", err
			}
		case model.FormationAssignmentTypeRuntimeContext:
			if err := e.validateRuntimeContextProviderSubaccount(ctx, formationAssignment, destination.SubaccountID); err != nil {
				return "", err
			}
		default:
			return "", errors.Errorf("Unknown formation assignment type: %q", formationAssignment.TargetType)
		}

		subaccountID = destination.SubaccountID
	}

	return subaccountID, nil
}

func (e *ConstraintEngine) getConsumerTenant(ctx context.Context, formationAssignment *webhook.FormationAssignment) (string, error) {
	labelableObjType, err := determineLabelableObjectType(formationAssignment.TargetType)
	if err != nil {
		return "", err
	}

	labels, err := e.labelRepo.ListForObject(ctx, formationAssignment.TenantID, labelableObjType, formationAssignment.Target)
	if err != nil {
		return "", errors.Wrapf(err, "while getting labels for %s with ID: %q", formationAssignment.TargetType, formationAssignment.Target)
	}

	globalSubaccIDLbl, globalSubaccIDExists := labels[GlobalSubaccountLabelKey]
	if !globalSubaccIDExists {
		return "", errors.Errorf("%q label does not exists for: %q with ID: %q", GlobalSubaccountLabelKey, formationAssignment.TargetType, formationAssignment.Target)
	}

	globalSubaccIDLblValue, ok := globalSubaccIDLbl.Value.(string)
	if !ok {
		return "", errors.Errorf("unexpected type of %q label, expect: string, got: %T", GlobalSubaccountLabelKey, globalSubaccIDLbl.Value)
	}

	return globalSubaccIDLblValue, nil
}

func determineLabelableObjectType(assignmentType model.FormationAssignmentType) (model.LabelableObject, error) {
	switch assignmentType {
	case model.FormationAssignmentTypeApplication:
		return model.ApplicationLabelableObject, nil
	case model.FormationAssignmentTypeRuntime:
		return model.RuntimeLabelableObject, nil
	case model.FormationAssignmentTypeRuntimeContext:
		return model.RuntimeContextLabelableObject, nil
	default:
		return "", errors.Errorf("Couldn't determine the label-able object type from assignment type: %q", assignmentType)
	}
}

func (e *ConstraintEngine) getRegionLabel(ctx context.Context, tenantID string) (string, error) {
	regionLbl, err := e.labelRepo.GetByKey(ctx, tenantID, model.TenantLabelableObject, tenantID, regionLabelKey)
	if err != nil {
		return "", err
	}

	region, ok := regionLbl.Value.(string)
	if !ok {
		return "", errors.Errorf("unexpected type of %q label, expect: string, got: %T", regionLabelKey, regionLbl.Value)
	}
	return region, nil
}

func (e *ConstraintEngine) validateAppTemplateProviderSubaccount(ctx context.Context, formationAssignment *webhook.FormationAssignment, destinationSubaccountID string) error {
	app, err := e.applicationRepository.GetByID(ctx, formationAssignment.TenantID, formationAssignment.Target)
	if err != nil {
		return err
	}

	if app.ApplicationTemplateID != nil && *app.ApplicationTemplateID != "" {
		labels, err := e.labelRepo.ListForGlobalObject(ctx, model.AppTemplateLabelableObject, *app.ApplicationTemplateID)
		if err != nil {
			return errors.Wrapf(err, "while getting labels for application template with ID: %q", *app.ApplicationTemplateID)
		}

		subaccountLbl, subaccountLblExists := labels[GlobalSubaccountLabelKey]

		if !subaccountLblExists {
			return errors.Errorf("%q label should exist as part of the provider application template with ID: %q", GlobalSubaccountLabelKey, *app.ApplicationTemplateID)
		}

		subaccountLblValue, ok := subaccountLbl.Value.(string)
		if !ok {
			return errors.Errorf("unexpected type of %q label, expect: string, got: %T", GlobalSubaccountLabelKey, subaccountLbl.Value)
		}

		if destinationSubaccountID != subaccountLblValue {
			return errors.Errorf("The provided destination subaccount is different from the owner subaccount of the application template with ID: %q", *app.ApplicationTemplateID)
		}
	}

	return nil
}

func (e *ConstraintEngine) validateRuntimeProviderSubaccount(ctx context.Context, runtimeID, destinationSubaccountID string) error {
	exists, err := e.runtimeRepository.OwnerExists(ctx, destinationSubaccountID, runtimeID)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("The provided destination subaccount: %q is not provider of the runtime with ID: %q", destinationSubaccountID, runtimeID)
	}

	return nil
}

func (e *ConstraintEngine) validateRuntimeContextProviderSubaccount(ctx context.Context, formationAssignment *webhook.FormationAssignment, destinationSubaccountID string) error {
	rtmCtxID, err := e.runtimeCtxRepository.GetByID(ctx, formationAssignment.TenantID, formationAssignment.Target)
	if err != nil {
		return err
	}

	return e.validateRuntimeProviderSubaccount(ctx, rtmCtxID.RuntimeID, destinationSubaccountID)
}

func (e *ConstraintEngine) deleteDestinations(ctx context.Context) error { // todo::: propagate the assignmentID as parameter
	// todo::: extract the destination from the DB by assignment ID with the associated destination name and tenantID(subaccountID)
	// todo::: get subaccountID region label

	strURL, err := buildURL(e.destinationCfg, "region", "subaccountID", "destinationName", true)
	if err != nil {
		return errors.Wrapf(err, "while building destination URL")
	}

	req, err := http.NewRequest(http.MethodDelete, strURL, nil)
	req.Header.Set(ClientUserHeaderKey, "dummyValue?") // todo::: double check what should be the value of the header??

	log.C(ctx).Infof("Deleting destination with name: %q", "destinationName")
	resp, err := e.mtlsHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Errorf("Failed to read destination response body: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("Failed to delete destination, status: %d, body: %s", resp.StatusCode, body)
	}

	log.C(ctx).Infof("Successfully delete destination with name: %q", "destinationName")

	return nil
}

func buildURL(destinationCfg *DestinationConfig, region, subaccountID, destinationName string, isDeleteRequest bool) (string, error) {
	if region == "" || subaccountID == "" {
		return "", errors.Errorf("The provided region and/or subaccount for the URL couldn't be empty")
	}

	base, err := url.Parse(destinationCfg.BaseURL)
	if err != nil {
		return "", err
	}

	path := destinationCfg.Path

	regionalEndpoint := strings.Replace(path, fmt.Sprintf("{%s}", destinationCfg.RegionParam), region, 1)
	regionalEndpoint = strings.Replace(regionalEndpoint, fmt.Sprintf("{%s}", destinationCfg.SubaccountIDParam), subaccountID, 1)

	if isDeleteRequest {
		if destinationName == "" {
			return "", errors.Errorf("The destination name should not be empty in case of %s request", http.MethodDelete)
		}
		regionalEndpoint += fmt.Sprintf("/{%s}", destinationCfg.DestinationNameParam)
		regionalEndpoint = strings.Replace(regionalEndpoint, fmt.Sprintf("{%s}", destinationCfg.DestinationNameParam), destinationName, 1)
	}

	// Path params
	base.Path += regionalEndpoint

	return base.String(), nil
}

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.C(ctx).Errorf("An error has occurred while closing response body: %v", err)
	}
}

func validateDestinationReqBody(destinationReqBody *DestinationReqBody) error {
	if destinationReqBody.Name == "" {
		return errors.New("The name field of the destination request body is mandatory")
	}

	if destinationReqBody.Url == "" {
		return errors.New("The URL field of the destination request body is mandatory")
	}

	if destinationReqBody.Type == "" {
		return errors.New("The type field of the destination request body is mandatory")
	}

	if destinationReqBody.ProxyType == "" {
		return errors.New("The proxy type field of the destination request body is mandatory")
	}

	if destinationReqBody.AuthenticationType == "" {
		return errors.New("The authentication type field of the destination request body is mandatory")
	}

	return nil
}

// Destination Creator Operator types

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
	Type                 DestinationType      `json:"type"`
	Description          string               `json:"description,omitempty"`
	ProxyType            DestinationProxyType `json:"proxyType"`
	Authentication       DestinationAuthType  `json:"authentication"`
	Url                  string               `json:"url"`
	SubaccountID         string               `json:"subaccountId,omitempty"`
	AdditionalAttributes AdditionalAttributes `json:"additionalAttributes,omitempty"`
	// todo::: additional fields for KeyStore(phase 2)
}

// AdditionalAttributes todo::: add godoc
type AdditionalAttributes json.RawMessage

// Destination Creator API types

// DestinationReqBody // todo::: add godoc
type DestinationReqBody struct {
	Name               string               `json:"name"`
	Url                string               `json:"url"`
	Type               DestinationType      `json:"type"`
	ProxyType          DestinationProxyType `json:"proxyType"`
	AuthenticationType DestinationAuthType  `json:"authenticationType"`
	User               string               `json:"user"`
	Password           string               `json:"password"`
}

// DestinationErrorResponse // todo::: add godoc
type DestinationErrorResponse struct {
	Error DestinationError `json:"error"`
}

// DestinationError // todo::: add godoc
type DestinationError struct {
	Timestamp string `json:"timestamp"`
	Code      int    `json:"code"`
	Message   string `json:"message"`
}

// DestinationConfig // todo::: add godoc
type DestinationConfig struct {
	BaseURL              string `envconfig:"APP_DESTINATION_BASE_URL"`
	Path                 string `envconfig:"APP_DESTINATION_PATH"`                    // todo::: "/regions/{region}/subaccounts/{subaccountId}/destinations"
	RegionParam          string `envconfig:"APP_DESTINATION_REGION_PARAMETER"`        // todo::: "region"
	SubaccountIDParam    string `envconfig:"APP_DESTINATION_SUBACCOUNT_ID_PARAMETER"` // todo::: "subaccountId"
	DestinationNameParam string `envconfig:"APP_DESTINATION_NAME_PARAMETER"`          // todo::: "destinationName"
}
