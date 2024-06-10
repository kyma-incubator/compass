package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"github.com/kyma-incubator/compass/components/director/internal/uid"
	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"os"
	"path/filepath"
	"reflect"
	"strings"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"
	"github.com/pkg/errors"
)

const (
	applicationTemplatesDirectoryPath   = "/data/templates/"
	integrationSystemJSONKey            = "intSystem"
	integrationSystemNameJSONKey        = "name"
	integrationSystemDescriptionJSONKey = "description"
	managedAppProvisioningLabelKey      = "managed_app_provisioning"
	integrationSystemIDLabelKey         = "integrationSystemID"
)

//go:generate mockery --name=appTmplService --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type appTmplService interface {
	GetByNameAndRegion(ctx context.Context, name string, region interface{}) (*model.ApplicationTemplate, error)
	Create(ctx context.Context, in model.ApplicationTemplateInput) (string, error)
	Update(ctx context.Context, id string, override bool, in model.ApplicationTemplateUpdateInput) error
}

//go:generate mockery --name=webhookService --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type webhookService interface {
	ListForApplicationTemplate(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error)
	Create(ctx context.Context, owningResourceID string, in model.WebhookInput, objectType model.WebhookReferenceObjectType) (string, error)
	Update(ctx context.Context, id string, in model.WebhookInput, objectType model.WebhookReferenceObjectType) error
	Delete(ctx context.Context, id string, objectType model.WebhookReferenceObjectType) error
}

//go:generate mockery --name=intSysSvc --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type intSysSvc interface {
	Create(ctx context.Context, in model.IntegrationSystemInput) (string, error)
	List(ctx context.Context, pageSize int, cursor string) (model.IntegrationSystemPage, error)
}

//go:generate mockery --name=certSubjMappingService --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type certSubjMappingService interface {
	ListByConsumerID(ctx context.Context, consumerID string) ([]*model.CertSubjectMapping, error)
	Create(ctx context.Context, item *model.CertSubjectMapping) (string, error)
	Update(ctx context.Context, in *model.CertSubjectMapping) error
}

//go:generate mockery --name=uidService --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type uidService interface {
	Generate() string
}

// ManagedResource represents a managed resource
type ManagedResource struct {
	model.ApplicationTemplateInput
	CertSubjMappingInputs []model.CertSubjectMappingInput `json:"certSubjectMappings"`
}

// DataLoader loads and creates all the necessary data needed by system-fetcher
type DataLoader struct {
	transaction        persistence.Transactioner
	appTmplSvc         appTmplService
	webhookSvc         webhookService
	intSysSvc          intSysSvc
	certSubjMappingSvc certSubjMappingService
	uidSvc             uidService
	cfg                Config
}

// NewDataLoader creates new DataLoader
func NewDataLoader(tx persistence.Transactioner, cfg Config, appTmplSvc appTmplService, intSysSvc intSysSvc, webhookSvc webhookService, certSubjMappingSvc certSubjMappingService) *DataLoader {
	return &DataLoader{
		transaction:        tx,
		appTmplSvc:         appTmplSvc,
		intSysSvc:          intSysSvc,
		webhookSvc:         webhookSvc,
		certSubjMappingSvc: certSubjMappingSvc,
		uidSvc:             uid.NewService(),
		cfg:                cfg,
	}
}

// LoadData loads and creates all the necessary data needed by system-fetcher
func (d *DataLoader) LoadData(ctx context.Context, readDir func(dirname string) ([]os.DirEntry, error), readFile func(filename string) ([]byte, error)) error {
	managedResourcesMap, err := d.loadManagedResources(ctx, readDir, readFile)
	if err != nil {
		return errors.Wrap(err, "failed while loading managed resources")
	}

	tx, err := d.transaction.Begin()
	if err != nil {
		return errors.Wrap(err, "Error while beginning transaction")
	}
	defer d.transaction.RollbackUnlessCommitted(ctx, tx)
	ctxWithTx := persistence.SaveToContext(ctx, tx)

	managedResources, err := d.createDependentEntities(ctxWithTx, managedResourcesMap)
	if err != nil {
		return errors.Wrap(err, "failed while creating dependent entities")
	}

	appTemplateToCertSubjMappingsMap, err := d.upsertAppTemplates(ctxWithTx, managedResources)
	if err != nil {
		return errors.Wrap(err, "failed while upserting application templates")
	}

	if err = d.upsertCertSubjectMappings(ctxWithTx, appTemplateToCertSubjMappingsMap); err != nil {
		return errors.Wrap(err, "failed while upserting certificate subject mappings")
	}

	if err := tx.Commit(); err != nil {
		return errors.Wrap(err, "while committing transaction")
	}

	return nil
}

