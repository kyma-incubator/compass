package apptemplate

import (
	"context"
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"regexp"
	"sort"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/apiclient"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/selfregmanager"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

const (
	globalSubaccountIDLabelKey = "global_subaccount_id"
	sapProviderName            = "SAP"
	displayNameLabelKey        = "displayName"
)

var defaultSlisFilterValueForManagedByProperty = []interface{}{
	map[string]interface{}{
		"key":       "$.additionalAttributes.managedBy",
		"value":     []string{"SAP Cloud"},
		"operation": "exclude",
	},
}

// ApplicationTemplateService missing godoc
//
//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateService interface {
	Create(ctx context.Context, in model.ApplicationTemplateInput) (string, error)
	CreateWithLabels(ctx context.Context, in model.ApplicationTemplateInput, labels map[string]interface{}) (string, error)
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	GetByFilters(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.ApplicationTemplate, error)
	GetByNameAndRegion(ctx context.Context, name string, region interface{}) (*model.ApplicationTemplate, error)
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (model.ApplicationTemplatePage, error)
	ListByName(ctx context.Context, name string) ([]*model.ApplicationTemplate, error)
	ListByFilters(ctx context.Context, filter []*labelfilter.LabelFilter) ([]*model.ApplicationTemplate, error)
	Update(ctx context.Context, id string, override bool, in model.ApplicationTemplateUpdateInput) error
	Delete(ctx context.Context, id string) error
	PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, values model.ApplicationFromTemplateInputValues) (string, error)
	ListLabels(ctx context.Context, appTemplateID string) (map[string]*model.Label, error)
	GetLabel(ctx context.Context, appTemplateID string, key string) (*model.Label, error)
}

// ApplicationTemplateConverter missing godoc
//
//go:generate mockery --name=ApplicationTemplateConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateConverter interface {
	ToGraphQL(in *model.ApplicationTemplate) (*graphql.ApplicationTemplate, error)
	MultipleToGraphQL(in []*model.ApplicationTemplate) ([]*graphql.ApplicationTemplate, error)
	InputFromGraphQL(in graphql.ApplicationTemplateInput) (model.ApplicationTemplateInput, error)
	UpdateInputFromGraphQL(in graphql.ApplicationTemplateUpdateInput) (model.ApplicationTemplateUpdateInput, error)
	ApplicationFromTemplateInputFromGraphQL(appTemplate *model.ApplicationTemplate, in graphql.ApplicationFromTemplateInput) (model.ApplicationFromTemplateInput, error)
}

// ApplicationConverter missing godoc
//
//go:generate mockery --name=ApplicationConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
	CreateRegisterInputJSONToGQL(in string) (graphql.ApplicationRegisterInput, error)
	CreateInputFromGraphQL(ctx context.Context, in graphql.ApplicationRegisterInput) (model.ApplicationRegisterInput, error)
}

// ApplicationService missing godoc
//
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	Create(ctx context.Context, in model.ApplicationRegisterInput) (string, error)
	CreateFromTemplate(ctx context.Context, in model.ApplicationRegisterInput, appTemplateID *string, systemFieldDiscoveryValue bool) (string, error)
	Get(ctx context.Context, id string) (*model.Application, error)
}

// WebhookService missing godoc
//
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookService interface {
	ListForApplicationTemplate(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error)
	EnrichWebhooksWithTenantMappingWebhooks(in []*graphql.WebhookInput) ([]*graphql.WebhookInput, error)
}

// WebhookConverter missing godoc
//
//go:generate mockery --name=WebhookConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookConverter interface {
	MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error)
	MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error)
}

// LabelService is responsible for Label operations
//
//go:generate mockery --name=LabelService --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelService interface {
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

// SelfRegisterManager missing godoc
//
//go:generate mockery --name=SelfRegisterManager --output=automock --outpkg=automock --case=underscore --disable-version-string
type SelfRegisterManager interface {
	IsSelfRegistrationFlow(ctx context.Context, labels map[string]interface{}) (bool, error)
	PrepareForSelfRegistration(ctx context.Context, resourceType resource.Type, labels map[string]interface{}, id string, validate func() error) (map[string]interface{}, error)
	CleanupSelfRegistration(ctx context.Context, selfRegisterLabelValue, region string) error
	GetSelfRegDistinguishingLabelKey() string
}

// Resolver missing godoc
type Resolver struct {
	transact persistence.Transactioner

	appSvc                  ApplicationService
	appConverter            ApplicationConverter
	appTemplateSvc          ApplicationTemplateService
	appTemplateConverter    ApplicationTemplateConverter
	webhookSvc              WebhookService
	webhookConverter        WebhookConverter
	labelSvc                LabelService
	selfRegManager          SelfRegisterManager
	uidService              UIDService
	appTemplateProductLabel string
	certSubjectMappingSvc   CertSubjectMappingService
	ordClient               *apiclient.ORDClient
	envConsumerSubjects     []string
}

