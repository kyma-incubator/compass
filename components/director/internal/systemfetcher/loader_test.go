package systemfetcher_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"os"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
	pAutomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const tempFileName = "tmp.json"

func TestLoadData(t *testing.T) {
	testErr := errors.New("testErr")
	whID1 := "123456789"
	whID2 := "123456789-new"
	whID3 := "other"
	modeAsyncCallback := model.WebhookModeAsyncCallback

	applicationTemplateName := "app-tmpl-name"
	applicationTemplatesJSON := "[{" +
		"\"name\":\"" + applicationTemplateName + "\"," +
		"\"description\":\"app-tmpl-desc\"}]"
	applicationTemplatesWithIntSysJSON := "[{" +
		"\"name\":\"" + applicationTemplateName + "\"," +
		"\"applicationInputJSON\":\"{\\\"name\\\": \\\"name\\\", \\\"labels\\\": {\\\"legacy\\\": \\\"true\\\"}}\"," +
		"\"intSystem\":{" +
		"\"name\":\"int-sys-name\"," +
		"\"description\":\"int-sys-desc\"}," +
		"\"description\":\"app-tmpl-desc\"}]"

	applicationTemplatesWithIntSysJSONAndPlaceholders := "[{" +
		"\"name\":\"" + applicationTemplateName + "\"," +
		"\"applicationInputJSON\":\"{\\\"name\\\": \\\"name\\\", \\\"labels\\\": {\\\"legacy\\\":\\\"true\\\"}}\"," +
		"\"placeholders\":[{\"name\":\"name\",\"description\": \"description\",\"jsonPath\": \"jsonPath\"}]," +
		"\"labels\":{\"managed_app_provisioning\":false}," +
		"\"intSystem\":{" +
		"\"name\":\"int-sys-name\"," +
		"\"description\":\"int-sys-desc\"}," +
		"\"description\":\"app-tmpl-desc\"}]"

	applicationTemplatesWithIntSysJSONAndWebhooks := "[{" +
		"\"name\":\"" + applicationTemplateName + "\"," +
		"\"webhooks\":[{\"id\":\"" + whID1 + "\", \"objectID\":\"" + whID1 + "\", \"objectType\": \"ApplicationWebhook\", \"type\": \"CONFIGURATION_CHANGED\", \"mode\": \"ASYNC_CALLBACK\" },{\"id\":\"" + whID2 + "\", \"objectID\":\"" + whID2 + "\", \"objectType\": \"ApplicationWebhook\", \"type\": \"UNREGISTER_APPLICATION\", \"mode\": \"ASYNC_CALLBACK\", \"auth\": {\"SecretRef\":{\"SecretName\": \"name\", \"SecretKey\": \"key\"}} }]," +
		"\"applicationInputJSON\":\"{\\\"name\\\": \\\"name\\\", \\\"labels\\\": {\\\"legacy\\\":\\\"true\\\"}}\"," +
		"\"placeholders\":[{\"name\":\"name\",\"description\": \"description\",\"jsonPath\": \"jsonPath\"}]," +
		"\"intSystem\":{" +
		"\"name\":\"int-sys-name\"," +
		"\"description\":\"int-sys-desc\"}," +
		"\"description\":\"app-tmpl-desc\"}]"

	pageInfo := &pagination.Page{
		StartCursor: "",
		EndCursor:   "",
		HasNextPage: false,
	}

	intSysPage := model.IntegrationSystemPage{
		Data: []*model.IntegrationSystem{
			{
				ID:          "id",
				Name:        "name",
				Description: str.Ptr("desc"),
			},
		},
		PageInfo:   pageInfo,
		TotalCount: 0,
	}

	type testCase struct {
		name              string
		mockTransactioner func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner)
		appTmplSvc        func() *automock.AppTmplService
		intSysSvc         func() *automock.IntSysSvc
		webhookSvc        func() *automock.WebhookService
		readDirFunc       func(path string) ([]os.DirEntry, error)
		readFileFunc      func(path string) ([]byte, error)
		expectedErr       error
	}
	tests := []testCase{
		{
			name: "load application templates failed - read dir returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntStartTransaction()
			},
			appTmplSvc: mockAppTmplService,
			intSysSvc:  mockIntSysService,
			webhookSvc: mockWebhookService,
			readDirFunc: func(path string) ([]os.DirEntry, error) {
				return nil, testErr
			},
			readFileFunc: mockReadFile,
			expectedErr:  testErr,
		},
		{
			name: "load application templates failed - unsupported file type",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntStartTransaction()
			},
			appTmplSvc: mockAppTmplService,
			intSysSvc:  mockIntSysService,
			webhookSvc: mockWebhookService,
			readDirFunc: func(path string) ([]os.DirEntry, error) {
				file := FakeFile{name: "test.txt"}
				return []os.DirEntry{&file}, nil
			},
			readFileFunc: mockReadFile,
			expectedErr:  fmt.Errorf("unsupported file format \".txt\", supported format: json"),
		},
		{
			name: "load application templates failed - read file returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntStartTransaction()
			},
			appTmplSvc:  mockAppTmplService,
			intSysSvc:   mockIntSysService,
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte("[]"), testErr
			},
			expectedErr: testErr,
		},
		{
			name: "begin transaction failed",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(testErr).ThatFailsOnBegin()
			},
			appTmplSvc:   mockAppTmplService,
			intSysSvc:    mockIntSysService,
			webhookSvc:   mockWebhookService,
			readDirFunc:  mockReadDir,
			readFileFunc: mockReadFile,
			expectedErr:  testErr,
		},
		{
			name: "upsert application templates failed - GetByNameAndRegion returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc: func() *automock.AppTmplService {
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(nil, testErr).Once()
				return appTmplSvc
			},
			intSysSvc:   mockIntSysService,
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesJSON), nil
			},
			expectedErr: testErr,
		},
		{
			name: "upsert application templates failed - create returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc: func() *automock.AppTmplService {
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(nil, fmt.Errorf("Object not found")).Once()
				appTmplSvc.On("Create", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationTemplateInput")).Return("", testErr).Once()
				return appTmplSvc
			},
			intSysSvc:   mockIntSysService,
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesJSON), nil
			},
			expectedErr: testErr,
		},
		{
			name: "commit returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(testErr).ThatFailsOnCommit()
			},
			appTmplSvc: func() *automock.AppTmplService {
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(nil, fmt.Errorf("Object not found")).Once()
				appTmplSvc.On("Create", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationTemplateInput")).Return("", nil).Once()
				return appTmplSvc
			},
			intSysSvc:   mockIntSysService,
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesJSON), nil
			},
			expectedErr: testErr,
		},
		{
			name: "create application templates dependent entities failed - invalid intSys json object",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc:  mockAppTmplService,
			intSysSvc:   mockIntSysService,
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				applicationTemplatesWithInvalidIntSysJSON := "[{" +
					"\"name\":\"" + applicationTemplateName + "\"," +
					"\"intSystem\":123," +
					"\"description\":\"app-tmpl-desc\"}]"
				return []byte(applicationTemplatesWithInvalidIntSysJSON), nil
			},
			expectedErr: errors.New("the type of the integration system is float64 instead of map[string]interface{}. map[]"),
		},
		{
			name: "extract integration system failed - invalid intSys name json field",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc:  mockAppTmplService,
			intSysSvc:   mockIntSysService,
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				applicationTemplatesWithInvalidIntSysNameJSON := "[{" +
					"\"name\":\"" + applicationTemplateName + "\"," +
					"\"intSystem\":{" +
					"\"name\":123," +
					"\"description\":\"int-sys-desc\"}," +
					"\"description\":\"app-tmpl-desc\"}]"
				return []byte(applicationTemplatesWithInvalidIntSysNameJSON), nil
			},
			expectedErr: errors.New("integration system name value must be string"),
		},
		{
			name: "extract integration system failed - invalid intSys description json field",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc:  mockAppTmplService,
			intSysSvc:   mockIntSysService,
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				applicationTemplatesWithInvalidIntSysDescJSON := "[{" +
					"\"name\":\"" + applicationTemplateName + "\"," +
					"\"intSystem\":{" +
					"\"name\":\"int-sys-name\"," +
					"\"description\":123}," +
					"\"description\":\"app-tmpl-desc\"}]"
				return []byte(applicationTemplatesWithInvalidIntSysDescJSON), nil
			},
			expectedErr: errors.New("integration system description value must be string"),
		},
		{
			name: "extract integration system failed - missing intSys name json field",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc:  mockAppTmplService,
			intSysSvc:   mockIntSysService,
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				applicationTemplatesWithMissingIntSysNameJSON := "[{" +
					"\"name\":\"" + applicationTemplateName + "\"," +
					"\"intSystem\":{" +
					"\"description\":123}," +
					"\"description\":\"app-tmpl-desc\"}]"
				return []byte(applicationTemplatesWithMissingIntSysNameJSON), nil
			},
			expectedErr: errors.New("integration system name is missing"),
		},
		{
			name: "extract integration system failed - missing intSys description json field",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc:  mockAppTmplService,
			intSysSvc:   mockIntSysService,
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				applicationTemplatesWithMissingIntSysDescJSON := "[{" +
					"\"name\":\"" + applicationTemplateName + "\"," +
					"\"intSystem\":{" +
					"\"name\":\"int-sys-name\"}," +
					"\"description\":\"app-tmpl-desc\"}]"
				return []byte(applicationTemplatesWithMissingIntSysDescJSON), nil
			},
			expectedErr: errors.New("integration system description is missing"),
		},
		{
			name: "list integration systems failed - list returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc: mockAppTmplService,
			intSysSvc: func() *automock.IntSysSvc {
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(model.IntegrationSystemPage{}, testErr).Once()
				return intSysSvc
			},
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSON), nil
			},
			expectedErr: testErr,
		},
		{
			name: "create app templates dependent entities failed - create integration system returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc: mockAppTmplService,
			intSysSvc: func() *automock.IntSysSvc {
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				intSysSvc.On("Create", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.IntegrationSystemInput")).Return("", testErr).Once()
				return intSysSvc
			},
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSON), nil
			},
			expectedErr: testErr,
		},
		{
			name: "upsert application template failed - update returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc: func() *automock.AppTmplService {
				template := model.ApplicationTemplate{
					ID:                   "id",
					ApplicationInputJSON: "{\"test\":\"test\"}",
				}
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(&template, nil).Once()
				appTmplSvc.On("Update", txtest.CtxWithDBMatcher(), "id", false, mock.AnythingOfType("model.ApplicationTemplateUpdateInput")).Return(testErr).Once()
				return appTmplSvc
			},
			intSysSvc: func() *automock.IntSysSvc {
				intSysPage := model.IntegrationSystemPage{
					Data: []*model.IntegrationSystem{
						{
							ID:          "id",
							Name:        "int-sys-name",
							Description: str.Ptr("int-sys-desc"),
						},
					},
					PageInfo:   pageInfo,
					TotalCount: 0,
				}
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				return intSysSvc
			},
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSON), nil
			},
			expectedErr: testErr,
		},
		{
			name: "Success",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
			},
			appTmplSvc: func() *automock.AppTmplService {
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(nil, fmt.Errorf("Object not found")).Once()
				appTmplSvc.On("Create", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationTemplateInput")).Return("", nil).Once()
				return appTmplSvc
			},
			intSysSvc: func() *automock.IntSysSvc {
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				intSysSvc.On("Create", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.IntegrationSystemInput")).Return("int-sys-id", nil).Once()
				return intSysSvc
			},
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSON), nil
			},
		},
		{
			name: "Success - integration system already exists",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
			},
			appTmplSvc: func() *automock.AppTmplService {
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(nil, fmt.Errorf("Object not found")).Once()
				appTmplSvc.On("Create", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationTemplateInput")).Return("", nil).Once()
				return appTmplSvc
			},
			intSysSvc: func() *automock.IntSysSvc {
				intSysPage := model.IntegrationSystemPage{
					Data: []*model.IntegrationSystem{
						{
							ID:          "id",
							Name:        "int-sys-name",
							Description: str.Ptr("int-sys-desc"),
						},
					},
					PageInfo:   pageInfo,
					TotalCount: 0,
				}
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				return intSysSvc
			},
			webhookSvc:  mockWebhookService,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSON), nil
			},
		},
		{
			name: "Success - application template already exists, update triggered",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
			},
			appTmplSvc: func() *automock.AppTmplService {
				template := model.ApplicationTemplate{
					ID:                   "id",
					ApplicationInputJSON: "{\"test\":\"test\"}",
				}
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(&template, nil).Once()
				appTmplSvc.On("Update", txtest.CtxWithDBMatcher(), "id", false, mock.AnythingOfType("model.ApplicationTemplateUpdateInput")).Return(nil).Once()
				return appTmplSvc
			},
			intSysSvc: func() *automock.IntSysSvc {
				intSysPage := model.IntegrationSystemPage{
					Data: []*model.IntegrationSystem{
						{
							ID:          "id",
							Name:        "int-sys-name",
							Description: str.Ptr("int-sys-desc"),
						},
					},
					PageInfo:   pageInfo,
					TotalCount: 0,
				}
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				return intSysSvc
			},
			webhookSvc: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), "id").Return(nil, nil)
				return svc
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSON), nil
			},
		},
		{
			name: "Success - application template already exists, update triggered when only labels are different",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
			},
			appTmplSvc: func() *automock.AppTmplService {
				template := model.ApplicationTemplate{
					ID:                   "id",
					Name:                 "app-tmpl-name",
					Description:          str.Ptr("app-tmpl-desc"),
					ApplicationInputJSON: "{\"integrationSystemID\":\"id\",\"labels\":{\"legacy\":\"true\"},\"name\":\"name\"}",
					Placeholders: []model.ApplicationTemplatePlaceholder{
						{
							Name:        "name",
							Description: str.Ptr("description"),
							JSONPath:    str.Ptr("jsonPath"),
						},
					},
					Labels: map[string]interface{}{
						"managed_app_provisioning": true,
					},
				}
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(&template, nil).Once()
				appTmplSvc.On("Update", txtest.CtxWithDBMatcher(), "id", false, mock.AnythingOfType("model.ApplicationTemplateUpdateInput")).Return(nil).Once()
				return appTmplSvc
			},
			intSysSvc: func() *automock.IntSysSvc {
				intSysPage := model.IntegrationSystemPage{
					Data: []*model.IntegrationSystem{
						{
							ID:          "id",
							Name:        "int-sys-name",
							Description: str.Ptr("int-sys-desc"),
						},
					},
					PageInfo:   pageInfo,
					TotalCount: 0,
				}
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				return intSysSvc
			},
			webhookSvc: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), "id").Return(nil, nil)
				return svc
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSONAndPlaceholders), nil
			},
		},
		{
			name: "Success - integration system already exists, missing labels in applicationInputJSON",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
			},
			appTmplSvc: func() *automock.AppTmplService {
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(nil, fmt.Errorf("Object not found")).Once()
				appTmplSvc.On("Create", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationTemplateInput")).Return("", nil).Once()
				return appTmplSvc
			},
			intSysSvc: func() *automock.IntSysSvc {
				intSysPage := model.IntegrationSystemPage{
					Data: []*model.IntegrationSystem{
						{
							ID:          "id",
							Name:        "int-sys-name",
							Description: str.Ptr("int-sys-desc"),
						},
					},
					PageInfo:   pageInfo,
					TotalCount: 0,
				}
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				return intSysSvc
			},
			webhookSvc: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), "id").Return(nil, nil)
				return svc
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				applicationTemplatesWithIntSysAndMissingLabelsJSON := "[{" +
					"\"name\":\"" + applicationTemplateName + "\"," +
					"\"applicationInputJSON\":\"{\\\"name\\\": \\\"name\\\"}\"," +
					"\"intSystem\":{" +
					"\"name\":\"int-sys-name\"," +
					"\"description\":\"int-sys-desc\"}," +
					"\"description\":\"app-tmpl-desc\"}]"
				return []byte(applicationTemplatesWithIntSysAndMissingLabelsJSON), nil
			},
		},
		{
			name: "Success - application template already exists, update triggered when only webhooks length is different",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
			},
			appTmplSvc: func() *automock.AppTmplService {
				template := model.ApplicationTemplate{
					ID:                   "id",
					Name:                 "app-tmpl-name",
					Description:          str.Ptr("app-tmpl-desc"),
					ApplicationInputJSON: "{\"integrationSystemID\":\"id\",\"labels\":{\"legacy\":\"true\"},\"name\":\"name\"}",
					Placeholders: []model.ApplicationTemplatePlaceholder{
						{
							Name:        "name",
							Description: str.Ptr("description"),
							JSONPath:    str.Ptr("jsonPath"),
						},
					},
					Labels: map[string]interface{}{
						"managed_app_provisioning": false,
					},
					Webhooks: []model.Webhook{
						fixWebhookModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged),
						fixWebhookModel(whID3, modeAsyncCallback, model.WebhookTypeOpenResourceDiscovery),
					},
				}
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(&template, nil).Once()
				return appTmplSvc
			},
			intSysSvc: func() *automock.IntSysSvc {
				intSysPage := model.IntegrationSystemPage{
					Data: []*model.IntegrationSystem{
						{
							ID:          "id",
							Name:        "int-sys-name",
							Description: str.Ptr("int-sys-desc"),
						},
					},
					PageInfo:   pageInfo,
					TotalCount: 0,
				}
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				return intSysSvc
			},
			webhookSvc: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				wh1 := fixWebhookModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged)
				wh2 := fixWebhookModel(whID3, modeAsyncCallback, model.WebhookTypeOpenResourceDiscovery)
				whs := []*model.Webhook{
					&wh1, &wh2,
				}
				whInputCreate := fixWebhookInputModel(whID2, modeAsyncCallback, model.WebhookTypeDeleteApplication)
				setWebhookAuth(&whInputCreate)
				whInputUpdate := fixWebhookInputModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged)
				svc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), "id").Return(whs, nil)
				svc.On("Create", txtest.CtxWithDBMatcher(), "id", whInputCreate, model.ApplicationTemplateWebhookReference).Return(whID2, nil)
				svc.On("Update", txtest.CtxWithDBMatcher(), whID1, whInputUpdate, model.ApplicationTemplateWebhookReference).Return(nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), whID3, model.ApplicationTemplateWebhookReference).Return(nil)
				return svc
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSONAndWebhooks), nil
			},
		},
		{
			name: "Success - application template already exists, update triggered when only webhooks internals are different",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
			},
			appTmplSvc: func() *automock.AppTmplService {
				template := model.ApplicationTemplate{
					ID:                   "id",
					Name:                 "app-tmpl-name",
					Description:          str.Ptr("app-tmpl-desc"),
					ApplicationInputJSON: "{\"integrationSystemID\":\"id\",\"labels\":{\"legacy\":\"true\"},\"name\":\"name\"}",
					Placeholders: []model.ApplicationTemplatePlaceholder{
						{
							Name:        "name",
							Description: str.Ptr("description"),
							JSONPath:    str.Ptr("jsonPath"),
						},
					},
					Labels: map[string]interface{}{
						"managed_app_provisioning": false,
					},
					Webhooks: []model.Webhook{
						fixWebhookModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged),
						fixWebhookModel(whID2, model.WebhookModeSync, model.WebhookTypeDeleteApplication),
					},
				}
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(&template, nil).Once()
				return appTmplSvc
			},
			intSysSvc: func() *automock.IntSysSvc {
				intSysPage := model.IntegrationSystemPage{
					Data: []*model.IntegrationSystem{
						{
							ID:          "id",
							Name:        "int-sys-name",
							Description: str.Ptr("int-sys-desc"),
						},
					},
					PageInfo:   pageInfo,
					TotalCount: 0,
				}
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				return intSysSvc
			},
			webhookSvc: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				wh1 := fixWebhookModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged)
				wh2 := fixWebhookModel(whID2, model.WebhookModeSync, model.WebhookTypeDeleteApplication)
				whs := []*model.Webhook{&wh1, &wh2}
				whInput1Update := fixWebhookInputModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged)
				whInput2Update := fixWebhookInputModel(whID2, modeAsyncCallback, model.WebhookTypeDeleteApplication)
				setWebhookAuth(&whInput2Update)
				svc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), "id").Return(whs, nil)
				svc.On("Update", txtest.CtxWithDBMatcher(), whID1, whInput1Update, model.ApplicationTemplateWebhookReference).Return(nil)
				svc.On("Update", txtest.CtxWithDBMatcher(), whID2, whInput2Update, model.ApplicationTemplateWebhookReference).Return(nil)
				return svc
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSONAndWebhooks), nil
			},
		},
		{
			name: "Fail - application template already exists, update triggered when only webhooks length is different but creating fails",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc: func() *automock.AppTmplService {
				template := model.ApplicationTemplate{
					ID:                   "id",
					Name:                 "app-tmpl-name",
					Description:          str.Ptr("app-tmpl-desc"),
					ApplicationInputJSON: "{\"integrationSystemID\":\"id\",\"labels\":{\"legacy\":\"true\"},\"name\":\"name\"}",
					Placeholders: []model.ApplicationTemplatePlaceholder{
						{
							Name:        "name",
							Description: str.Ptr("description"),
							JSONPath:    str.Ptr("jsonPath"),
						},
					},
					Labels: map[string]interface{}{
						"managed_app_provisioning": false,
					},
					Webhooks: []model.Webhook{
						fixWebhookModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged),
						fixWebhookModel(whID3, modeAsyncCallback, model.WebhookTypeOpenResourceDiscovery),
					},
				}
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(&template, nil).Once()
				return appTmplSvc
			},
			intSysSvc: func() *automock.IntSysSvc {
				intSysPage := model.IntegrationSystemPage{
					Data: []*model.IntegrationSystem{
						{
							ID:          "id",
							Name:        "int-sys-name",
							Description: str.Ptr("int-sys-desc"),
						},
					},
					PageInfo:   pageInfo,
					TotalCount: 0,
				}
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				return intSysSvc
			},
			webhookSvc: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				wh1 := fixWebhookModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged)
				wh2 := fixWebhookModel(whID3, modeAsyncCallback, model.WebhookTypeOpenResourceDiscovery)
				whs := []*model.Webhook{
					&wh1, &wh2,
				}
				whInputCreate := fixWebhookInputModel(whID2, modeAsyncCallback, model.WebhookTypeDeleteApplication)
				setWebhookAuth(&whInputCreate)
				whInputUpdate := fixWebhookInputModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged)
				svc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), "id").Return(whs, nil)
				svc.On("Create", txtest.CtxWithDBMatcher(), "id", whInputCreate, model.ApplicationTemplateWebhookReference).Return("", testErr)
				svc.On("Update", txtest.CtxWithDBMatcher(), whID1, whInputUpdate, model.ApplicationTemplateWebhookReference).Return(nil)
				return svc
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSONAndWebhooks), nil
			},
			expectedErr: testErr,
		},
		{
			name: "Fail - application template already exists, update triggered when only webhooks length is different but updating fails",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc: func() *automock.AppTmplService {
				template := model.ApplicationTemplate{
					ID:                   "id",
					Name:                 "app-tmpl-name",
					Description:          str.Ptr("app-tmpl-desc"),
					ApplicationInputJSON: "{\"integrationSystemID\":\"id\",\"labels\":{\"legacy\":\"true\"},\"name\":\"name\"}",
					Placeholders: []model.ApplicationTemplatePlaceholder{
						{
							Name:        "name",
							Description: str.Ptr("description"),
							JSONPath:    str.Ptr("jsonPath"),
						},
					},
					Labels: map[string]interface{}{
						"managed_app_provisioning": false,
					},
					Webhooks: []model.Webhook{
						fixWebhookModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged),
						fixWebhookModel(whID3, modeAsyncCallback, model.WebhookTypeOpenResourceDiscovery),
					},
				}
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(&template, nil).Once()
				return appTmplSvc
			},
			intSysSvc: func() *automock.IntSysSvc {
				intSysPage := model.IntegrationSystemPage{
					Data: []*model.IntegrationSystem{
						{
							ID:          "id",
							Name:        "int-sys-name",
							Description: str.Ptr("int-sys-desc"),
						},
					},
					PageInfo:   pageInfo,
					TotalCount: 0,
				}
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				return intSysSvc
			},
			webhookSvc: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				wh1 := fixWebhookModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged)
				wh2 := fixWebhookModel(whID3, modeAsyncCallback, model.WebhookTypeOpenResourceDiscovery)
				whs := []*model.Webhook{
					&wh1, &wh2,
				}
				whInputUpdate := fixWebhookInputModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged)
				svc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), "id").Return(whs, nil)
				svc.On("Update", txtest.CtxWithDBMatcher(), whID1, whInputUpdate, model.ApplicationTemplateWebhookReference).Return(testErr)
				return svc
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSONAndWebhooks), nil
			},
			expectedErr: testErr,
		},
		{
			name: "Fail - application template already exists, update triggered when only webhooks length is different but deleting fails",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc: func() *automock.AppTmplService {
				template := model.ApplicationTemplate{
					ID:                   "id",
					Name:                 "app-tmpl-name",
					Description:          str.Ptr("app-tmpl-desc"),
					ApplicationInputJSON: "{\"integrationSystemID\":\"id\",\"labels\":{\"legacy\":\"true\"},\"name\":\"name\"}",
					Placeholders: []model.ApplicationTemplatePlaceholder{
						{
							Name:        "name",
							Description: str.Ptr("description"),
							JSONPath:    str.Ptr("jsonPath"),
						},
					},
					Labels: map[string]interface{}{
						"managed_app_provisioning": false,
					},
					Webhooks: []model.Webhook{
						fixWebhookModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged),
						fixWebhookModel(whID3, modeAsyncCallback, model.WebhookTypeOpenResourceDiscovery),
					},
				}
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(&template, nil).Once()
				return appTmplSvc
			},
			intSysSvc: func() *automock.IntSysSvc {
				intSysPage := model.IntegrationSystemPage{
					Data: []*model.IntegrationSystem{
						{
							ID:          "id",
							Name:        "int-sys-name",
							Description: str.Ptr("int-sys-desc"),
						},
					},
					PageInfo:   pageInfo,
					TotalCount: 0,
				}
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				return intSysSvc
			},
			webhookSvc: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				wh1 := fixWebhookModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged)
				wh2 := fixWebhookModel(whID3, modeAsyncCallback, model.WebhookTypeOpenResourceDiscovery)
				whs := []*model.Webhook{
					&wh1, &wh2,
				}
				whInputCreate := fixWebhookInputModel(whID2, modeAsyncCallback, model.WebhookTypeDeleteApplication)
				setWebhookAuth(&whInputCreate)
				whInputUpdate := fixWebhookInputModel(whID1, modeAsyncCallback, model.WebhookTypeConfigurationChanged)
				svc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), "id").Return(whs, nil)
				svc.On("Create", txtest.CtxWithDBMatcher(), "id", whInputCreate, model.ApplicationTemplateWebhookReference).Return(whID2, nil)
				svc.On("Update", txtest.CtxWithDBMatcher(), whID1, whInputUpdate, model.ApplicationTemplateWebhookReference).Return(nil)
				svc.On("Delete", txtest.CtxWithDBMatcher(), whID3, model.ApplicationTemplateWebhookReference).Return(testErr)
				return svc
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSONAndWebhooks), nil
			},
			expectedErr: testErr,
		},
		{
			name: "Fail - application template already exists, update triggered when only labels are different but listing webhooks fails",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc: func() *automock.AppTmplService {
				template := model.ApplicationTemplate{
					ID:                   "id",
					Name:                 "app-tmpl-name",
					Description:          str.Ptr("app-tmpl-desc"),
					ApplicationInputJSON: "{\"integrationSystemID\":\"id\",\"labels\":{\"legacy\":\"true\"},\"name\":\"name\"}",
					Placeholders: []model.ApplicationTemplatePlaceholder{
						{
							Name:        "name",
							Description: str.Ptr("description"),
							JSONPath:    str.Ptr("jsonPath"),
						},
					},
					Labels: map[string]interface{}{
						"managed_app_provisioning": true,
					},
				}
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(&template, nil).Once()
				appTmplSvc.On("Update", txtest.CtxWithDBMatcher(), "id", false, mock.AnythingOfType("model.ApplicationTemplateUpdateInput")).Return(nil).Once()
				return appTmplSvc
			},
			intSysSvc: func() *automock.IntSysSvc {
				intSysPage := model.IntegrationSystemPage{
					Data: []*model.IntegrationSystem{
						{
							ID:          "id",
							Name:        "int-sys-name",
							Description: str.Ptr("int-sys-desc"),
						},
					},
					PageInfo:   pageInfo,
					TotalCount: 0,
				}
				intSysSvc := &automock.IntSysSvc{}
				intSysSvc.On("List", txtest.CtxWithDBMatcher(), 200, "").Return(intSysPage, nil).Once()
				return intSysSvc
			},
			webhookSvc: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListForApplicationTemplate", txtest.CtxWithDBMatcher(), "id").Return(nil, testErr)
				return svc
			},
			readDirFunc: mockReadDir,
			expectedErr: testErr,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSONAndPlaceholders), nil
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			appTmplSvc := testCase.appTmplSvc()
			intSysSvc := testCase.intSysSvc()
			webhookSvc := testCase.webhookSvc()
			mockedTx, transactioner := testCase.mockTransactioner()
			defer mock.AssertExpectationsForObjects(t, appTmplSvc, intSysSvc, mockedTx, transactioner)

			kubeClient, err := systemfetcher.NewKubernetesClient(context.TODO(), systemfetcher.KubeConfig{UseKubernetes: "false"})
			require.NoError(t, err)

			dataLoader := systemfetcher.NewDataLoader(transactioner, systemfetcher.Config{}, appTmplSvc, intSysSvc, webhookSvc, kubeClient)
			err = dataLoader.LoadData(context.TODO(), testCase.readDirFunc, testCase.readFileFunc)

			if testCase.expectedErr != nil {
				require.Contains(t, err.Error(), testCase.expectedErr.Error())
			} else {
				require.NoError(t, err)
			}
		})
	}
}

func mockReadFile(_ string) ([]byte, error) {
	return []byte("[]"), nil
}

func mockReadDir(_ string) ([]os.DirEntry, error) {
	file := FakeFile{name: tempFileName}
	return []os.DirEntry{&file}, nil
}

func mockAppTmplService() *automock.AppTmplService {
	return &automock.AppTmplService{}
}

func mockIntSysService() *automock.IntSysSvc {
	return &automock.IntSysSvc{}
}

func mockWebhookService() *automock.WebhookService {
	return &automock.WebhookService{}
}

type FakeFile struct {
	name string
}

func (f *FakeFile) Type() fs.FileMode {
	return 0
}

func (f *FakeFile) Info() (fs.FileInfo, error) {
	return nil, nil
}

func (f *FakeFile) Name() string {
	return f.name
}

func (f *FakeFile) IsDir() bool {
	return false
}
