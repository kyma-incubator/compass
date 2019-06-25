package webhook_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelInput := fixModelWebhookInput("foo")

	webhookModel := mock.MatchedBy(func(webhook *model.ApplicationWebhook) bool {
		return webhook.Type == modelInput.Type && webhook.URL == modelInput.URL
	})

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name          string
		RepositoryFn  func() *automock.WebhookRepository
		Input         model.ApplicationWebhookInput
		ApplicationID string
		ExpectedErr   error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("Create", webhookModel).Return(nil).Once()
				return repo
			},
			Input:         *modelInput,
			ApplicationID: "1",
			ExpectedErr:   nil,
		},
		{
			Name: "Error",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("Create", webhookModel).Return(testErr).Once()
				return repo
			},
			Input:         *modelInput,
			ApplicationID: "1",
			ExpectedErr:   testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo)

			// when
			result, err := svc.Create(ctx, testCase.ApplicationID, testCase.Input)

			// then
			assert.IsType(t, "string", result)
			if testCase.ExpectedErr == nil {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	url := "bar"

	webhookModel := fixModelWebhook("1", id, url)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		Input              model.ApplicationWebhookInput
		InputID            string
		ExpectedWebhook    *model.ApplicationWebhook
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", id).Return(webhookModel, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedWebhook:    webhookModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedWebhook:    webhookModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo)

			// when
			webhook, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedWebhook, webhook)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_List(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelWebhooks := []*model.ApplicationWebhook{
		fixModelWebhook("1", "foo", "Foo"),
		fixModelWebhook("2", "bar", "Bar"),
	}
	applicationID := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		ExpectedResult     []*model.ApplicationWebhook
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByApplicationID", applicationID).Return(modelWebhooks, nil).Once()
				return repo
			},
			ExpectedResult:     modelWebhooks,
			ExpectedErrMessage: "",
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByApplicationID", applicationID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo)

			// when
			webhooks, err := svc.List(ctx, applicationID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, webhooks)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	url := "foo"
	id := "bar"
	modelInput := fixModelWebhookInput(url)

	inputWebhookModel := mock.MatchedBy(func(webhook *model.ApplicationWebhook) bool {
		return webhook.URL == modelInput.URL
	})

	webhookModel := fixModelWebhook("1", id, url)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		Input              model.ApplicationWebhookInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", id).Return(webhookModel, nil).Once()
				repo.On("Update", inputWebhookModel).Return(nil).Once()
				return repo
			},
			InputID:            id,
			Input:              *modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Update Error",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", id).Return(webhookModel, nil).Once()
				repo.On("Update", inputWebhookModel).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			Input:              *modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			Input:              *modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo)

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	url := "bar"

	webhookModel := fixModelWebhook("1", id, url)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, "tenant")

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		Input              model.ApplicationWebhookInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", id).Return(webhookModel, nil).Once()
				repo.On("Delete", webhookModel).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Delete Error",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", id).Return(webhookModel, nil).Once()
				repo.On("Delete", webhookModel).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Get Error",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo)

			// when
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}
