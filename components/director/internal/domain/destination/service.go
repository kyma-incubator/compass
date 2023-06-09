package destination

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/domain/client"
	"github.com/kyma-incubator/compass/components/director/internal/domain/destination/destinationcreator"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/pkg/errors"
)

const (
	clientUserHeaderKey        = "CLIENT_USER"
	contentTypeHeaderKey       = "Content-Type"
	contentTypeApplicationJson = "application/json;charset=UTF-8"
	globalSubaccountLabelKey   = "global_subaccount_id"
	regionLabelKey             = "region"
)

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

//go:generate mockery --exported --name=destinationRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type destinationRepository interface {
	DeleteByTenantIDAndAssignmentID(ctx context.Context, tenantID, formationAssignmentID string) error
	ListByTenantIDAndAssignmentID(ctx context.Context, tenantID, formationAssignmentID string) ([]*model.Destination, error)
	UpsertWithEmbeddedTenant(ctx context.Context, destination *model.Destination) error
}

//go:generate mockery --exported --name=tenantRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type tenantRepository interface {
	GetByExternalTenant(ctx context.Context, externalTenant string) (*model.BusinessTenantMapping, error)
}

// UIDService generates UUIDs for new entities
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// Service todo::: add godoc
type Service struct {
	mtlsHTTPClient        *http.Client
	destinationCreatorCfg *destinationcreator.Config
	transact              persistence.Transactioner
	applicationRepository applicationRepository
	runtimeRepository     runtimeRepository
	runtimeCtxRepository  runtimeCtxRepository
	labelRepo             labelRepository
	destinationRepo       destinationRepository
	tenantRepo            tenantRepository
	uidSvc                UIDService
}

// NewService todo::: add godoc
func NewService(mtlsHTTPClient *http.Client, destinationCreatorCfg *destinationcreator.Config, transact persistence.Transactioner, applicationRepository applicationRepository, runtimeRepository runtimeRepository, runtimeCtxRepository runtimeCtxRepository, labelRepo labelRepository, destinationRepository destinationRepository, tenantRepository tenantRepository, uidSvc UIDService) *Service {
	return &Service{
		mtlsHTTPClient:        mtlsHTTPClient,
		destinationCreatorCfg: destinationCreatorCfg,
		transact:              transact,
		applicationRepository: applicationRepository,
		runtimeRepository:     runtimeRepository,
		runtimeCtxRepository:  runtimeCtxRepository,
		labelRepo:             labelRepo,
		destinationRepo:       destinationRepository,
		tenantRepo:            tenantRepository,
		uidSvc:                uidSvc,
	}
}

// CreateBasicCredentialDestinations todo::: add godoc
func (s *Service) CreateBasicCredentialDestinations(ctx context.Context, destinationDetails operators.Destination, basicAuthenticationCredentials operators.BasicAuthentication, formationAssignment *webhook.FormationAssignment) (defaultStatusCode int, err error) {
	subaccountID, err := s.validateDestinationSubaccount(ctx, destinationDetails.SubaccountID, formationAssignment)
	if err != nil {
		return defaultStatusCode, err
	}

	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return defaultStatusCode, err
	}

	strURL, err := buildURL(s.destinationCreatorCfg, region, subaccountID, "", false)
	if err != nil {
		return defaultStatusCode, errors.Wrapf(err, "while building destination URL")
	}

	reqBody, err := s.prepareRequestBody(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment)
	if err != nil {
		return defaultStatusCode, err
	}

	destinationName := destinationDetails.Name
	log.C(ctx).Infof("Creating inbound basic destination with name: %q, subaccount ID: %q and assignment ID: %q in the destination service", destinationName, subaccountID, formationAssignment.ID)
	statusCode, err := s.executeCreateRequest(ctx, strURL, reqBody, destinationName)
	if err != nil {
		return defaultStatusCode, errors.Wrapf(err, "while creating inbound basic destination with name: %q in the destination service", destinationName)
	}

	if statusCode == http.StatusConflict {
		return statusCode, nil
	}

	if transactionErr := s.transaction(ctx, func(ctxWithTransact context.Context) error {
		t, err := s.tenantRepo.GetByExternalTenant(ctx, subaccountID)
		if err != nil {
			return errors.Wrapf(err, "while getting tenant by external ID: %q", subaccountID)
		}

		destModel := &model.Destination{
			ID:                    s.uidSvc.Generate(),
			Name:                  reqBody.Name,
			Type:                  reqBody.Type,
			URL:                   reqBody.URL,
			Authentication:        reqBody.AuthenticationType,
			SubaccountID:          t.ID,
			FormationAssignmentID: &formationAssignment.ID,
		}

		if err = s.destinationRepo.UpsertWithEmbeddedTenant(ctx, destModel); err != nil {
			return errors.Wrapf(err, "while upserting basic destination with name: %q and assignment ID: %q in the DB", destinationName, formationAssignment.ID)
		}
		return nil
	}); transactionErr != nil {
		return defaultStatusCode, transactionErr
	}

	return statusCode, nil
}

