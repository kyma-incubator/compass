package apptemplate

import (
	"context"
	"encoding/json"
	"regexp"
	"strings"

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

// ApplicationTemplateService missing godoc
//go:generate mockery --name=ApplicationTemplateService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateService interface {
	Create(ctx context.Context, in model.ApplicationTemplateInput) (string, error)
	CreateWithLabels(ctx context.Context, in model.ApplicationTemplateInput, labels map[string]interface{}) (string, error)
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	GetByName(ctx context.Context, name string) (*model.ApplicationTemplate, error)
	List(ctx context.Context, pageSize int, cursor string) (model.ApplicationTemplatePage, error)
	Update(ctx context.Context, id string, in model.ApplicationTemplateUpdateInput) error
	Delete(ctx context.Context, id string) error
	PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, values model.ApplicationFromTemplateInputValues) (string, error)
	ListLabels(ctx context.Context, appTemplateID string) (map[string]*model.Label, error)
	GetLabel(ctx context.Context, appTemplateID string, key string) (*model.Label, error)
}

// ApplicationTemplateConverter missing godoc
//go:generate mockery --name=ApplicationTemplateConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateConverter interface {
	ToGraphQL(in *model.ApplicationTemplate) (*graphql.ApplicationTemplate, error)
	MultipleToGraphQL(in []*model.ApplicationTemplate) ([]*graphql.ApplicationTemplate, error)
	InputFromGraphQL(in graphql.ApplicationTemplateInput) (model.ApplicationTemplateInput, error)
	UpdateInputFromGraphQL(in graphql.ApplicationTemplateUpdateInput) (model.ApplicationTemplateUpdateInput, error)
	ApplicationFromTemplateInputFromGraphQL(in graphql.ApplicationFromTemplateInput) model.ApplicationFromTemplateInput
}

// ApplicationConverter missing godoc
//go:generate mockery --name=ApplicationConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationConverter interface {
	ToGraphQL(in *model.Application) *graphql.Application
	CreateInputJSONToGQL(in string) (graphql.ApplicationRegisterInput, error)
	CreateInputFromGraphQL(ctx context.Context, in graphql.ApplicationRegisterInput) (model.ApplicationRegisterInput, error)
}

// ApplicationService missing godoc
//go:generate mockery --name=ApplicationService --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationService interface {
	Create(ctx context.Context, in model.ApplicationRegisterInput) (string, error)
	CreateFromTemplate(ctx context.Context, in model.ApplicationRegisterInput, appTemplateID *string) (string, error)
	Get(ctx context.Context, id string) (*model.Application, error)
}

// WebhookService missing godoc
//go:generate mockery --name=WebhookService --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookService interface {
	ListForApplicationTemplate(ctx context.Context, applicationTemplateID string) ([]*model.Webhook, error)
}

// WebhookConverter missing godoc
//go:generate mockery --name=WebhookConverter --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookConverter interface {
	MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error)
	MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error)
}

// SelfRegisterManager missing godoc
//go:generate mockery --name=SelfRegisterManager --output=automock --outpkg=automock --case=underscore --disable-version-string
type SelfRegisterManager interface {
	PrepareForSelfRegistration(ctx context.Context, resourceType resource.Type, labels map[string]interface{}, id string, validate func() error) (map[string]interface{}, error)
	CleanupSelfRegistration(ctx context.Context, selfRegisterLabelValue, region string) error
	GetSelfRegDistinguishingLabelKey() string
}

// Resolver missing godoc
type Resolver struct {
	transact persistence.Transactioner

	appSvc               ApplicationService
	appConverter         ApplicationConverter
	appTemplateSvc       ApplicationTemplateService
	appTemplateConverter ApplicationTemplateConverter
	webhookSvc           WebhookService
	webhookConverter     WebhookConverter
	selfRegManager       SelfRegisterManager
	uidService           UIDService
}

// NewResolver missing godoc
func NewResolver(transact persistence.Transactioner, appSvc ApplicationService, appConverter ApplicationConverter, appTemplateSvc ApplicationTemplateService, appTemplateConverter ApplicationTemplateConverter, webhookService WebhookService, webhookConverter WebhookConverter, selfRegisterManager SelfRegisterManager, uidService UIDService) *Resolver {
	return &Resolver{
		transact:             transact,
		appSvc:               appSvc,
		appConverter:         appConverter,
		appTemplateSvc:       appTemplateSvc,
		appTemplateConverter: appTemplateConverter,
		webhookSvc:           webhookService,
		webhookConverter:     webhookConverter,
		selfRegManager:       selfRegisterManager,
		uidService:           uidService,
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
		if apperrors.IsNotFoundError(err) {
			return nil, tx.Commit()
		}
		return nil, err
	}

	err = tx.Commit()
	if err != nil {
		return nil, err
	}

	out, err := r.appTemplateConverter.ToGraphQL(appTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while converting application template to graphql")
	}

	return out, nil
}

