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
		appRegisterInputJSON = `{ "name": "test1"}`

		displayNameKey = "display-name"
	)

	appTemplate := &model.ApplicationTemplate{
		ID: appTemplateID,
		Placeholders: []model.ApplicationTemplatePlaceholder{
			{Name: displayNameKey},
		},
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
		inputValues := inputValuesForSystem(testSystem)
		svc.On("Get", context.TODO(), testSystem.TemplateID).Return(appTemplate, nil).Once()
		svc.On("PrepareApplicationCreateInputJSON", appTemplate, inputValues).Return(appRegisterInputJSON, nil).Once()
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
		setupAppTemplateSvc func(testSystem systemfetcher.System, potentialErr error) *automock.ApplicationTemplateService
		setupAppConverter   func(testSystem systemfetcher.System, potentialErr error) *automock.ApplicationConverter
	}
	tests := []testCase{
		{
			name:                "Succeeds",
			system:              testSystem,
			setupAppTemplateSvc: appTemplateSvcNoErrors,
			setupAppConverter:   appConvSvcNoErrors,
		},
		{
			name:        "Fails when template cannot be fetched",
			system:      testSystem,
			expectedErr: errors.New("cannot get template"),
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
			name:        "Fails when app input JSON cannot be prepared",
			system:      testSystem,
			expectedErr: errors.New("cannot prepare input json"),
			setupAppTemplateSvc: func(testSystem systemfetcher.System, err error) *automock.ApplicationTemplateService {
				svc := &automock.ApplicationTemplateService{}
				inputValues := inputValuesForSystem(testSystem)
				svc.On("Get", context.TODO(), testSystem.TemplateID).Return(appTemplate, nil).Once()
				svc.On("PrepareApplicationCreateInputJSON", appTemplate, inputValues).Return("", err).Once()
				return svc
			},
			setupAppConverter: func(testSystem systemfetcher.System, _ error) *automock.ApplicationConverter {
				return &automock.ApplicationConverter{}
			},
		},
		{
			name:                "Fails when app input JSON cannot be prepared",
			system:              testSystem,
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
			tr := systemfetcher.NewTemplateRenderer(templateSvc, convSvc)

			// WHEN
			regIn, err := tr.ApplicationRegisterInputFromTemplate(context.TODO(), testSystem)

			// THEN
			if test.expectedErr != nil {
				require.Error(t, err)
				require.ErrorIs(t, err, test.expectedErr)
			} else {
				require.NoError(t, err)
				require.Equal(t, fixAppInputBySystem(test.system), *regIn)
			}
		})
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
