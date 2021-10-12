package systemfetcher

import (
	"context"
	"encoding/json"
	"fmt"
	"reflect"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/tidwall/gjson"
)

//go:generate mockery --name=applicationTemplateService --output=automock --outpkg=automock --case=underscore --exported=true
type applicationTemplateService interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, values model.ApplicationFromTemplateInputValues) (string, error)
}

//go:generate mockery --name=applicationConverter --output=automock --outpkg=automock --case=underscore --exported=true
type applicationConverter interface {
	CreateInputJSONToModel(ctx context.Context, in string) (model.ApplicationRegisterInput, error)
}

// PlaceholderMapping is the mapping we have between a placeholder key we use in templates,
// and input field from the external system provider.
type PlaceholderMapping struct {
	PlaceholderName string `json:"placeholder_name"`
	SystemKey       string `json:"system_key"`
}

type renderer struct {
	appTemplateService applicationTemplateService
	appConverter       applicationConverter

	appInputOverride     string
	placeholdersMapping  []PlaceholderMapping
	placeholdersOverride []model.ApplicationTemplatePlaceholder
}

// NewTemplateRenderer returns a new application input renderer by a given application template.
func NewTemplateRenderer(appTemplateService applicationTemplateService, appConverter applicationConverter, appInputOverride string, mapping []PlaceholderMapping) *renderer {
	placeholders := make([]model.ApplicationTemplatePlaceholder, 0)
	for _, p := range mapping {
		placeholders = append(placeholders, model.ApplicationTemplatePlaceholder{
			Name: p.PlaceholderName,
		})
	}

	return &renderer{
		appTemplateService:   appTemplateService,
		appConverter:         appConverter,
		appInputOverride:     appInputOverride,
		placeholdersMapping:  mapping,
		placeholdersOverride: placeholders,
	}
}

func (r *renderer) ApplicationRegisterInputFromTemplate(ctx context.Context, sc System) (*model.ApplicationRegisterInput, error) {
	appTemplate, err := r.appTemplateService.Get(ctx, sc.TemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting application template with ID %s", sc.TemplateID)
	}

	inputValues, err := r.getTemplateInputs(sc)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting template inputs for Application Template with name %s", appTemplate.Name)
	}

	appTemplate.Placeholders = r.placeholdersOverride
	appTemplate.ApplicationInputJSON, err = r.mergedApplicationInput(appTemplate.ApplicationInputJSON, r.appInputOverride)
	if err != nil {
		return nil, errors.Wrap(err, "while merging application input from template and override application input")
	}
	appRegisterInputJSON, err := r.appTemplateService.PrepareApplicationCreateInputJSON(appTemplate, *inputValues)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing ApplicationRegisterInput JSON from Application Template with name %s", appTemplate.Name)
	}

	appRegisterInput, err := r.appConverter.CreateInputJSONToModel(ctx, appRegisterInputJSON)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing ApplicationRegisterInput model from Application Template with name %s", appTemplate.Name)
	}

	return &appRegisterInput, nil
}

func (r *renderer) getTemplateInputs(s System) (*model.ApplicationFromTemplateInputValues, error) {
	systemJSON, err := json.Marshal(s)
	if err != nil {
		return nil, err
	}

	inputValues := model.ApplicationFromTemplateInputValues{}
	for _, pm := range r.placeholdersMapping {
		placeholderInput := gjson.GetBytes(systemJSON, pm.SystemKey).String()
		inputValues = append(inputValues, &model.ApplicationTemplateValueInput{
			Placeholder: pm.PlaceholderName,
			Value:       placeholderInput,
		})
	}

	return &inputValues, nil
}

func (r *renderer) mergedApplicationInput(originalAppInputJSON, overrideAppInputJSON string) (string, error) {
	var originalAppInput map[string]interface{}
	var overrideAppInput map[string]interface{}

	if err := json.Unmarshal([]byte(originalAppInputJSON), &originalAppInput); err != nil {
		return "", errors.Wrapf(err, "while unmarshaling original application input")
	}

	if err := json.Unmarshal([]byte(overrideAppInputJSON), &overrideAppInput); err != nil {
		return "", errors.Wrapf(err, "while unmarshaling override application input")
	}

	for field, overrideV := range overrideAppInput {
		if originalV, ok := originalAppInput[field]; ok && originalV != nil {
			if reflect.TypeOf(originalV) != reflect.TypeOf(overrideV) {
				return "", fmt.Errorf("values %v and %v of key %s have different types - %T and %T", originalV, overrideV, field, originalV, overrideV)
			}

			if field == "labels" {
				newLabels, err := enrichLabels(originalV, overrideV)
				if err != nil {
					return "", err
				}

				originalAppInput[field] = newLabels
				continue
			}
		}

		originalAppInput[field] = overrideV
	}

	merged, err := json.Marshal(originalAppInput)
	if err != nil {
		return "", errors.Wrapf(err, "while marshalling merged app input")
	}
	return string(merged), nil
}

func enrichLabels(originalLabels, overrideLabels interface{}) (map[string]interface{}, error) {
	templateLabelsMap, ok := originalLabels.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("app template labels are with type %T instead of map[string]interface{}", originalLabels)
	}
	overrideLabelsMap, ok := overrideLabels.(map[string]interface{})
	if !ok {
		return nil, fmt.Errorf("overrides labels are with type %T instead of map[string]interface{}", overrideLabels)
	}

	for labelKey, labelValue := range overrideLabelsMap {
		templateLabelsMap[labelKey] = labelValue
	}
	return templateLabelsMap, nil
}
