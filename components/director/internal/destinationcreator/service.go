package destinationcreator

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"

	destinationcreatorpkg "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

	"github.com/kyma-incubator/compass/components/director/internal/domain/client"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/pkg/errors"
	"github.com/tidwall/sjson"
)

const (
	clientUserHeaderKey        = "CLIENT_USER"
	contentTypeHeaderKey       = "Content-Type"
	contentTypeApplicationJSON = "application/json;charset=UTF-8"
	// GlobalSubaccountLabelKey is a label that holds the external subaccount ID of an entity
	GlobalSubaccountLabelKey = "global_subaccount_id"
	// RegionLabelKey holds the region value of an entity
	RegionLabelKey = "region"
	// DepthLimit is the recursion depth in case of conflict during destination creation
	DepthLimit = 2
)

//go:generate mockery --exported --name=httpClient --output=automock --outpkg=automock --case=underscore --disable-version-string
type httpClient interface {
	Do(request *http.Request) (*http.Response, error)
}

//go:generate mockery --exported --name=applicationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type applicationRepository interface {
	ListByScenariosNoPaging(ctx context.Context, tenant string, scenarios []string) ([]*model.Application, error)
	GetByID(ctx context.Context, tenant, id string) (*model.Application, error)
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

//go:generate mockery --exported --name=runtimeRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeRepository interface {
	OwnerExists(ctx context.Context, tenant, id string) (bool, error)
}

//go:generate mockery --exported --name=runtimeCtxRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type runtimeCtxRepository interface {
	GetByID(ctx context.Context, tenant, id string) (*model.RuntimeContext, error)
}

//go:generate mockery --exported --name=labelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelRepository interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	ListForGlobalObject(ctx context.Context, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
}

//go:generate mockery --exported --name=tenantRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantRepository interface {
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
}

// UIDService generates UUIDs for new entities
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// Service consists of a service-level operations related to the destination creator remote microservice
type Service struct {
	mtlsHTTPClient        httpClient
	config                *Config
	applicationRepository applicationRepository
	runtimeRepository     runtimeRepository
	runtimeCtxRepository  runtimeCtxRepository
	labelRepo             labelRepository
	tenantRepo            tenantRepository
}

// NewService creates a new Service
func NewService(
	mtlsHTTPClient httpClient,
	config *Config,
	applicationRepository applicationRepository,
	runtimeRepository runtimeRepository,
	runtimeCtxRepository runtimeCtxRepository,
	labelRepo labelRepository,
	tenantRepository tenantRepository,
) *Service {
	return &Service{
		mtlsHTTPClient:        mtlsHTTPClient,
		config:                config,
		applicationRepository: applicationRepository,
		runtimeRepository:     runtimeRepository,
		runtimeCtxRepository:  runtimeCtxRepository,
		labelRepo:             labelRepo,
		tenantRepo:            tenantRepository,
	}
}

// CreateDesignTimeDestinations is responsible to create so-called design time(destinationcreator.AuthTypeNoAuth) destination resource in the DB as well as in the remote destination service
func (s *Service) CreateDesignTimeDestinations(
	ctx context.Context,
	destinationDetails operators.Destination,
	formationAssignment *model.FormationAssignment,
	depth uint8,
) error {
	subaccountID := destinationDetails.SubaccountID
	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting region label for tenant with ID: %s", subaccountID)
	}

	strURL, err := buildDestinationURL(ctx, s.config.DestinationAPIConfig, URLParameters{
		EntityName:   "",
		Region:       region,
		SubaccountID: subaccountID,
		InstanceID:   destinationDetails.InstanceID,
	}, false)
	if err != nil {
		return errors.Wrapf(err, "while building destination URL")
	}

	destinationName := destinationDetails.Name
	destReqBody := &NoAuthDestinationRequestBody{
		BaseDestinationRequestBody: BaseDestinationRequestBody{
			Name:                 destinationName,
			URL:                  destinationDetails.URL,
			Type:                 destinationcreatorpkg.Type(destinationDetails.Type),
			ProxyType:            destinationcreatorpkg.ProxyType(destinationDetails.ProxyType),
			AuthenticationType:   destinationcreatorpkg.AuthType(destinationDetails.Authentication),
			AdditionalProperties: destinationDetails.AdditionalProperties,
		},
	}

	if err := destReqBody.Validate(); err != nil {
		return errors.Wrapf(err, "while validating no authentication destination request body")
	}

	log.C(ctx).Infof("Creating design time destination with name: %q, subaccount ID: %q and assignment ID: %q in the destination service", destinationName, subaccountID, formationAssignment.ID)
	_, statusCode, err := s.executeCreateRequest(ctx, strURL, destReqBody, destinationName)
	if err != nil {
		return errors.Wrapf(err, "while creating design time destination with name: %q in the destination service", destinationName)
	}

	if statusCode == http.StatusConflict {
		log.C(ctx).Infof("The destination with name: %q already exists. Will be deleted and created again...", destinationName)
		depth++
		if depth > DepthLimit {
			return errors.Errorf("Destination creator service retry limit: %d is exceeded", DepthLimit)
		}

		if err := s.DeleteDestination(ctx, destinationName, subaccountID, destinationDetails.InstanceID, formationAssignment); err != nil {
			return errors.Wrapf(err, "while deleting destination with name: %q and subaccount ID: %q", destinationName, subaccountID)
		}

		return s.CreateDesignTimeDestinations(ctx, destinationDetails, formationAssignment, depth)
	}

	return nil
}