// NewResolver missing godoc
func NewResolver(transact persistence.Transactioner, appSvc ApplicationService, appConverter ApplicationConverter, appTemplateSvc ApplicationTemplateService, appTemplateConverter ApplicationTemplateConverter, webhookService WebhookService, webhookConverter WebhookConverter, labelSvc LabelService, selfRegisterManager SelfRegisterManager, uidService UIDService, certSubjectMappingSvc CertSubjectMappingService, appTemplateProductLabel string, ordAggregatorClientConfig apiclient.OrdAggregatorClientConfig, environmentConsumerSubjects []string) *Resolver {
	return &Resolver{
		transact:                transact,
		appSvc:                  appSvc,
		appConverter:            appConverter,
		appTemplateSvc:          appTemplateSvc,
		appTemplateConverter:    appTemplateConverter,
		webhookSvc:              webhookService,
		webhookConverter:        webhookConverter,
		labelSvc:                labelSvc,
		selfRegManager:          selfRegisterManager,
		uidService:              uidService,
		appTemplateProductLabel: appTemplateProductLabel,
		certSubjectMappingSvc:   certSubjectMappingSvc,
		ordClient:               apiclient.NewORDClient(ordAggregatorClientConfig),
		envConsumerSubjects:     environmentConsumerSubjects,
	}
}

// ApplicationTemplate missing godoc
func (r *Resolver) ApplicationTemplate(ctx context.Context, id string) (*graphql.ApplicationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	appTemplate, err := r.appTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	out, err := r.appTemplateConverter.ToGraphQL(appTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting application template to graphql")
	}

	return out, nil
}

// ApplicationTemplates missing godoc
func (r *Resolver) ApplicationTemplates(ctx context.Context, filter []*graphql.LabelFilter, first *int, after *graphql.PageCursor) (*graphql.ApplicationTemplatePage, error) {
	labelFilter := labelfilter.MultipleFromGraphQL(filter)
	var cursor string
	if after != nil {
		cursor = string(*after)
	}
	if first == nil {
		return nil, apperrors.NewInvalidDataError("missing required parameter 'first'")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	appTemplatePage, err := r.appTemplateSvc.List(ctx, labelFilter, *first, cursor)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlAppTemplate, err := r.appTemplateConverter.MultipleToGraphQL(appTemplatePage.Data)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting application templates to graphql")
	}

	return &graphql.ApplicationTemplatePage{
		Data:       gqlAppTemplate,
		TotalCount: appTemplatePage.TotalCount,
		PageInfo: &graphql.PageInfo{
			StartCursor: graphql.PageCursor(appTemplatePage.PageInfo.StartCursor),
			EndCursor:   graphql.PageCursor(appTemplatePage.PageInfo.EndCursor),
			HasNextPage: appTemplatePage.PageInfo.HasNextPage,
		},
	}, nil
}

func (r *Resolver) validateLabels(labels map[string]interface{}) error {
	systemRolesLabel, hasSystemRoles := labels[r.appTemplateProductLabel]
	slisFilterLabel, hasSlisFilter := labels[systemfetcher.SlisFilterLabelKey]

	if !hasSystemRoles && hasSlisFilter {
		return errors.New("system role is required when slis filter is defined")
	}

	if !hasSystemRoles {
		return nil
	}

	systemRolesLabelValue, ok := systemRolesLabel.([]interface{})
	if !ok {
		return errors.New("invalid format of system roles label")
	}

	if len(systemRolesLabelValue) == 0 && hasSlisFilter {
		return errors.New("system role must not be empty when slis filter is defined")
	}

	if !hasSlisFilter {
		return nil
	}

	systemRoles, err := str.ConvertToStringArray(systemRolesLabelValue)
	if err != nil {
		return err
	}

	slisFilterLabelValue, ok := slisFilterLabel.([]interface{})
	if !ok {
		return errors.Errorf("invalid format of slis filter label")
	}

	productIds := make([]string, 0)

	for _, slisFilterValue := range slisFilterLabelValue {
		filter, ok := slisFilterValue.(map[string]interface{})
		if !ok {
			return errors.New("invalid format of slis filter value")
		}

		productID, ok := filter[systemfetcher.ProductIDKey]
		if !ok {
			return errors.New("missing productId in slis filter")
		}

		productIDStr, ok := productID.(string)
		if !ok {
			return errors.New("invalid format of productId value")
		}

		productIds = append(productIds, productIDStr)
	}

	systemRolesCount := len(systemRoles)
	slisFilterProductIdsCount := len(productIds)

	if systemRolesCount != slisFilterProductIdsCount {
		return errors.New("system roles count does not match the product ids count in slis filter")
	}

	sort.Strings(systemRoles)
	sort.Strings(productIds)

	for i, systemRole := range systemRoles {
		if systemRole != productIds[i] {
			return errors.New("system roles don't match with product ids in slis filter")
		}
	}

	return nil
}

