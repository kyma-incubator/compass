package label

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/jsonschema"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// LabelRepository missing godoc
//
//go:generate mockery --name=LabelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelRepository interface {
	Create(ctx context.Context, tenant string, label *model.Label) error
	Upsert(ctx context.Context, tenant string, label *model.Label) error
	UpsertGlobal(ctx context.Context, label *model.Label) error
	UpdateWithVersion(ctx context.Context, tenant string, label *model.Label) error
	GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	GetByKeyGlobal(ctx context.Context, objectType model.LabelableObject, objectID, key string) (*model.Label, error)
	Delete(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID string, key string) error
	ListForObject(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
}

// LabelDefinitionRepository missing godoc
//
//go:generate mockery --name=LabelDefinitionRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type LabelDefinitionRepository interface {
	Create(ctx context.Context, def model.LabelDefinition) error
	Exists(ctx context.Context, tenant string, key string) (bool, error)
	GetByKey(ctx context.Context, tenant string, key string) (*model.LabelDefinition, error)
}

// UIDService missing godoc
//
//go:generate mockery --name=UIDService --output=automock --outpkg=automock --case=underscore --disable-version-string
type UIDService interface {
	Generate() string
}

type labelService struct {
	labelRepo           LabelRepository
	labelDefinitionRepo LabelDefinitionRepository
	uidService          UIDService
}

// NewLabelService missing godoc
func NewLabelService(labelRepo LabelRepository, labelDefinitionRepo LabelDefinitionRepository, uidService UIDService) *labelService {
	return &labelService{labelRepo: labelRepo, labelDefinitionRepo: labelDefinitionRepo, uidService: uidService}
}

// UpsertMultipleLabels upserts multiple labels for a given tenant and object
func (s *labelService) UpsertMultipleLabels(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string, labels map[string]interface{}) error {
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

// CreateLabel missing godoc
func (s *labelService) CreateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error {
	if err := s.validateLabel(ctx, tenant, labelInput); err != nil {
		return err
	}
	label := labelInput.ToLabel(id, tenant)

	if err := s.labelRepo.Create(ctx, tenant, label); err != nil {
		return errors.Wrapf(err, "while creating label with ID: %s for %s with ID: %s", label.ID, label.ObjectType, label.ObjectID)
	}
	log.C(ctx).Debugf("Successfully created label with ID: %s for %s with ID: %s", label.ID, label.ObjectType, label.ObjectID)

	return nil
}

// UpsertLabel missing godoc
func (s *labelService) UpsertLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error {
	if err := s.validateLabel(ctx, tenant, labelInput); err != nil {
		return err
	}

	label := labelInput.ToLabel(s.uidService.Generate(), tenant)

	if err := s.labelRepo.Upsert(ctx, tenant, label); err != nil {
		return errors.Wrapf(err, "while upserting label with ID: %s for %s with ID: %s", label.ID, label.ObjectType, label.ObjectID)
	}
	log.C(ctx).Debugf("Successfully upserted label with ID: %s for %s with ID: %s", label.ID, label.ObjectType, label.ObjectID)

	return nil
}

// UpsertLabelGlobal missing godoc
func (s *labelService) UpsertLabelGlobal(ctx context.Context, labelInput *model.LabelInput) error {
	if err := s.validateLabel(ctx, "", labelInput); err != nil {
		return err
	}

	label := labelInput.ToLabel(s.uidService.Generate(), "")

	if err := s.labelRepo.UpsertGlobal(ctx, label); err != nil {
		return errors.Wrapf(err, "while upserting label with ID: %q for %q with ID: %q", label.ID, label.ObjectType, label.ObjectID)
	}
	log.C(ctx).Debugf("Successfully upserted label with ID: %q for %q with ID: %q", label.ID, label.ObjectType, label.ObjectID)

	return nil
}

// UpdateLabel missing godoc
func (s *labelService) UpdateLabel(ctx context.Context, tenant, id string, labelInput *model.LabelInput) error {
	if err := s.validateLabel(ctx, tenant, labelInput); err != nil {
		return err
	}
	label := labelInput.ToLabel(id, tenant)

	if err := s.labelRepo.UpdateWithVersion(ctx, tenant, label); err != nil {
		return errors.Wrapf(err, "while updating label with ID: %s for %s with ID: %s", label.ID, label.ObjectType, label.ObjectID)
	}
	log.C(ctx).Debugf("Successfully updated label with ID: %s for %s with ID: %s", label.ID, label.ObjectType, label.ObjectID)

	return nil
}

// GetLabel missing godoc
func (s *labelService) GetLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) (*model.Label, error) {
	label, err := s.labelRepo.GetByKey(ctx, tenant, labelInput.ObjectType, labelInput.ObjectID, labelInput.Key)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting Label with key %s for %s with ID: %s", labelInput.Key, labelInput.ObjectType, labelInput.ObjectID)
	}
	return label, nil
}

// GetByKey returns label for a given tenant, object and key
func (s *labelService) GetByKey(ctx context.Context, tenant string, objectType model.LabelableObject, objectID, key string) (*model.Label, error) {
	return s.labelRepo.GetByKey(ctx, tenant, objectType, objectID, key)
}

// Delete deletes a label with given key, objectID and objectType
func (s *labelService) Delete(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID, key string) error {
	if err := s.labelRepo.Delete(ctx, tenantID, objectType, objectID, key); err != nil {
		return errors.Wrapf(err, "while deleting label with key: %s for %s with ID: %s", key, objectType, objectID)
	}
	return nil
}

// ListForObject retrieves all labels for the provided parameters.
// The returned map contains the label key as the map's key and for the value - the label itself
func (s *labelService) ListForObject(ctx context.Context, tenantID string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error) {
	labels, err := s.labelRepo.ListForObject(ctx, tenantID, objectType, objectID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting labels for %s with ID: %s", objectType, objectID)
	}
	return labels, nil
}

func (s *labelService) validateLabel(ctx context.Context, tenant string, labelInput *model.LabelInput) error {
	// we should validate only scenario labels
	if labelInput.Key != model.ScenariosKey {
		return nil
	}

	labelDef, err := s.labelDefinitionRepo.GetByKey(ctx, tenant, labelInput.Key)
	if err != nil {
		return errors.Wrapf(err, "while getting LabelDefinition for key '%s'", labelInput.Key)
	}

	if err := s.validateLabelInputValue(labelInput, labelDef); err != nil {
		return errors.Wrapf(err, "while validating Label value for '%s'", labelInput.Key)
	}
	return nil
}

func (s *labelService) validateLabelInputValue(labelInput *model.LabelInput, labelDef *model.LabelDefinition) error {
	if labelDef == nil || labelDef.Schema == nil {
		// nothing to validate
		return nil
	}

	validator, err := jsonschema.NewValidatorFromRawSchema(*labelDef.Schema)
	if err != nil {
		return errors.Wrapf(err, "while creating JSON Schema validator for schema %+v", *labelDef.Schema)
	}

	jsonSchema, err := json.Marshal(*labelDef.Schema)
	if err != nil {
		return apperrors.InternalErrorFrom(err, "while marshalling json schema")
	}

	result, err := validator.ValidateRaw(labelInput.Value)
	if err != nil {
		return apperrors.InternalErrorFrom(err, "while validating value=%+v against JSON Schema=%s", labelInput.Value, string(jsonSchema))
	}
	if !result.Valid {
		return apperrors.NewInvalidDataError(fmt.Sprintf("input value=%+v, key=%s, is not valid against JSON Schema=%s,result=%s", labelInput.Value, labelInput.Key, jsonSchema, result.Error.Error()))
	}

	return nil
}