// CreateDesignTimeDestinations todo:: go doc
func (s *Service) CreateDesignTimeDestinations(ctx context.Context, destinationDetails operators.Destination, formationAssignment *webhook.FormationAssignment) (defaultStatusCode int, err error) {
	subaccountID, err := s.validateDestinationSubaccount(ctx, destinationDetails.SubaccountID, formationAssignment)
	if err != nil {
		return defaultStatusCode, err
	}

	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return defaultStatusCode, err
	}

	strURL, err := buildURL(s.destinationCreatorCfg, region, subaccountID, "", false)
	if err != nil {
		return defaultStatusCode, errors.Wrapf(err, "while building destination URL")
	}

	destinationName := destinationDetails.Name
	destReqBody := &destinationcreator.NoAuthRequestBody{
		BaseDestinationRequestBody: destinationcreator.BaseDestinationRequestBody{
			Name:                 destinationDetails.Name,
			URL:                  destinationDetails.URL,
			Type:                 destinationDetails.Type,
			ProxyType:            destinationDetails.ProxyType,
			AuthenticationType:   destinationDetails.Authentication,
			AdditionalAttributes: destinationDetails.AdditionalAttributes,
		},
	}

	if err := destReqBody.Validate(); err != nil {
		return defaultStatusCode, errors.Wrapf(err, "while validating no authentication destination request body")
	}

	log.C(ctx).Infof("Creating design time destination with name: %q, subaccount ID: %q and assignment ID: %q in the destination service", destinationName, subaccountID, formationAssignment.ID)
	statusCode, err := s.executeCreateRequest(ctx, strURL, destReqBody, destinationName)
	if err != nil {
		return defaultStatusCode, errors.Wrapf(err, "while creating design time destination with name: %q in the destination service", destinationName)
	}

	t, err := s.tenantRepo.GetByExternalTenant(ctx, subaccountID)
	if err != nil {
		return defaultStatusCode, errors.Wrapf(err, "while getting tenant by external ID: %q", subaccountID)
	}

	destModel := &model.Destination{
		ID:                    s.uidSvc.Generate(),
		Name:                  destReqBody.Name,
		Type:                  destReqBody.Type,
		URL:                   destReqBody.URL,
		Authentication:        destReqBody.AuthenticationType,
		SubaccountID:          t.ID,
		FormationAssignmentID: &formationAssignment.ID,
	}

	if transactionErr := s.transaction(ctx, func(ctxWithTransact context.Context) error {
		if err = s.destinationRepo.UpsertWithEmbeddedTenant(ctx, destModel); err != nil {
			return errors.Wrapf(err, "while upserting basic destination with name: %q and assignment ID: %q in the DB", destinationName, formationAssignment.ID)
		}
		return nil
	}); transactionErr != nil {
		return defaultStatusCode, transactionErr
	}

	return statusCode, nil
}

