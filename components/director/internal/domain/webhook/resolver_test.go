package webhook_test

import (
	"context"
	"errors"

	"github.com/stretchr/testify/require"

	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestResolver_AddWebhook(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"
	gqlWebhookInput := fixGQLWebhookInput("foo")
	modelWebhookInput := fixModelWebhookInput("foo")

	gqlWebhook := fixGQLWebhook(id, "", "")
	modelWebhook := fixModelWebhook(id, applicationID, "foo")

	testCases := []struct {
		Name               string
		ServiceFn          func() *automock.WebhookService
		AppServiceFn       func() *automock.ApplicationService
		ConverterFn        func() *automock.WebhookConverter
		InputApplicationID string
		InputWebhook       graphql.ApplicationWebhookInput
		ExpectedWebhook    *graphql.ApplicationWebhook
		ExpectedErr        error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", context.TODO(), applicationID, *modelWebhookInput).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(modelWebhook, nil).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), applicationID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    gqlWebhook,
			ExpectedErr:        nil,
		},
		{
			Name: "Returns error when application not exist",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), applicationID).Return(false, nil)
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    nil, ExpectedErr: errors.New("Cannot add Webhook to not existing Application"),
		},
		{
			Name: "Returns error when application existence check failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), applicationID).Return(false, testErr).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    nil,
			ExpectedErr:        testErr,
		},
		{
			Name: "Returns error when webhook creation failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", context.TODO(), applicationID, *modelWebhookInput).Return("", testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), applicationID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    nil,
			ExpectedErr:        testErr,
		},
		{
			Name: "Returns error when webhook retrieval failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Create", context.TODO(), applicationID, *modelWebhookInput).Return(id, nil).Once()
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			AppServiceFn: func() *automock.ApplicationService {
				appSvc := &automock.ApplicationService{}
				appSvc.On("Exist", context.TODO(), applicationID).Return(true, nil).Once()
				return appSvc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputApplicationID: applicationID,
			InputWebhook:       *gqlWebhookInput,
			ExpectedWebhook:    nil,
			ExpectedErr:        testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			appSvc := testCase.AppServiceFn()
			converter := testCase.ConverterFn()

			resolver := webhook.NewResolver(svc, appSvc, converter)

			// when
			result, err := resolver.AddApplicationWebhook(context.TODO(), testCase.InputApplicationID, testCase.InputWebhook)

			// then
			assert.Equal(t, testCase.ExpectedWebhook, result)
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			svc.AssertExpectations(t)
			appSvc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_UpdateWebhook(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"
	gqlWebhookInput := fixGQLWebhookInput("foo")
	modelWebhookInput := fixModelWebhookInput("foo")
	gqlWebhook := fixGQLWebhook(id, "", "")
	modelWebhook := fixModelWebhook(id, applicationID, "foo")

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.WebhookService
		ConverterFn     func() *automock.WebhookConverter
		InputWebhookID  string
		InputWebhook    graphql.ApplicationWebhookInput
		ExpectedWebhook *graphql.ApplicationWebhook
		ExpectedErr     error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", context.TODO(), id, *modelWebhookInput).Return(nil).Once()
				svc.On("Get", context.TODO(), id).Return(modelWebhook, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: gqlWebhook,
			ExpectedErr:     nil,
		},
		{
			Name: "Returns error when webhook update failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", context.TODO(), id, *modelWebhookInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Returns error when webhook retrieval failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Update", context.TODO(), id, *modelWebhookInput).Return(nil).Once()
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("InputFromGraphQL", gqlWebhookInput).Return(modelWebhookInput).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := webhook.NewResolver(svc, nil, converter)

			// when
			result, err := resolver.UpdateApplicationWebhook(context.TODO(), testCase.InputWebhookID, testCase.InputWebhook)

			// then
			assert.Equal(t, testCase.ExpectedWebhook, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}

func TestResolver_DeleteWebhook(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	applicationID := "foo"
	id := "bar"

	gqlWebhookInput := fixGQLWebhookInput("foo")

	gqlWebhook := fixGQLWebhook(id, "", "")
	modelWebhook := fixModelWebhook(id, applicationID, "foo")

	testCases := []struct {
		Name            string
		ServiceFn       func() *automock.WebhookService
		ConverterFn     func() *automock.WebhookConverter
		InputWebhookID  string
		InputWebhook    graphql.ApplicationWebhookInput
		ExpectedWebhook *graphql.ApplicationWebhook
		ExpectedErr     error
	}{
		{
			Name: "Success",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", context.TODO(), id).Return(modelWebhook, nil).Once()
				svc.On("Delete", context.TODO(), id).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: gqlWebhook,
			ExpectedErr:     nil,
		},
		{
			Name: "Returns error when webhook retrieval failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", context.TODO(), id).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Returns error when webhook deletion failed",
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("Get", context.TODO(), id).Return(modelWebhook, nil).Once()
				svc.On("Delete", context.TODO(), id).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				conv := &automock.WebhookConverter{}
				conv.On("ToGraphQL", modelWebhook).Return(gqlWebhook).Once()
				return conv
			},
			InputWebhookID:  id,
			InputWebhook:    *gqlWebhookInput,
			ExpectedWebhook: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := webhook.NewResolver(svc, nil, converter)

			// when
			result, err := resolver.DeleteApplicationWebhook(context.TODO(), testCase.InputWebhookID)

			// then
			assert.Equal(t, testCase.ExpectedWebhook, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			svc.AssertExpectations(t)
			converter.AssertExpectations(t)
		})
	}
}
