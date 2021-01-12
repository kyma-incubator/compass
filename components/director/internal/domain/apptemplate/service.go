package apptemplate

import (
	"context"
	"fmt"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

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
	log.C(ctx).Debugf("ID %s generated for Application Template with name %s", id, in.Name)

	appTemplate := in.ToApplicationTemplate(id)

	err := s.appTemplateRepo.Create(ctx, appTemplate)
	if err != nil {
		return "", errors.Wrapf(err, "while creating Application Template with name %s", in.Name)
	}

	return id, nil
}

func (s *service) Get(ctx context.Context, id string) (*model.ApplicationTemplate, error) {
	appTemplate, err := s.appTemplateRepo.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application Template with id %s", id)
	}

	return appTemplate, nil
}

func (s *service) GetByName(ctx context.Context, name string) (*model.ApplicationTemplate, error) {
	appTemplate, err := s.appTemplateRepo.GetByName(ctx, name)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Application Template with name %s", name)
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
	if pageSize < 1 || pageSize > 200 {
		return model.ApplicationTemplatePage{}, apperrors.NewInvalidDataError("page size must be between 1 and 200")
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
