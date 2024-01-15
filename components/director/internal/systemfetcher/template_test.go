package systemfetcher_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var emptyCtx = context.Background()

func TestNewTemplateRenderer(t *testing.T) {
	t.Run("Creates a new renderer", func(t *testing.T) {
		appInputOverride := `{"name":"{{name}}"}`
		templateSvc := &automock.ApplicationTemplateService{}
		convSvc := &automock.ApplicationConverter{}
		convSvc.On("CreateInputJSONToModel", emptyCtx, appInputOverride).Return(model.ApplicationRegisterInput{}, nil).Once()
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
		convSvc.On("CreateInputJSONToModel", emptyCtx, invalidOverrides).Return(model.ApplicationRegisterInput{}, expectedErr).Once()
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
		appInputOverride     = `{"name":"testtest"}`
	)
	var (
		optionalFalse = false
	)

	placeholdersMappings := []systemfetcher.PlaceholderMapping{
		{
			PlaceholderName: "name",
			SystemKey:       "$.displayName",
		},
	}

	appTemplate := &model.ApplicationTemplate{
		ID:                   appTemplateID,
		ApplicationInputJSON: `{"name":"testtest"}`,
	}

	appTemplateWithOverrides := &model.ApplicationTemplate{
		ID: appTemplateID,
		Placeholders: []model.ApplicationTemplatePlaceholder{
			{Name: "name", JSONPath: str.Ptr("$.displayName"), Optional: &optionalFalse},
		},
		ApplicationInputJSON: appInputOverride,
	}

	testSystem := systemfetcher.System{
		SystemPayload: map[string]interface{}{
			"systemNumber":           "123",
			"displayName":            "test",
			"productDescription":     "test",
			"baseUrl":                "http://test",
			"infrastructureProvider": "test",
		},
		TemplateID:      "123",
		StatusCondition: model.ApplicationStatusConditionConnected,
	}

	appTemplateSvcNoErrors := func(testSystem systemfetcher.System, _ error) *automock.ApplicationTemplateService {
		svc := &automock.ApplicationTemplateService{}
		inputValues := fixInputValuesForSystem(t, testSystem)
		template := &model.ApplicationTemplate{
			ID:                   appTemplateID,
			ApplicationInputJSON: `{ "name": "testtest"}`,
		}
		svc.On("Get", emptyCtx, testSystem.TemplateID).Return(template, nil).Once()
		svc.On("PrepareApplicationCreateInputJSON", appTemplateWithOverrides, inputValues).Return(appRegisterInputJSON, nil).Once()
		return svc
	}
	appConvSvcNoErrors := func(testSystem systemfetcher.System, _ error) *automock.ApplicationConverter {
		appInput := fixAppInputBySystem(t, testSystem)
		conv := &automock.ApplicationConverter{}
		conv.On("CreateInputJSONToModel", emptyCtx, appRegisterInputJSON).Return(appInput, nil).Once()
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
				resultTemplate.ApplicationInputJSON = `{"integrationSystemID":"a8396508-66be-4dc7-b463-577809289941","labels":{"legacy":"true","tenant":"123"},"name":"test1"}`
				appTemplateFromDB := model.ApplicationTemplate{
					ID:                   appTemplateID,
					ApplicationInputJSON: `{ "name": "testtest"}`,
				}
				appTemplateFromDB.ApplicationInputJSON = `{"name": "test1","labels":{"tenant":"123"},"integrationSystemID":"a8396508-66be-4dc7-b463-577809289941"}`

				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", emptyCtx, testSystem.TemplateID).Return(&appTemplateFromDB, nil).Once()
				svc.On("PrepareApplicationCreateInputJSON", &resultTemplate, fixInputValuesForSystem(t, testSystem)).Return(appRegisterInputJSON, nil).Once()
				return svc
			},
			setupAppConverter: appConvSvcNoErrors,
		},
		{
			name:                "Succeeds when app template has placeholders",
			system:              testSystem,
			appInputOverride:    `{"name":"{{name}}"}`,
			placeholderMappings: placeholdersMappings,
			setupAppTemplateSvc: func(testSystem systemfetcher.System, _ error) *automock.ApplicationTemplateService {
				resultTemplate := model.ApplicationTemplate{
					ID: appTemplateID,
					Placeholders: []model.ApplicationTemplatePlaceholder{
						{Name: "name", JSONPath: str.Ptr("$.displayName"), Optional: &optionalFalse},
					},
					ApplicationInputJSON: `{"integrationSystemID":"a8396508-66be-4dc7-b463-577809289941","labels":{"tenant":"123"},"name":"test1"}`,
				}
				appTemplateFromDB := model.ApplicationTemplate{
					ID: appTemplateID,
					Placeholders: []model.ApplicationTemplatePlaceholder{
						{Name: "name", JSONPath: str.Ptr("$.displayName"), Optional: &optionalFalse},
					},
					ApplicationInputJSON: `{ "name": "test1","labels":{"tenant":"123"},"integrationSystemID":"a8396508-66be-4dc7-b463-577809289941"}`,
				}

				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", emptyCtx, testSystem.TemplateID).Return(&appTemplateFromDB, nil).Once()
				svc.On("PrepareApplicationCreateInputJSON", &resultTemplate, fixInputValuesForSystemWhichAppTemplateHasPlaceholders(t, testSystem)).Return(`{"name":"test"}`, nil).Once()
				return svc
			},
			setupAppConverter: func(testSystem systemfetcher.System, _ error) *automock.ApplicationConverter {
				appInput := fixAppInputBySystem(t, testSystem)
				conv := &automock.ApplicationConverter{}
				conv.On("CreateInputJSONToModel", emptyCtx, `{"name":"test"}`).Return(appInput, nil).Once()
				return conv
			},
		},
		{
			name:             "Fails when app template has placeholders without assigned values",
			system:           testSystem,
			expectedErr:      errors.New("missing or empty key \"$.nonexistentKey\" in system payload"),
			appInputOverride: appInputOverride,
			placeholderMappings: []systemfetcher.PlaceholderMapping{
				{
					PlaceholderName: "name",
					SystemKey:       "$.displayName",
				},
				{
					PlaceholderName: "description",
					SystemKey:       "$.nonexistentKey",
				},
			},
			setupAppTemplateSvc: func(testSystem systemfetcher.System, _ error) *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", emptyCtx, testSystem.TemplateID).Return(appTemplate, nil).Once()
				return svc
			},
			setupAppConverter: func(testSystem systemfetcher.System, _ error) *automock.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
		},
		{
			name:                "Fails when app input from template cannot be unmarshalled into a map",
			system:              testSystem,
			expectedErr:         errors.New("while unmarshaling original application input"),
			appInputOverride:    `{"description":"{{description}}","labels":"hello"}`,
			placeholderMappings: placeholdersMappings,
			setupAppTemplateSvc: func(testSystem systemfetcher.System, _ error) *automock.ApplicationTemplateService {
				appTemplateFromDB := *appTemplate
				appTemplateFromDB.ApplicationInputJSON = ``

				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", emptyCtx, testSystem.TemplateID).Return(&appTemplateFromDB, nil).Once()
				return svc
			},
			setupAppConverter: func(testSystem systemfetcher.System, _ error) *automock.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
		},
		{
			name:                "Fails when app input from override cannot be unmarshalled into a map",
			system:              testSystem,
			expectedErr:         errors.New("while unmarshaling override application input"),
			appInputOverride:    ``,
			placeholderMappings: placeholdersMappings,
			setupAppTemplateSvc: func(testSystem systemfetcher.System, _ error) *automock.ApplicationTemplateService {
				appTemplateFromDB := *appTemplate
				appTemplateFromDB.ApplicationInputJSON = `{"name": "{{display-name}}"}`

				svc := &automock.ApplicationTemplateService{}
				svc.On("Get", emptyCtx, testSystem.TemplateID).Return(&appTemplateFromDB, nil).Once()
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
				svc.On("Get", emptyCtx, testSystem.TemplateID).Return(nil, err).Once()
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
				inputValues := fixInputValuesForSystem(t, testSystem)
				template := &model.ApplicationTemplate{
					ID:                   appTemplateID,
					ApplicationInputJSON: `{ "name": "testtest"}`,
				}
				svc.On("Get", emptyCtx, testSystem.TemplateID).Return(template, nil).Once()
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
				conv.On("CreateInputJSONToModel", emptyCtx, appRegisterInputJSON).Return(model.ApplicationRegisterInput{}, err).Once()
				return conv
			},
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			// GIVEN
			templateSvc := test.setupAppTemplateSvc(test.system, test.expectedErr)
			convSvc := test.setupAppConverter(test.system, test.expectedErr)
			convSvc.On("CreateInputJSONToModel", emptyCtx, test.appInputOverride).Return(model.ApplicationRegisterInput{}, nil).Once()
			tr, err := systemfetcher.NewTemplateRenderer(templateSvc, convSvc, test.appInputOverride, test.placeholderMappings)
			require.NoError(t, err)

			defer mock.AssertExpectationsForObjects(t, templateSvc, convSvc)

			// WHEN
			regIn, err := tr.ApplicationRegisterInputFromTemplate(emptyCtx, testSystem)

			// THEN
			if test.expectedErr != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.expectedErr.Error())
			} else {
				require.NoError(t, err)
				require.Equal(t, fixAppInputBySystem(t, test.system), *regIn)
			}
		})
	}
}
