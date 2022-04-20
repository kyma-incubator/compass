package systemfetcher_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestNewTemplateRenderer(t *testing.T) {
	t.Run("Creates a new renderer", func(t *testing.T) {
		appInputOverride := `{"name":"{{name}}"}`
		templateSvc := &automock.ApplicationTemplateService{}
		convSvc := &automock.ApplicationConverter{}
		convSvc.On("CreateInputJSONToModel", context.TODO(), appInputOverride).Return(model.ApplicationRegisterInput{}, nil).Once()
		defer mock.AssertExpectationsForObjects(t, templateSvc, convSvc)
		tr, err := systemfetcher.NewTemplateRenderer(templateSvc, convSvc, appInputOverride, []systemfetcher.PlaceholderMapping{})

		require.NoError(t, err)
		require.NotNil(t, tr)
	})
	t.Run("Fails to create a new renderer when the override application input is not valid", func(t *testing.T) {
		invalidOverrides := "invalid"
		expectedErr := errors.New("test err")

		templateSvc := &automock.ApplicationTemplateService{}
		convSvc := &automock.ApplicationConverter{}
		convSvc.On("CreateInputJSONToModel", context.TODO(), invalidOverrides).Return(model.ApplicationRegisterInput{}, expectedErr).Once()
		defer mock.AssertExpectationsForObjects(t, templateSvc, convSvc)

		tr, err := systemfetcher.NewTemplateRenderer(templateSvc, convSvc, invalidOverrides, []systemfetcher.PlaceholderMapping{})
		require.ErrorIs(t, err, expectedErr)
		require.Nil(t, tr)
	})
}
func TestApplicationRegisterInputFromTemplate(t *testing.T) {
	const (
		appTemplateID        = "appTmp1"
		appRegisterInputJSON = `{"name":"test"}`
		appInputOverride     = `{"name":"{{name}}"}`
	)

	placeholdersMappings := []systemfetcher.PlaceholderMapping{
		{
			PlaceholderName: "name",
			SystemKey:       "displayName",
		},
	}

	appTemplate := &model.ApplicationTemplate{
		ID:                   appTemplateID,
		ApplicationInputJSON: `{ "name": "testtest"}`,
	}

	appTemplateWithOverrides := &model.ApplicationTemplate{
		ID: appTemplateID,
		Placeholders: []model.ApplicationTemplatePlaceholder{
			{Name: "name"},
		},
		ApplicationInputJSON: appInputOverride,
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
			appInputOverride:    `{"name":"{{name}}","labels":{"legacy":"true"}}`,
			placeholderMappings: placeholdersMappings,
			setupAppTemplateSvc: func(testSystem systemfetcher.System, _ error) *automock.ApplicationTemplateService {
				resultTemplate := *appTemplateWithOverrides
				resultTemplate.ApplicationInputJSON = `{"labels":{"legacy":"true","tenant":"123"},"name":"{{name}}"}`
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
			name:             "Fails when app template has placeholders without assigned values",
			system:           testSystem,
			expectedErr:      errors.New("missing or empty key \"nonexistentKey\" in system input"),
			appInputOverride: appInputOverride,
			placeholderMappings: []systemfetcher.PlaceholderMapping{
				{
					PlaceholderName: "name",
					SystemKey:       "displayName",
				},
				{
					PlaceholderName: "description",
					SystemKey:       "nonexistentKey",
				},
			},
			setupAppTemplateSvc: func(testSystem systemfetcher.System, _ error) *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				//inputValues := fixInputValuesForSystem(testSystem)
				svc.On("Get", context.TODO(), testSystem.TemplateID).Return(appTemplate, nil).Once()
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
			convSvc.On("CreateInputJSONToModel", context.TODO(), test.appInputOverride).Return(model.ApplicationRegisterInput{}, nil).Once()
			tr, err := systemfetcher.NewTemplateRenderer(templateSvc, convSvc, test.appInputOverride, test.placeholderMappings)
			require.NoError(t, err)

			defer mock.AssertExpectationsForObjects(t, templateSvc, convSvc)

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
