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

	webhookModel := mock.MatchedBy(func(webhook *model.Webhook) bool {
		return webhook.Type == modelInput.Type && webhook.URL == modelInput.URL
	})

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant())

	testCases := []struct {
		Name          string
		RepositoryFn  func() *automock.WebhookRepository
		UIDServiceFn  func() *automock.UIDService
		Input         model.WebhookInput
		ID            string
		ApplicationID string
		ExpectedErr   error
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("Create", ctx, webhookModel).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return("foo").Once()
				return svc
			},
			Input:         *modelInput,
			ApplicationID: "1",
			ExpectedErr:   nil,
		},
		{
			Name: "Returns error when webhook creation failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("Create", ctx, webhookModel).Return(testErr).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return("").Once()
				return svc
			},
			Input:         *modelInput,
			ApplicationID: "1",
			ExpectedErr:   testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			uidSvc := testCase.UIDServiceFn()

			svc := webhook.NewService(repo, uidSvc)

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
			uidSvc.AssertExpectations(t)
		})
	}
}

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	url := "bar"

	webhookModel := fixModelWebhook("1", id, givenTenant(), url)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant())

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		Input              model.WebhookInput
		InputID            string
		ExpectedWebhook    *model.Webhook
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(webhookModel, nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedWebhook:    webhookModel,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when webhook retrieval failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(nil, testErr).Once()
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
			svc := webhook.NewService(repo, nil)

			// when
			actual, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedWebhook, actual)
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

	modelWebhooks := []*model.Webhook{
		fixModelWebhook("1", "foo", givenTenant(), "Foo"),
		fixModelWebhook("2", "bar", givenTenant(), "Bar"),
	}
	applicationID := "foo"

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant())

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		ExpectedResult     []*model.Webhook
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByApplicationID", ctx, givenTenant(), applicationID).Return(modelWebhooks, nil).Once()
				return repo
			},
			ExpectedResult:     modelWebhooks,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when webhook listing failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("ListByApplicationID", ctx, givenTenant(), applicationID).Return(nil, testErr).Once()
				return repo
			},
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo, nil)

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

	inputWebhookModel := mock.MatchedBy(func(webhook *model.Webhook) bool {
		return webhook.URL == modelInput.URL
	})

	webhookModel := fixModelWebhook("1", id, givenTenant(), url)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant())

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		Input              model.WebhookInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(webhookModel, nil).Once()
				repo.On("Update", ctx, inputWebhookModel).Return(nil).Once()
				return repo
			},
			InputID:            id,
			Input:              *modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when webhook update failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(webhookModel, nil).Once()
				repo.On("Update", ctx, inputWebhookModel).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			Input:              *modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when webhook retrieval failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(nil, testErr).Once()
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
			svc := webhook.NewService(repo, nil)

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

	webhookModel := fixModelWebhook("1", id, givenTenant(), url)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant())

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		Input              model.WebhookInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(webhookModel, nil).Once()
				repo.On("Delete", ctx, webhookModel.Tenant, webhookModel.ID).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when webhook deletion failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(webhookModel, nil).Once()
				repo.On("Delete", ctx, webhookModel.Tenant, webhookModel.ID).Return(testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when webhook retrieval failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(nil, testErr).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo, nil)

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