// CreateBasicCredentialDestinations is responsible to create a basic destination resource in the remote destination service
func (s *Service) CreateBasicCredentialDestinations(
	ctx context.Context,
	destinationDetails operators.Destination,
	basicAuthenticationCredentials operators.BasicAuthentication,
	formationAssignment *model.FormationAssignment,
	correlationIDs []string,
	depth uint8,
) error {
	subaccountID := destinationDetails.SubaccountID
	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting region label for tenant with ID: %s", subaccountID)
	}

	strURL, err := buildDestinationURL(ctx, s.config.DestinationAPIConfig, URLParameters{
		EntityName:   "",
		Region:       region,
		SubaccountID: subaccountID,
		InstanceID:   destinationDetails.InstanceID,
	}, false)
	if err != nil {
		return errors.Wrapf(err, "while building destination URL")
	}

	reqBody, err := s.PrepareBasicRequestBody(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs)
	if err != nil {
		return err
	}

	destinationName := destinationDetails.Name
	log.C(ctx).Infof("Creating inbound basic destination with name: %q, subaccount ID: %q and assignment ID: %q in the destination service", destinationName, subaccountID, formationAssignment.ID)
	_, statusCode, err := s.executeCreateRequest(ctx, strURL, reqBody, destinationName)
	if err != nil {
		return errors.Wrapf(err, "while creating inbound basic destination with name: %q in the destination service", destinationName)
	}

	if statusCode == http.StatusConflict {
		log.C(ctx).Infof("The destination with name: %q already exists. Will be deleted and created again...", destinationName)
		depth++
		if depth > DepthLimit {
			return errors.Errorf("Destination creator service retry limit: %d is exceeded", DepthLimit)
		}

		if err := s.DeleteDestination(ctx, destinationName, subaccountID, destinationDetails.InstanceID, formationAssignment); err != nil {
			return errors.Wrapf(err, "while deleting destination with name: %q and subaccount ID: %q", destinationName, subaccountID)
		}

		return s.CreateBasicCredentialDestinations(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs, depth)
	}

	return nil
}

// CreateSAMLAssertionDestination is responsible to create SAML Assertion destination resource in the DB as well as in the remote destination service
func (s *Service) CreateSAMLAssertionDestination(
	ctx context.Context,
	destinationDetails operators.Destination,
	samlAuthCreds *operators.SAMLAssertionAuthentication,
	formationAssignment *model.FormationAssignment,
	correlationIDs []string,
	depth uint8,
) error {
	subaccountID := destinationDetails.SubaccountID
	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting region label for tenant with ID: %s", subaccountID)
	}

	strURL, err := buildDestinationURL(ctx, s.config.DestinationAPIConfig, URLParameters{
		EntityName:   "",
		Region:       region,
		SubaccountID: subaccountID,
		InstanceID:   destinationDetails.InstanceID,
	}, false)
	if err != nil {
		return errors.Wrapf(err, "while building destination URL")
	}

	certName, err := GetDestinationCertificateName(ctx, destinationcreatorpkg.AuthTypeSAMLAssertion, formationAssignment.ID)
	if err != nil {
		return errors.Wrapf(err, "while getting destination certificate name for destination auth type: %s", destinationcreatorpkg.AuthTypeSAMLAssertion)
	}

	destinationName := destinationDetails.Name
	destReqBody := &SAMLAssertionDestinationRequestBody{
		BaseDestinationRequestBody: BaseDestinationRequestBody{
			Name:               destinationName,
			URL:                samlAuthCreds.URL,
			Type:               destinationcreatorpkg.TypeHTTP,
			ProxyType:          destinationcreatorpkg.ProxyTypeInternet,
			AuthenticationType: destinationcreatorpkg.AuthTypeSAMLAssertion,
		},
		KeyStoreLocation: certName + destinationcreatorpkg.JavaKeyStoreFileExtension,
	}

	if destinationDetails.Type != "" {
		destReqBody.Type = destinationcreatorpkg.Type(destinationDetails.Type)
	}

	if destinationDetails.ProxyType != "" {
		destReqBody.ProxyType = destinationcreatorpkg.ProxyType(destinationDetails.ProxyType)
	}

	if destinationDetails.Authentication != "" && destinationcreatorpkg.AuthType(destinationDetails.Authentication) != destinationcreatorpkg.AuthTypeSAMLAssertion {
		return errors.Errorf("The provided authentication type: %s in the destination details is invalid. It should be %s", destinationDetails.Authentication, destinationcreatorpkg.AuthTypeSAMLAssertion)
	}

	enrichedProperties, err := enrichDestinationAdditionalPropertiesWithCorrelationIDs(s.config, correlationIDs, destinationDetails.AdditionalProperties)
	if err != nil {
		return err
	}
	destReqBody.AdditionalProperties = enrichedProperties

	app, err := s.applicationRepository.GetByID(ctx, formationAssignment.TenantID, formationAssignment.Source)
	if err != nil {
		return errors.Wrapf(err, "while getting application with ID: %q", formationAssignment.Source)
	}
	if app.BaseURL != nil {
		destReqBody.Audience = *app.BaseURL
	}

	if err := destReqBody.Validate(s.config); err != nil {
		return errors.Wrapf(err, "while validating SAML assertion destination request body")
	}

	log.C(ctx).Infof("Creating SAML assertion destination with name: %q, subaccount ID: %q and assignment ID: %q in the destination service", destinationName, subaccountID, formationAssignment.ID)
	_, statusCode, err := s.executeCreateRequest(ctx, strURL, destReqBody, destinationName)
	if err != nil {
		return errors.Wrapf(err, "while creating SAML assertion destination with name: %q in the destination service", destinationName)
	}

	if statusCode == http.StatusConflict {
		log.C(ctx).Infof("The destination with name: %q already exists. Will be deleted and created again...", destinationName)
		depth++
		if depth > DepthLimit {
			return errors.Errorf("Destination creator service retry limit: %d is exceeded", DepthLimit)
		}

		if err := s.DeleteDestination(ctx, destinationName, subaccountID, destinationDetails.InstanceID, formationAssignment); err != nil {
			return errors.Wrapf(err, "while deleting destination with name: %q and subaccount ID: %q", destinationName, subaccountID)
		}

		return s.CreateSAMLAssertionDestination(ctx, destinationDetails, samlAuthCreds, formationAssignment, correlationIDs, depth)
	}

	return nil
}

