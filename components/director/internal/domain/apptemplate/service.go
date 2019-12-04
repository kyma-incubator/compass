package apptemplate

import (
	"context"
	"fmt"
	"strings"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
)

//go:generate mockery -name=ApplicationTemplateRepository -output=automock -outpkg=automock -case=underscore
type ApplicationTemplateRepository interface {
	Create(ctx context.Context, item model.ApplicationTemplate) error
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	GetByName(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	Exists(ctx context.Context, id string) (bool, error)
	List(ctx context.Context, pageSize int, cursor string) (model.ApplicationTemplatePage, error)
	Update(ctx context.Context, model model.ApplicationTemplate) error
	Delete(ctx context.Context, id string) error
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type service struct {
	appTemplateRepo ApplicationTemplateRepository

	uidService UIDService
}

func NewService(appTemplateRepo ApplicationTemplateRepository, uidService UIDService) *service {
	return &service{
		appTemplateRepo: appTemplateRepo,
		uidService:      uidService,
	}
}

func (s *service) Create(ctx context.Context, in model.ApplicationTemplateInput) (string, error) {
	id := s.uidService.Generate()
	appTemplate := in.ToApplicationTemplate(id)

	err := s.appTemplateRepo.Create(ctx, appTemplate)
	if err != nil {
		return "", errors.Wrap(err, "while creating Application Template")
	}

	return id, nil
}

func (s *service) Get(ctx context.Context, id string) (*model.ApplicationTemplate, error) {
	appTemplate, err := s.appTemplateRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application Template with ID %s", id)
	}

	return appTemplate, nil
}

func (s *service) GetByName(ctx context.Context, name string) (*model.ApplicationTemplate, error) {
	appTemplate, err := s.appTemplateRepo.GetByName(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application Template with [name=%s]", name)
	}

	return appTemplate, nil
}

func (s *service) Exists(ctx context.Context, id string) (bool, error) {
	exist, err := s.appTemplateRepo.Exists(ctx, id)
	if err != nil {
		return false, errors.Wrapf(err, "while getting Application Template with ID %s", id)
	}

	return exist, nil
}

func (s *service) List(ctx context.Context, pageSize int, cursor string) (model.ApplicationTemplatePage, error) {
	if pageSize < 1 || pageSize > 100 {
		return model.ApplicationTemplatePage{}, errors.New("page size must be between 1 and 100")
	}

	return s.appTemplateRepo.List(ctx, pageSize, cursor)
}

func (s *service) Update(ctx context.Context, id string, in model.ApplicationTemplateInput) error {
	appTemplate := in.ToApplicationTemplate(id)

	err := s.appTemplateRepo.Update(ctx, appTemplate)
	if err != nil {
		return errors.Wrapf(err, "while updating Application Template with ID %s", id)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	err := s.appTemplateRepo.Delete(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Application Template with ID %s", id)
	}

	return nil
}

func (s *service) PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, templatePlaceholderValues []*model.ApplicationTemplateValueInput) (string, error) {
	appCreateInputJSON := appTemplate.ApplicationInputJSON
	for _, placeholder := range appTemplate.Placeholders {
		newValue, err := s.lookForPlaceholderValue(templatePlaceholderValues, placeholder.Name)
		if err != nil {
			return "", err
		}
		appCreateInputJSON = strings.ReplaceAll(appCreateInputJSON, fmt.Sprintf("{{%s}}", placeholder.Name), newValue)
	}
	return appCreateInputJSON, nil
}

func (s *service) lookForPlaceholderValue(values []*model.ApplicationTemplateValueInput, placeholderName string) (string, error) {
	for _, value := range values {
		if value.Placeholder == placeholderName {
			return value.Value, nil
		}
	}
	return "", errors.Errorf("required placeholder [name=%s] value not provided", placeholderName)
}
