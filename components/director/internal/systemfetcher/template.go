package systemfetcher

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"

	"k8s.io/client-go/util/jsonpath"

	"github.com/imdario/mergo"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

//go:generate mockery --name=applicationTemplateService --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type applicationTemplateService interface {
	Get(ctx context.Context, id string) (*model.ApplicationTemplate, error)
	PrepareApplicationCreateInputJSON(appTemplate *model.ApplicationTemplate, values model.ApplicationFromTemplateInputValues) (string, error)
}

//go:generate mockery --name=applicationConverter --output=automock --outpkg=automock --case=underscore --exported=true --disable-version-string
type applicationConverter interface {
	CreateInputJSONToModel(ctx context.Context, in string) (model.ApplicationRegisterInput, error)
}

// PlaceholderMapping is the mapping we have between a placeholder key we use in templates,
// and input field from the external system provider.
type PlaceholderMapping struct {
	PlaceholderName string `json:"placeholder_name"`
	SystemKey       string `json:"system_key"`
	Optional        bool   `json:"optional"`
}

type renderer struct {
	appTemplateService applicationTemplateService
	appConverter       applicationConverter

	appInputOverride     string
	placeholdersMapping  []PlaceholderMapping
	placeholdersOverride []model.ApplicationTemplatePlaceholder
}

// NewTemplateRenderer returns a new application input renderer by a given application template.
func NewTemplateRenderer(appTemplateService applicationTemplateService, appConverter applicationConverter, appInputOverride string, mapping []PlaceholderMapping) (*renderer, error) {
	if _, err := appConverter.CreateInputJSONToModel(context.Background(), appInputOverride); err != nil {
		return nil, errors.Wrapf(err, "while converting override application input JSON into application input")
	}
	placeholders := make([]model.ApplicationTemplatePlaceholder, 0)
	for i := 0; i < len(mapping); i++ {
		placeholders = append(placeholders, model.ApplicationTemplatePlaceholder{
			Name:     mapping[i].PlaceholderName,
			JSONPath: &mapping[i].SystemKey,
			Optional: &mapping[i].Optional,
		})
	}

	return &renderer{
		appTemplateService:   appTemplateService,
		appConverter:         appConverter,
		appInputOverride:     appInputOverride,
		placeholdersMapping:  mapping,
		placeholdersOverride: placeholders,
	}, nil
}

