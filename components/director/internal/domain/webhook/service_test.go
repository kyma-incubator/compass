package webhook_test

import (
	"context"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	"github.com/kyma-incubator/compass/components/director/internal/domain/webhook/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const testCaseErrorOnLoadingTenant = "Returns error on loading tenant"

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	modelInput := fixModelWebhookInput("foo")

	webhookModel := mock.MatchedBy(func(webhook *model.Webhook) bool {
		return webhook.Type == modelInput.Type && webhook.URL == modelInput.URL
	})

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant(), givenExternalTenant())

	testCases := []struct {
		Name         string
		RepositoryFn func() *automock.WebhookRepository
		UIDServiceFn func() *automock.UIDService
		ExpectedErr  error
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
			ExpectedErr: nil,
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
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			uidSvc := testCase.UIDServiceFn()

			svc := webhook.NewService(repo, uidSvc)

			// when
			result, err := svc.Create(ctx, givenApplicationID(), *modelInput)

			// then

			if testCase.ExpectedErr == nil {
				assert.NotEmpty(t, result)
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			repo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}

	t.Run(testCaseErrorOnLoadingTenant, func(t *testing.T) {
		svc := webhook.NewService(nil, nil)
		// when
		_, err := svc.Create(context.TODO(), givenApplicationID(), *modelInput)
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
}

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	url := "bar"

	webhookModel := fixModelWebhook("1", id, givenTenant(), url)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant(), givenExternalTenant())

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
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
			ExpectedWebhook:    webhookModel,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo, nil)

			// when
			actual, err := svc.Get(ctx, id)

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

	t.Run(testCaseErrorOnLoadingTenant, func(t *testing.T) {
		svc := webhook.NewService(nil, nil)
		// when
		_, err := svc.Get(context.TODO(), givenApplicationID())
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
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
	ctx = tenant.SaveToContext(ctx, givenTenant(), givenExternalTenant())

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

	t.Run(testCaseErrorOnLoadingTenant, func(t *testing.T) {
		svc := webhook.NewService(nil, nil)
		// when
		_, err := svc.List(context.TODO(), givenApplicationID())
		assert.True(t, apperrors.IsCannotReadTenant(err))
	})
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
	ctx = tenant.SaveToContext(ctx, givenTenant(), givenExternalTenant())

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
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
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when webhook retrieval failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo, nil)

			// when
			err := svc.Update(ctx, id, *modelInput)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run(testCaseErrorOnLoadingTenant, func(t *testing.T) {
		svc := webhook.NewService(nil, nil)
		// when
		err := svc.Update(context.TODO(), givenApplicationID(), *modelInput)
		assert.EqualError(t, err, fmt.Sprintf("while getting Webhook: %s", apperrors.NewCannotReadTenantError().Error()))
	})
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"
	url := "bar"

	webhookModel := fixModelWebhook("1", id, givenTenant(), url)

	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, givenTenant(), givenExternalTenant())

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.WebhookRepository
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(webhookModel, nil).Once()
				repo.On("Delete", ctx, webhookModel.TenantID, webhookModel.ID).Return(nil).Once()
				return repo
			},
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when webhook deletion failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(webhookModel, nil).Once()
				repo.On("Delete", ctx, webhookModel.TenantID, webhookModel.ID).Return(testErr).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when webhook retrieval failed",
			RepositoryFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("GetByID", ctx, givenTenant(), id).Return(nil, testErr).Once()
				return repo
			},
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()
			svc := webhook.NewService(repo, nil)

			// when
			err := svc.Delete(ctx, id)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}

	t.Run(testCaseErrorOnLoadingTenant, func(t *testing.T) {
		svc := webhook.NewService(nil, nil)
		// when
		err := svc.Delete(context.TODO(), id)
		assert.EqualError(t, err, fmt.Sprintf("while getting Webhook: %s", apperrors.NewCannotReadTenantError()))
	})
}
