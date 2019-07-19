package application_test

import (
	"context"
	"testing"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/internal/domain/application"
	"github.com/kyma-incubator/compass/components/director/internal/domain/application/automock"
	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/tenant"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestService_Create(t *testing.T) {
	// given
	testErr := errors.New("Test error")
	modelInput := model.ApplicationInput{
		Name: "foo.bar-not",
		Webhooks: []*model.WebhookInput{
			{URL: "test.foo.com"},
			{URL: "test.bar.com"},
		},
		Documents: []*model.DocumentInput{
			{Title: "foo", Description: "test"},
			{Title: "bar", Description: "test"},
		},
		Apis: []*model.APIDefinitionInput{
			{Name: "foo"}, {Name: "bar"},
		},
		EventAPIs: []*model.EventAPIDefinitionInput{
			{Name: "foo"}, {Name: "bar"},
		},
	}
	id := "foo"

	appModel := modelFromInput(modelInput, id)

	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name           string
		AppRepoFn      func() *automock.ApplicationRepository
		WebhookRepoFn  func() *automock.WebhookRepository
		APIRepoFn      func() *automock.APIRepository
		EventAPIRepoFn func() *automock.EventAPIRepository
		DocumentRepoFn func() *automock.DocumentRepository
		UIDServiceFn   func() *automock.UIDService
		Input          model.ApplicationInput
		ExpectedErr    error
	}{
		{
			Name: "Success",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Create", mock.MatchedBy(appModel.ApplicationMatcherFn)).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("CreateMany", mock.Anything).Return(nil).Once()
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("CreateMany", mock.Anything).Return(nil).Once()
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("CreateMany", mock.Anything).Return(nil).Once()
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("CreateMany", mock.Anything).Return(nil).Once()
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return(id)
				return svc
			},
			Input:       modelInput,
			ExpectedErr: nil,
		},
		{
			Name: "Returns error when application name is empty",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       model.ApplicationInput{Name: ""},
			ExpectedErr: errors.New("a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character"),
		},
		{
			Name: "Returns error when application name contains uppercase letter",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				return svc
			},
			Input:       model.ApplicationInput{Name: "upperCase"},
			ExpectedErr: errors.New("a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character"),
		},
		{
			Name: "Returns error when application creation failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Create", mock.MatchedBy(appModel.ApplicationMatcherFn)).Return(testErr).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			UIDServiceFn: func() *automock.UIDService {
				svc := &automock.UIDService{}
				svc.On("Generate").Return("").Once()
				return svc
			},
			Input:       modelInput,
			ExpectedErr: testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := testCase.AppRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			apiRepo := testCase.APIRepoFn()
			eventAPIRepo := testCase.EventAPIRepoFn()
			documentRepo := testCase.DocumentRepoFn()
			uidSvc := testCase.UIDServiceFn()
			svc := application.NewService(appRepo, webhookRepo, apiRepo, eventAPIRepo, documentRepo, uidSvc)

			// when
			result, err := svc.Create(ctx, testCase.Input)

			// then
			assert.IsType(t, "string", result)
			if err == nil {
				require.Nil(t, testCase.ExpectedErr)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			}

			appRepo.AssertExpectations(t)
			webhookRepo.AssertExpectations(t)
			apiRepo.AssertExpectations(t)
			eventAPIRepo.AssertExpectations(t)
			documentRepo.AssertExpectations(t)
			uidSvc.AssertExpectations(t)
		})
	}
}

func TestService_CreateWithInvalidNames(t *testing.T) {
	//GIVEN
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name               string
		InputID            string
		Input              model.ApplicationInput
		ExpectedErrMessage string
	}{
		{
			Name:               "Returns error when application name is empty",
			InputID:            "foo",
			Input:              model.ApplicationInput{Name: ""},
			ExpectedErrMessage: "a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character",
		},
		{
			Name:               "Returns error when application name contains upper case letter",
			InputID:            "foo",
			Input:              model.ApplicationInput{Name: "upperCase"},
			ExpectedErrMessage: "a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character",
		}}

	for _, testCase := range (testCases) {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := application.NewService(nil, nil, nil, nil, nil, nil)

			//WHEN
			_, err := svc.Create(ctx, testCase.Input)

			//THEN
			assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
		})

	}
}

