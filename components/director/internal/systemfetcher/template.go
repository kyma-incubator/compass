package systemfetcher

import (
	"context"
	"strings"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
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

type renderer struct {
	appTemplateService applicationTemplateService
	appConverter       applicationConverter
}

// NewTemplateRenderer returns a new application input renderer by a given application template.
func NewTemplateRenderer(appTemplateService applicationTemplateService, appConverter applicationConverter) *renderer {
	return &renderer{
		appTemplateService: appTemplateService,
		appConverter:       appConverter,
	}
}

func (r *renderer) ApplicationRegisterInputFromTemplate(ctx context.Context, sc System) (*model.ApplicationRegisterInput, error) {
	appTemplate, err := r.appTemplateService.Get(ctx, sc.TemplateID)
	if err != nil {
		return nil, errors.Wrapf(err, "while getting application template with ID %s", sc.TemplateID)
	}

	inputValues := getTemplateInputs(appTemplate, sc)
	appRegisterInputJSON, err := r.appTemplateService.PrepareApplicationCreateInputJSON(appTemplate, inputValues)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing ApplicationRegisterInput JSON from Application Template with name %s", appTemplate.Name)
	}

	appRegisterInput, err := r.appConverter.CreateInputJSONToModel(ctx, appRegisterInputJSON)
	if err != nil {
		return nil, errors.Wrapf(err, "while preparing ApplicationRegisterInput model from Application Template with name %s", appTemplate.Name)
	}

	return &appRegisterInput, nil
}

func getTemplateInputs(t *model.ApplicationTemplate, s System) model.ApplicationFromTemplateInputValues {
	inputValues := model.ApplicationFromTemplateInputValues{}
	for _, p := range t.Placeholders {
		inputValues = append(inputValues, &model.ApplicationTemplateValueInput{
			Placeholder: p.Name,
			Value:       getPlaceholderInput(p, s),
		})
	}

	return inputValues
}

func getPlaceholderInput(p model.ApplicationTemplatePlaceholder, s System) string {
	if strings.Contains(strings.ToLower(p.Name), "name") {
		return s.DisplayName
	}
	if strings.Contains(strings.ToLower(p.Name), "url") {
		return s.BaseURL
	}
	if strings.Contains(strings.ToLower(p.Name), "description") {
		return s.ProductDescription
	}
	return ""
}
