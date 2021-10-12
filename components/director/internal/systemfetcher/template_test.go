package systemfetcher_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/require"
)

func TestApplicationRegisterInputFromTemplate(t *testing.T) {
	const (
		appTemplateID        = "appTmp1"
		appRegisterInputJSON = `{"description":"{{description}}","name":"test1"}`
		appInputOverride     = `{ "description": "{{description}}"}`
	)

	placeholdersMappings := []systemfetcher.PlaceholderMapping{
		{
			PlaceholderName: "description",
			SystemKey:       "productDescription",
		},
	}

	appTemplate := &model.ApplicationTemplate{
		ID:                   appTemplateID,
		ApplicationInputJSON: `{ "name": "test1"}`,
	}

	appTemplateWithOverrides := &model.ApplicationTemplate{
		ID: appTemplateID,
		Placeholders: []model.ApplicationTemplatePlaceholder{
			{Name: "description"},
		},
		ApplicationInputJSON: appRegisterInputJSON,
	}

	testSystem := systemfetcher.System{
		SystemBase: systemfetcher.SystemBase{
			SystemNumber:           "123",
			DisplayName:            "test",
			ProductDescription:     "test",
			BaseURL:                "http://test",
			InfrastructureProvider: "test",
		},
		TemplateID: "123",
	}

	appTemplateSvcNoErrors := func(testSystem systemfetcher.System, _ error) *automock.ApplicationTemplateService {
		svc := &automock.ApplicationTemplateService{}
		inputValues := fixInputValuesForSystem(testSystem)
		svc.On("Get", context.TODO(), testSystem.TemplateID).Return(appTemplate, nil).Once()
		svc.On("PrepareApplicationCreateInputJSON", appTemplateWithOverrides, inputValues).Return(appRegisterInputJSON, nil).Once()
		return svc
	}
	appConvSvcNoErrors := func(testSystem systemfetcher.System, _ error) *automock.ApplicationConverter {
		appInput := fixAppInputBySystem(testSystem)
		conv := &automock.ApplicationConverter{}
		conv.On("CreateInputJSONToModel", context.TODO(), appRegisterInputJSON).Return(appInput, nil).Once()
		return conv
	}

	type testCase struct {
		name                string
		system              systemfetcher.System
		expectedErr         error
		appInputOverride    string
		placeholderMappings []systemfetcher.PlaceholderMapping
		setupAppTemplateSvc func(testSystem systemfetcher.System, potentialErr error) *automock.ApplicationTemplateService
		setupAppConverter   func(testSystem systemfetcher.System, potentialErr error) *automock.ApplicationConverter
	}
	tests := []testCase{
		{
			name:                "Succeeds",
			system:              testSystem,
			appInputOverride:    appInputOverride,
			placeholderMappings: placeholdersMappings,
			setupAppTemplateSvc: appTemplateSvcNoErrors,
			setupAppConverter:   appConvSvcNoErrors,
		},
		{
			name:                "Succeeds when app template should be enriched with additional labels",
			system:              testSystem,
			appInputOverride:    `{"description":"{{description}}","labels":{"legacy":"true"}}`,
			placeholderMappings: placeholdersMappings,
			setupAppTemplateSvc: func(testSystem systemfetcher.System, _ error) *automock.ApplicationTemplateService {
				resultTemplate := *appTemplateWithOverrides
				resultTemplate.ApplicationInputJSON = `{"description":"{{description}}","labels":{"legacy":"true","tenant":"123"},"name":"test1"}`
				appTemplateFromDB := *appTemplate
				appTemplateFromDB.ApplicationInputJSON = `{ "name": "test1","labels":{"tenant":"123"}}`

				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", context.TODO(), testSystem.TemplateID).Return(&appTemplateFromDB, nil).Once()
				svc.On("PrepareApplicationCreateInputJSON", &resultTemplate, fixInputValuesForSystem(testSystem)).Return(appRegisterInputJSON, nil).Once()
				return svc
			},
			setupAppConverter: appConvSvcNoErrors,
		},
		{
			name:                "Fails when app input from template and overrides have the same field with different types",
			system:              testSystem,
			expectedErr:         errors.New("values map[tenant:123] and hello of key labels have different types - map[string]interface {} and string"),
			appInputOverride:    `{"description":"{{description}}","labels":"hello"}`,
			placeholderMappings: placeholdersMappings,
			setupAppTemplateSvc: func(testSystem systemfetcher.System, _ error) *automock.ApplicationTemplateService {
				resultTemplate := *appTemplateWithOverrides
				resultTemplate.ApplicationInputJSON = `{"description":"{{description}}","labels":{"legacy":"true","tenant":"123"},"name":"test1"}`
				appTemplateFromDB := *appTemplate
				appTemplateFromDB.ApplicationInputJSON = `{ "name": "test1","labels":{"tenant":"123"}}`

				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", context.TODO(), testSystem.TemplateID).Return(&appTemplateFromDB, nil).Once()
				return svc
			},
			setupAppConverter: func(testSystem systemfetcher.System, _ error) *automock.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
		},
		{
			name:                "Fails when app input from template and overrides have a field with incorrect type",
			system:              testSystem,
			expectedErr:         errors.New("app template labels are with type string instead of map[string]interface{}"),
			appInputOverride:    `{"description":"{{description}}","labels":"hello"}`,
			placeholderMappings: placeholdersMappings,
			setupAppTemplateSvc: func(testSystem systemfetcher.System, _ error) *automock.ApplicationTemplateService {
				resultTemplate := *appTemplateWithOverrides
				resultTemplate.ApplicationInputJSON = `{"description":"{{description}}","labels":"hello2","name":"test1"}`
				appTemplateFromDB := *appTemplate
				appTemplateFromDB.ApplicationInputJSON = `{ "name": "test1","labels":"hello"}`

				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", context.TODO(), testSystem.TemplateID).Return(&appTemplateFromDB, nil).Once()
				return svc
			},
			setupAppConverter: func(testSystem systemfetcher.System, _ error) *automock.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
		},
		{
			name:                "Fails when template cannot be fetched",
			system:              testSystem,
			appInputOverride:    appInputOverride,
			placeholderMappings: placeholdersMappings,
			expectedErr:         errors.New("cannot get template"),
			setupAppTemplateSvc: func(testSystem systemfetcher.System, err error) *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", context.TODO(), testSystem.TemplateID).Return(nil, err).Once()
				return svc
			},
			setupAppConverter: func(testSystem systemfetcher.System, _ error) *automock.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
		},
		{
			name:                "Fails when app input JSON cannot be prepared",
			system:              testSystem,
			appInputOverride:    appInputOverride,
			placeholderMappings: placeholdersMappings,
			expectedErr:         errors.New("cannot prepare input json"),
			setupAppTemplateSvc: func(testSystem systemfetcher.System, err error) *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				inputValues := fixInputValuesForSystem(testSystem)
				svc.On("Get", context.TODO(), testSystem.TemplateID).Return(appTemplate, nil).Once()
				svc.On("PrepareApplicationCreateInputJSON", appTemplateWithOverrides, inputValues).Return("", err).Once()
				return svc
			},
			setupAppConverter: func(testSystem systemfetcher.System, _ error) *automock.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
		},
		{
			name:                "Fails when app input cannot be prepared",
			system:              testSystem,
			appInputOverride:    appInputOverride,
			placeholderMappings: placeholdersMappings,
			expectedErr:         errors.New("cannot prepare input"),
			setupAppTemplateSvc: appTemplateSvcNoErrors,
			setupAppConverter: func(testSystem systemfetcher.System, err error) *automock.ApplicationConverter {
				conv := &automock.ApplicationConverter{}
				conv.On("CreateInputJSONToModel", context.TODO(), appRegisterInputJSON).Return(model.ApplicationRegisterInput{}, err).Once()
				return conv
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// GIVEN
			templateSvc := test.setupAppTemplateSvc(test.system, test.expectedErr)
			convSvc := test.setupAppConverter(test.system, test.expectedErr)
			tr := systemfetcher.NewTemplateRenderer(templateSvc, convSvc, test.appInputOverride, test.placeholderMappings)

			// WHEN
			regIn, err := tr.ApplicationRegisterInputFromTemplate(context.TODO(), testSystem)

			// THEN
			if test.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, fixAppInputBySystem(test.system), *regIn)
			}
		})
	}
}

func fixInputValuesForSystem(s systemfetcher.System) model.ApplicationFromTemplateInputValues {
	return model.ApplicationFromTemplateInputValues{
		{
			Placeholder: "description",
			Value:       s.ProductDescription,
		},
	}
}

func fixAppInputBySystem(system systemfetcher.System) model.ApplicationRegisterInput {
	initStatusCond := model.ApplicationStatusConditionInitial
	return model.ApplicationRegisterInput{
		Name:            system.DisplayName,
		Description:     &system.ProductDescription,
		BaseURL:         &system.BaseURL,
		ProviderName:    &system.InfrastructureProvider,
		SystemNumber:    &system.SystemNumber,
		StatusCondition: &initStatusCond,
		Labels: map[string]interface{}{
			"managed": "true",
		},
	}
}
