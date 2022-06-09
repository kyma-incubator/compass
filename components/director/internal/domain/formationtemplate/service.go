package formationtemplate

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/pkg/errors"
)

// FormationTemplateRepository missing godoc
//go:generate mockery --name=FormationTemplateRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type FormationTemplateRepository interface {
	Create(ctx context.Context, item *model.FormationTemplate) error
	Get(ctx context.Context, id string) (*model.FormationTemplate, error)
	List(ctx context.Context, pageSize int, cursor string) (*model.FormationTemplatePage, error)
	Update(ctx context.Context, model *model.FormationTemplate) error
	Delete(ctx context.Context, id string) error
	Exists(ctx context.Context, id string) (bool, error)
}

// UIDService missing godoc
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type service struct {
	repo      FormationTemplateRepository
	uidSvc    UIDService
	converter FormationTemplateConverter
}

// NewService missing godoc
func NewService(repo FormationTemplateRepository, uidSvc UIDService, converter FormationTemplateConverter) *service {
	return &service{
		repo:      repo,
		uidSvc:    uidSvc,
		converter: converter,
	}
}

func (s *service) Create(ctx context.Context, in *model.FormationTemplateInput) (string, error) {
	formationTemplateID := s.uidSvc.Generate()

	log.C(ctx).Debugf("ID %s generated for Formation Template with name %s", formationTemplateID, in.Name)

	err := s.repo.Create(ctx, s.converter.FromModelInputToModel(in, formationTemplateID))
	if err != nil {
		return "", errors.Wrapf(err, "while creating Formation Template with name %s", in.Name)
	}

	return formationTemplateID, nil
}

func (s *service) Get(ctx context.Context, id string) (*model.FormationTemplate, error) {
	formationTemplate, err := s.repo.Get(ctx, id)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Formation Template with id %s", id)
	}

	return formationTemplate, nil
}

func (s *service) List(ctx context.Context, pageSize int, cursor string) (*model.FormationTemplatePage, error) {
	if pageSize < 1 || pageSize > 200 {
		return nil, apperrors.NewInvalidDataError("page size must be between 1 and 200")
	}

	return s.repo.List(ctx, pageSize, cursor)
}

func (s *service) Update(ctx context.Context, id string, in *model.FormationTemplateInput) error {
	exists, err := s.repo.Exists(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while ensuring Formation Template with ID %s exists", id)
	} else if !exists {
		return apperrors.NewNotFoundError(resource.FormationTemplate, id)
	}
	err = s.repo.Update(ctx, s.converter.FromModelInputToModel(in, id))
	if err != nil {
		return errors.Wrapf(err, "while updating Formation Template with ID %s", id)
	}

	return nil
}

func (s *service) Delete(ctx context.Context, id string) error {
	err := s.repo.Delete(ctx, id)
	if err != nil {
		return errors.Wrapf(err, "while deleting Formation Template with ID %s", id)
	}

	return nil
}
