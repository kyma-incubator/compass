package systemfetcher_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"testing"
	"time"

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
		readDirFunc       func(path string) ([]fs.FileInfo, error)
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
			readDirFunc: func(path string) ([]fs.FileInfo, error) {
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
			readDirFunc: func(path string) ([]fs.FileInfo, error) {
				file := FakeFile{name: "test.txt"}
				return []fs.FileInfo{&file}, nil
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
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSON), nil
			},
			expectedErr: testErr,
		},
		{
			name: "enrich with integration system id label failed - incorrect app input json labels type",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc: mockAppTmplService,
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
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				applicationTemplatesWithIncorrectLabelsTypeAndIntSysJSON := "[{" +
					"\"name\":\"" + applicationTemplateName + "\"," +
					"\"applicationInputJSON\":\"{\\\"name\\\": \\\"name\\\", \\\"labels\\\": \\\"\\\"}\"," +
					"\"intSystem\":{" +
					"\"name\":\"int-sys-name\"," +
					"\"description\":\"int-sys-desc\"}," +
					"\"description\":\"app-tmpl-desc\"}]"
				return []byte(applicationTemplatesWithIncorrectLabelsTypeAndIntSysJSON), nil
			},
			expectedErr: errors.New("app input json labels are type map[string]interface {} instead of map[string]interface{}"),
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
				appTmplSvc.On("Update", txtest.CtxWithDBMatcher(), "id", mock.AnythingOfType("model.ApplicationTemplateUpdateInput")).Return(testErr).Once()
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
				appTmplSvc.On("Update", txtest.CtxWithDBMatcher(), "id", mock.AnythingOfType("model.ApplicationTemplateUpdateInput")).Return(nil).Once()
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
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(applicationTemplatesWithIntSysJSON), nil
			},
		},
		{
			name: "Success - integration systeem already exists, missing labels in applicationInputJSON",
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
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			appTmplSvc := testCase.appTmplSvc()
			intSysSvc := testCase.intSysSvc()
			mockedTx, transactioner := testCase.mockTransactioner()
			defer mock.AssertExpectationsForObjects(t, appTmplSvc, intSysSvc, mockedTx, transactioner)

			dataLoader := systemfetcher.NewDataLoader(transactioner, appTmplSvc, intSysSvc)
			err := dataLoader.LoadData(context.TODO(), testCase.readDirFunc, testCase.readFileFunc)

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

func mockReadDir(_ string) ([]fs.FileInfo, error) {
	file := FakeFile{name: tempFileName}
	return []fs.FileInfo{&file}, nil
}

func mockAppTmplService() *automock.AppTmplService {
	return &automock.AppTmplService{}
}

func mockIntSysService() *automock.IntSysSvc {
	return &automock.IntSysSvc{}
}

type FakeFile struct {
	name string
}

func (f *FakeFile) Name() string {
	return f.name
}

func (f *FakeFile) Size() int64 {
	return 0
}

func (f *FakeFile) Mode() fs.FileMode {
	return 0
}

func (f *FakeFile) ModTime() time.Time {
	return time.Time{}
}

func (f *FakeFile) IsDir() bool {
	return false
}

func (f *FakeFile) Sys() any {
	return nil
}