// CreateClientCertificateDestination is responsible to create client certificate destination resource in the DB as well as in the remote destination service
func (s *Service) CreateClientCertificateDestination(
	ctx context.Context,
	destinationDetails operators.Destination,
	clientCertAuthCreds *operators.ClientCertAuthentication,
	formationAssignment *model.FormationAssignment,
	correlationIDs []string,
	depth uint8,
) error {
	subaccountID := destinationDetails.SubaccountID
	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting region label for tenant with ID: %s", subaccountID)
	}

	strURL, err := buildDestinationURL(ctx, s.config.DestinationAPIConfig, URLParameters{
		EntityName:   "",
		Region:       region,
		SubaccountID: subaccountID,
		InstanceID:   destinationDetails.InstanceID,
	}, false)
	if err != nil {
		return errors.Wrapf(err, "while building destination URL")
	}

	certName, err := GetDestinationCertificateName(ctx, destinationcreatorpkg.AuthTypeClientCertificate, formationAssignment.ID)
	if err != nil {
		return errors.Wrapf(err, "while getting destination certificate name for destination auth type: %s", destinationcreatorpkg.AuthTypeClientCertificate)
	}

	destinationName := destinationDetails.Name
	destReqBody := &ClientCertAuthDestinationRequestBody{
		BaseDestinationRequestBody: BaseDestinationRequestBody{
			Name:               destinationName,
			URL:                clientCertAuthCreds.URL,
			Type:               destinationcreatorpkg.TypeHTTP,
			ProxyType:          destinationcreatorpkg.ProxyTypeInternet,
			AuthenticationType: destinationcreatorpkg.AuthTypeClientCertificate,
		},
		KeyStoreLocation: certName + destinationcreatorpkg.JavaKeyStoreFileExtension,
	}

	if destinationDetails.Type != "" {
		destReqBody.Type = destinationcreatorpkg.Type(destinationDetails.Type)
	}

	if destinationDetails.ProxyType != "" {
		destReqBody.ProxyType = destinationcreatorpkg.ProxyType(destinationDetails.ProxyType)
	}

	if destinationDetails.Authentication != "" && destinationcreatorpkg.AuthType(destinationDetails.Authentication) != destinationcreatorpkg.AuthTypeClientCertificate {
		return errors.Errorf("The provided authentication type: %s in the destination details is invalid. It should be %s", destinationDetails.Authentication, destinationcreatorpkg.AuthTypeClientCertificate)
	}

	enrichedProperties, err := enrichDestinationAdditionalPropertiesWithCorrelationIDs(s.config, correlationIDs, destinationDetails.AdditionalProperties)
	if err != nil {
		return err
	}
	destReqBody.AdditionalProperties = enrichedProperties

	if err := destReqBody.Validate(s.config); err != nil {
		return errors.Wrapf(err, "while validating client certificate destination request body")
	}

	log.C(ctx).Infof("Creating client certificate authentication destination with name: %q, subaccount ID: %q and assignment ID: %q in the destination service", destinationName, subaccountID, formationAssignment.ID)
	_, statusCode, err := s.executeCreateRequest(ctx, strURL, destReqBody, destinationName)
	if err != nil {
		return errors.Wrapf(err, "while creating client certificate authentication destination with name: %q in the destination service", destinationName)
	}

	if statusCode == http.StatusConflict {
		log.C(ctx).Infof("The destination with name: %q already exists. Will be deleted and created again...", destinationName)
		depth++
		if depth > DepthLimit {
			return errors.Errorf("Destination creator service retry limit: %d is exceeded", DepthLimit)
		}

		if err := s.DeleteDestination(ctx, destinationName, subaccountID, destinationDetails.InstanceID, formationAssignment); err != nil {
			return errors.Wrapf(err, "while deleting destination with name: %q and subaccount ID: %q", destinationName, subaccountID)
		}

		return s.CreateClientCertificateDestination(ctx, destinationDetails, clientCertAuthCreds, formationAssignment, correlationIDs, depth)
	}

	return nil
}