func (d *DataLoader) syncWebhooks(ctx context.Context, appTemplateID string, webhooksModel []*model.Webhook, webhooksInput []*model.WebhookInput) error {
	webhooksModelMap := make(map[model.WebhookType]*model.Webhook)

	for _, webhook := range webhooksModel {
		webhooksModelMap[webhook.Type] = webhook
	}

	for _, webhookInput := range webhooksInput {
		webhookModel, exists := webhooksModelMap[webhookInput.Type]
		var err error
		if exists {
			log.C(ctx).Infof("Webhook of type %s exists. Will update it...", webhookInput.Type)
			err = d.webhookSvc.Update(ctx, webhookModel.ID, *webhookInput, model.ApplicationTemplateWebhookReference)
			delete(webhooksModelMap, webhookInput.Type) // remove the item from the map after updating
		} else {
			log.C(ctx).Infof("Webhook of type %s does not exist. Will create it...", webhookInput.Type)
			_, err = d.webhookSvc.Create(ctx, appTemplateID, *webhookInput, model.ApplicationTemplateWebhookReference)
		}

		if err != nil {
			return err
		}
	}

	for _, webhookModel := range webhooksModelMap {
		log.C(ctx).Infof("Webhook of type %s is missing in the input. Will delete it...", webhookModel.Type)
		if err := d.webhookSvc.Delete(ctx, webhookModel.ID, model.ApplicationTemplateWebhookReference); err != nil {
			return err
		}
	}

	return nil
}

func (d *DataLoader) loadManagedResources(ctx context.Context, readDir func(dirname string) ([]os.DirEntry, error), readFile func(filename string) ([]byte, error)) ([]map[string]interface{}, error) {
	appTemplatesFileLocation := applicationTemplatesDirectoryPath
	if len(d.cfg.TemplatesFileLocation) > 0 {
		appTemplatesFileLocation = d.cfg.TemplatesFileLocation
	}

	files, err := readDir(appTemplatesFileLocation)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading directory with application templates files [%s]", appTemplatesFileLocation)
	}

	var managedResources []map[string]interface{}
	for _, f := range files {
		log.C(ctx).Infof("Loading managed resources from file: %s", f.Name())

		if filepath.Ext(f.Name()) != ".json" {
			return nil, apperrors.NewInternalError(fmt.Sprintf("unsupported file format %q, supported format: json", filepath.Ext(f.Name())))
		}

		bytes, err := readFile(appTemplatesFileLocation + f.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "while reading application templates file %q", appTemplatesFileLocation+f.Name())
		}

		var resourcesFromFile []map[string]interface{}
		if err := json.Unmarshal(bytes, &resourcesFromFile); err != nil {
			return nil, errors.Wrapf(err, "while unmarshalling managed resources from file %s", appTemplatesFileLocation+f.Name())
		}
		log.C(ctx).Infof("Successfully loaded application templates from file: %s", f.Name())
		managedResources = append(managedResources, resourcesFromFile...)
	}

	return managedResources, nil
}

