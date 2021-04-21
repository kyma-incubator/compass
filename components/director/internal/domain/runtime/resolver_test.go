package runtime_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"

	"github.com/kyma-incubator/compass/components/director/pkg/resource"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

var contextParam = mock.MatchedBy(func(ctx context.Context) bool {
	persistenceOp, err := persistence.FromCtx(ctx)
	return err == nil && persistenceOp != nil
})

func TestResolver_CreateRuntime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Lorem ipsum")
	gqlRuntime := fixGQLRuntime(t, "foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter

		Input           graphql.RuntimeInput
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Create", contextParam, modelInput).Return("foo", nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			Input:           gqlInput,
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name: "Returns error when runtime creation failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Create", contextParam, modelInput).Return("", testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:           gqlInput,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Returns error when runtime retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Create", contextParam, modelInput).Return("foo", nil).Once()
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			Input:           gqlInput,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.RegisterRuntime(context.TODO(), testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_UpdateRuntime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Lorem ipsum")
	gqlRuntime := fixGQLRuntime(t, "foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.RuntimeInput{
		Name:        "Foo",
		Description: &desc,
	}
	runtimeID := "foo"

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		RuntimeID       string
		Input           graphql.RuntimeInput
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Update", contextParam, runtimeID, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			RuntimeID:       runtimeID,
			Input:           gqlInput,
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name: "Returns error when runtime update failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Update", contextParam, runtimeID, modelInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			RuntimeID:       runtimeID,
			Input:           gqlInput,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Returns error when runtime retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Update", contextParam, runtimeID, modelInput).Return(nil).Once()
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("InputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			RuntimeID:       runtimeID,
			Input:           gqlInput,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.UpdateRuntime(context.TODO(), testCase.RuntimeID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_DeleteRuntime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Bar")
	gqlRuntime := fixGQLRuntime(t, "foo", "Foo", "Bar")
	testErr := errors.New("Test error")
	scenariosNotFoundErr := apperrors.NewNotFoundError(resource.Label, "")
	scenarioAssignmentNotFoundErr := apperrors.NewNotFoundError(resource.AutomaticScenarioAssigment, "")
	txGen := txtest.NewTransactionContextGenerator(testErr)
	testAuths := fixOAuths()
	emptyScenariosLabel := &model.Label{Key: model.ScenariosKey, Value: []interface{}{}}
	singleScenarioLabel := &model.Label{Key: model.ScenariosKey, Value: []interface{}{"scenario-0"}}
	multiScenariosLabel := &model.Label{Key: model.ScenariosKey, Value: []interface{}{"scenario-0", "scenario-1", "scenario-2", "scenario-3"}}

	testCases := []struct {
		Name                    string
		TransactionerFn         func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn               func() *automock.RuntimeService
		ScenarioAssignmentFn    func() *automock.ScenarioAssignmentService
		SysAuthServiceFn        func() *automock.SystemAuthService
		OAuth20ServiceFn        func() *automock.OAuth20Service
		ConverterFn             func() *automock.RuntimeConverter
		BundleInstanceAuthSvcFn func() *automock.BundleInstanceAuthService
		InputID                 string
		ExpectedRuntime         *graphql.Runtime
		ExpectedErr             error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(nil, scenariosNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				auth := &model.BundleInstanceAuth{
					RuntimeID: &modelRuntime.ID,
					Status: &model.BundleInstanceAuthStatus{
						Condition: model.BundleInstanceAuthStatusConditionSucceeded,
					},
				}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{auth}, nil)
				svc.On("Update", contextParam, auth).Return(nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns error when runtime deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", contextParam, "foo").Return(testErr).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(nil, scenariosNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				auth := &model.BundleInstanceAuth{
					RuntimeID: &modelRuntime.ID,
					Status: &model.BundleInstanceAuthStatus{
						Condition: model.BundleInstanceAuthStatusConditionSucceeded,
					},
				}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{auth}, nil)
				svc.On("Update", contextParam, auth).Return(nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when runtime retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when transaction starting failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", contextParam, modelRuntime.ID).Return(nil)
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(nil, scenariosNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Return error when listing all system auths failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(nil, testErr)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Fails when cannot list bundle instance auths by runtime id",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return(nil, testErr)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Fails when cannot update bundle instance auths status to unused",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				auth := &model.BundleInstanceAuth{
					RuntimeID: &modelRuntime.ID,
					Status: &model.BundleInstanceAuthStatus{
						Condition: model.BundleInstanceAuthStatusConditionSucceeded,
					},
				}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{auth}, nil)
				svc.On("Update", contextParam, auth).Return(testErr)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Return error when removing oauth from hydra",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(nil, scenariosNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(testErr)
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns error when listing scenarios label",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(nil, testErr)
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns empty scenarios when listing scenarios label should succeed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(emptyScenariosLabel, nil)
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()

				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)

				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns scenario when listing scenarios label and error when listing scenario assignments should fail",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(singleScenarioLabel, nil)
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				scenarios, err := label.ValueToStringsSlice(singleScenarioLabel.Value)
				assert.NoError(t, err)

				svc.On("GetForScenarioName", contextParam, scenarios[0]).Return(model.AutomaticScenarioAssignment{}, testErr)
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns scenario when listing scenarios label and not found when listing scenario assignments should succeed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(singleScenarioLabel, nil)
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				scenarios, err := label.ValueToStringsSlice(singleScenarioLabel.Value)
				assert.NoError(t, err)
				svc.On("GetForScenarioName", contextParam, scenarios[0]).Return(model.AutomaticScenarioAssignment{}, scenarioAssignmentNotFoundErr)
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns scenario when listing scenarios label and scenario assignment when listing scenario assignments but fails on deletion of scenario assignment should fail",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(singleScenarioLabel, nil)
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				scenarios, err := label.ValueToStringsSlice(singleScenarioLabel.Value)
				assert.NoError(t, err)
				scenarioAssignment := model.AutomaticScenarioAssignment{}
				svc.On("GetForScenarioName", contextParam, scenarios[0]).Return(scenarioAssignment, nil)
				svc.On("Delete", contextParam, scenarioAssignment).Return(testErr)
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name:            "Returns scenario when listing scenarios label and scenario assignment when listing scenario assignments and succeeds on deletion of scenario assignment should succeed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(singleScenarioLabel, nil)
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				scenarios, err := label.ValueToStringsSlice(singleScenarioLabel.Value)
				assert.NoError(t, err)
				scenarioAssignment := model.AutomaticScenarioAssignment{}
				svc.On("GetForScenarioName", contextParam, scenarios[0]).Return(scenarioAssignment, nil)
				svc.On("Delete", contextParam, scenarioAssignment).Return(nil)
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns multiple scenarios when listing scenarios label and only some are created by a scenario assignment should succeed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(multiScenariosLabel, nil)
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				scenarios, err := label.ValueToStringsSlice(multiScenariosLabel.Value)
				assert.NoError(t, err)

				emptyAssignment := model.AutomaticScenarioAssignment{}
				scenarioAssignment1 := model.AutomaticScenarioAssignment{ScenarioName: scenarios[1]}
				scenarioAssignment2 := model.AutomaticScenarioAssignment{ScenarioName: scenarios[2]}

				svc.On("GetForScenarioName", contextParam, scenarios[0]).Return(emptyAssignment, scenarioAssignmentNotFoundErr)
				svc.On("GetForScenarioName", contextParam, scenarios[1]).Return(scenarioAssignment1, nil)
				svc.On("GetForScenarioName", contextParam, scenarios[2]).Return(scenarioAssignment2, nil)
				svc.On("GetForScenarioName", contextParam, scenarios[3]).Return(emptyAssignment, scenarioAssignmentNotFoundErr)

				svc.On("Delete", contextParam, scenarioAssignment1).Return(nil).Once()
				svc.On("Delete", contextParam, scenarioAssignment2).Return(nil).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns multiple scenarios when listing scenarios label and all are created by a scenario assignment should succeed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(multiScenariosLabel, nil)
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				scenarios, err := label.ValueToStringsSlice(multiScenariosLabel.Value)
				assert.NoError(t, err)

				scenarioAssignment0 := model.AutomaticScenarioAssignment{ScenarioName: scenarios[0]}
				scenarioAssignment1 := model.AutomaticScenarioAssignment{ScenarioName: scenarios[1]}
				scenarioAssignment2 := model.AutomaticScenarioAssignment{ScenarioName: scenarios[2]}
				scenarioAssignment3 := model.AutomaticScenarioAssignment{ScenarioName: scenarios[3]}

				svc.On("GetForScenarioName", contextParam, scenarios[0]).Return(scenarioAssignment0, nil)
				svc.On("GetForScenarioName", contextParam, scenarios[1]).Return(scenarioAssignment1, nil)
				svc.On("GetForScenarioName", contextParam, scenarios[2]).Return(scenarioAssignment2, nil)
				svc.On("GetForScenarioName", contextParam, scenarios[3]).Return(scenarioAssignment3, nil)

				svc.On("Delete", contextParam, scenarioAssignment0).Return(nil).Once()
				svc.On("Delete", contextParam, scenarioAssignment1).Return(nil).Once()
				svc.On("Delete", contextParam, scenarioAssignment2).Return(nil).Once()
				svc.On("Delete", contextParam, scenarioAssignment3).Return(nil).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name:            "Returns multiple scenarios when listing scenarios label and none are created by a scenario assignment should succeed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(multiScenariosLabel, nil)
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				scenarios, err := label.ValueToStringsSlice(multiScenariosLabel.Value)
				assert.NoError(t, err)

				emptyAssignment := model.AutomaticScenarioAssignment{}

				svc.On("GetForScenarioName", contextParam, scenarios[0]).Return(emptyAssignment, scenarioAssignmentNotFoundErr)
				svc.On("GetForScenarioName", contextParam, scenarios[1]).Return(emptyAssignment, scenarioAssignmentNotFoundErr)
				svc.On("GetForScenarioName", contextParam, scenarios[2]).Return(emptyAssignment, scenarioAssignmentNotFoundErr)
				svc.On("GetForScenarioName", contextParam, scenarios[3]).Return(emptyAssignment, scenarioAssignmentNotFoundErr)

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, model.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: func() *automock.OAuth20Service {
				svc := &automock.OAuth20Service{}
				svc.On("DeleteMultipleClientCredentials", contextParam, testAuths).Return(nil)
				return svc
			},
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			InputID:         "foo",
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
	}
	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			scenarioAssignmentSvc := testCase.ScenarioAssignmentFn()
			converter := testCase.ConverterFn()
			sysAuthSvc := testCase.SysAuthServiceFn()
			oAuth20Svc := testCase.OAuth20ServiceFn()
			bundleInstanceAuthSvc := testCase.BundleInstanceAuthSvcFn()

			resolver := runtime.NewResolver(transact, svc, scenarioAssignmentSvc, sysAuthSvc, oAuth20Svc, converter, nil, nil, bundleInstanceAuthSvc)

			// when
			result, err := resolver.DeleteRuntime(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			if testCase.ExpectedErr != nil {
				assert.EqualError(t, testCase.ExpectedErr, err.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, svc, scenarioAssignmentSvc, converter, transact, persistTx, sysAuthSvc, oAuth20Svc)
		})
	}
}

func TestResolver_Runtime(t *testing.T) {
	// given
	modelRuntime := fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Bar")
	gqlRuntime := fixGQLRuntime(t, "foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		InputID         string
		ExpectedRuntime *graphql.Runtime
		ExpectedErr     error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			InputID:         "foo",
			ExpectedRuntime: gqlRuntime,
			ExpectedErr:     nil,
		},
		{
			Name: "Success when runtime not found returns nil",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, apperrors.NewNotFoundError(resource.Runtime, "foo")).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     nil,
		},
		{
			Name: "Returns error when runtime retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputID:         "foo",
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.Runtime(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_Runtimes(t *testing.T) {
	// given
	modelRuntimes := []*model.Runtime{
		fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Lorem Ipsum"),
		fixModelRuntime(t, "bar", "tenant-bar", "Bar", "Lorem Ipsum"),
	}

	gqlRuntimes := []*graphql.Runtime{
		fixGQLRuntime(t, "foo", "Foo", "Lorem Ipsum"),
		fixGQLRuntime(t, "bar", "Bar", "Lorem Ipsum"),
	}

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"
	filter := []*labelfilter.LabelFilter{{Key: ""}}
	gqlFilter := []*graphql.LabelFilter{{Key: ""}}
	testErr := errors.New("Test error")

	testCases := []struct {
		Name              string
		PersistenceFn     func() *persistenceautomock.PersistenceTx
		TransactionerFn   func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn         func() *automock.RuntimeService
		ConverterFn       func() *automock.RuntimeConverter
		InputLabelFilters []*graphql.LabelFilter
		InputFirst        *int
		InputAfter        *graphql.PageCursor
		ExpectedResult    *graphql.RuntimePage
		ExpectedErr       error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("List", contextParam, filter, first, after).Return(fixRuntimePage(modelRuntimes), nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("MultipleToGraphQL", modelRuntimes).Return(gqlRuntimes).Once()
				return conv
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult:    fixGQLRuntimePage(gqlRuntimes),
			ExpectedErr:       nil,
		},
		{
			Name: "Returns error when runtime listing failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("List", contextParam, filter, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputFirst:        &first,
			InputAfter:        &gqlAfter,
			InputLabelFilters: gqlFilter,
			ExpectedResult:    nil,
			ExpectedErr:       testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.Runtimes(context.TODO(), testCase.InputLabelFilters, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_SetRuntimeLabel(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	runtimeID := "foo"
	labelKey := "key"
	labelValue := []string{"foo", "bar"}
	gqlLabel := &graphql.Label{
		Key:   labelKey,
		Value: labelValue,
	}
	modelLabelInput := &model.LabelInput{
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	modelLabel := &model.Label{
		ID:         "baz",
		Tenant:     "quaz",
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		InputRuntimeID  string
		InputKey        string
		InputValue      interface{}
		ExpectedLabel   *graphql.Label
		ExpectedErr     error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("SetLabel", contextParam, modelLabelInput).Return(nil).Once()
				svc.On("GetLabel", contextParam, runtimeID, modelLabelInput.Key).Return(modelLabel, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			InputValue:     gqlLabel.Value,
			ExpectedLabel:  gqlLabel,
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when adding label to runtime failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("SetLabel", contextParam, modelLabelInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			InputValue:     gqlLabel.Value,
			ExpectedLabel:  nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.SetRuntimeLabel(context.TODO(), testCase.InputRuntimeID, testCase.InputKey, testCase.InputValue)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}

	t.Run("Returns error when Label input validation failed", func(t *testing.T) {
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// when
		result, err := resolver.SetRuntimeLabel(context.TODO(), "", "", "")

		// then
		require.Nil(t, result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "value=cannot be blank")
		assert.Contains(t, err.Error(), "key=cannot be blank")
	})
}

func TestResolver_DeleteRuntimeLabel(t *testing.T) {
	// given
	testErr := errors.New("Test error")

	runtimeID := "foo"

	gqlLabel := &graphql.Label{
		Key:   "key",
		Value: []string{"foo", "bar"},
	}
	modelLabel := &model.Label{
		Key:   "key",
		Value: []string{"foo", "bar"},
	}

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		ConverterFn     func() *automock.RuntimeConverter
		InputRuntimeID  string
		InputKey        string
		ExpectedLabel   *graphql.Label
		ExpectedErr     error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, gqlLabel.Key).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", contextParam, runtimeID, gqlLabel.Key).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			ExpectedLabel:  gqlLabel,
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when label retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, gqlLabel.Key).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			ExpectedLabel:  nil,
			ExpectedErr:    testErr,
		},
		{
			Name: "Returns error when deleting runtime's label failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()

				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, gqlLabel.Key).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", contextParam, runtimeID, gqlLabel.Key).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				return conv
			},
			InputRuntimeID: runtimeID,
			InputKey:       gqlLabel.Key,
			ExpectedLabel:  nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil)

			// when
			result, err := resolver.DeleteRuntimeLabel(context.TODO(), testCase.InputRuntimeID, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_Labels(t *testing.T) {
	// given
	id := "foo"
	tenant := "tenant"
	labelKey := "key"
	labelValue := "val"

	gqlRuntime := fixGQLRuntime(t, id, "name", "desc")

	modelLabels := map[string]*model.Label{
		"abc": {
			ID:         "abc",
			Tenant:     tenant,
			Key:        labelKey,
			Value:      labelValue,
			ObjectID:   id,
			ObjectType: model.RuntimeLabelableObject,
		},
		"def": {
			ID:         "def",
			Tenant:     tenant,
			Key:        labelKey,
			Value:      labelValue,
			ObjectID:   id,
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	gqlLabels := graphql.Labels{
		labelKey: labelValue,
		labelKey: labelValue,
	}

	testErr := errors.New("Test error")

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		InputRuntime    *graphql.Runtime
		InputKey        string
		ExpectedResult  graphql.Labels
		ExpectedErr     error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListLabels", contextParam, id).Return(modelLabels, nil).Once()
				return svc
			},
			InputKey:       labelKey,
			ExpectedResult: gqlLabels,
			ExpectedErr:    nil,
		},
		{
			Name: "Success returns nil when labels not found",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListLabels", contextParam, id).Return(nil, errors.New("doesn't exist")).Once()
				return svc
			},
			InputKey:       labelKey,
			ExpectedResult: nil,
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when label listing failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListLabels", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			InputKey:       labelKey,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			svc := testCase.ServiceFn()
			transact := testCase.TransactionerFn(persistTx)

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil)

			// when
			result, err := resolver.Labels(context.TODO(), gqlRuntime, &testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, transact, persistTx)
		})
	}
}

func TestResolver_GetLabel(t *testing.T) {
	// given
	runtimeID := "37e89317-9ace-441d-9dc0-badf09b035b4"
	tenant := "tenant"
	labelKey := runtime.IsNormalizedLabel
	labelValue := "true"

	modelLabel := &model.Label{
		ID:         "abc",
		Tenant:     tenant,
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	gqlLabels := &graphql.Labels{
		labelKey: labelValue,
	}

	testErr := errors.New("Test error")

	testCases := []struct {
		Name            string
		PersistenceFn   func() *persistenceautomock.PersistenceTx
		TransactionerFn func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn       func() *automock.RuntimeService
		InputRuntime    *graphql.Runtime
		InputKey        string
		ExpectedResult  *graphql.Labels
		ExpectedErr     error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, labelKey).Return(modelLabel, nil).Once()
				return svc
			},
			InputKey:       labelKey,
			ExpectedResult: gqlLabels,
			ExpectedErr:    nil,
		},
		{
			Name: "Success returns nil when label not found",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, labelKey).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID)).Once()
				return svc
			},
			InputKey:       labelKey,
			ExpectedResult: nil,
			ExpectedErr:    nil,
		},
		{
			Name: "Returns error when label listing fails",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner {
				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return().Once()
				return transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, labelKey).Return(nil, testErr).Once()
				return svc
			},
			InputKey:       labelKey,
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			svc := testCase.ServiceFn()
			transact := testCase.TransactionerFn(persistTx)

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil)

			// when
			result, err := resolver.GetLabel(context.TODO(), runtimeID, labelKey)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, transact, persistTx)
		})
	}
}

