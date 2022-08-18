package systemfetcher_test

import (
	"context"
	"errors"
	"fmt"
	"io/fs"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher"
	"github.com/kyma-incubator/compass/components/director/internal/systemfetcher/automock"
	pAutomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	integrationSystemsDirectoryPath = "/data/int-systems/"
	tempFileName                    = "tmp.json"
)

func TestLoadData(t *testing.T) {
	testErr := errors.New("testErr")
	integrationSystemId := "sys-id"
	integrationSystemsJson := "[{" +
		"\"id\":\"" + integrationSystemId + "\"," +
		"\"name\":\"sys-name\"," +
		"\"description\":\"sys-desc\"" +
		"}]"

	applicationTemplateName := "app-tmpl-name"
	applicationTemplatesJson := "[{" +
		"\"name\":\"" + applicationTemplateName + "\"," +
		"\"description\":\"app-tmpl-desc\"}]"

	type testCase struct {
		name              string
		mockTransactioner func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner)
		appTmplSvc        func() *automock.AppTmplService
		intSysRepo        func() *automock.IntSysRepo
		readDirFunc       func(path string) ([]fs.FileInfo, error)
		readFileFunc      func(path string) ([]byte, error)
		expectedErr       error
	}
	tests := []testCase{
		{
			name: "load integration systems failed - read dir returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntStartTransaction()
			},
			appTmplSvc: mockAppTmplService,
			intSysRepo: mockIntSysRepository,
			readDirFunc: func(path string) ([]fs.FileInfo, error) {
				return nil, testErr
			},
			readFileFunc: mockReadFile,
			expectedErr:  testErr,
		},
		{
			name: "load integration systems failed - unsupported file type",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntStartTransaction()
			},
			appTmplSvc: mockAppTmplService,
			intSysRepo: mockIntSysRepository,
			readDirFunc: func(path string) ([]fs.FileInfo, error) {
				file := FakeFile{name: "test.txt"}
				return []fs.FileInfo{&file}, nil
			},
			readFileFunc: mockReadFile,
			expectedErr:  fmt.Errorf("unsupported file format \".txt\", supported format: json"),
		},
		{
			name: "load integration systems failed - read file returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntStartTransaction()
			},
			appTmplSvc:  mockAppTmplService,
			intSysRepo:  mockIntSysRepository,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte("[]"), testErr
			},
			expectedErr: testErr,
		},
		{
			name: "load application templates failed - read dir returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntStartTransaction()
			},
			appTmplSvc: mockAppTmplService,
			intSysRepo: mockIntSysRepository,
			readDirFunc: func(path string) ([]fs.FileInfo, error) {
				if path == integrationSystemsDirectoryPath {
					return mockReadDir("")
				}
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
			intSysRepo: mockIntSysRepository,
			readDirFunc: func(path string) ([]fs.FileInfo, error) {
				if path == integrationSystemsDirectoryPath {
					return mockReadDir("")
				}
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
			intSysRepo:  mockIntSysRepository,
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				if path == integrationSystemsDirectoryPath+tempFileName {
					return mockReadFile("")
				}
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
			intSysRepo:   mockIntSysRepository,
			readDirFunc:  mockReadDir,
			readFileFunc: mockReadFile,
			expectedErr:  testErr,
		},

		{
			name: "upsert integration systems failed - exists returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc: mockAppTmplService,
			intSysRepo: func() *automock.IntSysRepo {
				intSysRepo := &automock.IntSysRepo{}
				intSysRepo.On("Exists", txtest.CtxWithDBMatcher(), integrationSystemId).Return(false, testErr).Once()
				return intSysRepo
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(integrationSystemsJson), nil
			},
			expectedErr: testErr,
		},
		{
			name: "upsert integration systems failed - create returns error",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatDoesntExpectCommit()
			},
			appTmplSvc: mockAppTmplService,
			intSysRepo: func() *automock.IntSysRepo {
				intSysRepo := &automock.IntSysRepo{}
				intSysRepo.On("Exists", txtest.CtxWithDBMatcher(), integrationSystemId).Return(false, nil).Once()
				intSysRepo.On("Create", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.IntegrationSystem")).Return(testErr).Once()
				return intSysRepo
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				return []byte(integrationSystemsJson), nil
			},
			expectedErr: testErr,
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
			intSysRepo: func() *automock.IntSysRepo {
				intSysRepo := &automock.IntSysRepo{}
				intSysRepo.On("Exists", txtest.CtxWithDBMatcher(), integrationSystemId).Return(true, nil).Once()
				return intSysRepo
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				if path == integrationSystemsDirectoryPath+tempFileName {
					return []byte(integrationSystemsJson), nil
				}
				return []byte(applicationTemplatesJson), nil
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
			intSysRepo: func() *automock.IntSysRepo {
				intSysRepo := &automock.IntSysRepo{}
				intSysRepo.On("Exists", txtest.CtxWithDBMatcher(), integrationSystemId).Return(true, nil).Once()
				return intSysRepo
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				if path == integrationSystemsDirectoryPath+tempFileName {
					return []byte(integrationSystemsJson), nil
				}
				return []byte(applicationTemplatesJson), nil
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
			intSysRepo: func() *automock.IntSysRepo {
				intSysRepo := &automock.IntSysRepo{}
				intSysRepo.On("Exists", txtest.CtxWithDBMatcher(), integrationSystemId).Return(true, nil).Once()
				return intSysRepo
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				if path == integrationSystemsDirectoryPath+tempFileName {
					return []byte(integrationSystemsJson), nil
				}
				return []byte(applicationTemplatesJson), nil
			},
			expectedErr: testErr,
		},
		{
			name: "Success with empty array of integration systems and application templates",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
			},
			appTmplSvc:   mockAppTmplService,
			intSysRepo:   mockIntSysRepository,
			readDirFunc:  mockReadDir,
			readFileFunc: mockReadFile,
		},
		{
			name: "success",
			mockTransactioner: func() (*pAutomock.PersistenceTx, *pAutomock.Transactioner) {
				return txtest.NewTransactionContextGenerator(nil).ThatSucceeds()
			},
			appTmplSvc: func() *automock.AppTmplService {
				appTmplSvc := &automock.AppTmplService{}
				appTmplSvc.On("GetByNameAndRegion", txtest.CtxWithDBMatcher(), applicationTemplateName, nil).Return(nil, fmt.Errorf("Object not found")).Once()
				appTmplSvc.On("Create", txtest.CtxWithDBMatcher(), mock.AnythingOfType("model.ApplicationTemplateInput")).Return("", nil).Once()
				return appTmplSvc
			},
			intSysRepo: func() *automock.IntSysRepo {
				intSysRepo := &automock.IntSysRepo{}
				intSysRepo.On("Exists", txtest.CtxWithDBMatcher(), integrationSystemId).Return(true, nil).Once()
				return intSysRepo
			},
			readDirFunc: mockReadDir,
			readFileFunc: func(path string) ([]byte, error) {
				if path == integrationSystemsDirectoryPath+tempFileName {
					return []byte(integrationSystemsJson), nil
				}
				return []byte(applicationTemplatesJson), nil
			},
		},
	}

	for _, testCase := range tests {
		t.Run(testCase.name, func(t *testing.T) {
			appTmplSvc := testCase.appTmplSvc()
			inySysRepo := testCase.intSysRepo()
			mockedTx, transactioner := testCase.mockTransactioner()
			defer mock.AssertExpectationsForObjects(t, appTmplSvc, inySysRepo, mockedTx, transactioner)

			dataLoader := systemfetcher.NewDataLoader(transactioner, appTmplSvc, inySysRepo)
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

func mockIntSysRepository() *automock.IntSysRepo {
	return &automock.IntSysRepo{}
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