func TestService_Update(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	modelInput := model.ApplicationInput{
		Name: "bar",
	}
	id := "foo"

	appModel := modelFromInput(modelInput, id)

	inputApplicationModel := mock.MatchedBy(func(app *model.Application) bool {
		return app.Name == modelInput.Name
	})

	applicationModel := &model.Application{
		ID:          id,
		Name:        "foo",
		Description: &desc,
	}

	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name               string
		AppRepoFn          func() *automock.ApplicationRepository
		WebhookRepoFn      func() *automock.WebhookRepository
		APIRepoFn          func() *automock.APIRepository
		EventAPIRepoFn     func() *automock.EventAPIRepository
		DocumentRepoFn     func() *automock.DocumentRepository
		Input              model.ApplicationInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, "foo").Return(applicationModel, nil).Once()
				repo.On("Update", inputApplicationModel).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(nil).Once()
				repo.On("CreateMany", appModel.Webhooks).Return(nil).Once()
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(nil).Once()
				repo.On("CreateMany", appModel.Apis).Return(nil).Once()
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(nil).Once()
				repo.On("CreateMany", appModel.EventAPIs).Return(nil).Once()
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(nil).Once()
				repo.On("CreateMany", appModel.Documents).Return(nil).Once()
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when application update failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, "foo").Return(applicationModel, nil).Once()
				repo.On("Update", inputApplicationModel).Return(testErr).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application retrieval failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, "foo").Return(nil, testErr).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when deleting apllication's subresource failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, "foo").Return(applicationModel, nil).Once()
				repo.On("Update", inputApplicationModel).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(testErr).Once()
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			InputID:            "foo",
			Input:              modelInput,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := testCase.AppRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			apiRepo := testCase.APIRepoFn()
			eventAPIRepo := testCase.EventAPIRepoFn()
			documentRepo := testCase.DocumentRepoFn()

			svc := application.NewService(appRepo, webhookRepo, apiRepo, eventAPIRepo, documentRepo, nil)

			// when
			err := svc.Update(ctx, testCase.InputID, testCase.Input)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			appRepo.AssertExpectations(t)
			webhookRepo.AssertExpectations(t)
			apiRepo.AssertExpectations(t)
			eventAPIRepo.AssertExpectations(t)
			documentRepo.AssertExpectations(t)
		})
	}
}

func TestService_UpdateWithInvalidNames(t *testing.T) {
	//GIVEN
	tnt := "tenant"
	appID := ""
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name               string
		InputID            string
		Input              model.ApplicationInput
		ExpectedErrMessage string
	}{
		{
			Name:               "Returns error when application name is empty",
			InputID:            "foo",
			Input:              model.ApplicationInput{Name: ""},
			ExpectedErrMessage: "a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character",
		},
		{
			Name:               "Returns error when application name contains upper case letter",
			InputID:            "foo",
			Input:              model.ApplicationInput{Name: "upperCase"},
			ExpectedErrMessage: "a DNS-1123 subdomain must consist of lower case alphanumeric characters, '-' or '.', and must start and end with an alphanumeric character",
		}}

	for _, testCase := range (testCases) {
		t.Run(testCase.Name, func(t *testing.T) {
			svc := application.NewService(nil, nil, nil, nil, nil, nil)

			//WHEN
			err := svc.Update(ctx, appID, testCase.Input)

			//THEN
			assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
		})

	}
}

func TestService_Delete(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"

	desc := "Lorem ipsum"

	applicationModel := &model.Application{
		ID:          id,
		Name:        "foo",
		Description: &desc,
	}

	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name               string
		AppRepoFn          func() *automock.ApplicationRepository
		WebhookRepoFn      func() *automock.WebhookRepository
		APIRepoFn          func() *automock.APIRepository
		EventAPIRepoFn     func() *automock.EventAPIRepository
		DocumentRepoFn     func() *automock.DocumentRepository
		Input              model.ApplicationInput
		InputID            string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, id).Return(applicationModel, nil).Once()
				repo.On("Delete", applicationModel).Return(nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(nil).Once()
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(nil).Once()
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(nil).Once()
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when application deletion failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, id).Return(applicationModel, nil).Once()
				repo.On("Delete", applicationModel).Return(testErr).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(nil).Once()
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(nil).Once()
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(nil).Once()
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(nil).Once()
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when deleting application's subresource failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, id).Return(applicationModel, nil).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				repo.On("DeleteAllByApplicationID", id).Return(testErr).Once()
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application retrieval failed",
			AppRepoFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, id).Return(nil, testErr).Once()
				return repo
			},
			WebhookRepoFn: func() *automock.WebhookRepository {
				repo := &automock.WebhookRepository{}
				return repo
			},
			APIRepoFn: func() *automock.APIRepository {
				repo := &automock.APIRepository{}
				return repo
			},
			EventAPIRepoFn: func() *automock.EventAPIRepository {
				repo := &automock.EventAPIRepository{}
				return repo
			},
			DocumentRepoFn: func() *automock.DocumentRepository {
				repo := &automock.DocumentRepository{}
				return repo
			},
			InputID:            id,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			appRepo := testCase.AppRepoFn()
			webhookRepo := testCase.WebhookRepoFn()
			apiRepo := testCase.APIRepoFn()
			eventAPIRepo := testCase.EventAPIRepoFn()
			documentRepo := testCase.DocumentRepoFn()

			svc := application.NewService(appRepo, webhookRepo, apiRepo, eventAPIRepo, documentRepo, nil)

			// when
			err := svc.Delete(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			appRepo.AssertExpectations(t)
			webhookRepo.AssertExpectations(t)
			apiRepo.AssertExpectations(t)
			eventAPIRepo.AssertExpectations(t)
			documentRepo.AssertExpectations(t)
		})
	}
}