func TestResolver_Auths(t *testing.T) {
	// GIVEN
	tnt := "tnt"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	parentRuntime := fixGQLRuntime(t, "foo", "bar", "baz")

	modelSysAuths := []model.SystemAuth{
		fixModelSystemAuth("bar", tnt, parentRuntime.ID, fixModelAuth()),
		fixModelSystemAuth("baz", tnt, parentRuntime.ID, fixModelAuth()),
		fixModelSystemAuth("faz", tnt, parentRuntime.ID, fixModelAuth()),
	}

	gqlSysAuths := []*graphql.RuntimeSystemAuth{
		fixGQLSystemAuth("bar", fixGQLAuth()),
		fixGQLSystemAuth("baz", fixGQLAuth()),
		fixGQLSystemAuth("faz", fixGQLAuth()),
	}

	testErr := errors.New("this is a test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		SysAuthSvcFn    func() *automock.SystemAuthService
		SysAuthConvFn   func() *automock.SystemAuthConverter
		ExpectedOutput  []*graphql.RuntimeSystemAuth
		ExpectedError   error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				sysAuthSvc.On("ListForObject", txtest.CtxWithDBMatcher(), model.RuntimeReference, parentRuntime.ID).Return(modelSysAuths, nil).Once()
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				sysAuthConv.On("ToGraphQL", &modelSysAuths[0]).Return(gqlSysAuths[0], nil).Once()
				sysAuthConv.On("ToGraphQL", &modelSysAuths[1]).Return(gqlSysAuths[1], nil).Once()
				sysAuthConv.On("ToGraphQL", &modelSysAuths[2]).Return(gqlSysAuths[2], nil).Once()
				return sysAuthConv
			},
			ExpectedOutput: gqlSysAuths,
			ExpectedError:  nil,
		},
		{
			Name:            "Error when listing for object",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				sysAuthSvc.On("ListForObject", txtest.CtxWithDBMatcher(), model.RuntimeReference, parentRuntime.ID).Return(nil, testErr).Once()
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				sysAuthSvc.On("ListForObject", txtest.CtxWithDBMatcher(), model.RuntimeReference, parentRuntime.ID).Return(modelSysAuths, nil).Once()
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			sysAuthSvc := testCase.SysAuthSvcFn()
			sysAuthConv := testCase.SysAuthConvFn()

			resolver := runtime.NewResolver(transact, nil, nil, sysAuthSvc, nil, nil, sysAuthConv, nil, nil)

			// WHEN
			result, err := resolver.Auths(ctx, parentRuntime)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, sysAuthSvc, sysAuthConv, transact, persist)
		})
	}

	t.Run("Error when parent object is nil", func(t *testing.T) {
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		result, err := resolver.Auths(context.TODO(), nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Runtime cannot be empty")
		assert.Nil(t, result)
	})
}

