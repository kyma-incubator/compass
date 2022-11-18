package apptemplate

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const applicationTypeLabelKey = "applicationType"
const otherSystemType = "Other System Type"

// ApplicationTemplateRepository missing godoc
//
//go:generate mockery --name=ApplicationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type ApplicationTemplateRepository interface {
	Create(ctx context.Context, item model.ApplicationTemplate) error
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	GetByFilters(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.ApplicationTemplate, error)
	Exists(ctx context.Context, id string) (bool, error)
	List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (model.ApplicationTemplatePage, error)
	ListByName(ctx context.Context, id string) ([]*model.ApplicationTemplate, error)
	ListByFilters(ctx context.Context, filter []*labelfilter.LabelFilter) ([]*model.ApplicationTemplate, error)
	Update(ctx context.Context, model model.ApplicationTemplate) error
	Delete(ctx context.Context, id string) error
}

// UIDService missing godoc
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

// WebhookRepository missing godoc
//
//go:generate mockery --name=WebhookRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type WebhookRepository interface {
	CreateMany(ctx context.Context, tenant string, items []*model.Webhook) error
}

// LabelUpsertService missing godoc
//
//go:generate mockery --name=LabelUpsertService --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelUpsertService interface {
	UpsertMultipleLabels(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labels map[string]interface{}) error
}

// LabelRepository missing godoc
//
//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelRepository interface {
	ListForGlobalObject(ctx context.Context, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

type service struct {
	appTemplateRepo    ApplicationTemplateRepository
	webhookRepo        WebhookRepository
	uidService         UIDService
	labelUpsertService LabelUpsertService
	labelRepo          LabelRepository
}

// NewService missing godoc
func NewService(appTemplateRepo ApplicationTemplateRepository, webhookRepo WebhookRepository, uidService UIDService, labelUpsertService LabelUpsertService, labelRepo LabelRepository) *service {
	return &service{
		appTemplateRepo:    appTemplateRepo,
		webhookRepo:        webhookRepo,
		uidService:         uidService,
		labelUpsertService: labelUpsertService,
		labelRepo:          labelRepo,
	}
}

// Create missing godoc
func (s *service) Create(ctx context.Context, in model.ApplicationTemplateInput) (string, error) {
	appTemplateID := s.uidService.Generate()
	if len(str.PtrStrToStr(in.ID)) > 0 {
		appTemplateID = *in.ID
	}

	if in.Labels == nil {
		in.Labels = map[string]interface{}{}
	}

	log.C(ctx).Debugf("ID %s generated for Application Template with name %s", appTemplateID, in.Name)

	appInputJSON, err := enrichWithApplicationTypeLabel(in.ApplicationInputJSON, in.Name)
	if err != nil {
		return "", err
	}
	in.ApplicationInputJSON = appInputJSON

	region := in.Labels[tenant.RegionLabelKey]
	_, err = s.GetByNameAndRegion(ctx, in.Name, region)
	if err != nil && !apperrors.IsNotFoundError(err) {
		return "", errors.Wrapf(err, "while checking if application template with name %q and region %v exists", in.Name, region)
	}
	if err == nil {
		return "", fmt.Errorf("application template with name %q and region %v already exists", in.Name, region)
	}

	appTemplate := in.ToApplicationTemplate(appTemplateID)

	err = s.appTemplateRepo.Create(ctx, appTemplate)
	if err != nil {
		return "", errors.Wrapf(err, "while creating Application Template with name %s", in.Name)
	}

	webhooks := make([]*model.Webhook, 0, len(in.Webhooks))
	for _, item := range in.Webhooks {
		webhooks = append(webhooks, item.ToWebhook(s.uidService.Generate(), appTemplateID, model.ApplicationTemplateWebhookReference))
	}
	if err = s.webhookRepo.CreateMany(ctx, "", webhooks); err != nil {
		return "", errors.Wrapf(err, "while creating Webhooks for applicationTemplate")
	}

	err = s.labelUpsertService.UpsertMultipleLabels(ctx, "", model.AppTemplateLabelableObject, appTemplateID, in.Labels)
	if err != nil {
		return appTemplateID, errors.Wrapf(err, "while creating multiple labels for Application Template with id %s", appTemplateID)
	}

	return appTemplateID, nil
}

// CreateWithLabels Creates an AppTemplate with provided labels
func (s *service) CreateWithLabels(ctx context.Context, in model.ApplicationTemplateInput, labels map[string]interface{}) (string, error) {
	for key, val := range labels {
		in.Labels[key] = val
	}

	appTemplateID, err := s.Create(ctx, in)
	if err != nil {
		return "", errors.Wrapf(err, "while creating Application Template")
	}

	return appTemplateID, nil
}

// Get missing godoc
func (s *service) Get(ctx context.Context, id string) (*model.ApplicationTemplate, error) {
	appTemplate, err := s.appTemplateRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application Template with id %s", id)
	}

	return appTemplate, nil
}

