package labeldef

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

type scenariosService struct {
	repo       Repository
	uidService UIDService
}

func NewScenariosService(r Repository, uidService UIDService) *scenariosService {
	return &scenariosService{
		repo:       r,
		uidService: uidService,
	}
}

func (s *scenariosService) EnsureScenariosLabelDefinitionExists(ctx context.Context, tenant string) error {
	ldExists, err := s.repo.Exists(ctx, tenant, model.ScenariosKey)
	if err != nil {
		return errors.Wrapf(err, "while checking if Label Definition with key %s exists", model.ScenariosKey)
	}
	if !ldExists {
		var scenariosSchema interface{} = model.ScenariosSchema
		scenariosLD := model.LabelDefinition{
			ID:     s.uidService.Generate(),
			Tenant: tenant,
			Key:    model.ScenariosKey,
			Schema: &scenariosSchema,
		}
		err = s.repo.Create(ctx, scenariosLD)
		if err != nil {
			return errors.Wrapf(err, "while creating Label Definition with key %s", model.ScenariosKey)
		}
	}
	return nil
}