func (d *DataLoader) createDependentEntities(ctx context.Context, managedResourcesMap []map[string]interface{}) ([]ManagedResource, error) {
	managedResources := make([]ManagedResource, 0, len(managedResourcesMap))
	for _, managedResource := range managedResourcesMap {
		var input ManagedResource
		managedResourceJson, err := json.Marshal(managedResource)
		if err != nil {
			return nil, errors.Wrap(err, "while marshaling managed resources")
		}

		if err = json.Unmarshal(managedResourceJson, &input); err != nil {
			return nil, errors.Wrap(err, "while unmarshalling managed resources into a struct")
		}

		log.C(ctx).Errorf("KALO- app template input %v", input.ApplicationTemplateInput)
		log.C(ctx).Errorf("KALO- cert subj mapping inputs %v", input.CertSubjMappingInputs)

		intSystem, ok := managedResource[integrationSystemJSONKey]
		if ok {
			intSystemData, ok := intSystem.(map[string]interface{})
			if !ok {
				return nil, fmt.Errorf("the type of the integration system is %T instead of map[string]interface{}. %v", intSystem, intSystemData)
			}

			appTmplIntSystem, err := extractIntegrationSystem(intSystemData)
			if err != nil {
				return nil, err
			}

			intSystemsFromDB, err := d.listIntegrationSystems(ctx)
			if err != nil {
				return nil, err
			}

			var intSysID string
			for _, is := range intSystemsFromDB {
				if is.Name == appTmplIntSystem.Name && str.PtrStrToStr(is.Description) == str.PtrStrToStr(appTmplIntSystem.Description) {
					intSysID = is.ID
					break
				}
			}
			if intSysID == "" {
				id, err := d.intSysSvc.Create(ctx, *appTmplIntSystem)
				if err != nil {
					return nil, errors.Wrapf(err, "while creating integration system with name %s", appTmplIntSystem.Name)
				}
				intSysID = id
				log.C(ctx).Infof("Successfully created integration system with id: %s", intSysID)
			}
			input.ApplicationInputJSON, err = enrichWithIntegrationSystemIDLabel(input.ApplicationInputJSON, intSysID)
			if err != nil {
				return nil, err
			}
		}
		managedResources = append(managedResources, input)
	}

	return enrichApplicationTemplateInput(managedResources), nil
}

func (d *DataLoader) listIntegrationSystems(ctx context.Context) ([]*model.IntegrationSystem, error) {
	pageSize := 200
	pageCursor := ""
	hasNextPage := true

	var integrationSystems []*model.IntegrationSystem
	for hasNextPage {
		intSystemsPage, err := d.intSysSvc.List(ctx, pageSize, pageCursor)
		if err != nil {
			return nil, errors.Wrapf(err, "while listing integration systems")
		}

		integrationSystems = append(integrationSystems, intSystemsPage.Data...)

		pageCursor = intSystemsPage.PageInfo.EndCursor
		hasNextPage = intSystemsPage.PageInfo.HasNextPage
	}

	return integrationSystems, nil
}

