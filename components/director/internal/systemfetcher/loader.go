package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

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

// DataLoader loads and creates all the necessary data needed by system-fetcher
type DataLoader struct {
	transaction persistence.Transactioner
	appTmplSvc  appTmplService
	webhookSvc  webhookService
	intSysSvc   intSysSvc
	cfg         Config
	kubeClient  KubeClient
}

// NewDataLoader creates new DataLoader
func NewDataLoader(tx persistence.Transactioner, cfg Config, appTmplSvc appTmplService, intSysSvc intSysSvc, webhookSvc webhookService, kubeClient KubeClient) *DataLoader {
	return &DataLoader{
		transaction: tx,
		appTmplSvc:  appTmplSvc,
		intSysSvc:   intSysSvc,
		webhookSvc:  webhookSvc,
		cfg:         cfg,
		kubeClient:  kubeClient,
	}
}

// LoadData loads and creates all the necessary data needed by system-fetcher
func (d *DataLoader) LoadData(ctx context.Context, readDir func(dirname string) ([]os.DirEntry, error), readFile func(filename string) ([]byte, error)) error {
	appTemplateInputsMap, err := d.loadAppTemplates(ctx, readDir, readFile)
	if err != nil {
		return errors.Wrap(err, "failed while loading application templates")
	}

	tx, err := d.transaction.Begin()
	if err != nil {
		return errors.Wrap(err, "Error while beginning transaction")
	}
	defer d.transaction.RollbackUnlessCommitted(ctx, tx)
	ctxWithTx := persistence.SaveToContext(ctx, tx)

	appTemplateInputs, err := d.createAppTemplatesDependentEntities(ctxWithTx, appTemplateInputsMap)
	if err != nil {
		return errors.Wrap(err, "failed while creating application templates dependent entities")
	}

	if err = d.upsertAppTemplates(ctxWithTx, appTemplateInputs); err != nil {
		return errors.Wrap(err, "failed while upserting application templates")
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

func (d *DataLoader) loadAppTemplates(ctx context.Context, readDir func(dirname string) ([]os.DirEntry, error), readFile func(filename string) ([]byte, error)) ([]map[string]interface{}, error) {
	appTemplatesFileLocation := applicationTemplatesDirectoryPath
	if len(d.cfg.TemplatesFileLocation) > 0 {
		appTemplatesFileLocation = d.cfg.TemplatesFileLocation
	}

	files, err := readDir(appTemplatesFileLocation)
	if err != nil {
		return nil, errors.Wrapf(err, "while reading directory with application templates files [%s]", appTemplatesFileLocation)
	}

	var appTemplateInputs []map[string]interface{}
	for _, f := range files {
		log.C(ctx).Infof("Loading application templates from file: %s", f.Name())

		if filepath.Ext(f.Name()) != ".json" {
			return nil, apperrors.NewInternalError(fmt.Sprintf("unsupported file format %q, supported format: json", filepath.Ext(f.Name())))
		}

		bytes, err := readFile(appTemplatesFileLocation + f.Name())
		if err != nil {
			return nil, errors.Wrapf(err, "while reading application templates file %q", appTemplatesFileLocation+f.Name())
		}

		var templatesFromFile []map[string]interface{}
		if err := json.Unmarshal(bytes, &templatesFromFile); err != nil {
			return nil, errors.Wrapf(err, "while unmarshalling application templates from file %s", appTemplatesFileLocation+f.Name())
		}
		log.C(ctx).Infof("Successfully loaded application templates from file: %s", f.Name())
		appTemplateInputs = append(appTemplateInputs, templatesFromFile...)
	}

	return appTemplateInputs, nil
}

func (d *DataLoader) createAppTemplatesDependentEntities(ctx context.Context, appTmplInputs []map[string]interface{}) ([]model.ApplicationTemplateInput, error) {
	appTemplateInputs := make([]model.ApplicationTemplateInput, 0, len(appTmplInputs))
	for _, appTmplInput := range appTmplInputs {
		var input model.ApplicationTemplateInput
		appTmplInputJSON, err := json.Marshal(appTmplInput)
		if err != nil {
			return nil, errors.Wrap(err, "while marshaling application template input")
		}

		if err = json.Unmarshal(appTmplInputJSON, &input); err != nil {
			return nil, errors.Wrap(err, "while unmarshalling application template input")
		}

		intSystem, ok := appTmplInput[integrationSystemJSONKey]
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
		appTemplateInputs = append(appTemplateInputs, input)
	}

	return enrichApplicationTemplateInput(appTemplateInputs), nil
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

func (d *DataLoader) upsertAppTemplates(ctx context.Context, appTemplateInputs []model.ApplicationTemplateInput) error {
	for _, appTmplInput := range appTemplateInputs {
		var region interface{}
		region, err := retrieveRegion(appTmplInput.Labels)
		if err != nil {
			return err
		}
		if region == "" {
			region = nil
		}

		log.C(ctx).Infof("Retrieving application template with name %q and region %s", appTmplInput.Name, region)
		appTemplate, err := d.appTmplSvc.GetByNameAndRegion(ctx, appTmplInput.Name, region)
		if err != nil {
			if !strings.Contains(err.Error(), "Object not found") {
				return errors.Wrapf(err, "error while getting application template with name %q and region %s", appTmplInput.Name, region)
			}

			log.C(ctx).Infof("Cannot find application template with name %q and region %s. Creation triggered...", appTmplInput.Name, region)
			if err := d.setWebhooksCredentialsFromSecrets(ctx, appTmplInput.Webhooks); err != nil {
				return errors.Wrap(err, "error while setting webhooks credentials from secret")
			}
			templateID, err := d.appTmplSvc.Create(ctx, appTmplInput)
			if err != nil {
				return errors.Wrapf(err, "error while creating application template with name %q and region %s", appTmplInput.Name, region)
			}
			log.C(ctx).Infof("Successfully created application template with id: %q", templateID)
			continue
		}

		if !areAppTemplatesEqual(appTemplate, appTmplInput) {
			log.C(ctx).Infof("Updating application template with id %q", appTemplate.ID)
			appTemplateUpdateInput := model.ApplicationTemplateUpdateInput{
				Name:                 appTmplInput.Name,
				Description:          appTmplInput.Description,
				ApplicationNamespace: appTmplInput.ApplicationNamespace,
				ApplicationInputJSON: appTmplInput.ApplicationInputJSON,
				Placeholders:         appTmplInput.Placeholders,
				AccessLevel:          appTmplInput.AccessLevel,
				Labels:               appTmplInput.Labels,
			}
			if err := d.appTmplSvc.Update(ctx, appTemplate.ID, false, appTemplateUpdateInput); err != nil {
				return errors.Wrapf(err, "while updating application template with id %q", appTemplate.ID)
			}
			log.C(ctx).Infof("Successfully updated application template with id %q", appTemplate.ID)
		}

		webhooks, err := d.webhookSvc.ListForApplicationTemplate(ctx, appTemplate.ID)
		if err != nil {
			return err
		}

		if !areWebhooksEqual(webhooks, appTmplInput.Webhooks) {
			if err := d.setWebhooksCredentialsFromSecrets(ctx, appTmplInput.Webhooks); err != nil {
				return errors.Wrap(err, "error while setting webhooks credentials from secret")
			}
			if err := d.syncWebhooks(ctx, appTemplate.ID, webhooks, appTmplInput.Webhooks); err != nil {
				return errors.Wrapf(err, "while updating webhooks for application tempate with id %q", appTemplate.ID)
			}

			log.C(ctx).Infof("Successfully updated the webhooks for application template with id %q", appTemplate.ID)
		}
	}

	return nil
}
func (d *DataLoader) setWebhooksCredentialsFromSecrets(ctx context.Context, webhooks []*model.WebhookInput) error {

	secretNameToSecretKeyToSecretData, err := d.prepareWebhooksCredentialsMappings(ctx, webhooks)
	if err != nil {
		return errors.Wrap(err, "while preparing webhooks credentials mappings")
	}
	for i, wh := range webhooks {
		if wh.Auth != nil && wh.Auth.SecretRef != nil && wh.Auth.SecretRef.SecretName != "" && wh.Auth.SecretRef.SecretKey != "" {
			secretName := wh.Auth.SecretRef.SecretName
			secretKey := wh.Auth.SecretRef.SecretKey

			log.C(ctx).Infof("Setting credentials for webhook with id: %q for key: %q from secret with name: %q", wh.ID, secretKey, secretName)
			secret, exists := secretNameToSecretKeyToSecretData[secretName]
			if !exists {
				log.C(ctx).Warnf("There is no secret with name: %q. No credential from secret will be set for webhook with id: %q", secretName, wh.ID)
				continue
			}
			whAuth, exists := secret[secretKey]
			if !exists {
				log.C(ctx).Warnf("There is no credentials for key: %q in secret with name: %q. No credentials from secret will be set for webhook with id: %q", secretKey, secretName, wh.ID)
				continue
			}
			webhooks[i].Auth = whAuth
			log.C(ctx).Infof("Successfully set credentials for webhook with id: %q for key: %q from secret with name: %q", wh.ID, secretKey, secretName)
		}
	}
	return nil
}

func (d *DataLoader) prepareWebhooksCredentialsMappings(ctx context.Context, webhooks []*model.WebhookInput) (map[string]map[string]*model.AuthInput, error) {
	log.C(ctx).Infof("Preparing webhooks credentials mappings...")

	secretNameToSecretKeyToSecretData := make(map[string]map[string]*model.AuthInput, 0)
	for _, wh := range webhooks {
		if wh.Auth != nil && wh.Auth.SecretRef != nil && wh.Auth.SecretRef.SecretName != "" {
			secretName := wh.Auth.SecretRef.SecretName

			if _, exists := secretNameToSecretKeyToSecretData[secretName]; !exists {
				secretDataBytes, err := d.kubeClient.GetSystemFetcherSecretData(ctx, secretName)
				if err != nil {
					log.C(ctx).Warnf(err.Error())
					continue
				}
				secretDataCredentials := make(map[string]*model.AuthInput, 0)
				if err := json.Unmarshal(secretDataBytes, &secretDataCredentials); err != nil {
					return nil, errors.Wrapf(err, "while unmarshalling secret data from secret with name: %q", secretName)
				}
				secretNameToSecretKeyToSecretData[secretName] = secretDataCredentials
			}
		}
	}
	return secretNameToSecretKeyToSecretData, nil
}

func enrichApplicationTemplateInput(appTemplateInputs []model.ApplicationTemplateInput) []model.ApplicationTemplateInput {
	enriched := make([]model.ApplicationTemplateInput, 0, len(appTemplateInputs))
	for _, appTemplateInput := range appTemplateInputs {
		if appTemplateInput.Description == nil {
			appTemplateInput.Description = str.Ptr(appTemplateInput.Name)
		}

		if appTemplateInput.Placeholders == nil || len(appTemplateInput.Placeholders) == 0 {
			appTemplateInput.Placeholders = []model.ApplicationTemplatePlaceholder{
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

		if appTemplateInput.AccessLevel == "" {
			appTemplateInput.AccessLevel = model.GlobalApplicationTemplateAccessLevel
		}

		if appTemplateInput.Labels == nil {
			appTemplateInput.Labels = map[string]interface{}{managedAppProvisioningLabelKey: false}
		}
		enriched = append(enriched, appTemplateInput)
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