func (r *Resolver) validateApplicationTemplate(in graphql.ApplicationTemplateInput) error {
	if err := in.Validate(); err != nil {
		return err
	}

	if err := r.validateLabels(in.Labels); err != nil {
		return err
	}

	if err := validateAppTemplateNameBasedOnProvider(in.Name, in.ApplicationInput); err != nil {
		return err
	}

	return nil
}

// CreateApplicationTemplate missing godoc
func (r *Resolver) CreateApplicationTemplate(ctx context.Context, in graphql.ApplicationTemplateInput) (*graphql.ApplicationTemplate, error) {
	log.C(ctx).Infof("Validating graphql input for Application Template with name %s", in.Name)
	if err := r.validateApplicationTemplate(in); err != nil {
		return nil, err
	}

	log.C(ctx).Info("Enriching webhooks with tenant mapping webhooks")
	webhooks, err := r.webhookSvc.EnrichWebhooksWithTenantMappingWebhooks(in.Webhooks)
	if err != nil {
		return nil, err
	}

	if in.Webhooks != nil {
		in.Webhooks = webhooks
	}
	convertedIn, err := r.appTemplateConverter.InputFromGraphQL(in)

	if err != nil {
		return nil, err
	}

	if convertedIn.Labels == nil {
		convertedIn.Labels = make(map[string]interface{})
	}

	selfRegID := r.uidService.Generate()
	convertedIn.ID = &selfRegID
	log.C(ctx).Infof("Generated ID %s for Application Template with name %s", selfRegID, in.Name)

	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while loading consumer")
	}

	labels := convertedIn.Labels
	if _, err := tenant.LoadFromContext(ctx); err == nil && consumerInfo.Flow.IsCertFlow() {
		isSelfReg, selfRegFlowErr := r.selfRegManager.IsSelfRegistrationFlow(ctx, labels)
		if selfRegFlowErr != nil {
			return nil, selfRegFlowErr
		}

		if isSelfReg {
			validate := func() error {
				return validateAppTemplateForSelfReg(in.ApplicationInput)
			}

			log.C(ctx).Info("Executing self registration flow for Application Template")
			labels, err = r.selfRegManager.PrepareForSelfRegistration(ctx, resource.ApplicationTemplate, convertedIn.Labels, selfRegID, validate)
			if err != nil {
				return nil, err
			}
		}

		labels[scenarioassignment.SubaccountIDKey] = consumerInfo.ConsumerID
	} else {
		selfRegLabel := r.selfRegManager.GetSelfRegDistinguishingLabelKey()
		if _, distinguishLabelExists := labels[selfRegLabel]; distinguishLabelExists {
			log.C(ctx).Errorf("Label %s is forbidden in a non-cert flow.", selfRegLabel)
			return nil, errors.Errorf("label %s is forbidden when creating Application Template in a non-cert flow.", selfRegLabel)
		}
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}

	defer func() {
		didRollback := r.transact.RollbackUnlessCommitted(ctx, tx)
		if didRollback {
			labelVal := str.CastOrEmpty(convertedIn.Labels[r.selfRegManager.GetSelfRegDistinguishingLabelKey()])
			if labelVal != "" {
				label, ok := labels[selfregmanager.RegionLabel].(string)
				if !ok {
					log.C(ctx).Errorf("An error occurred while casting region label value to string")
				} else {
					r.cleanupAndLogOnError(ctx, selfRegID, label)
				}
			}
		}
	}()

	ctx = persistence.SaveToContext(ctx, tx)
	if err := r.checkProviderAppTemplateExistence(ctx, labels, convertedIn); err != nil {
		return nil, err
	}

	systemRole, hasSystemRole := labels[r.appTemplateProductLabel]
	_, slisFilterLabelExists := labels[systemfetcher.SlisFilterLabelKey]

	if hasSystemRole && !slisFilterLabelExists {
		log.C(ctx).Infof("Application Template with name %s has system role, but doesn't have slis filter defined, creating it...", convertedIn.Name)

		systemRoleValues, ok := systemRole.([]interface{})
		if !ok {
			return nil, errors.Errorf("invalid format of system roles for application template with ID %s", convertedIn.Name)
		}

		filtersFromSystemRoles := make([]interface{}, 0)

		for _, systemRoleValue := range systemRoleValues {
			systemRoleValueStr, ok := systemRoleValue.(string)
			if !ok {
				return nil, errors.Errorf("system role value should be a string for application template with ID %s", convertedIn.Name)
			}
			slisFilter := map[string]interface{}{
				"productId": systemRoleValueStr,
				"filter":    defaultSlisFilterValueForManagedByProperty,
			}

			filtersFromSystemRoles = append(filtersFromSystemRoles, slisFilter)
		}

		labels[systemfetcher.SlisFilterLabelKey] = filtersFromSystemRoles
	}

	log.C(ctx).Infof("Creating an Application Template with name %s", convertedIn.Name)
	id, err := r.appTemplateSvc.CreateWithLabels(ctx, convertedIn, labels)
	if err != nil {
		return nil, err
	}
	log.C(ctx).Infof("Successfully created an Application Template with name %s and id %s", convertedIn.Name, id)

	if consumerInfo.Flow.IsCertFlow() && consumerInfo.Subject != "" {
		log.C(ctx).Infof("Flow is cert. Preparing to create a certificate subject mapping.")
		if err = r.prepareCertSubjectMapping(ctx, id, consumerInfo.Subject); err != nil {
			return nil, err
		}
	}

	appTemplate, err := r.appTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlAppTemplate, err := r.appTemplateConverter.ToGraphQL(appTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting Application Template with id %s to GraphQL", id)
	}

	for _, wh := range convertedIn.Webhooks {
		if wh.Type == model.WebhookTypeOpenResourceDiscoveryStatic {
			log.C(ctx).Infof("Executing aggregation API call for Application Template with ID %s", id)
			if err := r.ordClient.Aggregate(ctx, "", id); err != nil {
				log.C(ctx).WithError(err).Errorf("Error while calling aggregate API with AppTemplateID %q", id)
			}
			break
		}
	}

	return gqlAppTemplate, nil
}