// DeleteDestinations todo::: go doc
func (s *Service) DeleteDestinations(ctx context.Context, formationAssignment *webhook.FormationAssignment) error {
	externalDestSubaccountID, err := s.getConsumerTenant(ctx, formationAssignment)
	if err != nil {
		return err
	}

	formationAssignmentID := formationAssignment.ID

	t, err := s.tenantRepo.GetByExternalTenant(ctx, externalDestSubaccountID)
	if err != nil {
		return errors.Wrapf(err, "while getting tenant by external ID: %q", externalDestSubaccountID)
	}

	destinations, err := s.destinationRepo.ListByTenantIDAndAssignmentID(ctx, t.ID, formationAssignmentID)
	if err != nil {
		return err
	}

	log.C(ctx).Infof("There is/are %d destination(s) in the DB", len(destinations))
	if len(destinations) == 0 {
		return nil
	}

	for _, destination := range destinations {
		if err := s.DeleteDestinationFromDestinationService(ctx, destination.Name, externalDestSubaccountID, formationAssignment); err != nil {
			return err
		}
	}

	if transactionErr := s.transaction(ctx, func(ctxWithTransact context.Context) error {
		if err := s.destinationRepo.DeleteByTenantIDAndAssignmentID(ctx, t.ID, formationAssignmentID); err != nil {
			return errors.Wrapf(err, "while deleting destination(s) by internal tenant ID: %q and assignment ID: %q from the DB", t.ID, formationAssignmentID)
		}
		return nil
	}); transactionErr != nil {
		return transactionErr
	}

	return nil
}

// DeleteDestinationFromDestinationService todo::: go doc
func (s *Service) DeleteDestinationFromDestinationService(ctx context.Context, destinationName, externalDestSubaccountID string, formationAssignment *webhook.FormationAssignment) error {
	subaccountID, err := s.validateDestinationSubaccount(ctx, externalDestSubaccountID, formationAssignment)
	if err != nil {
		return err
	}

	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return err
	}

	strURL, err := buildURL(s.destinationCreatorCfg, region, subaccountID, destinationName, true)
	if err != nil {
		return errors.Wrapf(err, "while building destination URL")
	}

	clientUser, err := client.LoadFromContext(ctx)
	if err != nil || clientUser == "" {
		log.C(ctx).Warn("unable to provide client_user. Using correlation ID as client_user header...")
		clientUser = correlation.CorrelationIDFromContext(ctx)
	}

	req, err := http.NewRequest(http.MethodDelete, strURL, nil)
	if err != nil {
		return errors.Wrap(err, "while preparing request for deleting destination from destination service")
	}
	req.Header.Set(clientUserHeaderKey, clientUser)
	req.Header.Set(contentTypeHeaderKey, contentTypeApplicationJson)

	log.C(ctx).Infof("Deleting destination with name: %q and subaccount ID: %q from destination service", destinationName, subaccountID)
	resp, err := s.mtlsHTTPClient.Do(req)
	if err != nil {
		return err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return errors.Errorf("Failed to read destination delete response body: %v", err)
	}

	if resp.StatusCode != http.StatusNoContent {
		return errors.Errorf("Failed to delete destination from destination service, status: %d, body: %s", resp.StatusCode, body)
	}

	log.C(ctx).Infof("Successfully deleted destination with name: %q and subaccount ID: %q from destination service", destinationName, subaccountID)

	return nil
}

func (s *Service) validateDestinationSubaccount(ctx context.Context, externalDestSubaccountID string, formationAssignment *webhook.FormationAssignment) (string, error) {
	var subaccountID string
	if externalDestSubaccountID == "" {
		consumerSubaccountID, err := s.getConsumerTenant(ctx, formationAssignment)
		if err != nil {
			return "", err
		}
		subaccountID = consumerSubaccountID

		log.C(ctx).Info("There was no subaccount ID provided in the destination but the consumer is validated successfully")
		return subaccountID, nil
	}

	if externalDestSubaccountID != "" {
		consumerSubaccountID, err := s.getConsumerTenant(ctx, formationAssignment)
		if err != nil {
			log.C(ctx).Warnf("Couldn't validate the if the provided destination subaccount ID: %q is a consumer subaccount. Validating if it's a provider one...", externalDestSubaccountID)
		}

		if consumerSubaccountID != "" && externalDestSubaccountID == consumerSubaccountID {
			log.C(ctx).Infof("Successfully validated the provided destination subaccount ID: %q is a consumer subaccount", externalDestSubaccountID)
			return consumerSubaccountID, nil
		}

		switch formationAssignment.TargetType {
		case model.FormationAssignmentTypeApplication:
			if err := s.validateAppTemplateProviderSubaccount(ctx, formationAssignment, externalDestSubaccountID); err != nil {
				return "", err
			}
		case model.FormationAssignmentTypeRuntime:
			if err := s.validateRuntimeProviderSubaccount(ctx, formationAssignment.Target, externalDestSubaccountID); err != nil {
				return "", err
			}
		case model.FormationAssignmentTypeRuntimeContext:
			if err := s.validateRuntimeContextProviderSubaccount(ctx, formationAssignment, externalDestSubaccountID); err != nil {
				return "", err
			}
		default:
			return "", errors.Errorf("Unknown formation assignment type: %q", formationAssignment.TargetType)
		}

		subaccountID = externalDestSubaccountID
	}

	return subaccountID, nil
}