// CreateCertificate is responsible to create certificate resource in the remote destination service
func (s *Service) CreateCertificate(
	ctx context.Context,
	destinationsDetails []operators.Destination,
	destinationAuthType destinationcreatorpkg.AuthType,
	formationAssignment *model.FormationAssignment,
	depth uint8,
) (*operators.CertificateData, error) {
	if err := s.EnsureDestinationSubaccountIDsCorrectness(ctx, destinationsDetails, formationAssignment); err != nil {
		return nil, err
	}

	subaccountID := destinationsDetails[0].SubaccountID // we're getting the subaccount ID from the first element, but before that we ensured they are all equal
	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting region label for tenant with ID: %s", subaccountID)
	}

	strURL, err := buildCertificateURL(ctx, s.config.CertificateAPIConfig, URLParameters{
		EntityName:   "",
		Region:       region,
		SubaccountID: subaccountID,
		InstanceID:   destinationsDetails[0].InstanceID,
	}, false)
	if err != nil {
		return nil, errors.Wrapf(err, "while building certificate URL")
	}

	certName, err := GetDestinationCertificateName(ctx, destinationAuthType, formationAssignment.ID)
	if err != nil {
		return nil, err
	}

	certReqBody := &CertificateRequestBody{Name: certName}

	if err := certReqBody.Validate(); err != nil {
		return nil, errors.Wrapf(err, "while validating certificate request body")
	}

	log.C(ctx).Infof("Creating certificate with name: %q for subaccount with ID: %q in the destination service for %q destination", certName, subaccountID, destinationAuthType)
	respBody, statusCode, err := s.executeCreateRequest(ctx, strURL, certReqBody, certName)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating certificate with name: %q for subaccount with ID: %q in the destination service", certName, subaccountID)
	}

	if statusCode == http.StatusConflict {
		log.C(ctx).Infof("The certificate with name: %q already exists. Will be deleted and created again...", certName)
		depth++
		if depth > DepthLimit {
			return nil, errors.Errorf("Destination creator service retry limit: %d is exceeded", DepthLimit)
		}

		if err := s.DeleteCertificate(ctx, certName, subaccountID, destinationsDetails[0].InstanceID, formationAssignment); err != nil {
			return nil, errors.Wrapf(err, "while deleting certificate with name: %q and subaccount ID: %q", certName, subaccountID)
		}

		return s.CreateCertificate(ctx, destinationsDetails, destinationAuthType, formationAssignment, depth)
	}

	var certResp CertificateResponse
	err = json.Unmarshal(respBody, &certResp)
	if err != nil {
		return nil, errors.Wrap(err, "while unmarshalling certificate response")
	}

	if err := certResp.Validate(); err != nil {
		return nil, errors.Wrap(err, "while validation destination certificate data")
	}

	certData := &operators.CertificateData{
		FileName:         certResp.FileName,
		CommonName:       certResp.CommonName,
		CertificateChain: certResp.CertificateChain,
	}

	return certData, nil
}

// GetDestinationCertificateName return a certificate name based on the destination authentication type and formation assignment ID.
// And if needed to truncate the certificate name, so it can be compliant with the max destination name constraint.
func GetDestinationCertificateName(ctx context.Context, destinationAuthentication destinationcreatorpkg.AuthType, formationAssignmentID string) (string, error) {
	var certName string
	switch destinationAuthentication {
	case destinationcreatorpkg.AuthTypeSAMLAssertion:
		certName = certificateSAMLAssertionDestinationPrefix + formationAssignmentID
	case destinationcreatorpkg.AuthTypeSAMLBearerAssertion:
		certName = certificateSAMLBearerAssertionDestinationPrefix + formationAssignmentID
	case destinationcreatorpkg.AuthTypeClientCertificate:
		certName = certificateClientCertificateDestinationPrefix + formationAssignmentID
	default:
		return "", errors.Errorf("Invalid destination authentication type: %q for certificate creation", destinationAuthentication)
	}

	if len(certName) > destinationcreatorpkg.MaxDestinationNameLength {
		log.C(ctx).Warnf("The certificate name: %s is more than %d symbols. Truncating it...", certName, destinationcreatorpkg.MaxDestinationNameLength)
		certName = certName[:destinationcreatorpkg.MaxDestinationNameLength]
	}

	return certName, nil
}