func TestService_Get(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	id := "foo"

	desc := "Lorem ipsum"

	applicationModel := &model.Application{
		ID:          "foo",
		Name:        "foo",
		Description: &desc,
	}

	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name                string
		RepositoryFn        func() *automock.ApplicationRepository
		Input               model.ApplicationInput
		InputID             string
		ExpectedApplication *model.Application
		ExpectedErrMessage  string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, id).Return(applicationModel, nil).Once()
				return repo
			},
			InputID:             id,
			ExpectedApplication: applicationModel,
			ExpectedErrMessage:  "",
		},
		{
			Name: "Returns error when application retrieval failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, id).Return(nil, testErr).Once()
				return repo
			},
			InputID:             id,
			ExpectedApplication: applicationModel,
			ExpectedErrMessage:  testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo, nil, nil, nil, nil, nil)

			// when
			app, err := svc.Get(ctx, testCase.InputID)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedApplication, app)
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

	modelApplications := []*model.Application{
		fixModelApplication("foo", "foo", "Lorem Ipsum"),
		fixModelApplication("bar", "bar", "Lorem Ipsum"),
	}
	applicationPage := &model.ApplicationPage{
		Data:       modelApplications,
		TotalCount: len(modelApplications),
		PageInfo: &pagination.Page{
			HasNextPage: false,
			EndCursor:   "end",
			StartCursor: "start",
		},
	}

	first := 2
	after := "test"
	filter := []*labelfilter.LabelFilter{
		{Label: "", Values: []string{"foo", "bar"}, Operator: labelfilter.FilterOperatorAll},
	}

	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		InputLabelFilters  []*labelfilter.LabelFilter
		InputPageSize      *int
		InputCursor        *string
		ExpectedResult     *model.ApplicationPage
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("List", tnt, filter, &first, &after).Return(applicationPage, nil).Once()
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      &first,
			InputCursor:        &after,
			ExpectedResult:     applicationPage,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when application listing failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("List", tnt, filter, &first, &after).Return(nil, testErr).Once()
				return repo
			},
			InputLabelFilters:  filter,
			InputPageSize:      &first,
			InputCursor:        &after,
			ExpectedResult:     nil,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo, nil, nil, nil, nil, nil)

			// when
			app, err := svc.List(ctx, testCase.InputLabelFilters, testCase.InputPageSize, testCase.InputCursor)

			// then
			if testCase.ExpectedErrMessage == "" {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, app)
			} else {
				assert.Contains(t, err.Error(), testCase.ExpectedErrMessage)
			}

			repo.AssertExpectations(t)
		})
	}
}

func TestService_Exist(t *testing.T) {
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)
	testError := errors.New("Test error")

	applicationID := "id"

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		InputApplicationID string
		ExptectedValue     bool
		ExpectedError      error
	}{
		{
			Name: "Application exits",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exist", tnt, applicationID).Return(true, nil)
				return repo
			},
			InputApplicationID: applicationID,
			ExptectedValue:     true,
			ExpectedError:      nil,
		},
		{
			Name: "Application not exits",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exist", tnt, applicationID).Return(false, nil)
				return repo
			},
			InputApplicationID: applicationID,
			ExptectedValue:     false,
			ExpectedError:      nil,
		},
		{
			Name: "Returns error",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("Exist", tnt, applicationID).Return(false, testError)
				return repo
			},
			InputApplicationID: applicationID,
			ExptectedValue:     false,
			ExpectedError:      testError,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			appRepo := testCase.RepositoryFn()
			svc := application.NewService(appRepo, nil, nil, nil, nil, nil)

			// WHEN
			value, err := svc.Exist(ctx, testCase.InputApplicationID)

			// THEN
			if testCase.ExpectedError != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				require.Nil(t, err)
			}

			assert.Equal(t, testCase.ExptectedValue, value)
			appRepo.AssertExpectations(t)
		})
	}
}