// GetByFilters gets a model.ApplicationTemplate by given slice of labelfilter.LabelFilter
func (s *service) GetByFilters(ctx context.Context, filter []*labelfilter.LabelFilter) (*model.ApplicationTemplate, error) {
	appTemplate, err := s.appTemplateRepo.GetByFilters(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while getting Application Template by filters")
	}

	return appTemplate, nil
}

// ListByName retrieves all Application Templates by given name
func (s *service) ListByName(ctx context.Context, name string) ([]*model.ApplicationTemplate, error) {
	appTemplates, err := s.appTemplateRepo.ListByName(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing application templates with name %q", name)
	}

	return appTemplates, nil
}

// ListByFilters retrieves all Application Templates by given slice of labelfilter.LabelFilter
func (s *service) ListByFilters(ctx context.Context, filter []*labelfilter.LabelFilter) ([]*model.ApplicationTemplate, error) {
	appTemplates, err := s.appTemplateRepo.ListByFilters(ctx, filter)
	if err != nil {
		return nil, errors.Wrap(err, "while listing application templates by filters")
	}

	return appTemplates, nil
}

// GetByNameAndRegion retrieves Application Template by given name and region
func (s *service) GetByNameAndRegion(ctx context.Context, name string, region interface{}) (*model.ApplicationTemplate, error) {
	appTemplates, err := s.appTemplateRepo.ListByName(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing application templates with name %q", name)
	}

	for _, appTemplate := range appTemplates {
		appTmplRegion, err := s.retrieveLabel(ctx, appTemplate.ID, tenant.RegionLabelKey)
		if err != nil && !apperrors.IsNotFoundError(err) {
			return nil, err
		}

		if region == appTmplRegion {
			log.C(ctx).Infof("Found Application Template with name %q and region label %v", name, region)
			return appTemplate, nil
		}
	}

	return nil, apperrors.NewNotFoundErrorWithType(resource.ApplicationTemplate)
}

// ListLabels retrieves all labels for application template
func (s *service) ListLabels(ctx context.Context, appTemplateID string) (map[string]*model.Label, error) {
	appTemplateExists, err := s.appTemplateRepo.Exists(ctx, appTemplateID)
	if err != nil {
		return nil, errors.Wrap(err, "while checking Application Template existence")
	}

	if !appTemplateExists {
		return nil, fmt.Errorf("application template with ID %s doesn't exist", appTemplateID)
	}

	labels, err := s.labelRepo.ListForGlobalObject(ctx, model.AppTemplateLabelableObject, appTemplateID) // tenent is not needed for AppTemplateLabelableObject
	if err != nil {
		return nil, errors.Wrap(err, "while getting labels for Application Template")
	}

	return labels, nil
}

// GetLabel gets a given label for application template
func (s *service) GetLabel(ctx context.Context, appTemplateID string, key string) (*model.Label, error) {
	labels, err := s.ListLabels(ctx, appTemplateID)
	if err != nil {
		return nil, err
	}

	label, ok := labels[key]
	if !ok {
		return nil, apperrors.NewNotFoundErrorWithMessage(resource.Label, "", fmt.Sprintf("label %s for application template with ID %s doesn't exist", key, appTemplateID))
	}

	return label, nil
}

// Exists missing godoc
func (s *service) Exists(ctx context.Context, id string) (bool, error) {
	exist, err := s.appTemplateRepo.Exists(ctx, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Application Template with ID %s", id)
	}

	return exist, nil
}

