package datainputbuilder

import (
	"context"
	"encoding/json"
	"strconv"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery --exported --name=labelRepository --output=automock --outpkg=automock --case=underscore --disable-version-string
type labelRepository interface {
	ListForObject(ctx context.Context, tenant string, objectType model.LabelableObject, objectID string) (map[string]*model.Label, error)
	ListForObjectIDs(ctx context.Context, tenant string, objectType model.LabelableObject, objectIDs []string) (map[string]map[string]interface{}, error)
}

// WebhookLabelBuilder takes care to get and build labels for objects
type WebhookLabelBuilder struct {
	labelRepository labelRepository
}

// NewWebhookLabelBuilder creates a WebhookLabelBuilder
func NewWebhookLabelBuilder(labelRepository labelRepository) *WebhookLabelBuilder {
	return &WebhookLabelBuilder{
		labelRepository: labelRepository,
	}
}

// GetLabelsForObject builds labels for object with ID objectID
func (b *WebhookLabelBuilder) GetLabelsForObject(ctx context.Context, tenant, objectID string, objectType model.LabelableObject) (map[string]string, error) {
	labels, err := b.labelRepository.ListForObject(ctx, tenant, objectType, objectID)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing labels for %q with ID: %q", objectType, objectID)
	}
	labelsMap := make(map[string]string, len(labels))
	for _, l := range labels {
		labelBytes, err := json.Marshal(l.Value)
		if err != nil {
			return nil, errors.Wrap(err, "while unmarshaling label value")
		}

		stringLabel := string(labelBytes)
		unquotedLabel, err := strconv.Unquote(stringLabel)
		if err != nil {
			labelsMap[l.Key] = stringLabel
		} else {
			labelsMap[l.Key] = unquotedLabel
		}
	}
	return labelsMap, nil
}

// GetLabelsForObjects builds labels for objects with IDs objectIDs
func (b *WebhookLabelBuilder) GetLabelsForObjects(ctx context.Context, tenant string, objectIDs []string, objectType model.LabelableObject) (map[string]map[string]string, error) {
	labelsForResources, err := b.labelRepository.ListForObjectIDs(ctx, tenant, objectType, objectIDs)
	if err != nil {
		return nil, errors.Wrapf(err, "while listing labels for %q with IDs: %q", objectType, objectIDs)
	}
	labelsForResourcesMap := make(map[string]map[string]string, len(labelsForResources))
	for resourceID, labels := range labelsForResources {
		for key, value := range labels {
			labelBytes, err := json.Marshal(value)
			if err != nil {
				return nil, errors.Wrap(err, "while marshaling label value")
			}

			if _, ok := labelsForResourcesMap[resourceID]; !ok {
				labelsForResourcesMap[resourceID] = make(map[string]string)
			}

			stringLabel := string(labelBytes)
			unquotedLabel, err := strconv.Unquote(stringLabel)
			if err != nil {
				labelsForResourcesMap[resourceID][key] = stringLabel
			} else {
				labelsForResourcesMap[resourceID][key] = unquotedLabel
			}
		}
	}
	return labelsForResourcesMap, nil
}