func TestService_AddLabel(t *testing.T) {
	// given
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testErr := errors.New("Test error")

	desc := "Lorem ipsum"

	applicationID := "foo"
	modifiedApplicationModel := fixModelApplicationWithLabels(applicationID, "foo", map[string][]string{
		"key": {"value1"},
	})
	modifiedApplicationModel.Description = &desc

	labelKey := "key"
	labelValues := []string{"value1"}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		InputApplicationID string
		InputKey           string
		InputValues        []string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, applicationID).Return(fixModelApplication(applicationID, "foo", desc), nil).Once()
				repo.On("Update", modifiedApplicationModel).Return(nil).Once()

				return repo
			},
			InputApplicationID: applicationID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when application update failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, applicationID).Return(fixModelApplication(applicationID, "foo", desc), nil).Once()
				repo.On("Update", modifiedApplicationModel).Return(testErr).Once()

				return repo
			},
			InputApplicationID: applicationID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application retrieval failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, applicationID).Return(nil, testErr).Once()

				return repo
			},
			InputApplicationID: applicationID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo, nil, nil, nil, nil, nil)

			// when
			err := svc.AddLabel(ctx, testCase.InputApplicationID, testCase.InputKey, testCase.InputValues)

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

func TestService_DeleteLabel(t *testing.T) {
	// given
	tnt := "tenant"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt)

	testErr := errors.New("Test error")

	applicationID := "foo"
	modifiedApplicationModel := fixModelApplicationWithLabels(applicationID, "foo", map[string][]string{})

	labelKey := "key"
	labelValues := []string{"value1", "value2"}

	testCases := []struct {
		Name               string
		RepositoryFn       func() *automock.ApplicationRepository
		InputApplicationID string
		InputKey           string
		InputValues        []string
		ExpectedErrMessage string
	}{
		{
			Name: "Success",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, applicationID).Return(
					fixModelApplicationWithLabels(applicationID, "foo", map[string][]string{
						"key": {"value1", "value2"},
					}), nil).Once()
				repo.On("Update", modifiedApplicationModel).Return(nil).Once()

				return repo
			},
			InputApplicationID: applicationID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: "",
		},
		{
			Name: "Returns error when application update failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, applicationID).Return(
					fixModelApplicationWithLabels(applicationID, "foo", map[string][]string{
						"key": {"value1", "value2"},
					}), nil).Once()
				repo.On("Update", modifiedApplicationModel).Return(testErr).Once()

				return repo
			},
			InputApplicationID: applicationID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: testErr.Error(),
		},
		{
			Name: "Returns error when application retrieval failed",
			RepositoryFn: func() *automock.ApplicationRepository {
				repo := &automock.ApplicationRepository{}
				repo.On("GetByID", tnt, applicationID).Return(nil, testErr).Once()

				return repo
			},
			InputApplicationID: applicationID,
			InputKey:           labelKey,
			InputValues:        labelValues,
			ExpectedErrMessage: testErr.Error(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			repo := testCase.RepositoryFn()

			svc := application.NewService(repo, nil, nil, nil, nil, nil)

			// when
			err := svc.DeleteLabel(ctx, testCase.InputApplicationID, testCase.InputKey, testCase.InputValues)

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

type testModel struct {
	ApplicationMatcherFn func(app *model.Application) bool
	Webhooks             []*model.Webhook
	Apis                 []*model.APIDefinition
	EventAPIs            []*model.EventAPIDefinition
	Documents            []*model.Document
}

func modelFromInput(in model.ApplicationInput, applicationID string) testModel {
	applicationModelMatcherFn := func(app *model.Application) bool {
		return app.Name == in.Name && app.Description == in.Description
	}

	var webhooksModel []*model.Webhook
	for _, item := range in.Webhooks {
		webhooksModel = append(webhooksModel, item.ToWebhook(uuid.New().String(), applicationID))
	}

	var apisModel []*model.APIDefinition
	for _, item := range in.Apis {
		apisModel = append(apisModel, item.ToAPIDefinition(uuid.New().String(), applicationID))
	}

	var eventAPIsModel []*model.EventAPIDefinition
	for _, item := range in.EventAPIs {
		eventAPIsModel = append(eventAPIsModel, item.ToEventAPIDefinition(uuid.New().String(), applicationID))
	}

	var documentsModel []*model.Document
	for _, item := range in.Documents {
		documentsModel = append(documentsModel, item.ToDocument(uuid.New().String(), applicationID))
	}

	return testModel{
		ApplicationMatcherFn: applicationModelMatcherFn,
		Documents:            documentsModel,
		Apis:                 apisModel,
		EventAPIs:            eventAPIsModel,
		Webhooks:             webhooksModel,
	}
}