func (d *DataLoader) upsertAppTemplates(ctx context.Context, managedResources []ManagedResource) (map[string][]model.CertSubjectMappingInput, error) {
	appTemplateToCertSubjMappingsMap := make(map[string][]model.CertSubjectMappingInput)
	for _, managedResource := range managedResources {
		var region interface{}
		region, err := retrieveRegion(managedResource.ApplicationTemplateInput.Labels)
		if err != nil {
			return nil, err
		}
		if region == "" {
			region = nil
		}

		log.C(ctx).Infof("Retrieving application template with name %q and region %s", managedResource.ApplicationTemplateInput.Name, region)
		appTemplate, err := d.appTmplSvc.GetByNameAndRegion(ctx, managedResource.ApplicationTemplateInput.Name, region)
		if err != nil {
			if !strings.Contains(err.Error(), "Object not found") {
				return nil, errors.Wrapf(err, "error while getting application template with name %q and region %s", managedResource.ApplicationTemplateInput.Name, region)
			}

			log.C(ctx).Infof("Cannot find application template with name %q and region %s. Creation triggered...", managedResource.ApplicationTemplateInput.Name, region)
			templateID, err := d.appTmplSvc.Create(ctx, managedResource.ApplicationTemplateInput)
			if err != nil {
				return nil, errors.Wrapf(err, "error while creating application template with name %q and region %s", managedResource.ApplicationTemplateInput.Name, region)
			}
			log.C(ctx).Infof("Successfully created application template with id: %q", templateID)

			if len(managedResource.CertSubjMappingInputs) > 0 {
				appTemplateToCertSubjMappingsMap[templateID] = managedResource.CertSubjMappingInputs
			}
			continue
		}

		if len(managedResource.CertSubjMappingInputs) > 0 {
			appTemplateToCertSubjMappingsMap[appTemplate.ID] = managedResource.CertSubjMappingInputs
		}
		if !areAppTemplatesEqual(appTemplate, managedResource.ApplicationTemplateInput) {
			log.C(ctx).Infof("Updating application template with id %q", appTemplate.ID)
			appTemplateUpdateInput := model.ApplicationTemplateUpdateInput{
				Name:                 managedResource.ApplicationTemplateInput.Name,
				Description:          managedResource.ApplicationTemplateInput.Description,
				ApplicationNamespace: managedResource.ApplicationTemplateInput.ApplicationNamespace,
				ApplicationInputJSON: managedResource.ApplicationTemplateInput.ApplicationInputJSON,
				Placeholders:         managedResource.ApplicationTemplateInput.Placeholders,
				AccessLevel:          managedResource.ApplicationTemplateInput.AccessLevel,
				Labels:               managedResource.ApplicationTemplateInput.Labels,
			}
			if err := d.appTmplSvc.Update(ctx, appTemplate.ID, false, appTemplateUpdateInput); err != nil {
				return nil, errors.Wrapf(err, "while updating application template with id %q", appTemplate.ID)
			}
			log.C(ctx).Infof("Successfully updated application template with id %q", appTemplate.ID)
		}

		webhooks, err := d.webhookSvc.ListForApplicationTemplate(ctx, appTemplate.ID)
		if err != nil {
			return nil, err
		}

		if !areWebhooksEqual(webhooks, managedResource.ApplicationTemplateInput.Webhooks) {
			if err := d.syncWebhooks(ctx, appTemplate.ID, webhooks, managedResource.ApplicationTemplateInput.Webhooks); err != nil {
				return nil, errors.Wrapf(err, "while updating webhooks for application tempate with id %q", appTemplate.ID)
			}

			log.C(ctx).Infof("Successfully updated the webhooks for application template with id %q", appTemplate.ID)
		}
	}

	return appTemplateToCertSubjMappingsMap, nil
}

func (d *DataLoader) upsertCertSubjectMappings(ctx context.Context, appTemplateToCertSubjMappingsMap map[string][]model.CertSubjectMappingInput) error {
	for templateID, certSubjectMappingInputs := range appTemplateToCertSubjMappingsMap {
		existingCertSubjMappings, err := d.certSubjMappingSvc.ListByConsumerID(ctx, templateID)
		if err != nil {
			return errors.Wrapf(err, "error while listing certificate subject mappings by consumer id %q", templateID)
		}

		for _, csmi := range certSubjectMappingInputs {
			if exists, csm := certSubjectMappingExists(csmi.Subject, existingCertSubjMappings); exists {
				csm.ConsumerType = csmi.ConsumerType
				csm.TenantAccessLevels = csmi.TenantAccessLevels

				if err = d.certSubjMappingSvc.Update(ctx, csm); err != nil {
					return errors.Wrapf(err, "error while updating certificate subject mapping with id %q", csm.ID)
				}
				continue
			}

			csm := csmi.ToModel(d.uidSvc.Generate())
			csm.CreatedAt = time.Now()
			csm.InternalConsumerID = &templateID

			if _, err = d.certSubjMappingSvc.Create(ctx, csm); err != nil {
				return errors.Wrapf(err, "error while creating certificate subject mapping")
			}
		}
	}

	return nil
}

func certSubjectMappingExists(subject string, certSubjMappings []*model.CertSubjectMapping) (bool, *model.CertSubjectMapping) {
	for _, csm := range certSubjMappings {
		if cert.SubjectsMatch(subject, csm.Subject) {
			return true, csm
		}
	}
	return false, nil
}

