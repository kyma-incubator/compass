package destination

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/internal/domain/client"
	"github.com/kyma-incubator/compass/components/director/internal/domain/destination/destinationcreator"
	"github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/correlation"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
	"github.com/pkg/errors"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
)

const (
	clientUserHeaderKey      = "CLIENT_USER"
	globalSubaccountLabelKey = "global_subaccount_id"
	regionLabelKey           = "region"
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
	DeleteByAssignmentID(ctx context.Context, destinationName, tenantID, formationAssignmentID string) error
	CreateDestination(ctx context.Context, destination *model.Destination) error
}

// UIDService generates UUIDs for new entities
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type Service struct {
	mtlsHTTPClient        *http.Client
	destinationCreatorCfg *destinationcreator.Config
	transact              persistence.Transactioner
	applicationRepository applicationRepository
	runtimeRepository     runtimeRepository
	runtimeCtxRepository  runtimeCtxRepository
	labelRepo             labelRepository
	destinationRepo       destinationRepository
	uidSvc                UIDService
}

func NewService(mtlsHTTPClient *http.Client, destinationCreatorCfg *destinationcreator.Config, transact persistence.Transactioner, applicationRepository applicationRepository, runtimeRepository runtimeRepository, runtimeCtxRepository runtimeCtxRepository, labelRepo labelRepository, destinationRepository destinationRepository, uidSvc UIDService) *Service {
	return &Service{
		mtlsHTTPClient:        mtlsHTTPClient,
		destinationCreatorCfg: destinationCreatorCfg,
		transact:              transact,
		applicationRepository: applicationRepository,
		runtimeRepository:     runtimeRepository,
		runtimeCtxRepository:  runtimeCtxRepository,
		labelRepo:             labelRepo,
		destinationRepo:       destinationRepository,
		uidSvc:                uidSvc,
	}
}

func (s *Service) CreateBasicCredentialDestinations(ctx context.Context, destinationDetails operators.Destination, basicAuthenticationCredentials operators.BasicAuthentication, formationAssignment *webhook.FormationAssignment) (statusCode int, err error) {
	subaccountID, err := s.validateDestinationSubaccount(ctx, destinationDetails.SubaccountID, formationAssignment)
	if err != nil {
		return statusCode, err
	}

	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return statusCode, err
	}

	strURL, err := buildURL(s.destinationCreatorCfg, region, subaccountID, "", false)
	if err != nil {
		return statusCode, errors.Wrapf(err, "while building destination URL")
	}

	reqBody, err := s.prepareRequestBody(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment)
	if err != nil {
		return statusCode, err
	}

	destinationName := destinationDetails.Name
	log.C(ctx).Infof("Creating inbound basic destination with name: %q in the destination service", destinationName)
	statusCode, err = s.executeCreateRequest(ctx, strURL, reqBody, destinationName)
	if err != nil {
		return statusCode, errors.Wrapf(err, "while creating inbound basic destination with name: %q in the destination service", destinationName)
	}

	if transactionErr := s.transaction(ctx, func(ctxWithTransact context.Context) error {
		destModel := &model.Destination{
			ID:                    s.uidSvc.Generate(),
			Name:                  reqBody.Name,
			Type:                  reqBody.Type,
			Url:                   reqBody.Url,
			Authentication:        reqBody.AuthenticationType,
			SubaccountID:          subaccountID,
			FormationAssignmentID: &formationAssignment.ID,
			Revision:              s.uidSvc.Generate(),
		}

		if err = s.destinationRepo.CreateDestination(ctx, destModel); err != nil {
			return errors.Wrapf(err, "while creating destination with name: %q and assignment ID: %q in the DB", destinationName, formationAssignment.ID)
		}
		return nil
	}); transactionErr != nil {
		return statusCode, transactionErr
	}

	log.C(ctx).Infof("Successfully create inbound basic destination with name: %q and assignment ID: %q in the DB", destinationName, formationAssignment.ID)

	return statusCode, nil
}

// todo:: will be implemented with the second phase of the destination operator. Uncomment and make the needed adaptation
func (s *Service) CreateDesignTimeDestinations(ctx context.Context, destinationDetails operators.Destination, formationAssignment *webhook.FormationAssignment) (statusCode int, err error) {
	subaccountID, err := s.validateDestinationSubaccount(ctx, destinationDetails.SubaccountID, formationAssignment)
	if err != nil {
		return statusCode, err
	}

	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return statusCode, err
	}

	strURL, err := buildURL(s.destinationCreatorCfg, region, subaccountID, "", false)
	if err != nil {
		return statusCode, errors.Wrapf(err, "while building destination URL")
	}

	destinationName := destinationDetails.Name
	destReqBody := &destinationcreator.RequestBody{
		Name: destinationName,
	}

	if err := validate(destReqBody); err != nil {
		return statusCode, errors.Wrapf(err, "while validating destination request body")
	}

	log.C(ctx).Infof("Creating design time destination with name: %q in the destination service", destinationName)
	statusCode, err = s.executeCreateRequest(ctx, strURL, destReqBody, destinationName)
	if err != nil {
		return statusCode, errors.Wrapf(err, "while creating design time destination with name: %q in the destination service", destinationName)
	}

	if transactionErr := s.transaction(ctx, func(ctxWithTransact context.Context) error {
		if err = s.destinationRepo.CreateDestination(ctx, nil); err != nil { // todo:: adapt with the second phase
			return errors.Wrapf(err, "while creating destination with name: %q and assignment ID: %q in the DB", destinationName, formationAssignment.ID)
		}
		return nil
	}); transactionErr != nil {
		return statusCode, transactionErr
	}
	log.C(ctx).Infof("Successfully create design time destination with name: %q and assignment ID: %q in the DB", destinationName, formationAssignment.ID)

	return statusCode, nil
}