func (r *renderer) ApplicationRegisterInputFromTemplate(ctx context.Context, sc System) (*model.ApplicationRegisterInput, error) {
	appTemplate, err := r.appTemplateService.Get(ctx, sc.TemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting application template with ID %s", sc.TemplateID)
	}

	inputValues, err := r.getTemplateInputs(sc.SystemPayload, appTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting template inputs for Application Template with name %s", appTemplate.Name)
	}
	placeholdersOverride, err := r.extendPlaceholdersOverride(sc.SystemPayload, appTemplate)
	if err != nil {
		return nil, errors.Wrapf(err, "while extending placeholders override for Application Template with name %s", appTemplate.Name)
	}

	appTemplate.Placeholders = placeholdersOverride
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

func (r *renderer) getTemplateInputs(systemPayload map[string]interface{}, appTemplate *model.ApplicationTemplate) (*model.ApplicationFromTemplateInputValues, error) {
	parser := jsonpath.New("parser")

	placeholdersMappingInputValues := model.ApplicationFromTemplateInputValues{}
	for _, pm := range r.placeholdersMapping {
		if err := parser.Parse(fmt.Sprintf("{%s}", pm.SystemKey)); err != nil {
			return nil, errors.Wrapf(err, "while parsing placeholder mapping with name %s and system key: %s", pm.PlaceholderName, pm.SystemKey)
		}

		placeholderInput := new(bytes.Buffer)
		if err := parser.Execute(placeholderInput, systemPayload); err != nil && !pm.Optional {
			return nil, errors.Wrapf(err, "missing or empty key %q in system payload.", pm.SystemKey)
		}

		placeholdersMappingInputValues = append(placeholdersMappingInputValues, &model.ApplicationTemplateValueInput{
			Placeholder: pm.PlaceholderName,
			Value:       placeholderInput.String(),
		})
	}
	var appTemplatePlaceholdersInputValues model.ApplicationFromTemplateInputValues
	for _, placeholder := range appTemplate.Placeholders {
		if placeholder.JSONPath != nil && len(*placeholder.JSONPath) > 0 {
			if err := parser.Parse(fmt.Sprintf("{%s}", *placeholder.JSONPath)); err != nil {
				return nil, errors.Wrapf(err, "while parsing placeholder jsonPath with name: %s and path: %s for ap template with id: %s", placeholder.Name, *placeholder.JSONPath, appTemplate.ID)
			}

			placeholderInput := new(bytes.Buffer)
			if err := parser.Execute(placeholderInput, systemPayload); err != nil {
				return nil, errors.Wrapf(err, "placeholder value with name: %s and path: %s for app template with id: %s not found in system payload", placeholder.Name, *placeholder.JSONPath, appTemplate.ID)
			}

			appTemplatePlaceholdersInputValues = append(appTemplatePlaceholdersInputValues, &model.ApplicationTemplateValueInput{
				Placeholder: placeholder.Name,
				Value:       placeholderInput.String(),
			})
		}
	}
	inputValues := mergeInputValues(appTemplatePlaceholdersInputValues, placeholdersMappingInputValues)
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

	if err := mergo.Merge(&originalAppInput, overrideAppInput); err != nil {
		return "", errors.Wrapf(err, "while merging original app input: %v into destination app input: %v", originalAppInput, overrideAppInputJSON)
	}
	merged, err := json.Marshal(originalAppInput)
	if err != nil {
		return "", errors.Wrapf(err, "while marshalling merged app input")
	}
	return string(merged), nil
}

func (r *renderer) extendPlaceholdersOverride(systemPayload map[string]interface{}, appTemplate *model.ApplicationTemplate) ([]model.ApplicationTemplatePlaceholder, error) {
	parser := jsonpath.New("parser")

	var appTemplatePlaceholdersOverride []model.ApplicationTemplatePlaceholder
	for _, placeholder := range appTemplate.Placeholders {
		if placeholder.JSONPath != nil && len(*placeholder.JSONPath) > 0 {
			if err := parser.Parse(fmt.Sprintf("{%s}", *placeholder.JSONPath)); err != nil {
				return nil, errors.Wrapf(err, "while parsing placeholder jsonPath with name: %s and path: %s for ap template with id: %s", placeholder.Name, *placeholder.JSONPath, appTemplate.ID)
			}

			placeholderInput := new(bytes.Buffer)
			if err := parser.Execute(placeholderInput, systemPayload); err != nil {
				return nil, errors.Wrapf(err, "placeholder value with name: %s and path: %s for app template with id: %s not found in system payload", placeholder.Name, *placeholder.JSONPath, appTemplate.ID)
			}

			appTemplatePlaceholdersOverride = append(appTemplatePlaceholdersOverride, model.ApplicationTemplatePlaceholder{
				Name:     placeholder.Name,
				JSONPath: placeholder.JSONPath,
				Optional: placeholder.Optional,
			})
		}
	}
	placeholdersOverride := mergePlaceholders(appTemplatePlaceholdersOverride, r.placeholdersOverride)

	return placeholdersOverride, nil
}

func mergePlaceholders(appTemplatePlaceholdersOverride []model.ApplicationTemplatePlaceholder, placeholdersOverride []model.ApplicationTemplatePlaceholder) []model.ApplicationTemplatePlaceholder {
	placeholdersMap := make(map[string]model.ApplicationTemplatePlaceholder)
	for _, placeholder := range appTemplatePlaceholdersOverride {
		placeholdersMap[placeholder.Name] = placeholder
	}
	for _, placeholder := range placeholdersOverride {
		if _, exists := placeholdersMap[placeholder.Name]; !exists {
			placeholdersMap[placeholder.Name] = placeholder
		}
	}
	mergedPlaceholders := make([]model.ApplicationTemplatePlaceholder, 0)
	for _, p := range placeholdersMap {
		mergedPlaceholders = append(mergedPlaceholders, p)
	}
	return mergedPlaceholders
}

func mergeInputValues(appTemplatePlaceholdersInputValues model.ApplicationFromTemplateInputValues, placeholdersMappingInputValues model.ApplicationFromTemplateInputValues) model.ApplicationFromTemplateInputValues {
	inputsMap := make(map[string]*model.ApplicationTemplateValueInput)
	for _, input := range appTemplatePlaceholdersInputValues {
		inputsMap[input.Placeholder] = input
	}
	for _, input := range placeholdersMappingInputValues {
		if _, exists := inputsMap[input.Placeholder]; !exists {
			inputsMap[input.Placeholder] = input
		}
	}
	mergedInputs := make(model.ApplicationFromTemplateInputValues, 0)
	for _, input := range inputsMap {
		mergedInputs = append(mergedInputs, input)
	}
	return mergedInputs
}