func enrichApplicationTemplateInput(managedResources []ManagedResource) []ManagedResource {
	enriched := make([]ManagedResource, 0, len(managedResources))
	for _, managedResource := range managedResources {
		if managedResource.ApplicationTemplateInput.Description == nil {
			managedResource.ApplicationTemplateInput.Description = str.Ptr(managedResource.ApplicationTemplateInput.Name)
		}

		if managedResource.ApplicationTemplateInput.Placeholders == nil || len(managedResource.ApplicationTemplateInput.Placeholders) == 0 {
			managedResource.ApplicationTemplateInput.Placeholders = []model.ApplicationTemplatePlaceholder{
				{
					Name:        "name",
					Description: str.Ptr("Application’s technical name"),
					JSONPath:    str.Ptr("$.displayName"),
				},
				{
					Name:        "display-name",
					Description: str.Ptr("Application’s display name"),
					JSONPath:    str.Ptr("$.displayName"),
				},
			}
		}

		if managedResource.ApplicationTemplateInput.AccessLevel == "" {
			managedResource.ApplicationTemplateInput.AccessLevel = model.GlobalApplicationTemplateAccessLevel
		}

		if managedResource.ApplicationTemplateInput.Labels == nil {
			managedResource.ApplicationTemplateInput.Labels = map[string]interface{}{managedAppProvisioningLabelKey: false}
		}
		enriched = append(enriched, managedResource)
	}
	return enriched
}

func enrichWithIntegrationSystemIDLabel(applicationInputJSON, intSystemID string) (string, error) {
	var appInput map[string]interface{}

	if err := json.Unmarshal([]byte(applicationInputJSON), &appInput); err != nil {
		return "", errors.Wrapf(err, "while unmarshaling application input json")
	}

	appInput[integrationSystemIDLabelKey] = intSystemID

	inputJSON, err := json.Marshal(appInput)
	if err != nil {
		return "", errors.Wrapf(err, "while marshalling app input")
	}
	return string(inputJSON), nil
}

func extractIntegrationSystem(intSysMap map[string]interface{}) (*model.IntegrationSystemInput, error) {
	intSysName, ok := intSysMap[integrationSystemNameJSONKey]
	if !ok {
		return nil, fmt.Errorf("integration system name is missing")
	}
	intSysNameValue, ok := intSysName.(string)
	if !ok {
		return nil, fmt.Errorf("integration system name value must be string")
	}

	intSysDesc, ok := intSysMap[integrationSystemDescriptionJSONKey]
	if !ok {
		return nil, fmt.Errorf("integration system description is missing")
	}
	intSysDescValue, ok := intSysDesc.(string)
	if !ok {
		return nil, fmt.Errorf("integration system description value must be string")
	}

	return &model.IntegrationSystemInput{
		Name:        intSysNameValue,
		Description: str.Ptr(intSysDescValue),
	}, nil
}

func retrieveRegion(labels map[string]interface{}) (string, error) {
	if labels == nil {
		return "", nil
	}

	region, exists := labels[tenant.RegionLabelKey]
	if !exists {
		return "", nil
	}

	regionValue, ok := region.(string)
	if !ok {
		return "", fmt.Errorf("%q label value must be string", tenant.RegionLabelKey)
	}
	return regionValue, nil
}

func areAppTemplatesEqual(appTemplate *model.ApplicationTemplate, appTemplateInput model.ApplicationTemplateInput) bool {
	if appTemplate == nil {
		return false
	}

	isAppInputJSONEqual := appTemplate.ApplicationInputJSON == appTemplateInput.ApplicationInputJSON
	isLabelEqual := reflect.DeepEqual(appTemplate.Labels, appTemplateInput.Labels)
	isPlaceholderEqual := reflect.DeepEqual(appTemplate.Placeholders, appTemplateInput.Placeholders)

	if isAppInputJSONEqual && isLabelEqual && isPlaceholderEqual {
		return true
	}

	return false
}

func areWebhooksEqual(webhooksModel []*model.Webhook, webhooksInput []*model.WebhookInput) bool {
	if len(webhooksModel) != len(webhooksInput) {
		return false
	}
	foundWebhooksCounter := 0
	for _, whModel := range webhooksModel {
		for _, whInput := range webhooksInput {
			if reflect.DeepEqual(whModel, *whInput.ToWebhook(whModel.ID, whModel.ObjectID, whModel.ObjectType)) {
				foundWebhooksCounter++
				break
			}
		}
	}
	return foundWebhooksCounter == len(webhooksModel)
}