func (s *Service) DeleteDestinations(ctx context.Context, destinationDetails operators.Destination, formationAssignment *webhook.FormationAssignment) error {
	subaccountID, err := s.validateDestinationSubaccount(ctx, destinationDetails.SubaccountID, formationAssignment)
	if err != nil {
		return err
	}

	region, err := s.getRegionLabel(ctx, subaccountID)
	if err != nil {
		return err
	}

	destinationName := destinationDetails.Name
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
	req.Header.Set(clientUserHeaderKey, clientUser)

	log.C(ctx).Infof("Deleting destination with name: %q from destination service", destinationName)
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
	log.C(ctx).Infof("Successfully delete destination with name: %q from destination service", destinationName)

	if transactionErr := s.transaction(ctx, func(ctxWithTransact context.Context) error {
		if err = s.destinationRepo.DeleteByAssignmentID(ctxWithTransact, destinationName, subaccountID, formationAssignment.ID); err != nil {
			return errors.Wrapf(err, "while deleting destination with name: %q from the DB", destinationName)
		}
		return nil
	}); transactionErr != nil {
		return transactionErr
	}
	log.C(ctx).Infof("Successfully delete destination with name: %q and assignment ID: %q from the DB", destinationName, formationAssignment.ID)

	return nil
}

func (s *Service) validateDestinationSubaccount(ctx context.Context, destinationSubaccountID string, formationAssignment *webhook.FormationAssignment) (string, error) {
	consumerSubaccountID, err := s.getConsumerTenant(ctx, formationAssignment)
	if err != nil {
		return "", err
	}

	var subaccountID string
	subaccountID = consumerSubaccountID

	if destinationSubaccountID != "" && destinationSubaccountID != consumerSubaccountID {
		switch formationAssignment.TargetType {
		case model.FormationAssignmentTypeApplication:
			if err := s.validateAppTemplateProviderSubaccount(ctx, formationAssignment, destinationSubaccountID); err != nil {
				return "", err
			}
		case model.FormationAssignmentTypeRuntime:
			if err := s.validateRuntimeProviderSubaccount(ctx, formationAssignment.Target, destinationSubaccountID); err != nil {
				return "", err
			}
		case model.FormationAssignmentTypeRuntimeContext:
			if err := s.validateRuntimeContextProviderSubaccount(ctx, formationAssignment, destinationSubaccountID); err != nil {
				return "", err
			}
		default:
			return "", errors.Errorf("Unknown formation assignment type: %q", formationAssignment.TargetType)
		}

		subaccountID = destinationSubaccountID
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
	regionLbl, err := s.labelRepo.GetByKey(ctx, tenantID, model.TenantLabelableObject, tenantID, regionLabelKey)
	if err != nil {
		return "", err
	}

	region, ok := regionLbl.Value.(string)
	if !ok {
		return "", errors.Errorf("unexpected type of %q label, expect: string, got: %T", regionLabelKey, regionLbl.Value)
	}
	return region, nil
}

func (s *Service) validateAppTemplateProviderSubaccount(ctx context.Context, formationAssignment *webhook.FormationAssignment, destinationSubaccountID string) error {
	app, err := s.applicationRepository.GetByID(ctx, formationAssignment.TenantID, formationAssignment.Target)
	if err != nil {
		return err
	}

	if app.ApplicationTemplateID != nil && *app.ApplicationTemplateID != "" {
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

		if destinationSubaccountID != subaccountLblValue {
			return errors.Errorf("The provided destination subaccount is different from the owner subaccount of the application template with ID: %q", *app.ApplicationTemplateID)
		}
	}

	return nil
}

func (s *Service) validateRuntimeProviderSubaccount(ctx context.Context, runtimeID, destinationSubaccountID string) error {
	exists, err := s.runtimeRepository.OwnerExists(ctx, destinationSubaccountID, runtimeID)
	if err != nil {
		return err
	}

	if !exists {
		return errors.Errorf("The provided destination subaccount: %q is not provider of the runtime with ID: %q", destinationSubaccountID, runtimeID)
	}

	return nil
}

func (s *Service) validateRuntimeContextProviderSubaccount(ctx context.Context, formationAssignment *webhook.FormationAssignment, destinationSubaccountID string) error {
	rtmCtxID, err := s.runtimeCtxRepository.GetByID(ctx, formationAssignment.TenantID, formationAssignment.Target)
	if err != nil {
		return err
	}

	return s.validateRuntimeProviderSubaccount(ctx, rtmCtxID.RuntimeID, destinationSubaccountID)
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

func (s *Service) executeCreateRequest(ctx context.Context, url string, reqBody *destinationcreator.RequestBody, destinationName string) (statusCode int, err error) {
	reqBodyBytes, err := json.Marshal(reqBody)
	if err != nil {
		return statusCode, errors.Wrapf(err, "while marshalling destination request body")
	}

	clientUser, err := client.LoadFromContext(ctx)
	if err != nil || clientUser == "" {
		log.C(ctx).Warn("unable to provide client_user. Using correlation ID as client_user header...")
		clientUser = correlation.CorrelationIDFromContext(ctx)
	}

	req, err := http.NewRequest(http.MethodPost, url, bytes.NewBuffer(reqBodyBytes))
	req.Header.Set(clientUserHeaderKey, clientUser)

	resp, err := s.mtlsHTTPClient.Do(req)
	if err != nil {
		return statusCode, err
	}
	defer closeResponseBody(ctx, resp)

	body, err := ioutil.ReadAll(resp.Body)
	if err != nil {
		return statusCode, errors.Errorf("Failed to read destination response body: %v", err)
	}

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusConflict {
		return statusCode, errors.Errorf("Failed to create destination with name: %q, status: %d, body: %s", destinationName, resp.StatusCode, body)
	}

	if resp.StatusCode == http.StatusConflict {
		return http.StatusConflict, nil
	}
	log.C(ctx).Infof("Successfully create destination with name: %q in the destination service", destinationName)

	return http.StatusCreated, nil
}

func (s *Service) prepareRequestBody(ctx context.Context, destinationDetails operators.Destination, basicAuthenticationCredentials operators.BasicAuthentication, formationAssignment *webhook.FormationAssignment) (*destinationcreator.RequestBody, error) {
	reqBody := &destinationcreator.RequestBody{
		Name:               destinationDetails.Name,
		Type:               destinationcreator.TypeHTTP,
		ProxyType:          destinationcreator.ProxyTypeInternet,
		AuthenticationType: destinationcreator.AuthTypeBasic,
		User:               basicAuthenticationCredentials.Username,
		Password:           basicAuthenticationCredentials.Password,
	}

	if destinationDetails.Url != "" {
		reqBody.Url = destinationDetails.Url
	}

	if destinationDetails.Url == "" && basicAuthenticationCredentials.Url != "" {
		reqBody.Url = basicAuthenticationCredentials.Url
	}

	if destinationDetails.Url == "" && basicAuthenticationCredentials.Url == "" {
		app, err := s.applicationRepository.GetByID(ctx, formationAssignment.TenantID, formationAssignment.Target)
		if err != nil {
			return nil, err
		}
		if app.BaseURL != nil {
			reqBody.Url = *app.BaseURL
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

	if err := validate(reqBody); err != nil {
		return nil, err
	}

	return reqBody, nil
}

func validate(reqBody *destinationcreator.RequestBody) error {
	return validation.ValidateStruct(reqBody,
		validation.Field(&reqBody.Name, validation.Required, validation.Length(1, 200)),
		validation.Field(&reqBody.Url, validation.Required),
		validation.Field(&reqBody.Type, validation.In(destinationcreator.TypeHTTP, destinationcreator.TypeRFC, destinationcreator.TypeLDAP, destinationcreator.TypeMAIL)),
		validation.Field(&reqBody.ProxyType, validation.In(destinationcreator.ProxyTypeInternet, destinationcreator.ProxyTypeOnPremise, destinationcreator.ProxyTypePrivateLink)),
		validation.Field(&reqBody.AuthenticationType, validation.In(destinationcreator.AuthTypeNoAuth, destinationcreator.AuthTypeBasic, destinationcreator.AuthTypeSAMLBearer)),
		validation.Field(&reqBody.User, validation.Required, validation.Length(1, 256)),
	)
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

// todo::: delete
//func (reqBody *destinationcreator.RequestBody) validate() error {
//	return validation.ValidateStruct(reqBody,
//		validation.Field(&reqBody.Name, validation.Required, validation.Length(1, 200)),
//		validation.Field(&reqBody.Url, validation.Required),
//		validation.Field(&reqBody.Type, validation.In(TypeHTTP, TypeRFC, TypeLDAP, TypeMAIL)),
//		validation.Field(&reqBody.ProxyType, validation.In(ProxyTypeInternet, ProxyTypeOnPremise, ProxyTypePrivateLink)),
//		validation.Field(&reqBody.AuthenticationType, validation.In(AuthTypeNoAuth, AuthTypeBasic, AuthTypeSAMLBearer)),
//		validation.Field(&reqBody.User, validation.Required, validation.Length(1, 256)),
//	)
//}

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
