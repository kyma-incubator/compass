package labeldef

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

type scenariosService struct {
	repo       Repository
	uidService UIDService

	defaultScenarioEnabled bool
}

func NewScenariosService(r Repository, uidService UIDService, defaultScenarioEnabled bool) *scenariosService {
	return &scenariosService{
		repo:                   r,
		uidService:             uidService,
		defaultScenarioEnabled: defaultScenarioEnabled,
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

func (s *scenariosService) GetAvailableScenarios(ctx context.Context, tenantID string) ([]string, error) {
	def, err := s.repo.GetByKey(ctx, tenantID, model.ScenariosKey)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting `%s` label definition", model.ScenariosKey)
	}
	if def.Schema == nil {
		return nil, fmt.Errorf("missing schema for `%s` label definition", model.ScenariosKey)
	}

	b, err := json.Marshal(*def.Schema)
	if err != nil {
		return nil, errors.Wrapf(err, "while marshaling schema")
	}
	sd := ScenariosDefinition{}
	if err = json.Unmarshal(b, &sd); err != nil {
		return nil, errors.Wrapf(err, "while unmarshaling schema to %T", sd)
	}
	return sd.Items.Enum, nil
}

func (s *scenariosService) AddDefaultScenarioIfEnabled(ctx context.Context, labels *map[string]interface{}) {
	if labels == nil || !s.defaultScenarioEnabled {
		return
	}

	if _, ok := (*labels)[model.ScenariosKey]; !ok {
		if *labels == nil {
			*labels = map[string]interface{}{}
		}
		(*labels)[model.ScenariosKey] = model.ScenariosDefaultValue
		log.C(ctx).Debug("Successfully added Default scenario")
	}
}

type ScenariosDefinition struct {
	Items struct {
		Enum []string
	}
}