// DeleteCertificate is responsible to delete certificate resource from the remote destination service
func (s *Service) DeleteCertificate(
	ctx context.Context,
	certificateName, externalDestSubaccountID, instanceID string,
	formationAssignment *model.FormationAssignment,
) error {
	subaccountID, err := s.ValidateDestinationSubaccount(ctx, externalDestSubaccountID, formationAssignment)
	if err != nil {
		return err
	}

	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting region label for tenant with ID: %s", subaccountID)
	}

	strURL, err := buildCertificateURL(ctx, s.config.CertificateAPIConfig, URLParameters{
		EntityName:   certificateName,
		Region:       region,
		SubaccountID: subaccountID,
		InstanceID:   instanceID,
	}, true)
	if err != nil {
		return errors.Wrapf(err, "while building certificate URL")
	}

	log.C(ctx).Infof("Deleting destination certificate with name: %q and subaccount ID: %q from destination service", certificateName, subaccountID)
	err = s.executeDeleteRequest(ctx, strURL, certificateName, subaccountID)
	if err != nil {
		return err
	}

	return nil
}

// DeleteDestination is responsible to delete destination resource from the remote destination service
func (s *Service) DeleteDestination(
	ctx context.Context,
	destinationName, externalDestSubaccountID, instanceID string,
	formationAssignment *model.FormationAssignment,
) error {
	subaccountID, err := s.ValidateDestinationSubaccount(ctx, externalDestSubaccountID, formationAssignment)
	if err != nil {
		return err
	}

	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting region label for tenant with ID: %s", subaccountID)
	}

	strURL, err := buildDestinationURL(ctx, s.config.DestinationAPIConfig, URLParameters{
		EntityName:   destinationName,
		Region:       region,
		SubaccountID: subaccountID,
		InstanceID:   instanceID,
	}, true)
	if err != nil {
		return errors.Wrapf(err, "while building destination URL")
	}

	log.C(ctx).Infof("Deleting destination with name: %q and subaccount ID: %q from destination service", destinationName, subaccountID)
	err = s.executeDeleteRequest(ctx, strURL, destinationName, subaccountID)
	if err != nil {
		return err
	}

	return nil
}

// EnrichAssignmentConfigWithCertificateData is responsible to enrich the assignment configuration with the created certificate data for any type of destination.
func (s *Service) EnrichAssignmentConfigWithCertificateData(
	assignmentConfig json.RawMessage,
	destinationTypePath string,
	certData *operators.CertificateData,
) (json.RawMessage, error) {
	configStr := string(assignmentConfig)

	path := fmt.Sprintf("%s.certificate", destinationTypePath)
	configStr, err := sjson.Set(configStr, path, certData.CertificateChain)
	if err != nil {
		return nil, errors.Wrapf(err, "while enriching %q with certificate chain", destinationTypePath)
	}

	return json.RawMessage(configStr), nil
}

// EnrichAssignmentConfigWithSAMLCertificateData is responsible to enrich the assignment configuration with the necessary SAML details using created certificate data
func (s *Service) EnrichAssignmentConfigWithSAMLCertificateData(
	assignmentConfig json.RawMessage,
	destinationTypePath string,
	certData *operators.CertificateData,
) (json.RawMessage, error) {
	enrichedCfg, err := s.EnrichAssignmentConfigWithCertificateData(assignmentConfig, destinationTypePath, certData)
	if err != nil {
		return nil, err
	}

	path := fmt.Sprintf("%s.assertionIssuer", destinationTypePath)
	enrichedCfgStr, err := sjson.Set(string(enrichedCfg), path, certData.CommonName)
	if err != nil {
		return nil, errors.Wrapf(err, "while enriching %q with assertion issuer", destinationTypePath)
	}

	return json.RawMessage(enrichedCfgStr), nil
}

// ValidateDestinationSubaccount validates if the subaccount ID in the destination details is provided, it's the correct/valid one.
// If it's not provided, then we validate the subaccount ID from the formation assignment.
func (s *Service) ValidateDestinationSubaccount(
	ctx context.Context,
	externalDestSubaccountID string,
	formationAssignment *model.FormationAssignment,
) (string, error) {
	var subaccountID string
	if externalDestSubaccountID == "" {
		consumerSubaccountID, err := s.GetConsumerTenant(ctx, formationAssignment)
		if err != nil {
			return "", errors.Wrap(err, "An error occurred while getting consumer subaccount")
		}
		subaccountID = consumerSubaccountID

		log.C(ctx).Infof("There was no subaccount ID provided in the destination but the consumer: %q is validated successfully", subaccountID)
		return subaccountID, nil
	}

	if externalDestSubaccountID != "" {
		consumerSubaccountID, err := s.GetConsumerTenant(ctx, formationAssignment)
		if err != nil {
			log.C(ctx).Warnf("An error occurred while getting consumer subaccount: %s", err.Error())
		}

		if consumerSubaccountID != "" && externalDestSubaccountID == consumerSubaccountID {
			log.C(ctx).Infof("Successfully validated the provided destination subaccount ID: %q is a consumer subaccount", externalDestSubaccountID)
			return consumerSubaccountID, nil
		}

		switch formationAssignment.TargetType {
		case model.FormationAssignmentTypeApplication:
			if err := s.validateAppTemplateProviderSubaccount(ctx, formationAssignment, externalDestSubaccountID); err != nil {
				return "", errors.Wrapf(err, "while validating provider subaccount: %s for %s formation assignment type", externalDestSubaccountID, model.FormationAssignmentTypeApplication)
			}
		case model.FormationAssignmentTypeRuntime:
			if err := s.validateRuntimeProviderSubaccount(ctx, formationAssignment.Target, externalDestSubaccountID); err != nil {
				return "", errors.Wrapf(err, "while validating provider subaccount: %s for %s formation assignment type", externalDestSubaccountID, model.FormationAssignmentTypeRuntime)
			}
		case model.FormationAssignmentTypeRuntimeContext:
			if err := s.validateRuntimeContextProviderSubaccount(ctx, formationAssignment, externalDestSubaccountID); err != nil {
				return "", errors.Wrapf(err, "while validating provider subaccount: %s for %s formation assignment type", externalDestSubaccountID, model.FormationAssignmentTypeRuntimeContext)
			}
		default:
			return "", errors.Errorf("Unknown formation assignment type: %q", formationAssignment.TargetType)
		}

		subaccountID = externalDestSubaccountID
	}

	return subaccountID, nil
}