func (s *Service) getConsumerTenant(ctx context.Context, formationAssignment *webhook.FormationAssignment) (string, error) {
	labelableObjType, err := determineLabelableObjectType(formationAssignment.TargetType)
	if err != nil {
		return "", err
	}

	labels, err := s.labelRepo.ListForObject(ctx, formationAssignment.TenantID, labelableObjType, formationAssignment.Target)
	if err != nil {
		return "", errors.Wrapf(err, "while getting labels for %s with ID: %q", formationAssignment.TargetType, formationAssignment.Target)
	}

	globalSubaccIDLbl, globalSubaccIDExists := labels[globalSubaccountLabelKey]
	if !globalSubaccIDExists {
		return "", errors.Errorf("%q label does not exists for: %q with ID: %q", globalSubaccountLabelKey, formationAssignment.TargetType, formationAssignment.Target)
	}

	globalSubaccIDLblValue, ok := globalSubaccIDLbl.Value.(string)
	if !ok {
		return "", errors.Errorf("unexpected type of %q label, expect: string, got: %T", globalSubaccountLabelKey, globalSubaccIDLbl.Value)
	}

	return globalSubaccIDLblValue, nil
}

func (s *Service) getRegionLabel(ctx context.Context, tenantID string) (string, error) {
	t, err := s.tenantRepo.GetByExternalTenant(ctx, tenantID)
	if err != nil {
		return "", errors.Wrapf(err, "while getting tenant by external ID: %q", tenantID)
	}

	regionLbl, err := s.labelRepo.GetByKey(ctx, t.ID, model.TenantLabelableObject, tenantID, regionLabelKey)
	if err != nil {
		return "", err
	}

	region, ok := regionLbl.Value.(string)
	if !ok {
		return "", errors.Errorf("unexpected type of %q label, expect: string, got: %T", regionLabelKey, regionLbl.Value)
	}
	return region, nil
}