// Labels retrieve all labels for application template
func (r *Resolver) Labels(ctx context.Context, obj *graphql.ApplicationTemplate, key *string) (graphql.Labels, error) {
	if obj == nil {
		return nil, apperrors.NewInternalError("Application Template cannot be empty")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	itemMap, err := r.appTemplateSvc.ListLabels(ctx, obj.ID)
	if err != nil {
		if strings.Contains(err.Error(), "doesn't exist") {
			return nil, tx.Commit()
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	resultLabels := make(map[string]interface{})

	for _, label := range itemMap {
		if key == nil || label.Key == *key {
			resultLabels[label.Key] = label.Value
		}
	}

	var gqlLabels graphql.Labels = resultLabels
	return gqlLabels, nil
}

// RegisterApplicationFromTemplate registers an Application using Application Template
func (r *Resolver) RegisterApplicationFromTemplate(ctx context.Context, in graphql.ApplicationFromTemplateInput) (*graphql.Application, error) {
	consumerInfo, err := consumer.LoadFromContext(ctx)
	if err != nil {
		return nil, errors.Wrapf(err, "while fetching consumer info from context")
	}

	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Debugf("Extracting Application Template with name %q and consumer id REDACTED_%x from GraphQL input", in.TemplateName, sha256.Sum256([]byte(consumerInfo.ConsumerID)))
	appTemplate, err := r.retrieveAppTemplate(ctx, in.TemplateName, consumerInfo.ConsumerID, in.ID)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Registering an Application from Application Template with name %s", in.TemplateName)
	convertedIn, err := r.appTemplateConverter.ApplicationFromTemplateInputFromGraphQL(appTemplate, in)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Debugf("Preparing ApplicationCreateInput JSON from Application Template with name %s", in.TemplateName)
	appCreateInputJSON, err := r.appTemplateSvc.PrepareApplicationCreateInputJSON(appTemplate, convertedIn.Values)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing ApplicationCreateInput JSON from Application Template with name %s", in.TemplateName)
	}

	log.C(ctx).Debugf("Converting ApplicationCreateInput JSON to GraphQL ApplicationRegistrationInput from Application Template with name %s", in.TemplateName)
	appCreateInputGQL, err := r.appConverter.CreateRegisterInputJSONToGQL(appCreateInputJSON)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting ApplicationCreateInput JSON to GraphQL ApplicationRegistrationInput from Application Template with name %s", in.TemplateName)
	}

	log.C(ctx).Infof("Validating GraphQL ApplicationRegistrationInput from Application Template with name %s", convertedIn.TemplateName)
	if err := inputvalidation.Validate(appCreateInputGQL); err != nil {
		return nil, errors.Wrapf(err, "while validating application input from Application Template with name %s", convertedIn.TemplateName)
	}

	appCreateInputModel, err := r.appConverter.CreateInputFromGraphQL(ctx, appCreateInputGQL)
	if err != nil {
		return nil, errors.Wrap(err, "while converting ApplicationFromTemplate input")
	}

	if appCreateInputModel.Labels == nil {
		appCreateInputModel.Labels = make(map[string]interface{})
	}

	if _, exists := appCreateInputModel.Labels[application.ManagedLabelKey]; !exists {
		appCreateInputModel.Labels[application.ManagedLabelKey] = "false"
	}

	if convertedIn.Labels != nil {
		for k, v := range in.Labels {
			appCreateInputModel.Labels[k] = v
		}
	}

	applicationName, err := extractApplicationNameFromTemplateInput(appCreateInputJSON)
	if err != nil {
		return nil, err
	}

	log.C(ctx).Infof("Creating an Application with name %s from Application Template with name %s", applicationName, in.TemplateName)
	id, err := r.appSvc.CreateFromTemplate(ctx, appCreateInputModel, &appTemplate.ID, false)
	if err != nil {
		return nil, errors.Wrapf(err, "while creating an Application with name %s from Application Template with name %s", applicationName, in.TemplateName)
	}
	log.C(ctx).Infof("Application with name %s and id %s successfully created from Application Template with name %s", applicationName, id, in.TemplateName)

	app, err := r.appSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlApp := r.appConverter.ToGraphQL(app)

	if err := r.ordClient.Aggregate(ctx, app.ID, appTemplate.ID); err != nil {
		log.C(ctx).WithError(err).Errorf("Error while calling aggregate API with AppID %q and AppTemplateID %q", app.ID, id)
	}

	if err := r.ordClient.Aggregate(ctx, app.ID, ""); err != nil {
		log.C(ctx).WithError(err).Errorf("Error while calling aggregate API with AppID %q", app.ID)
	}

	return gqlApp, nil
}

// UpdateApplicationTemplate missing godoc
func (r *Resolver) UpdateApplicationTemplate(ctx context.Context, id string, override *bool, in graphql.ApplicationTemplateUpdateInput) (*graphql.ApplicationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := in.Validate(); err != nil {
		return nil, err
	}

	if err := validateAppTemplateNameBasedOnProvider(in.Name, in.ApplicationInput); err != nil {
		return nil, err
	}

	webhooks, err := r.webhookSvc.EnrichWebhooksWithTenantMappingWebhooks(in.Webhooks)
	if err != nil {
		return nil, err
	}

	if in.Webhooks != nil {
		in.Webhooks = webhooks
	}

	convertedIn, err := r.appTemplateConverter.UpdateInputFromGraphQL(in)
	if err != nil {
		return nil, err
	}

	labels, err := r.appTemplateSvc.ListLabels(ctx, id)
	if err != nil {
		return nil, err
	}

	resultLabels := make(map[string]interface{}, len(labels))
	for _, label := range labels {
		resultLabels[label.Key] = label.Value
	}

	isSelfRegFlow, err := r.selfRegManager.IsSelfRegistrationFlow(ctx, resultLabels)
	if err != nil {
		return nil, err
	}
	if isSelfRegFlow {
		if err := validateAppTemplateForSelfReg(in.ApplicationInput); err != nil {
			return nil, err
		}
	}

	shouldOverride := false
	if override != nil {
		shouldOverride = *override
	}

	log.C(ctx).Infof("Updating an Application Template with id %q", id)
	err = r.appTemplateSvc.Update(ctx, id, shouldOverride, convertedIn)
	if err != nil {
		return nil, err
	}

	appTemplate, err := r.appTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	gqlAppTemplate, err := r.appTemplateConverter.ToGraphQL(appTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting application template to graphql")
	}

	return gqlAppTemplate, nil
}

// DeleteApplicationTemplate missing godoc
func (r *Resolver) DeleteApplicationTemplate(ctx context.Context, id string) (*graphql.ApplicationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	appTemplate, err := r.appTemplateSvc.Get(ctx, id)
	if err != nil {
		return nil, err
	}

	_, err = r.appTemplateSvc.GetLabel(ctx, id, r.selfRegManager.GetSelfRegDistinguishingLabelKey())
	if err != nil {
		if !apperrors.IsNotFoundError(err) {
			return nil, errors.Wrapf(err, "while getting self register label")
		}
	} else {
		regionLabel, err := r.appTemplateSvc.GetLabel(ctx, id, selfregmanager.RegionLabel)
		if err != nil {
			return nil, errors.Wrapf(err, "while getting region label")
		}

		// Committing transaction as the cleanup sends request to external service
		if err = tx.Commit(); err != nil {
			return nil, err
		}

		regionValue, ok := regionLabel.Value.(string)
		if !ok {
			return nil, errors.Wrap(err, "while casting region label value to string")
		}

		log.C(ctx).Infof("Executing clean-up for self-registered app template with id %q", id)
		if err := r.selfRegManager.CleanupSelfRegistration(ctx, id, regionValue); err != nil {
			return nil, errors.Wrap(err, "An error occurred during cleanup of self-registered app template: ")
		}

		tx, err = r.transact.Begin()
		if err != nil {
			return nil, err
		}
		ctx = persistence.SaveToContext(ctx, tx)
	}

	log.C(ctx).Infof("Deleting an Application Template with id %q", id)
	err = r.appTemplateSvc.Delete(ctx, id)
	if err != nil {
		return nil, err
	}

	if err := r.certSubjectMappingSvc.DeleteByConsumerID(ctx, id); err != nil {
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	deletedAppTemplate, err := r.appTemplateConverter.ToGraphQL(appTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting application template to graphql")
	}

	return deletedAppTemplate, nil
}

// Webhooks missing godoc
func (r *Resolver) Webhooks(ctx context.Context, obj *graphql.ApplicationTemplate) ([]*graphql.Webhook, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	webhooks, err := r.webhookSvc.ListForApplicationTemplate(ctx, obj.ID)
	if err != nil {
		return nil, err
	}

	if err := tx.Commit(); err != nil {
		return nil, err
	}

	return r.webhookConverter.MultipleToGraphQL(webhooks)
}

func extractApplicationNameFromTemplateInput(applicationInputJSON string) (string, error) {
	b := []byte(applicationInputJSON)
	data := make(map[string]interface{})

	err := json.Unmarshal(b, &data)
	if err != nil {
		return "", errors.Wrap(err, "while unmarshalling application input JSON")
	}

	return data["name"].(string), nil
}

func (r *Resolver) cleanupAndLogOnError(ctx context.Context, id, region string) {
	if err := r.selfRegManager.CleanupSelfRegistration(ctx, id, region); err != nil {
		log.C(ctx).Errorf("An error occurred during cleanup of self-registered app template: %v", err)
	}
}

func (r *Resolver) retrieveAppTemplate(ctx context.Context,
	appTemplateName, consumerID string, appTemplateID *string) (*model.ApplicationTemplate, error) {
	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(globalSubaccountIDLabelKey, fmt.Sprintf("\"%s\"", consumerID)),
	}
	appTemplates, err := r.appTemplateSvc.ListByFilters(ctx, filters)
	if err != nil {
		return nil, err
	}

	for _, appTemplate := range appTemplates {
		if (appTemplateID == nil && appTemplate.Name == appTemplateName) ||
			(appTemplateID != nil && *appTemplateID == appTemplate.ID) {
			return appTemplate, nil
		}
	}

	appTemplates, err = r.appTemplateSvc.ListByName(ctx, appTemplateName)
	if err != nil {
		return nil, err
	}
	templates := make([]*model.ApplicationTemplate, 0, len(appTemplates))
	for _, appTemplate := range appTemplates {
		_, err := r.appTemplateSvc.GetLabel(ctx, appTemplate.ID, globalSubaccountIDLabelKey)
		if err != nil && !apperrors.IsNotFoundError(err) {
			return nil, errors.Wrapf(err, "while getting %q label", globalSubaccountIDLabelKey)
		}
		if err != nil && apperrors.IsNotFoundError(err) {
			templates = append(templates, appTemplate)
		}
	}

	if appTemplateID != nil {
		log.C(ctx).Infof("searching for application template with ID: %s", *appTemplateID)
		for _, appTemplate := range appTemplates {
			if appTemplate.ID == *appTemplateID {
				log.C(ctx).Infof("found application template with ID: %s", *appTemplateID)
				return appTemplate, nil
			}
		}
		return nil, errors.Errorf("application template with id %s and consumer id REDACTED_%x not found", *appTemplateID, sha256.Sum256([]byte(consumerID)))
	}
	if len(templates) < 1 {
		return nil, errors.Errorf("application template with name %q and consumer id REDACTED_%x not found", appTemplateName, sha256.Sum256([]byte(consumerID)))
	}
	if len(templates) > 1 {
		return nil, errors.Errorf("unexpected number of application templates. found %d", len(appTemplates))
	}
	return templates[0], nil
}

func validateAppTemplateForSelfReg(applicationInput *graphql.ApplicationJSONInput) error {
	appNameExists := applicationInput.Name != ""
	var appDisplayNameLabelExists bool

	if displayName, ok := applicationInput.Labels[displayNameLabelKey]; ok {
		displayNameValue, ok := displayName.(string)
		if !ok {
			return fmt.Errorf("%q label value must be string", displayNameLabelKey)
		}
		appDisplayNameLabelExists = displayNameValue != ""
	}

	if !appNameExists || !appDisplayNameLabelExists {
		return errors.Errorf("applicationInputJSON name property or applicationInputJSON displayName label is missing. They must be present in order to proceed.")
	}

	return nil
}

func validateAppTemplateNameBasedOnProvider(name string, appInput *graphql.ApplicationJSONInput) error {
	if appInput == nil || appInput.ProviderName == nil || str.PtrStrToStr(appInput.ProviderName) != sapProviderName {
		return nil
	}

	// Matches the following pattern - "SAP <product name>"
	r := regexp.MustCompile(`(^SAP\s)([A-Za-z0-9()_\- ]*)`)
	matches := r.FindStringSubmatch(name)
	if len(matches) == 0 {
		return errors.Errorf("application template name %q does not comply with the following naming convention: %q", name, "SAP <product name>")
	}

	return nil
}

func (r *Resolver) checkProviderAppTemplateExistence(ctx context.Context, labels map[string]interface{}, convertedIn model.ApplicationTemplateInput) error {
	if err := r.checkAppTemplateExistenceByDistinguishLabel(ctx, labels); err != nil {
		return err
	}

	if err := r.checkAppTemplateExistenceByProductLabel(ctx, labels, convertedIn); err != nil {
		return err
	}

	return nil
}

func (r *Resolver) checkAppTemplateExistenceByProductLabel(ctx context.Context, labels map[string]interface{}, appTemplateInput model.ApplicationTemplateInput) error {
	log.C(ctx).Infof("Checking Application Template existence by %q label", r.appTemplateProductLabel)

	regionLabelKey := selfregmanager.RegionLabel
	appTemplateRegion, isRegionalAppTemplate := labels[regionLabelKey]
	productLabel, productLabelExists := labels[r.appTemplateProductLabel]
	if !productLabelExists {
		log.C(ctx).Infof("%q label does not exist. Skipping the check.", r.appTemplateProductLabel)
		return nil
	}

	labelsKeyRegionLabel := fmt.Sprintf("%s.%s", labelsKey, regionLabelKey)
	hasRegionLabelInAppInputJSON := gjson.Get(appTemplateInput.ApplicationInputJSON, labelsKeyRegionLabel).Exists()

	if isRegionalAppTemplate {
		if !hasRegionLabelInAppInputJSON {
			return errors.Errorf("App Template with %q label has a missing %q label in the applicationInput", regionLabelKey, regionLabelKey)
		}

		if _, err := extractRegionPlaceholder(appTemplateInput.Placeholders); err != nil {
			return errors.Wrapf(err, "for regional Application Template input")
		}
	}

	productLabelArr, ok := productLabel.([]interface{})
	if !ok {
		return errors.Errorf("could not parse %q label for application template - it must be a string array", r.appTemplateProductLabel)
	}

	log.C(ctx).Infof("Getting application template for labels %q: %q", r.appTemplateProductLabel, productLabel)
	appTemplates, err := r.appTemplateSvc.ListByFilters(ctx, r.buildProductLabelFilter(productLabelArr))
	if err != nil {
		if apperrors.IsNotFoundError(err) {
			log.C(ctx).Infof("There are no application templates for the filter. Proceeding.")
			return nil
		}
		return errors.Wrapf(err, "while getting Application Template for labels %q: %q", r.appTemplateProductLabel, productLabel)
	}

	existingRegions := make([]string, 0)
	for _, appTemplate := range appTemplates {
		regionLabel, err := r.labelSvc.GetByKey(ctx, "", model.AppTemplateLabelableObject, appTemplate.ID, regionLabelKey)
		if err != nil {
			if apperrors.IsNotFoundError(err) {
				continue
			}
			return errors.Wrapf(err, "while getting %q label for Application Template", regionLabelKey)
		}

		existingRegions = append(existingRegions, regionLabel.Value.(string))
	}

	if len(appTemplates) > 0 {
		if len(existingRegions) == 0 {
			return errors.Errorf("Application Template with %q label is global and already exists", r.appTemplateProductLabel)
		}

		if len(existingRegions) > 0 && !isRegionalAppTemplate {
			return errors.Errorf("Existing application template with %q label is regional. The input application template should contain a %q label", r.appTemplateProductLabel, regionLabelKey)
		}

		if isRegionalAppTemplate {
			for _, existingRegion := range existingRegions {
				if existingRegion == appTemplateRegion.(string) {
					return errors.Errorf("Regional Application Template with %q label and %q: %q already exists", r.appTemplateProductLabel, regionLabelKey, existingRegion)
				}
			}

			isPlaceholderEqual, err := isRegionPlaceholderEqualToExistingPlaceholder(appTemplateInput.Placeholders, appTemplates[0].Placeholders)
			if err != nil {
				return err
			}

			if !isPlaceholderEqual {
				return errors.Errorf("Regional Application Template input with %q label has a different %q placeholder from the other Application Templates with the same label", r.appTemplateProductLabel, regionLabelKey)
			}
		}
	}

	return nil
}

func (r *Resolver) checkAppTemplateExistenceByDistinguishLabel(ctx context.Context, labels map[string]interface{}) error {
	selfRegisterDistinguishLabelKey := r.selfRegManager.GetSelfRegDistinguishingLabelKey()

	log.C(ctx).Infof("Checking Application Template existence by %q label", selfRegisterDistinguishLabelKey)

	regionLabelKey := selfregmanager.RegionLabel
	appTemplateRegion, regionExists := labels[regionLabelKey]
	appTemplateDistinguishLabel, exists := labels[selfRegisterDistinguishLabelKey]
	if !exists {
		log.C(ctx).Infof("%q label does not exist. Skipping the check.", selfRegisterDistinguishLabelKey)
		return nil
	}

	msg := fmt.Sprintf("%q: %q", selfRegisterDistinguishLabelKey, appTemplateDistinguishLabel)

	filters := []*labelfilter.LabelFilter{
		labelfilter.NewForKeyWithQuery(selfRegisterDistinguishLabelKey, fmt.Sprintf("\"%s\"", appTemplateDistinguishLabel)),
	}

	if regionExists {
		if _, ok := appTemplateRegion.(string); !ok {
			return errors.Errorf("%s label should be string", regionLabelKey)
		}
		filters = append(filters, labelfilter.NewForKeyWithQuery(regionLabelKey, fmt.Sprintf("\"%s\"", appTemplateRegion)))
		msg += fmt.Sprintf(" and %q: %q", regionLabelKey, appTemplateRegion)
	}

	log.C(ctx).Infof("Getting application template for labels %s", msg)
	appTemplate, err := r.appTemplateSvc.GetByFilters(ctx, filters)
	if err != nil && !apperrors.IsNotFoundError(err) {
		return errors.Wrap(err, fmt.Sprintf("Failed to get application template for labels %s", msg))
	}

	if appTemplate != nil {
		errMsg := fmt.Sprintf("Cannot have more than one application template with labels %s", msg)
		log.C(ctx).Error(errMsg)
		return errors.New(errMsg)
	}

	return nil
}

func (r *Resolver) buildProductLabelFilter(productLabelArr []interface{}) []*labelfilter.LabelFilter {
	filters := make([]*labelfilter.LabelFilter, 0, len(productLabelArr))
	for _, productLabelValue := range productLabelArr {
		productLabelStr, _ := productLabelValue.(string)
		query := fmt.Sprintf(`$[*] ? (@ == "%s")`, productLabelStr)
		filters = append(filters, labelfilter.NewForKeyWithQuery(r.appTemplateProductLabel, query))
	}

	return filters
}

func isRegionPlaceholderEqualToExistingPlaceholder(inputPlaceholders, existingPlaceholders []model.ApplicationTemplatePlaceholder) (bool, error) {
	inputRegionPlaceholder, err := extractRegionPlaceholder(inputPlaceholders)
	if err != nil {
		return false, errors.Wrapf(err, "for Application Template input")
	}
	existingRegionPlaceholder, err := extractRegionPlaceholder(existingPlaceholders)
	if err != nil {
		return false, errors.Wrapf(err, "for existing Application Template")
	}
	return inputRegionPlaceholder == existingRegionPlaceholder, nil
}

func extractRegionPlaceholder(placeholders []model.ApplicationTemplatePlaceholder) (string, error) {
	regionKey := selfregmanager.RegionLabel

	regionPlaceholder := ""
	for _, placeholder := range placeholders {
		if placeholder.Name == regionKey {
			regionPlaceholder = str.PtrStrToStr(placeholder.JSONPath)
			break
		}
	}

	if regionPlaceholder == "" {
		return "", errors.Errorf("%q placeholder should be present for regional Application Templates", regionKey)
	}

	return regionPlaceholder, nil
}

func (r *Resolver) prepareCertSubjectMapping(ctx context.Context, appTemplateID, subject string) error {
	for _, consumerSubject := range r.envConsumerSubjects {
		if cert.SubjectsMatch(subject, consumerSubject) {
			log.C(ctx).Info("Subject matches with a known environment consumer subject. Skipping certificate subject mapping creation.")
			return nil
		}
	}

	certSubjMappings, err := r.certSubjectMappingSvc.ListAll(ctx)
	if err != nil {
		return errors.Wrapf(err, "while checking if a certificate subject mapping exists with a subject: %s", subject)
	}

	for _, csm := range certSubjMappings {
		if cert.SubjectsMatch(subject, csm.Subject) {
			return fmt.Errorf("subject is already allow-listed. Not possible to associate app template consumer %q with already allow-listed subject", appTemplateID)
		}
	}

	id := r.uidService.Generate()
	model := &model.CertSubjectMapping{
		ID:                 id,
		Subject:            subject,
		ConsumerType:       string(consumer.ApplicationProvider),
		InternalConsumerID: &appTemplateID,
		TenantAccessLevels: []string{inputvalidation.GlobalAccessLevel},
	}

	if _, err := r.certSubjectMappingSvc.Create(ctx, model); err != nil {
		return errors.Wrapf(err, "while creating a cert subject mapping for app template consumer %q", appTemplateID)
	}

	log.C(ctx).Infof("Successfully created a certificate subject mapping for Application Template with ID %q", appTemplateID)

	return nil
}