// List missing godoc
func (s *service) List(ctx context.Context, filter []*labelfilter.LabelFilter, pageSize int, cursor string) (model.ApplicationTemplatePage, error) {
	if pageSize < 1 || pageSize > 200 {
		return model.ApplicationTemplatePage{}, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.appTemplateRepo.List(ctx, filter, pageSize, cursor)
}

// Update missing godoc
func (s *service) Update(ctx context.Context, id string, in model.ApplicationTemplateUpdateInput) error {
	oldAppTemplate, err := s.Get(ctx, id)
	if err != nil {
		return err
	}

	region, err := s.retrieveLabel(ctx, id, tenant.RegionLabelKey)
	if err != nil && !apperrors.IsNotFoundError(err) {
		return err
	}

	appInputJSON, err := enrichWithApplicationTypeLabel(in.ApplicationInputJSON, in.Name)
	if err != nil {
		return err
	}
	in.ApplicationInputJSON = appInputJSON

	if oldAppTemplate.Name != in.Name {
		_, err := s.GetByNameAndRegion(ctx, in.Name, region)
		if err != nil && !apperrors.IsNotFoundError(err) {
			return errors.Wrapf(err, "while checking if application template with name %q and region %v exists", in.Name, region)
		}
		if err == nil {
			return fmt.Errorf("application template with name %q and region %v already exists", in.Name, region)
		}
	}

	appTemplate := in.ToApplicationTemplate(id)

	err = s.appTemplateRepo.Update(ctx, appTemplate)
	if err != nil {
		return errors.Wrapf(err, "while updating Application Template with ID %s", id)
	}

	return nil
}

// Delete missing godoc
func (s *service) Delete(ctx context.Context, id string) error {
	err := s.appTemplateRepo.Delete(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Application Template with ID %s", id)
	}

	return nil
}

// PrepareApplicationCreateInputJSON missing godoc
func (s *service) PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, values model.ApplicationFromTemplateInputValues) (string, error) {
	appCreateInputJSON := appTemplate.ApplicationInputJSON
	for _, placeholder := range appTemplate.Placeholders {
		newValue, err := values.FindPlaceholderValue(placeholder.Name)
		if err != nil {
			return "", errors.Wrap(err, "required placeholder not provided")
		}
		appCreateInputJSON = strings.ReplaceAll(appCreateInputJSON, fmt.Sprintf("{{%s}}", placeholder.Name), newValue)
	}
	return appCreateInputJSON, nil
}

func (s *service) retrieveLabel(ctx context.Context, id string, labelKey string) (interface{}, error) {
	label, err := s.labelRepo.GetByKey(ctx, "", model.AppTemplateLabelableObject, id, labelKey)
	if err != nil {
		return nil, err
	}
	return label.Value, nil
}

func enrichWithApplicationTypeLabel(applicationInputJSON, applicationType string) (string, error) {
	var appInput map[string]interface{}

	if err := json.Unmarshal([]byte(applicationInputJSON), &appInput); err != nil {
		return "", errors.Wrapf(err, "while unmarshaling application input json")
	}

	labels, ok := appInput["labels"]
	if ok && labels != nil {
		labelsMap, ok := labels.(map[string]interface{})
		if !ok {
			return "", fmt.Errorf("app input json labels are type %T instead of map[string]interface{}. %v", labelsMap, labels)
		}

		if appType, ok := labelsMap[applicationTypeLabelKey]; ok {
			appTypeValue, ok := appType.(string)
			if !ok {
				return "", fmt.Errorf("%q label value must be string", applicationTypeLabelKey)
			}
			if applicationType != otherSystemType && appTypeValue != applicationType {
				return "", fmt.Errorf("%q label value does not match the application template name", applicationTypeLabelKey)
			}
			return applicationInputJSON, nil
		}

		labelsMap[applicationTypeLabelKey] = applicationType
		appInput["labels"] = labelsMap
	} else {
		appInput["labels"] = map[string]interface{}{applicationTypeLabelKey: applicationType}
	}

	inputJSON, err := json.Marshal(appInput)
	if err != nil {
		return "", errors.Wrapf(err, "while marshalling app input")
	}
	return string(inputJSON), nil
}