func (s *Service) validateAppTemplateProviderSubaccount(ctx context.Context, formationAssignment *webhook.FormationAssignment, externalDestSubaccountID string) error {
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

	subaccountLbl, subaccountLblExists := labels[globalSubaccountLabelKey]

	if !subaccountLblExists {
		return errors.Errorf("%q label should exist as part of the provider application template with ID: %q", globalSubaccountLabelKey, *app.ApplicationTemplateID)
	}

	subaccountLblValue, ok := subaccountLbl.Value.(string)
	if !ok {
		return errors.Errorf("unexpected type of %q label, expect: string, got: %T", globalSubaccountLabelKey, subaccountLbl.Value)
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

func (s *Service) validateRuntimeContextProviderSubaccount(ctx context.Context, formationAssignment *webhook.FormationAssignment, externalDestSubaccountID string) error {
	rtmCtxID, err := s.runtimeCtxRepository.GetByID(ctx, formationAssignment.TenantID, formationAssignment.Target)
	if err != nil {
		return err
	}

	return s.validateRuntimeProviderSubaccount(ctx, rtmCtxID.RuntimeID, externalDestSubaccountID)
}

func (s *Service) transaction(ctx context.Context, dbCall func(ctxWithTransact context.Context) error) error {
	tx, err := s.transact.Begin()
	if err != nil {
		log.C(ctx).WithError(err).Error("Failed to begin DB transaction")
		return err
	}
	defer s.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err = dbCall(ctx); err != nil {
		return err
	}

	if err = tx.Commit(); err != nil {
		log.C(ctx).WithError(err).Error("Failed to commit database transaction")
		return err
	}
	return nil
}

func (s *Service) executeCreateRequest(ctx context.Context, url string, reqBody interface{}, destinationName string) (defaultStatusCode int, err error) {
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return defaultStatusCode, errors.Wrapf(err, "while marshalling destination request body")
	}

	clientUser, err := client.LoadFromContext(ctx)
	if err != nil || clientUser == "" {
		log.C(ctx).Warn("unable to provide client_user. Using correlation ID as client_user header...")
		clientUser = correlation.CorrelationIDFromContext(ctx)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBodyBytes))
	if err != nil {
		return defaultStatusCode, errors.Wrap(err, "while preparing request for creation of destination in destination service")
	}
	req.Header.Set(clientUserHeaderKey, clientUser)
	req.Header.Set(contentTypeHeaderKey, contentTypeApplicationJson)

	resp, err := s.mtlsHTTPClient.Do(req)
	if err != nil {
		return defaultStatusCode, err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return defaultStatusCode, errors.Errorf("Failed to read destination response body: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return defaultStatusCode, errors.Errorf("Failed to create destination with name: %q, status: %d, body: %s", destinationName, resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusConflict {
		return http.StatusConflict, nil
	}
	log.C(ctx).Infof("Successfully created destination with name: %q in the destination service", destinationName)

	return http.StatusCreated, nil
}

func (s *Service) prepareRequestBody(ctx context.Context, destinationDetails operators.Destination, basicAuthenticationCredentials operators.BasicAuthentication, formationAssignment *webhook.FormationAssignment) (*destinationcreator.BasicRequestBody, error) {
	reqBody := &destinationcreator.BasicRequestBody{
		BaseDestinationRequestBody: destinationcreator.BaseDestinationRequestBody{
			Name:                 destinationDetails.Name,
			URL:                  "",
			Type:                 destinationcreator.TypeHTTP,
			ProxyType:            destinationcreator.ProxyTypeInternet,
			AuthenticationType:   destinationcreator.AuthTypeBasic,
			AdditionalAttributes: destinationDetails.AdditionalAttributes,
		},
		User:                 basicAuthenticationCredentials.Username,
		Password:             basicAuthenticationCredentials.Password,
	}

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
		reqBody.Type = destinationDetails.Type
	}

	if destinationDetails.ProxyType != "" {
		reqBody.ProxyType = destinationDetails.ProxyType
	}

	if destinationDetails.Authentication != destinationcreator.AuthTypeBasic {
		return nil, errors.Errorf("The provided authentication type is invalid: %s. It should be %s", reqBody.AuthenticationType, destinationcreator.AuthTypeBasic)
	}

	if err := reqBody.Validate(); err != nil {
		return nil, errors.Wrapf(err, "while validating basic destination request body")
	}

	return reqBody, nil
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

func buildURL(destinationCreatorCfg *destinationcreator.Config, region, subaccountID, destinationName string, isDeleteRequest bool) (string, error) {
	if region == "" || subaccountID == "" {
		return "", errors.Errorf("The provided region and/or subaccount for the URL couldn't be empty")
	}

	base, err := url.Parse(destinationCreatorCfg.BaseURL)
	if err != nil {
		return "", err
	}

	path := destinationCreatorCfg.Path

	regionalEndpoint := strings.Replace(path, fmt.Sprintf("{%s}", destinationCreatorCfg.RegionParam), region, 1)
	regionalEndpoint = strings.Replace(regionalEndpoint, fmt.Sprintf("{%s}", destinationCreatorCfg.SubaccountIDParam), subaccountID, 1)

	if isDeleteRequest {
		if destinationName == "" {
			return "", errors.Errorf("The destination name should not be empty in case of %s request", http.MethodDelete)
		}
		regionalEndpoint += fmt.Sprintf("/{%s}", destinationCreatorCfg.DestinationNameParam)
		regionalEndpoint = strings.Replace(regionalEndpoint, fmt.Sprintf("{%s}", destinationCreatorCfg.DestinationNameParam), destinationName, 1)
	}

	// Path params
	base.Path += regionalEndpoint

	return base.String(), nil
}