// PrepareBasicRequestBody constructs a basic destination request body with all the required field
func (s *Service) PrepareBasicRequestBody(
	ctx context.Context,
	destinationDetails operators.Destination,
	basicAuthenticationCredentials operators.BasicAuthentication,
	formationAssignment *model.FormationAssignment,
	correlationIDs []string,
) (*BasicAuthDestinationRequestBody, error) {
	reqBody := &BasicAuthDestinationRequestBody{
		BaseDestinationRequestBody: BaseDestinationRequestBody{
			Name:               destinationDetails.Name,
			URL:                "",
			Type:               destinationcreatorpkg.TypeHTTP,
			ProxyType:          destinationcreatorpkg.ProxyTypeInternet,
			AuthenticationType: destinationcreatorpkg.AuthTypeBasic,
		},
		User:     basicAuthenticationCredentials.Username,
		Password: basicAuthenticationCredentials.Password,
	}

	enrichedProperties, err := enrichDestinationAdditionalPropertiesWithCorrelationIDs(s.config, correlationIDs, destinationDetails.AdditionalProperties)
	if err != nil {
		return nil, err
	}
	reqBody.AdditionalProperties = enrichedProperties

	if destinationDetails.URL != "" {
		reqBody.URL = destinationDetails.URL
	}

	if destinationDetails.URL == "" && basicAuthenticationCredentials.URL != "" {
		reqBody.URL = basicAuthenticationCredentials.URL
	}

	if destinationDetails.URL == "" && basicAuthenticationCredentials.URL == "" {
		app, err := s.applicationRepository.GetByID(ctx, formationAssignment.TenantID, formationAssignment.Target)
		if err != nil {
			return nil, err
		}
		if app.BaseURL != nil {
			reqBody.URL = *app.BaseURL
		}
	}

	if destinationDetails.Type != "" {
		reqBody.Type = destinationcreatorpkg.Type(destinationDetails.Type)
	}

	if destinationDetails.ProxyType != "" {
		reqBody.ProxyType = destinationcreatorpkg.ProxyType(destinationDetails.ProxyType)
	}

	if destinationDetails.Authentication != "" && destinationcreatorpkg.AuthType(destinationDetails.Authentication) != destinationcreatorpkg.AuthTypeBasic {
		return nil, errors.Errorf("The provided authentication type: %s in the destination details is invalid. It should be %s", destinationDetails.Authentication, destinationcreatorpkg.AuthTypeBasic)
	}

	if err := reqBody.Validate(s.config); err != nil {
		return nil, errors.Wrapf(err, "while validating basic destination request body")
	}

	return reqBody, nil
}