func TestResolver_EventingConfiguration(t *testing.T) {
	// GIVEN
	tnt := "tnt"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	runtimeID := uuid.New()
	gqlRuntime := fixGQLRuntime(t, runtimeID.String(), "bar", "baz")

	testErr := errors.New("this is a test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	defaultEveningURL := "https://eventing.domain.local"
	modelRuntimeEventingCfg := fixModelRuntimeEventingConfiguration(t, defaultEveningURL)
	gqlRuntimeEventingCfg := fixGQLRuntimeEventingConfiguration(defaultEveningURL)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		EventingSvcFn   func() *automock.EventingService
		ExpectedOutput  *graphql.RuntimeEventingConfiguration
		ExpectedError   error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("GetForRuntime", txtest.CtxWithDBMatcher(), runtimeID).Return(modelRuntimeEventingCfg, nil).Once()

				return eventingSvc
			},
			ExpectedOutput: gqlRuntimeEventingCfg,
			ExpectedError:  nil,
		}, {
			Name:            "Error when getting the configuration for runtime failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("GetForRuntime", txtest.CtxWithDBMatcher(), runtimeID).Return(nil, testErr).Once()

				return eventingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				return eventingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		}, {
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("GetForRuntime", txtest.CtxWithDBMatcher(), runtimeID).Return(modelRuntimeEventingCfg, nil).Once()

				return eventingSvc
			},
			ExpectedOutput: nil,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			eventingSvc := testCase.EventingSvcFn()

			resolver := runtime.NewResolver(transact, nil, nil, nil, nil, nil, nil, eventingSvc, nil)

			// WHEN
			result, err := resolver.EventingConfiguration(ctx, gqlRuntime)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)

			mock.AssertExpectationsForObjects(t, eventingSvc, transact, persist)
		})
	}

	t.Run("Error when parent object ID is not a valid UUID", func(t *testing.T) {
		// GIVEN
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		result, err := resolver.EventingConfiguration(ctx, &graphql.Runtime{ID: "abc"})

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while parsing runtime ID as UUID")
		assert.Nil(t, result)
	})

	t.Run("Error when parent object is nil", func(t *testing.T) {
		// GIVEN
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		result, err := resolver.EventingConfiguration(context.TODO(), nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Runtime cannot be empty")
		assert.Nil(t, result)
	})
}

func fixOAuths() []model.SystemAuth {
	return []model.SystemAuth{
		{
			ID:       "foo",
			TenantID: str.Ptr("foo"),
			Value: &model.Auth{
				Credential: model.CredentialData{
					Basic: nil,
					Oauth: &model.OAuthCredentialData{
						ClientID:     "foo",
						ClientSecret: "foo",
						URL:          "foo",
					},
				},
			},
		},
		{
			ID:       "bar",
			TenantID: str.Ptr("bar"),
			Value:    nil,
		},
		{
			ID:       "test",
			TenantID: str.Ptr("test"),
			Value: &model.Auth{
				Credential: model.CredentialData{
					Basic: &model.BasicCredentialData{
						Username: "test",
						Password: "test",
					},
					Oauth: nil,
				},
			},
		},
	}
}