// ApplicationTemplates missing godoc
func (r *Resolver) ApplicationTemplates(ctx context.Context, first *int, after *graphql.PageCursor) (*graphql.ApplicationTemplatePage, error) {
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

	appTemplatePage, err := r.appTemplateSvc.List(ctx, *first, cursor)
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

// CreateApplicationTemplate missing godoc
func (r *Resolver) CreateApplicationTemplate(ctx context.Context, in graphql.ApplicationTemplateInput) (*graphql.ApplicationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := in.Validate(); err != nil {
		return nil, err
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
	validate := func() error {
		return validateAppTemplateInputForSelfReg(convertedIn)
	}
	labels, err := r.selfRegManager.PrepareForSelfRegistration(ctx, resource.ApplicationTemplate, convertedIn.Labels, selfRegID, validate)
	if err != nil {
		return nil, err
	}

	defer func() {
		didRollback := r.transact.RollbackUnlessCommitted(ctx, tx)
		if didRollback {
			labelVal := str.CastOrEmpty(convertedIn.Labels[r.selfRegManager.GetSelfRegDistinguishingLabelKey()])
			if labelVal != "" {
				label, ok := in.Labels[selfregmanager.RegionLabel].(string)
				if !ok {
					log.C(ctx).Errorf("An error occurred while casting region label value to string")
				} else {
					r.cleanupAndLogOnError(ctx, selfRegID, label)
				}
			}
		}
	}()

	log.C(ctx).Infof("Creating an Application Template with name %s", convertedIn.Name)
	id, err := r.appTemplateSvc.CreateWithLabels(ctx, convertedIn, labels)
	if err != nil {
		return nil, err
	}
	log.C(ctx).Infof("Successfully created an Application Template with name %s and id %s", convertedIn.Name, id)

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

// RegisterApplicationFromTemplate missing godoc
func (r *Resolver) RegisterApplicationFromTemplate(ctx context.Context, in graphql.ApplicationFromTemplateInput) (*graphql.Application, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	log.C(ctx).Infof("Registering an Application from Application Template with name %s", in.TemplateName)
	convertedIn := r.appTemplateConverter.ApplicationFromTemplateInputFromGraphQL(in)

	log.C(ctx).Debugf("Extracting Application Template with name %s from GraphQL input", in.TemplateName)
	appTemplate, err := r.appTemplateSvc.GetByName(ctx, convertedIn.TemplateName)
	if err != nil {
		return nil, err
	}
	applicationName, err := extractApplicationNameFromTemplateInput(appTemplate.ApplicationInputJSON)
	if err != nil {
		return nil, err
	}
	log.C(ctx).Infof("Registering an Application with name %s from Application Template with name %s", applicationName, in.TemplateName)

	log.C(ctx).Debugf("Preparing ApplicationCreateInput JSON from Application Template with name %s", in.TemplateName)
	appCreateInputJSON, err := r.appTemplateSvc.PrepareApplicationCreateInputJSON(appTemplate, convertedIn.Values)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing ApplicationCreateInput JSON from Application Template with name %s", in.TemplateName)
	}

	log.C(ctx).Debugf("Converting ApplicationCreateInput JSON to GraphQL ApplicationRegistrationInput from Application Template with name %s", in.TemplateName)
	appCreateInputGQL, err := r.appConverter.CreateInputJSONToGQL(appCreateInputJSON)
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
	appCreateInputModel.Labels["managed"] = "false"

	log.C(ctx).Infof("Creating an Application with name %s from Application Template with name %s", applicationName, in.TemplateName)
	id, err := r.appSvc.CreateFromTemplate(ctx, appCreateInputModel, &appTemplate.ID)
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
	return gqlApp, nil
}

// UpdateApplicationTemplate missing godoc
func (r *Resolver) UpdateApplicationTemplate(ctx context.Context, id string, in graphql.ApplicationTemplateUpdateInput) (*graphql.ApplicationTemplate, error) {
	tx, err := r.transact.Begin()
	if err != nil {
		return nil, err
	}
	defer r.transact.RollbackUnlessCommitted(ctx, tx)

	ctx = persistence.SaveToContext(ctx, tx)

	if err := in.Validate(); err != nil {
		return nil, err
	}

	convertedIn, err := r.appTemplateConverter.UpdateInputFromGraphQL(in)
	if err != nil {
		return nil, err
	}
	// TODO check name before update
	err = r.appTemplateSvc.Update(ctx, id, convertedIn)
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

	err = r.appTemplateSvc.Delete(ctx, id)
	if err != nil {
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

func validateAppTemplateInputForSelfReg(appTemplateInput model.ApplicationTemplateInput) error {
	if len(appTemplateInput.Placeholders) != 2 {
		return errors.Errorf("expecting %q and %q placeholders", "name", "display-name")
	}

	for _, placeholder := range appTemplateInput.Placeholders {
		if !(placeholder.Name == "name") && !(placeholder.Name == "display-name") {
			return errors.Errorf("unexpected placeholder with name %q found", placeholder.Name)
		}
	}

	// Matches the following pattern - "SAP <product name> (<region>)"
	r := regexp.MustCompile(`(^SAP\s)([A-Za-z0-9_\- ]*)\s[(]([A-Za-z0-9_\- ]*)[)]`)
	matches := r.FindStringSubmatch(appTemplateInput.Name)
	if len(matches) == 0 {
		return errors.Errorf("application template name %q does not comply with the following naming convention: %q", appTemplateInput.Name, "SAP <product name> (<region>)")
	}

	regionValue, exists := appTemplateInput.Labels[selfregmanager.RegionLabel]
	if !exists {
		return errors.Errorf("missing %q label", selfregmanager.RegionLabel)
	}

	region, ok := regionValue.(string)
	if !ok {
		return errors.Errorf("region value should be of type %q", "string")
	}

	if matches[len(matches)-1] != region {
		return errors.Errorf("the region specified in the application template name does not match %q", region)
	}

	return nil
}