// GetConsumerTenant is responsible to retrieve the "consumer" tenant ID from the given `formationAssignment`
func (s *Service) GetConsumerTenant(ctx context.Context, formationAssignment *model.FormationAssignment) (string, error) {
	labelableObjType, err := determineLabelableObjectType(formationAssignment.TargetType)
	if err != nil {
		return "", err
	}

	labels, err := s.labelRepo.ListForObject(ctx, formationAssignment.TenantID, labelableObjType, formationAssignment.Target)
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

// EnsureDestinationSubaccountIDsCorrectness is responsible to populate all the subaccount IDs in the destination details if they are not provided, and after that to check if they are all equal.
func (s *Service) EnsureDestinationSubaccountIDsCorrectness(ctx context.Context, destinationsDetails []operators.Destination, formationAssignment *model.FormationAssignment) error {
	if err := s.sanitizeDestinationSubaccountIDs(ctx, destinationsDetails, formationAssignment); err != nil {
		return errors.Wrap(err, "while sanitizing destination subaccount IDs")
	}

	destSubaccountIDs := make(map[string]struct{})
	for _, destDetails := range destinationsDetails {
		if destDetails.SubaccountID != "" {
			destSubaccountIDs[destDetails.SubaccountID] = struct{}{}
		}
	}

	if len(destSubaccountIDs) != 1 {
		err := errors.Errorf("All subaccount in the destinations details should have one and the same subaccount ID. Currently there is/are: %d subaacount ID(s)", len(destSubaccountIDs))
		log.C(ctx).Error(err)
		return err
	}

	return nil
}

// sanitizeDestinationSubaccountIDs is responsible to populate the subaccount ID of every destination element in the slice if it's not provided
func (s *Service) sanitizeDestinationSubaccountIDs(ctx context.Context, destinationsDetails []operators.Destination, formationAssignment *model.FormationAssignment) error {
	for i, destDetails := range destinationsDetails {
		if destDetails.SubaccountID == "" {
			subaccID, err := s.ValidateDestinationSubaccount(ctx, "", formationAssignment)
			if err != nil {
				return err
			}
			destinationsDetails[i].SubaccountID = subaccID
		}
	}

	return nil
}

func (s *Service) getRegionLabel(ctx context.Context, tenantID string) (string, error) {
	t, err := s.tenantRepo.GetByExternalTenant(ctx, tenantID)
	if err != nil {
		return "", errors.Wrapf(err, "while getting tenant by external ID: %q", tenantID)
	}

	regionLbl, err := s.labelRepo.GetByKey(ctx, t.ID, model.TenantLabelableObject, tenantID, RegionLabelKey)
	if err != nil {
		return "", err
	}

	region, ok := regionLbl.Value.(string)
	if !ok {
		return "", errors.Errorf("unexpected type of %q label, expect: string, got: %T", RegionLabelKey, regionLbl.Value)
	}
	return region, nil
}

func (s *Service) validateAppTemplateProviderSubaccount(ctx context.Context, formationAssignment *model.FormationAssignment, externalDestSubaccountID string) error {
	app, err := s.applicationRepository.GetByID(ctx, formationAssignment.TenantID, formationAssignment.Target)
	if err != nil {
		return err
	}

	if app.ApplicationTemplateID == nil || *app.ApplicationTemplateID == "" {
		return errors.Errorf("The application template ID for application ID: %q should not be empty", app.ID)
	}

	labels, err := s.labelRepo.ListForGlobalObject(ctx, model.AppTemplateLabelableObject, *app.ApplicationTemplateID)
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

	if externalDestSubaccountID != subaccountLblValue {
		return errors.Errorf("The provided destination subaccount is different from the owner subaccount of the application template with ID: %q", *app.ApplicationTemplateID)
	}

	log.C(ctx).Infof("Successfully validated that the provided destination subaccount: %q is a provider one - the owner of the application template", externalDestSubaccountID)

	return nil
}

func (s *Service) validateRuntimeProviderSubaccount(ctx context.Context, runtimeID, externalDestSubaccountID string) error {
	t, err := s.tenantRepo.GetByExternalTenant(ctx, externalDestSubaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting tenant by external ID: %q", externalDestSubaccountID)
	}

	if t.Type != tenant.Subaccount {
		return errors.Errorf("The provided destination external tenant ID: %q has invalid type, expected: %q, got: %q", externalDestSubaccountID, tenant.Subaccount, t.Type)
	}

	exists, err := s.runtimeRepository.OwnerExists(ctx, t.ID, runtimeID)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("The provided destination external subaccount: %q is not provider of the runtime with ID: %q", externalDestSubaccountID, runtimeID)
	}

	log.C(ctx).Infof("Successfully validated that the provided destination external subaccount: %q is a provider one - the owner of the runtime", externalDestSubaccountID)

	return nil
}

func (s *Service) validateRuntimeContextProviderSubaccount(ctx context.Context, formationAssignment *model.FormationAssignment, externalDestSubaccountID string) error {
	rtmCtx, err := s.runtimeCtxRepository.GetByID(ctx, formationAssignment.TenantID, formationAssignment.Target)
	if err != nil {
		return err
	}

	return s.validateRuntimeProviderSubaccount(ctx, rtmCtx.RuntimeID, externalDestSubaccountID)
}

func (s *Service) executeCreateRequest(
	ctx context.Context,
	url string,
	reqBody interface{},
	entityName string,
) (defaultRespBody []byte, defaultStatusCode int, err error) {
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return defaultRespBody, defaultStatusCode, errors.Wrapf(err, "while marshalling request body")
	}

	clientUser, err := client.LoadFromContext(ctx)
	if err != nil || clientUser == "" {
		log.C(ctx).Warn("unable to provide client_user. Using correlation ID as client_user header...")
		clientUser = correlation.CorrelationIDFromContext(ctx)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return defaultRespBody, defaultStatusCode, errors.Wrap(err, "while preparing destination service creation request")
	}
	req.Header.Set(clientUserHeaderKey, clientUser)
	req.Header.Set(contentTypeHeaderKey, contentTypeApplicationJSON)

	req = req.WithContext(ctx)
	resp, err := s.mtlsHTTPClient.Do(req)
	if err != nil {
		return defaultRespBody, defaultStatusCode, err
	}
	defer closeResponseBody(ctx, resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return defaultRespBody, defaultStatusCode, errors.Errorf("Failed to read response body: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return body, defaultStatusCode, errors.Errorf("Failed to create entity with name: %q, status: %d, body: %s", entityName, resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusConflict {
		log.C(ctx).Infof("The entity with name: %q already exists in the destination service. Returning conflict status code...", entityName)
		return body, http.StatusConflict, nil
	}
	log.C(ctx).Infof("Successfully created entity with name: %q in the destination service", entityName)

	return body, http.StatusCreated, nil
}

