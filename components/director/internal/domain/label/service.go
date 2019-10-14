package label

import (
	"context"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/pkg/errors"
)

//go:generate mockery -name=LabelRepository -output=automock -outpkg=automock -case=underscore
type LabelRepository interface {
	Upsert(ctx context.Context, label *model.Label) error
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
}

//go:generate mockery -name=LabelDefinitionRepository -output=automock -outpkg=automock -case=underscore
type LabelDefinitionRepository interface {
	Create(ctx context.Context, def model.LabelDefinition) error
	Exists(ctx context.Context, tenant string, key string) (bool, error)
	GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
}

//go:generate mockery -name=UIDService -output=automock -outpkg=automock -case=underscore
type UIDService interface {
	Generate() string
}

type labelUpsertService struct {
	labelRepo           LabelRepository
	labelDefinitionRepo LabelDefinitionRepository
	uidService          UIDService
}

func NewLabelUpsertService(labelRepo LabelRepository, labelDefinitionRepo LabelDefinitionRepository, uidService UIDService) *labelUpsertService {
	return &labelUpsertService{labelRepo: labelRepo, labelDefinitionRepo: labelDefinitionRepo, uidService: uidService}
}

func (s *labelUpsertService) UpsertMultipleLabels(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labels map[string]interface{}) error {
	for key, val := range labels {
		err := s.UpsertLabel(ctx, tenant, &model.LabelInput{
			Key:        key,
			Value:      val,
			ObjectID:   objectID,
			ObjectType: objectType,
		})
		if err != nil {
			return errors.Wrap(err, "while upserting multiple Labels")
		}
	}

	return nil
}

func (s *labelUpsertService) UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error {
	var labelDef *model.LabelDefinition

	labelDef, err := s.labelDefinitionRepo.GetByKey(ctx, tenant, labelInput.Key)
	if err != nil && !apperrors.IsNotFoundError(err) {
		return errors.Wrapf(err, "while reading LabelDefinition for key '%s'", labelInput.Key)
	}

	if labelDef == nil {
		// Create new LabelDefinition
		labelDefinitionID := s.uidService.Generate()
		labelDef = &model.LabelDefinition{
			ID:     labelDefinitionID,
			Tenant: tenant,
			Key:    labelInput.Key,
			Schema: nil,
		}
		err := s.labelDefinitionRepo.Create(ctx, *labelDef)
		if err != nil {
			return errors.Wrapf(err, "while creating  a new LabelDefinition for Label '%s'", labelInput.Key)
		}
	}

	err = s.validateLabelInputValue(ctx, tenant, labelInput, labelDef)
	if err != nil {
		return errors.Wrapf(err, "while validating Label value for '%s'", labelInput.Key)
	}

	label := labelInput.ToLabel(s.uidService.Generate(), tenant)
	err = s.labelRepo.Upsert(ctx, label)
	if err != nil {
		return errors.Wrapf(err, "while creating label '%s' for %s '%s'", labelInput.Key, labelInput.ObjectType, labelInput.ObjectID)
	}

	return nil
}

func (s *labelUpsertService) validateLabelInputValue(ctx context.Context, tenant string, labelInput *model.LabelInput, labelDef *model.LabelDefinition) error {
	if labelDef == nil || labelDef.Schema == nil {
		// nothing to validate
		return nil
	}

	validator, err := jsonschema.NewValidatorFromRawSchema(*labelDef.Schema)
	if err != nil {
		return errors.Wrapf(err, "while creating JSON Schema validator for schema %+v", *labelDef.Schema)
	}

	result, err := validator.ValidateRaw(labelInput.Value)
	if err != nil {
		return errors.Wrapf(err, "while validating value %+v against JSON Schema: %+v", labelInput.Value, *labelDef.Schema)
	}
	if !result.Valid {
		return errors.Wrapf(result.Error, "while validating value %+v against JSON Schema: %+v", labelInput.Value, *labelDef.Schema)
	}

	return nil
}