func (s *Service) executeDeleteRequest(ctx context.Context, url string, entityName, subaccountID string) error {
	clientUser, err := client.LoadFromContext(ctx)
	if err != nil || clientUser == "" {
		log.C(ctx).Warn("unable to provide client_user. Using correlation ID as client_user header...")
		clientUser = correlation.CorrelationIDFromContext(ctx)
	}

	req, err := http.NewRequest(http.MethodDelete, url, nil)
	if err != nil {
		return errors.Wrap(err, "while preparing destination service deletion request")
	}
	req.Header.Set(clientUserHeaderKey, clientUser)
	req.Header.Set(contentTypeHeaderKey, contentTypeApplicationJSON)

	req = req.WithContext(ctx)
	resp, err := s.mtlsHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return errors.Errorf("Failed to read destination delete response body: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("Failed to delete entity with name: %q from destination service, status: %d, body: %s", entityName, resp.StatusCode, body)
	}

	log.C(ctx).Infof("Successfully deleted entity with name: %q and subaccount ID: %q from destination service", entityName, subaccountID)

	return nil
}

func enrichDestinationAdditionalPropertiesWithCorrelationIDs(
	destinationCreatorCfg *Config,
	correlationIDs []string,
	destinationAdditionalProperties json.RawMessage,
) (json.RawMessage, error) {
	joinedCorrelationIDs := strings.Join(correlationIDs, ",")
	additionalProps, err := sjson.Set(string(destinationAdditionalProperties), destinationCreatorCfg.CorrelationIDsKey, joinedCorrelationIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "while setting the correlation IDs as additional properties of the destination")
	}

	return json.RawMessage(additionalProps), nil
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

func closeResponseBody(ctx context.Context, resp *http.Response) {
	if err := resp.Body.Close(); err != nil {
		log.C(ctx).Errorf("An error has occurred while closing response body: %v", err)
	}
}

func buildDestinationURL(ctx context.Context, destinationCfg *DestinationAPIConfig, parameters URLParameters, isDeleteRequest bool) (string, error) {
	return buildURL(ctx, URLConfig{
		BaseURL:             destinationCfg.BaseURL,
		SubaccountLevelPath: destinationCfg.SubaccountLevelPath,
		InstanceLevelPath:   destinationCfg.InstanceLevelPath,
		RegionParam:         destinationCfg.RegionParam,
		NameParam:           destinationCfg.DestinationNameParam,
		SubaccountIDParam:   destinationCfg.SubaccountIDParam,
		InstanceIDParam:     destinationCfg.InstanceIDParam,
	}, parameters, isDeleteRequest)
}

func buildCertificateURL(ctx context.Context, certificateCfg *CertificateAPIConfig, parameters URLParameters, isDeleteRequest bool) (string, error) {
	return buildURL(ctx, URLConfig{
		BaseURL:             certificateCfg.BaseURL,
		SubaccountLevelPath: certificateCfg.SubaccountLevelPath,
		InstanceLevelPath:   certificateCfg.InstanceLevelPath,
		RegionParam:         certificateCfg.RegionParam,
		NameParam:           certificateCfg.CertificateNameParam,
		SubaccountIDParam:   certificateCfg.SubaccountIDParam,
		InstanceIDParam:     certificateCfg.InstanceIDParam,
	}, parameters, isDeleteRequest)
}

func buildURL(ctx context.Context, urlConfig URLConfig, parameters URLParameters, isDeleteRequest bool) (string, error) {
	if parameters.Region == "" || parameters.SubaccountID == "" {
		return "", errors.Errorf("The provided region and/or subaccount for the URL couldn't be empty")
	}

	var path string
	if parameters.InstanceID == "" {
		path = urlConfig.SubaccountLevelPath
	} else {
		log.C(ctx).Infof("The entity with name: %q is on service instance level with ID: %q", parameters.EntityName, parameters.InstanceID)
		path = urlConfig.InstanceLevelPath
	}

	base, err := url.Parse(urlConfig.BaseURL)
	if err != nil {
		return "", err
	}

	regionalEndpoint := strings.Replace(path, fmt.Sprintf("{%s}", urlConfig.RegionParam), parameters.Region, 1)
	regionalEndpoint = strings.Replace(regionalEndpoint, fmt.Sprintf("{%s}", urlConfig.SubaccountIDParam), parameters.SubaccountID, 1)

	if isDeleteRequest {
		if parameters.EntityName == "" {
			return "", errors.Errorf("The entity name should not be empty in case of %s request", http.MethodDelete)
		}
		regionalEndpoint += fmt.Sprintf("/{%s}", urlConfig.NameParam)
		regionalEndpoint = strings.Replace(regionalEndpoint, fmt.Sprintf("{%s}", urlConfig.NameParam), parameters.EntityName, 1)
	}

	if parameters.InstanceID != "" {
		regionalEndpoint = strings.Replace(regionalEndpoint, fmt.Sprintf("{%s}", urlConfig.InstanceIDParam), parameters.InstanceID, 1)
	}

	// Path params
	base.Path += regionalEndpoint

	return base.String(), nil
}
