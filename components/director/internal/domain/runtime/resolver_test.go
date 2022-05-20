package runtime_test

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/scenarioassignment"

	dataloader "github.com/kyma-incubator/compass/components/director/internal/dataloaders"
	pkgmodel "github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/google/uuid"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/labelfilter"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/rtmtest"

	"github.com/kyma-incubator/compass/components/director/internal/domain/label"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/persistence"

	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/runtime"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
)

const RegionKey = "region"

var contextParam = mock.MatchedBy(func(ctx context.Context) bool {
	persistenceOp, err := persistence.FromCtx(ctx)
	return err == nil && persistenceOp != nil
})

func TestResolver_CreateRuntime(t *testing.T) {
	// GIVEN
	modelRuntime := fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Lorem ipsum")
	gqlRuntime := fixGQLRuntime(t, "foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	extSubaccountID := "extSubaccountID"
	desc := "Lorem ipsum"
	gqlInput := graphql.RuntimeRegisterInput{
		Name:        "Foo",
		Description: &desc,
		Labels:      graphql.Labels{RegionKey: "region"},
		Webhooks: []*graphql.WebhookInput{{
			Type: "test webhook",
		}},
	}
	gqlInputWithSubaccountLabel := graphql.RuntimeRegisterInput{
		Name:        "Foo",
		Description: &desc,
		Labels:      graphql.Labels{RegionKey: "region", scenarioassignment.SubaccountIDKey: extSubaccountID},
		Webhooks: []*graphql.WebhookInput{{
			Type: "test webhook",
		}},
	}
	gqlInputWithInvalidLabel := graphql.RuntimeRegisterInput{
		Name:        "Foo",
		Description: &desc,
		Labels:      graphql.Labels{RegionKey: "region", scenarioassignment.SubaccountIDKey: []string{"firstValue", "secondValue"}},
		Webhooks: []*graphql.WebhookInput{{
			Type: "test webhook",
		}},
	}
	modelInput := model.RuntimeRegisterInput{
		Name:        "Foo",
		Description: &desc,
		Webhooks: []*model.WebhookInput{{
			Type: "test webhook",
		}},
	}
	selfRegModelInput := model.RuntimeRegisterInput{
		Name:        "Foo",
		Description: &desc,
		Labels:      graphql.Labels{rtmtest.TestDistinguishLabel: "selfRegVal", RegionKey: "region"},
		Webhooks: []*model.WebhookInput{{
			Type: "test webhook",
		}},
	}
	labels := map[string]interface{}{"xsappnameCMPClone": "clone"}

	testCases := []struct {
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn        func() *automock.RuntimeService
		ConverterFn      func() *automock.RuntimeConverter
		SelfRegManagerFn func() *automock.SelfRegisterManager
		UUIDSvcFn        func() *automock.UidService
		TenantFetcherFn  func() *automock.TenantFetcher

		Input           graphql.RuntimeRegisterInput
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
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, testUUID).Return(modelRuntime, nil).Once()
				svc.On("CreateWithMandatoryLabels", contextParam, modelInput, testUUID, labels).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("RegisterInputFromGraphQL", gqlInput).Return(modelInput, nil).Once()
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			UUIDSvcFn: func() *automock.UidService {
				svc := &automock.UidService{}
				svc.On("Generate").Return(testUUID).Once()
				return svc
			},
			TenantFetcherFn: func() *automock.TenantFetcher {
				return &automock.TenantFetcher{}
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesPrepWithNoErrors(labels),
			Input:            gqlInput,
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
		},
		{
			Name: "Success with subaccount label",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, testUUID).Return(modelRuntime, nil).Once()
				svc.On("CreateWithMandatoryLabels", contextParam, modelInput, testUUID, labels).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("RegisterInputFromGraphQL", gqlInputWithSubaccountLabel).Return(modelInput, nil).Once()
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			UUIDSvcFn: func() *automock.UidService {
				svc := &automock.UidService{}
				svc.On("Generate").Return(testUUID).Once()
				return svc
			},
			TenantFetcherFn: func() *automock.TenantFetcher {
				svc := &automock.TenantFetcher{}
				svc.On("FetchOnDemand", extSubaccountID).Return(nil).Once()
				return svc
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesPrepWithNoErrors(labels),
			Input:            gqlInputWithSubaccountLabel,
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns error when runtime converting fails",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: txtest.NoopTransactioner,
			ServiceFn: func() *automock.RuntimeService {
				return &automock.RuntimeService{}
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("RegisterInputFromGraphQL", gqlInputWithInvalidLabel).Return(model.RuntimeRegisterInput{}, testErr).Once()
				return conv
			},
			UUIDSvcFn: func() *automock.UidService {
				return &automock.UidService{}
			},
			TenantFetcherFn: func() *automock.TenantFetcher {
				return &automock.TenantFetcher{}
			},
			SelfRegManagerFn: func() *automock.SelfRegisterManager {
				return &automock.SelfRegisterManager{}
			},
			Input:           gqlInputWithInvalidLabel,
			ExpectedRuntime: nil,
			ExpectedErr:     testErr,
		},
		{
			Name: "Returns error when subaccount id label has invalid value",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.NoopTransactioner,
			ServiceFn: func() *automock.RuntimeService {
				return &automock.RuntimeService{}
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("RegisterInputFromGraphQL", gqlInputWithInvalidLabel).Return(modelInput, nil).Once()
				return conv
			},
			UUIDSvcFn: func() *automock.UidService {
				svc := &automock.UidService{}
				svc.On("Generate").Return(testUUID).Once()
				return svc
			},
			TenantFetcherFn: func() *automock.TenantFetcher {
				return &automock.TenantFetcher{}
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesPrepWithNoErrors(labels),
			Input:            gqlInputWithInvalidLabel,
			ExpectedRuntime:  nil,
			ExpectedErr:      errors.New("while converting global_subaccount_id label: cannot cast label value"),
		},
		{
			Name: "Returns error when cannot fetch tenant details from tenant provider",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.NoopTransactioner,
			ServiceFn: func() *automock.RuntimeService {
				return &automock.RuntimeService{}
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("RegisterInputFromGraphQL", gqlInputWithSubaccountLabel).Return(modelInput, nil).Once()
				return conv
			},
			UUIDSvcFn: func() *automock.UidService {
				svc := &automock.UidService{}
				svc.On("Generate").Return(testUUID).Once()
				return svc
			},
			TenantFetcherFn: func() *automock.TenantFetcher {
				svc := &automock.TenantFetcher{}
				svc.On("FetchOnDemand", extSubaccountID).Return(testErr).Once()
				return svc
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesPrepWithNoErrors(labels),
			Input:            gqlInputWithSubaccountLabel,
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when runtime creation failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("CreateWithMandatoryLabels", contextParam, modelInput, testUUID, labels).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("RegisterInputFromGraphQL", gqlInput).Return(modelInput, nil).Once()
				return conv
			},
			UUIDSvcFn: func() *automock.UidService {
				svc := &automock.UidService{}
				svc.On("Generate").Return(testUUID).Once()
				return svc
			},
			TenantFetcherFn: func() *automock.TenantFetcher {
				return &automock.TenantFetcher{}
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesNotCleanupFunc(labels),
			Input:            gqlInput,
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when runtime self registration fails",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("CreateWithMandatoryLabels", contextParam, selfRegModelInput, testUUID, labels).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("RegisterInputFromGraphQL", gqlInput).Return(selfRegModelInput, nil).Once()
				return conv
			},
			UUIDSvcFn: func() *automock.UidService {
				svc := &automock.UidService{}
				svc.On("Generate").Return(testUUID).Once()
				return svc
			},
			TenantFetcherFn: func() *automock.TenantFetcher {
				return &automock.TenantFetcher{}
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatReturnsNoErrors(labels),
			Input:            gqlInput,
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when runtime cleanup fails",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, testUUID).Return(nil, testErr).Once()
				svc.On("CreateWithMandatoryLabels", contextParam, selfRegModelInput, testUUID, labels).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("RegisterInputFromGraphQL", gqlInput).Return(selfRegModelInput, nil).Once()
				return conv
			},
			UUIDSvcFn: func() *automock.UidService {
				svc := &automock.UidService{}
				svc.On("Generate").Return(testUUID).Once()
				return svc
			},
			TenantFetcherFn: func() *automock.TenantFetcher {
				return &automock.TenantFetcher{}
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatFailsOnCleanup(labels),
			Input:            gqlInput,
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when runtime retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("CreateWithMandatoryLabels", contextParam, modelInput, testUUID, labels).Return(nil).Once()
				svc.On("Get", contextParam, testUUID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("RegisterInputFromGraphQL", gqlInput).Return(modelInput, nil).Once()
				return conv
			},
			UUIDSvcFn: func() *automock.UidService {
				svc := &automock.UidService{}
				svc.On("Generate").Return(testUUID).Once()
				return svc
			},
			TenantFetcherFn: func() *automock.TenantFetcher {
				return &automock.TenantFetcher{}
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesNotCleanupFunc(labels),
			Input:            gqlInput,
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when self registered runtime retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("CreateWithMandatoryLabels", contextParam, selfRegModelInput, testUUID, labels).Return(nil).Once()
				svc.On("Get", contextParam, testUUID).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("RegisterInputFromGraphQL", gqlInput).Return(selfRegModelInput, nil).Once()
				return conv
			},
			UUIDSvcFn: func() *automock.UidService {
				svc := &automock.UidService{}
				svc.On("Generate").Return(testUUID).Once()
				return svc
			},
			TenantFetcherFn: func() *automock.TenantFetcher {
				return &automock.TenantFetcher{}
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatReturnsNoErrors(labels),
			Input:            gqlInput,
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when preparation for self-registration failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				return &persistenceautomock.PersistenceTx{}
			},
			TransactionerFn: txtest.NoopTransactioner,
			ServiceFn: func() *automock.RuntimeService {
				return &automock.RuntimeService{}
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("RegisterInputFromGraphQL", gqlInput).Return(modelInput, nil).Once()
				return conv
			},
			UUIDSvcFn: func() *automock.UidService {
				svc := &automock.UidService{}
				svc.On("Generate").Return(testUUID).Once()
				return svc
			},
			TenantFetcherFn: func() *automock.TenantFetcher {
				return &automock.TenantFetcher{}
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatReturnsErrorOnPrep,
			Input:            gqlInput,
			ExpectedRuntime:  nil,
			ExpectedErr:      errors.New(rtmtest.SelfRegErrorMsg),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			selfRegManager := testCase.SelfRegManagerFn()
			uuidSvc := testCase.UUIDSvcFn()
			fetcher := testCase.TenantFetcherFn()

			defer mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx, selfRegManager, fetcher)

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil, selfRegManager, uuidSvc, nil, nil, nil, nil, nil, fetcher, nil)

			// WHEN
			result, err := resolver.RegisterRuntime(context.TODO(), testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
		})
	}
}

func TestResolver_UpdateRuntime(t *testing.T) {
	// GIVEN
	modelRuntime := fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Lorem ipsum")
	gqlRuntime := fixGQLRuntime(t, "foo", "Foo", "Lorem ipsum")
	testErr := errors.New("Test error")

	desc := "Lorem ipsum"
	gqlInput := graphql.RuntimeUpdateInput{
		Name:        "Foo",
		Description: &desc,
	}
	modelInput := model.RuntimeUpdateInput{
		Name:        "Foo",
		Description: &desc,
	}
	runtimeID := "foo"
	emptyLabels := make(map[string]interface{})

	testCases := []struct {
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn        func() *automock.RuntimeService
		ConverterFn      func() *automock.RuntimeConverter
		SelfRegManagerFn func() *automock.SelfRegisterManager
		RuntimeID        string
		Input            graphql.RuntimeUpdateInput
		ExpectedRuntime  *graphql.Runtime
		ExpectedErr      error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Update", contextParam, runtimeID, modelInput).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("UpdateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatReturnsNoErrors(emptyLabels),
			RuntimeID:        runtimeID,
			Input:            gqlInput,
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns error when runtime update failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Update", contextParam, runtimeID, modelInput).Return(testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("UpdateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatReturnsNoErrors(emptyLabels),
			RuntimeID:        runtimeID,
			Input:            gqlInput,
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when runtime retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Update", contextParam, runtimeID, modelInput).Return(nil).Once()
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("UpdateInputFromGraphQL", gqlInput).Return(modelInput).Once()
				return conv
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatReturnsNoErrors(emptyLabels),
			RuntimeID:        runtimeID,
			Input:            gqlInput,
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			selfRegMng := testCase.SelfRegManagerFn()
			uuidSvc := &automock.UidService{}

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil, selfRegMng, uuidSvc, nil, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.UpdateRuntime(context.TODO(), testCase.RuntimeID, testCase.Input)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_DeleteRuntime(t *testing.T) {
	// GIVEN
	modelRuntime := fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Bar")
	gqlRuntime := fixGQLRuntime(t, "foo", "Foo", "Bar")
	testErr := errors.New("Test error")
	labelNotFoundErr := apperrors.NewNotFoundError(resource.Label, "")
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
		SelfRegManagerFn        func() *automock.SelfRegisterManager
		FormationsSvcFn         func() *automock.FormationService
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
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(nil, labelNotFoundErr).Once()
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
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
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesNotCleanup,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
		},
		{
			Name: "Success for self registered runtime",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Times(2)

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Times(2)
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

				return persistTx, transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(nil, labelNotFoundErr).Once()
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, nil).Once()
				svc.On("GetLabel", contextParam, "foo", RegionKey).Return(&model.Label{Value: testRegion}, nil).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
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
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesCleanupWithNoErrors,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns error when second transaction fails",
			TransactionerFn: func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTxSecond := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				persistTxSecond.On("Commit").Return(testErr).Once()

				transact := &persistenceautomock.Transactioner{}
				transact.On("Begin").Return(persistTx, nil).Once()
				transact.On("Begin").Return(persistTxSecond, nil).Once()
				transact.On("RollbackUnlessCommitted", mock.Anything, persistTxSecond).Return(false).Once()

				return persistTx, transact
			},
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(nil, labelNotFoundErr).Once()
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, nil).Once()
				svc.On("GetLabel", contextParam, "foo", RegionKey).Return(&model.Label{Value: testRegion}, nil).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
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
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesCleanupWithNoErrors,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when self-registration had been done but cleanup fails",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, nil).Once()
				svc.On("GetLabel", contextParam, "foo", RegionKey).Return(&model.Label{Value: testRegion}, nil).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn:          UnusedRuntimeConverter(),
			SysAuthServiceFn:     UnusedSysAuthService(),
			OAuth20ServiceFn:     UnusedOauth20Service(),
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
			SelfRegManagerFn: rtmtest.SelfRegManagerThatReturnsErrorOnCleanup,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      errors.New("An error occurred during cleanup of self-registered runtime"),
		},
		{
			Name:            "Returns error when region label retrieval fails",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, nil).Once()
				svc.On("GetLabel", contextParam, "foo", RegionKey).Return(nil, testErr).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn:          UnusedRuntimeConverter(),
			SysAuthServiceFn:     UnusedSysAuthService(),
			OAuth20ServiceFn:     UnusedOauth20Service(),
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				auth := &model.BundleInstanceAuth{
					RuntimeID: &modelRuntime.ID,
					Status: &model.BundleInstanceAuthStatus{
						Condition: model.BundleInstanceAuthStatusConditionSucceeded,
					},
				}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{auth}, nil)
				return svc
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesNotCleanup,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when runtime deletion failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", contextParam, "foo").Return(testErr).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(nil, labelNotFoundErr).Once()
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: UnusedOauth20Service(),
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
			SelfRegManagerFn: rtmtest.SelfRegManagerReturnsDistinguishingLabel,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when runtime retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()
				return svc
			},
			ScenarioAssignmentFn:    UnusedScenarioAssignmentService(),
			ConverterFn:             UnusedRuntimeConverter(),
			SysAuthServiceFn:        UnusedSysAuthService(),
			OAuth20ServiceFn:        UnusedOauth20Service(),
			BundleInstanceAuthSvcFn: UnusedBundleInstanceAuthService(),
			FormationsSvcFn:         UnusedFormationService(),
			SelfRegManagerFn:        rtmtest.NoopSelfRegManager,
			InputID:                 "foo",
			ExpectedRuntime:         nil,
			ExpectedErr:             testErr,
		},
		{
			Name:            "Returns error when transaction starting failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				return svc
			},
			ScenarioAssignmentFn:    UnusedScenarioAssignmentService(),
			ConverterFn:             UnusedRuntimeConverter(),
			SysAuthServiceFn:        UnusedSysAuthService(),
			OAuth20ServiceFn:        UnusedOauth20Service(),
			BundleInstanceAuthSvcFn: UnusedBundleInstanceAuthService(),
			FormationsSvcFn:         UnusedFormationService(),
			SelfRegManagerFn:        rtmtest.NoopSelfRegManager,
			InputID:                 "foo",
			ExpectedRuntime:         nil,
			ExpectedErr:             testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", contextParam, modelRuntime.ID).Return(nil)
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(nil, labelNotFoundErr).Once()
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
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
			SelfRegManagerFn: rtmtest.SelfRegManagerReturnsDistinguishingLabel,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Return error when listing all system auths failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn:          UnusedRuntimeConverter(),
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(nil, testErr)
				return svc
			},
			OAuth20ServiceFn: UnusedOauth20Service(),
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerReturnsDistinguishingLabel,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Fails when cannot list bundle instance auths by runtime id",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn:          UnusedRuntimeConverter(),
			SysAuthServiceFn:     UnusedSysAuthService(),
			OAuth20ServiceFn:     UnusedOauth20Service(),
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return(nil, testErr)
				return svc
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Fails when cannot list self-register distinguishing label because of error which is other than not found",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, testErr).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn:          UnusedRuntimeConverter(),
			SysAuthServiceFn:     UnusedSysAuthService(),
			OAuth20ServiceFn:     UnusedOauth20Service(),
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerReturnsDistinguishingLabel,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Fails when cannot update bundle instance auths status to unused",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn:          UnusedRuntimeConverter(),
			SysAuthServiceFn:     UnusedSysAuthService(),
			OAuth20ServiceFn:     UnusedOauth20Service(),
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
			SelfRegManagerFn: rtmtest.SelfRegManagerReturnsDistinguishingLabel,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Return error when removing oauth from hydra",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(nil, labelNotFoundErr).Once()
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
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
			SelfRegManagerFn: rtmtest.SelfRegManagerReturnsDistinguishingLabel,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns error when listing scenarios label",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(nil, testErr)
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn:          UnusedRuntimeConverter(),
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: UnusedOauth20Service(),
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerReturnsDistinguishingLabel,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns empty scenarios when listing scenarios label should succeed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(emptyScenariosLabel, nil)
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				return svc
			},
			ScenarioAssignmentFn: UnusedScenarioAssignmentService(),
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()

				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
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
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesNotCleanup,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
		},
		{
			Name:            "Returns scenario when listing scenarios label and error when listing scenario assignments should fail",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(singleScenarioLabel, nil)
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				scenarios, err := label.ValueToStringsSlice(singleScenarioLabel.Value)
				assert.NoError(t, err)

				svc.On("GetForScenarioName", contextParam, scenarios[0]).Return(model.AutomaticScenarioAssignment{}, testErr)
				return svc
			},
			ConverterFn: UnusedRuntimeConverter(),
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: UnusedOauth20Service(),
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerReturnsDistinguishingLabel,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns scenario when listing scenarios label and not found when listing scenario assignments should succeed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(singleScenarioLabel, nil)
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
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
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
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
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesNotCleanup,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
		},
		{
			Name:            "Returns scenario when listing scenarios label and scenario assignment when listing scenario assignments but fails on deletion of scenario assignment should fail",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(singleScenarioLabel, nil)
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				scenarios, err := label.ValueToStringsSlice(singleScenarioLabel.Value)
				assert.NoError(t, err)
				scenarioAssignment := model.AutomaticScenarioAssignment{}
				svc.On("GetForScenarioName", contextParam, scenarios[0]).Return(scenarioAssignment, nil)
				return svc
			},
			FormationsSvcFn: func() *automock.FormationService {
				svc := &automock.FormationService{}

				scenarioAssignment := model.AutomaticScenarioAssignment{}
				svc.On("DeleteAutomaticScenarioAssignment", contextParam, scenarioAssignment).Return(testErr)
				return svc
			},
			ConverterFn: UnusedRuntimeConverter(),
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
				return svc
			},
			OAuth20ServiceFn: UnusedOauth20Service(),
			BundleInstanceAuthSvcFn: func() *automock.BundleInstanceAuthService {
				svc := &automock.BundleInstanceAuthService{}
				svc.On("ListByRuntimeID", contextParam, modelRuntime.ID).Return([]*model.BundleInstanceAuth{}, nil)
				return svc
			},
			SelfRegManagerFn: rtmtest.SelfRegManagerReturnsDistinguishingLabel,
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
		{
			Name:            "Returns scenario when listing scenarios label and scenario assignment when listing scenario assignments and succeeds on deletion of scenario assignment should succeed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(singleScenarioLabel, nil)
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
				svc.On("Delete", contextParam, "foo").Return(nil).Once()
				return svc
			},
			ScenarioAssignmentFn: func() *automock.ScenarioAssignmentService {
				svc := &automock.ScenarioAssignmentService{}
				scenarios, err := label.ValueToStringsSlice(singleScenarioLabel.Value)
				assert.NoError(t, err)
				scenarioAssignment := model.AutomaticScenarioAssignment{}
				svc.On("GetForScenarioName", contextParam, scenarios[0]).Return(scenarioAssignment, nil)
				return svc
			},
			FormationsSvcFn: func() *automock.FormationService {
				svc := &automock.FormationService{}

				scenarioAssignment := model.AutomaticScenarioAssignment{}
				svc.On("DeleteAutomaticScenarioAssignment", contextParam, scenarioAssignment).Return(nil)
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
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
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesNotCleanup,
			InputID:          "foo",
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
		},
		{
			Name:            "Returns multiple scenarios when listing scenarios label and only some are created by a scenario assignment should succeed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(multiScenariosLabel, nil)
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
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

				return svc
			},
			FormationsSvcFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				scenarios, err := label.ValueToStringsSlice(multiScenariosLabel.Value)
				assert.NoError(t, err)

				scenarioAssignment1 := model.AutomaticScenarioAssignment{ScenarioName: scenarios[1]}
				scenarioAssignment2 := model.AutomaticScenarioAssignment{ScenarioName: scenarios[2]}
				svc.On("DeleteAutomaticScenarioAssignment", contextParam, scenarioAssignment1).Return(nil).Once()
				svc.On("DeleteAutomaticScenarioAssignment", contextParam, scenarioAssignment2).Return(nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
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
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesNotCleanup,
			InputID:          "foo",
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
		},
		{
			Name:            "Returns multiple scenarios when listing scenarios label and all are created by a scenario assignment should succeed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(multiScenariosLabel, nil)
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
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

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SysAuthServiceFn: func() *automock.SystemAuthService {
				svc := &automock.SystemAuthService{}
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
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
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesNotCleanup,
			FormationsSvcFn: func() *automock.FormationService {
				svc := &automock.FormationService{}
				scenarios, err := label.ValueToStringsSlice(multiScenariosLabel.Value)
				assert.NoError(t, err)

				scenarioAssignment0 := model.AutomaticScenarioAssignment{ScenarioName: scenarios[0]}
				scenarioAssignment1 := model.AutomaticScenarioAssignment{ScenarioName: scenarios[1]}
				scenarioAssignment2 := model.AutomaticScenarioAssignment{ScenarioName: scenarios[2]}
				scenarioAssignment3 := model.AutomaticScenarioAssignment{ScenarioName: scenarios[3]}

				svc.On("DeleteAutomaticScenarioAssignment", contextParam, scenarioAssignment0).Return(nil).Once()
				svc.On("DeleteAutomaticScenarioAssignment", contextParam, scenarioAssignment1).Return(nil).Once()
				svc.On("DeleteAutomaticScenarioAssignment", contextParam, scenarioAssignment2).Return(nil).Once()
				svc.On("DeleteAutomaticScenarioAssignment", contextParam, scenarioAssignment3).Return(nil).Once()

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
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
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
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
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
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesNotCleanup,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
		},
		{
			Name:            "Returns multiple scenarios when listing scenarios label and none are created by a scenario assignment should succeed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, nil).Once()
				svc.On("GetLabel", contextParam, "foo", model.ScenariosKey).Return(multiScenariosLabel, nil)
				svc.On("GetLabel", contextParam, "foo", rtmtest.TestDistinguishLabel).Return(nil, labelNotFoundErr).Once()
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
				svc.On("ListForObject", contextParam, pkgmodel.RuntimeReference, modelRuntime.ID).Return(testAuths, nil)
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
			SelfRegManagerFn: rtmtest.SelfRegManagerThatDoesNotCleanup,
			FormationsSvcFn:  UnusedFormationService(),
			InputID:          "foo",
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
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
			selfRegisterManager := testCase.SelfRegManagerFn()
			formationSvc := testCase.FormationsSvcFn()
			uuidSvc := &automock.UidService{}

			resolver := runtime.NewResolver(transact, svc, scenarioAssignmentSvc, sysAuthSvc, oAuth20Svc, converter, nil, nil, bundleInstanceAuthSvc, selfRegisterManager, uuidSvc, nil, nil, nil, nil, nil, nil, formationSvc)

			// WHEN
			result, err := resolver.DeleteRuntime(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			if testCase.ExpectedErr != nil {
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}

			mock.AssertExpectationsForObjects(t, svc, scenarioAssignmentSvc, converter, transact, persistTx, sysAuthSvc, oAuth20Svc, selfRegisterManager)
		})
	}
}

func TestResolver_Runtime(t *testing.T) {
	// GIVEN
	modelRuntime := fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Bar")
	gqlRuntime := fixGQLRuntime(t, "foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn        func() *automock.RuntimeService
		ConverterFn      func() *automock.RuntimeConverter
		SelfRegManagerFn func() *automock.SelfRegisterManager
		InputID          string
		ExpectedRuntime  *graphql.Runtime
		ExpectedErr      error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
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
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputID:          "foo",
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
		},
		{
			Name: "Success when runtime not found returns nil",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(modelRuntime, apperrors.NewNotFoundError(resource.Runtime, "foo")).Once()

				return svc
			},
			ConverterFn:      UnusedRuntimeConverter(),
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns error when runtime retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("Get", contextParam, "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn:      UnusedRuntimeConverter(),
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			selfRegManager := testCase.SelfRegManagerFn()
			uuidSvc := &automock.UidService{}

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil, selfRegManager, uuidSvc, nil, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.Runtime(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_Runtimes(t *testing.T) {
	// GIVEN
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
		SelfRegManagerFn  func() *automock.SelfRegisterManager
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
			TransactionerFn: txtest.TransactionerThatSucceeds,
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
			SelfRegManagerFn:  rtmtest.NoopSelfRegManager,
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
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("List", contextParam, filter, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn:       UnusedRuntimeConverter(),
			SelfRegManagerFn:  rtmtest.NoopSelfRegManager,
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
			selfRegManager := testCase.SelfRegManagerFn()
			uuidSvc := &automock.UidService{}

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil, selfRegManager, uuidSvc, nil, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.Runtimes(context.TODO(), testCase.InputLabelFilters, testCase.InputFirst, testCase.InputAfter)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_RuntimeByTokenIssuer(t *testing.T) {
	// GIVEN
	modelRuntime := fixModelRuntime(t, "foo", "tenant-foo", "Foo", "Bar")
	gqlRuntime := fixGQLRuntime(t, "foo", "Foo", "Bar")
	testErr := errors.New("Test error")

	testCases := []struct {
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn        func() *automock.RuntimeService
		ConverterFn      func() *automock.RuntimeConverter
		SelfRegManagerFn func() *automock.SelfRegisterManager
		InputID          string
		ExpectedRuntime  *graphql.Runtime
		ExpectedErr      error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetByTokenIssuer", contextParam, "foo").Return(modelRuntime, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeConverter {
				conv := &automock.RuntimeConverter{}
				conv.On("ToGraphQL", modelRuntime).Return(gqlRuntime).Once()
				return conv
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputID:          "foo",
			ExpectedRuntime:  gqlRuntime,
			ExpectedErr:      nil,
		},
		{
			Name: "Success when runtime not found returns nil",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetByTokenIssuer", contextParam, "foo").Return(modelRuntime, apperrors.NewNotFoundError(resource.Runtime, "foo")).Once()

				return svc
			},
			ConverterFn:      UnusedRuntimeConverter(),
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns error when runtime retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetByTokenIssuer", contextParam, "foo").Return(nil, testErr).Once()

				return svc
			},
			ConverterFn:      UnusedRuntimeConverter(),
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputID:          "foo",
			ExpectedRuntime:  nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			selfRegManager := testCase.SelfRegManagerFn()
			uuidSvc := &automock.UidService{}

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil, selfRegManager, uuidSvc, nil, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.RuntimeByTokenIssuer(context.TODO(), testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntime, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_SetRuntimeLabel(t *testing.T) {
	// GIVEN
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
		Key:        labelKey,
		Value:      labelValue,
		ObjectID:   runtimeID,
		ObjectType: model.RuntimeLabelableObject,
	}

	testCases := []struct {
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn        func() *automock.RuntimeService
		ConverterFn      func() *automock.RuntimeConverter
		SelfRegManagerFn func() *automock.SelfRegisterManager
		InputRuntimeID   string
		InputKey         string
		InputValue       interface{}
		ExpectedLabel    *graphql.Label
		ExpectedErr      error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("SetLabel", contextParam, modelLabelInput).Return(nil).Once()
				svc.On("GetLabel", contextParam, runtimeID, modelLabelInput.Key).Return(modelLabel, nil).Once()
				return svc
			},
			ConverterFn:      UnusedRuntimeConverter(),
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputRuntimeID:   runtimeID,
			InputKey:         gqlLabel.Key,
			InputValue:       gqlLabel.Value,
			ExpectedLabel:    gqlLabel,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns error when adding label to runtime failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("SetLabel", contextParam, modelLabelInput).Return(testErr).Once()
				return svc
			},
			ConverterFn:      UnusedRuntimeConverter(),
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputRuntimeID:   runtimeID,
			InputKey:         gqlLabel.Key,
			InputValue:       gqlLabel.Value,
			ExpectedLabel:    nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			selfRegManager := testCase.SelfRegManagerFn()
			uuidSvc := &automock.UidService{}

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil, selfRegManager, uuidSvc, nil, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.SetRuntimeLabel(context.TODO(), testCase.InputRuntimeID, testCase.InputKey, testCase.InputValue)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}

	t.Run("Returns error when Label input validation failed", func(t *testing.T) {
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		result, err := resolver.SetRuntimeLabel(context.TODO(), "", "", "")

		// then
		require.Nil(t, result)
		require.Error(t, err)
		assert.Contains(t, err.Error(), "value=cannot be blank")
		assert.Contains(t, err.Error(), "key=cannot be blank")
	})
}

func TestResolver_DeleteRuntimeLabel(t *testing.T) {
	// GIVEN
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
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn        func() *automock.RuntimeService
		ConverterFn      func() *automock.RuntimeConverter
		SelfRegManagerFn func() *automock.SelfRegisterManager
		InputRuntimeID   string
		InputKey         string
		ExpectedLabel    *graphql.Label
		ExpectedErr      error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, gqlLabel.Key).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", contextParam, runtimeID, gqlLabel.Key).Return(nil).Once()
				return svc
			},
			ConverterFn:      UnusedRuntimeConverter(),
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputRuntimeID:   runtimeID,
			InputKey:         gqlLabel.Key,
			ExpectedLabel:    gqlLabel,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns error when label retrieval failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, gqlLabel.Key).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn:      UnusedRuntimeConverter(),
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputRuntimeID:   runtimeID,
			InputKey:         gqlLabel.Key,
			ExpectedLabel:    nil,
			ExpectedErr:      testErr,
		},
		{
			Name: "Returns error when deleting runtime's label failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, gqlLabel.Key).Return(modelLabel, nil).Once()
				svc.On("DeleteLabel", contextParam, runtimeID, gqlLabel.Key).Return(testErr).Once()
				return svc
			},
			ConverterFn:      UnusedRuntimeConverter(),
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputRuntimeID:   runtimeID,
			InputKey:         gqlLabel.Key,
			ExpectedLabel:    nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			transact := testCase.TransactionerFn(persistTx)
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()
			selfRegManager := testCase.SelfRegManagerFn()
			uuidSvc := &automock.UidService{}

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, converter, nil, nil, nil, selfRegManager, uuidSvc, nil, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.DeleteRuntimeLabel(context.TODO(), testCase.InputRuntimeID, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedLabel, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persistTx)
		})
	}
}

func TestResolver_Webhooks(t *testing.T) {
	// GIVEN
	id := "foo"

	gqlRuntime := fixGQLRuntime(t, id, "name", "desc")

	modelWebhooks := []*model.Webhook{{Type: "test webhook"}}
	gqlWebhooks := []*graphql.Webhook{{Type: "test webhook"}}

	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.WebhookService
		ConverterFn     func() *automock.WebhookConverter
		ExpectedResult  []*graphql.Webhook
		ExpectedErr     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListForRuntime", contextParam, gqlRuntime.ID).Return(modelWebhooks, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				converter := &automock.WebhookConverter{}
				converter.Mock.On("MultipleToGraphQL", modelWebhooks).Return(gqlWebhooks, nil).Once()
				return converter
			},
			ExpectedResult: gqlWebhooks,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.WebhookService {
				return &automock.WebhookService{}
			},
			ConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when listing webhooks failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListForRuntime", contextParam, gqlRuntime.ID).Return([]*model.Webhook{}, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListForRuntime", contextParam, gqlRuntime.ID).Return(modelWebhooks, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				return &automock.WebhookConverter{}
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
		{
			Name:            "Returns error when webhook conversion failed",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.WebhookService {
				svc := &automock.WebhookService{}
				svc.On("ListForRuntime", contextParam, gqlRuntime.ID).Return(modelWebhooks, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.WebhookConverter {
				converter := &automock.WebhookConverter{}
				converter.Mock.On("MultipleToGraphQL", modelWebhooks).Return(nil, testErr).Once()
				return converter
			},
			ExpectedResult: nil,
			ExpectedErr:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, svc, converter, nil, nil)

			// WHEN
			result, err := resolver.Webhooks(context.TODO(), gqlRuntime)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, transact, persist, converter)
		})
	}
	t.Run("Returns error when Runtime object is missing", func(t *testing.T) {
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.Webhooks(context.TODO(), nil)
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Runtime cannot be empty")
	})
}

func TestResolver_Labels(t *testing.T) {
	// GIVEN
	id := "foo"
	labelKey1 := "key1"
	labelValue1 := "val1"
	labelKey2 := "key2"
	labelValue2 := "val2"

	gqlRuntime := fixGQLRuntime(t, id, "name", "desc")

	modelLabels := map[string]*model.Label{
		"abc": {
			ID:         "abc",
			Key:        labelKey1,
			Value:      labelValue1,
			ObjectID:   id,
			ObjectType: model.RuntimeLabelableObject,
		},
		"def": {
			ID:         "def",
			Key:        labelKey2,
			Value:      labelValue2,
			ObjectID:   id,
			ObjectType: model.RuntimeLabelableObject,
		},
	}

	gqlLabels := graphql.Labels{
		labelKey1: labelValue1,
		labelKey2: labelValue2,
	}

	gqlLabels1 := graphql.Labels{
		labelKey1: labelValue1,
	}

	testErr := errors.New("Test error")

	testCases := []struct {
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn        func() *automock.RuntimeService
		SelfRegManagerFn func() *automock.SelfRegisterManager
		InputRuntime     *graphql.Runtime
		InputKey         *string
		ExpectedResult   graphql.Labels
		ExpectedErr      error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListLabels", contextParam, id).Return(modelLabels, nil).Once()
				return svc
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputKey:         nil,
			ExpectedResult:   gqlLabels,
			ExpectedErr:      nil,
		},
		{
			Name: "Success when labels are filtered",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListLabels", contextParam, id).Return(modelLabels, nil).Once()
				return svc
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputKey:         &labelKey1,
			ExpectedResult:   gqlLabels1,
			ExpectedErr:      nil,
		},
		{
			Name: "Success returns nil when labels not found",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListLabels", contextParam, id).Return(nil, errors.New("doesn't exist")).Once()
				return svc
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputKey:         &labelKey1,
			ExpectedResult:   nil,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns error when label listing failed",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatDoesARollback,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("ListLabels", contextParam, id).Return(nil, testErr).Once()
				return svc
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputKey:         &labelKey1,
			ExpectedResult:   nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			svc := testCase.ServiceFn()
			transact := testCase.TransactionerFn(persistTx)
			selfRegManager := testCase.SelfRegManagerFn()
			uuidSvc := &automock.UidService{}

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, selfRegManager, uuidSvc, nil, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.Labels(context.TODO(), gqlRuntime, testCase.InputKey)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, transact, persistTx)
		})
	}
}

func TestResolver_GetLabel(t *testing.T) {
	// GIVEN
	runtimeID := "37e89317-9ace-441d-9dc0-badf09b035b4"
	labelKey := runtime.IsNormalizedLabel
	labelValue := "true"

	modelLabel := &model.Label{
		ID:         "abc",
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
		Name             string
		PersistenceFn    func() *persistenceautomock.PersistenceTx
		TransactionerFn  func(persistTx *persistenceautomock.PersistenceTx) *persistenceautomock.Transactioner
		ServiceFn        func() *automock.RuntimeService
		SelfRegManagerFn func() *automock.SelfRegisterManager
		InputRuntime     *graphql.Runtime
		InputKey         string
		ExpectedResult   *graphql.Labels
		ExpectedErr      error
	}{
		{
			Name: "Success",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, labelKey).Return(modelLabel, nil).Once()
				return svc
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputKey:         labelKey,
			ExpectedResult:   gqlLabels,
			ExpectedErr:      nil,
		},
		{
			Name: "Success returns nil when label not found",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				persistTx.On("Commit").Return(nil).Once()
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, labelKey).Return(nil, apperrors.NewNotFoundError(resource.Runtime, runtimeID)).Once()
				return svc
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputKey:         labelKey,
			ExpectedResult:   nil,
			ExpectedErr:      nil,
		},
		{
			Name: "Returns error when label listing fails",
			PersistenceFn: func() *persistenceautomock.PersistenceTx {
				persistTx := &persistenceautomock.PersistenceTx{}
				return persistTx
			},
			TransactionerFn: txtest.TransactionerThatSucceeds,
			ServiceFn: func() *automock.RuntimeService {
				svc := &automock.RuntimeService{}
				svc.On("GetLabel", contextParam, runtimeID, labelKey).Return(nil, testErr).Once()
				return svc
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			InputKey:         labelKey,
			ExpectedResult:   nil,
			ExpectedErr:      testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persistTx := testCase.PersistenceFn()
			svc := testCase.ServiceFn()
			transact := testCase.TransactionerFn(persistTx)
			selfRegManager := testCase.SelfRegManagerFn()
			uuidSvc := &automock.UidService{}

			resolver := runtime.NewResolver(transact, svc, nil, nil, nil, nil, nil, nil, nil, selfRegManager, uuidSvc, nil, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.GetLabel(context.TODO(), runtimeID, labelKey)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, transact, persistTx)
		})
	}
}

func TestResolver_RuntimeContext(t *testing.T) {
	// GIVEN
	id := "foo"
	key := "key"
	val := "value"
	runtimeID := "runtime_id"

	modelRuntimeContext := fixModelRuntimeContext(id, runtimeID, key, val)
	gqlRuntimeContext := fixGqlRuntimeContext(id, key, val)
	gqlRuntime := fixGQLRuntime(t, runtimeID, "runtime", "description")
	testErr := errors.New("Test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name                   string
		TransactionerFn        func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn              func() *automock.RuntimeContextService
		ConverterFn            func() *automock.RuntimeContextConverter
		InputID                string
		Runtime                *graphql.Runtime
		ExpectedRuntimeContext *graphql.RuntimeContext
		ExpectedErr            error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("GetForRuntime", contextParam, "foo", runtimeID).Return(modelRuntimeContext, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("ToGraphQL", modelRuntimeContext).Return(gqlRuntimeContext).Once()
				return conv
			},
			InputID:                id,
			Runtime:                gqlRuntime,
			ExpectedRuntimeContext: gqlRuntimeContext,
			ExpectedErr:            nil,
		},
		{
			Name:            "Returns when runtime context for runtime not found",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("GetForRuntime", contextParam, "foo", runtimeID).Return(modelRuntimeContext, apperrors.NewNotFoundError(resource.RuntimeContext, "foo")).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				return conv
			},
			InputID:                id,
			Runtime:                gqlRuntime,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            nil,
		},
		{
			Name:            "Returns error when runtime context retrieval failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("GetForRuntime", contextParam, "foo", runtimeID).Return(nil, testErr).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				return conv
			},
			InputID:                id,
			Runtime:                gqlRuntime,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name:            "Returns error when commit begin error",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				return conv
			},
			InputID:                "foo",
			Runtime:                gqlRuntime,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
		{
			Name:            "Returns error when commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("GetForRuntime", contextParam, "foo", runtimeID).Return(modelRuntimeContext, nil).Once()

				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				return conv
			},
			InputID:                "foo",
			Runtime:                gqlRuntime,
			ExpectedRuntimeContext: nil,
			ExpectedErr:            testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, svc, converter, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.RuntimeContext(context.TODO(), testCase.Runtime, testCase.InputID)

			// then
			assert.Equal(t, testCase.ExpectedRuntimeContext, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persist)
		})
	}
	t.Run("Returns error when Runtime object is missing", func(t *testing.T) {
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.RuntimeContext(context.TODO(), nil, "foo")
		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Runtime cannot be empty")
	})
}

func TestResolver_RuntimeContexts(t *testing.T) {
	id := "foo"
	key := "key"
	val := "value"
	firstRuntimeID := "runtime_id"
	secondRuntimeID := "runtime2_id"
	runtimeIDs := []string{firstRuntimeID, secondRuntimeID}

	testErr := errors.New("Test error")

	rtmCtxFirstRuntime := fixModelRuntimeContext(id, firstRuntimeID, key, val)
	rtmCtxSecondRuntime := fixModelRuntimeContext(id+"2", secondRuntimeID, key+"2", val+"2")

	rtmCtxsFirstRuntime := []*model.RuntimeContext{rtmCtxFirstRuntime}
	rtmCtxsSecondRuntime := []*model.RuntimeContext{rtmCtxSecondRuntime}

	gqlRtmCtxFirstRuntime := fixGqlRuntimeContext(id, key, val)
	gqlRtmCtxSecondRuntime := fixGqlRuntimeContext(id+"2", key+"2", val+"2")

	gqlRtmCtxsFirstRuntime := []*graphql.RuntimeContext{gqlRtmCtxFirstRuntime}
	gqlRtmCtxsSecondRuntime := []*graphql.RuntimeContext{gqlRtmCtxSecondRuntime}

	rtmCtxPageFirstRuntime := fixRtmCtxPage(rtmCtxsFirstRuntime)
	rtmCtxPageSecondRuntime := fixRtmCtxPage(rtmCtxsSecondRuntime)
	rtmCtxPages := []*model.RuntimeContextPage{rtmCtxPageFirstRuntime, rtmCtxPageSecondRuntime}

	gqlrtmCtxPageFirstApp := fixGQLRtmCtxPage(gqlRtmCtxsFirstRuntime)
	gqlrtmCtxPageSecondApp := fixGQLRtmCtxPage(gqlRtmCtxsSecondRuntime)
	gqlrtmCtxPages := []*graphql.RuntimeContextPage{gqlrtmCtxPageFirstApp, gqlrtmCtxPageSecondApp}
	// GIVEN

	txGen := txtest.NewTransactionContextGenerator(testErr)

	first := 2
	gqlAfter := graphql.PageCursor("test")
	after := "test"

	testCases := []struct {
		Name            string
		TransactionerFn func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		ServiceFn       func() *automock.RuntimeContextService
		ConverterFn     func() *automock.RuntimeContextConverter
		ExpectedResult  []*graphql.RuntimeContextPage
		ExpectedErr     []error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("ListByRuntimeIDs", txtest.CtxWithDBMatcher(), runtimeIDs, first, after).Return(rtmCtxPages, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("MultipleToGraphQL", rtmCtxsFirstRuntime).Return(gqlRtmCtxsFirstRuntime).Once()
				conv.On("MultipleToGraphQL", rtmCtxsSecondRuntime).Return(gqlRtmCtxsSecondRuntime).Once()
				return conv
			},
			ExpectedResult: gqlrtmCtxPages,
			ExpectedErr:    nil,
		},
		{
			Name:            "Returns error when transaction begin failed",
			TransactionerFn: txGen.ThatFailsOnBegin,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when runtime context listing failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("ListByRuntimeIDs", txtest.CtxWithDBMatcher(), runtimeIDs, first, after).Return(nil, testErr).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
		{
			Name:            "Returns error when transaction commit failed",
			TransactionerFn: txGen.ThatFailsOnCommit,
			ServiceFn: func() *automock.RuntimeContextService {
				svc := &automock.RuntimeContextService{}
				svc.On("ListByRuntimeIDs", txtest.CtxWithDBMatcher(), runtimeIDs, first, after).Return(rtmCtxPages, nil).Once()
				return svc
			},
			ConverterFn: func() *automock.RuntimeContextConverter {
				conv := &automock.RuntimeContextConverter{}
				conv.On("MultipleToGraphQL", rtmCtxsFirstRuntime).Return(gqlRtmCtxsFirstRuntime).Once()
				conv.On("MultipleToGraphQL", rtmCtxsSecondRuntime).Return(gqlRtmCtxsSecondRuntime).Once()
				return conv
			},
			ExpectedResult: nil,
			ExpectedErr:    []error{testErr},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			persist, transact := testCase.TransactionerFn()
			svc := testCase.ServiceFn()
			converter := testCase.ConverterFn()

			resolver := runtime.NewResolver(transact, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, svc, converter, nil, nil, nil, nil)
			firstRuntimeParams := dataloader.ParamRuntimeContext{ID: firstRuntimeID, Ctx: context.TODO(), First: &first, After: &gqlAfter}
			secondRuntimeParams := dataloader.ParamRuntimeContext{ID: secondRuntimeID, Ctx: context.TODO(), First: &first, After: &gqlAfter}
			keys := []dataloader.ParamRuntimeContext{firstRuntimeParams, secondRuntimeParams}

			// WHEN
			result, err := resolver.RuntimeContextsDataLoader(keys)

			// then
			assert.Equal(t, testCase.ExpectedResult, result)
			assert.Equal(t, testCase.ExpectedErr, err)

			mock.AssertExpectationsForObjects(t, svc, converter, transact, persist)
		})
	}

	t.Run("Returns error when there are no Runtimes", func(t *testing.T) {
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.RuntimeContextsDataLoader([]dataloader.ParamRuntimeContext{})
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInternalError("No Runtimes found").Error())
	})

	t.Run("Returns error when start cursor is nil", func(t *testing.T) {
		firstRuntimeParams := dataloader.ParamRuntimeContext{ID: firstRuntimeID, Ctx: context.TODO(), First: nil, After: &gqlAfter}
		keys := []dataloader.ParamRuntimeContext{firstRuntimeParams}

		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)
		// WHEN
		_, err := resolver.RuntimeContextsDataLoader(keys)
		// THEN
		require.Error(t, err[0])
		assert.EqualError(t, err[0], apperrors.NewInvalidDataError("missing required parameter 'first'").Error())
	})
}

func TestResolver_Auths(t *testing.T) {
	// GIVEN
	tnt := "tnt"
	externalTnt := "external-tnt"
	ctx := context.TODO()
	ctx = tenant.SaveToContext(ctx, tnt, externalTnt)

	parentRuntime := fixGQLRuntime(t, "foo", "bar", "baz")

	modelSysAuths := []pkgmodel.SystemAuth{
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
		Name             string
		TransactionerFn  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		SysAuthSvcFn     func() *automock.SystemAuthService
		SysAuthConvFn    func() *automock.SystemAuthConverter
		SelfRegManagerFn func() *automock.SelfRegisterManager
		ExpectedOutput   []*graphql.RuntimeSystemAuth
		ExpectedError    error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				sysAuthSvc.On("ListForObject", txtest.CtxWithDBMatcher(), pkgmodel.RuntimeReference, parentRuntime.ID).Return(modelSysAuths, nil).Once()
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				sysAuthConv.On("ToGraphQL", &modelSysAuths[0]).Return(gqlSysAuths[0], nil).Once()
				sysAuthConv.On("ToGraphQL", &modelSysAuths[1]).Return(gqlSysAuths[1], nil).Once()
				sysAuthConv.On("ToGraphQL", &modelSysAuths[2]).Return(gqlSysAuths[2], nil).Once()
				return sysAuthConv
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			ExpectedOutput:   gqlSysAuths,
			ExpectedError:    nil,
		},
		{
			Name:            "Error when listing for object",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				sysAuthSvc.On("ListForObject", txtest.CtxWithDBMatcher(), pkgmodel.RuntimeReference, parentRuntime.ID).Return(nil, testErr).Once()
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			ExpectedOutput:   nil,
			ExpectedError:    testErr,
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
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			ExpectedOutput:   nil,
			ExpectedError:    testErr,
		},
		{
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			SysAuthSvcFn: func() *automock.SystemAuthService {
				sysAuthSvc := &automock.SystemAuthService{}
				sysAuthSvc.On("ListForObject", txtest.CtxWithDBMatcher(), pkgmodel.RuntimeReference, parentRuntime.ID).Return(modelSysAuths, nil).Once()
				return sysAuthSvc
			},
			SysAuthConvFn: func() *automock.SystemAuthConverter {
				sysAuthConv := &automock.SystemAuthConverter{}
				return sysAuthConv
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			ExpectedOutput:   nil,
			ExpectedError:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			sysAuthSvc := testCase.SysAuthSvcFn()
			sysAuthConv := testCase.SysAuthConvFn()
			selfRegManager := testCase.SelfRegManagerFn()
			uuidSvc := &automock.UidService{}

			resolver := runtime.NewResolver(transact, nil, nil, sysAuthSvc, nil, nil, sysAuthConv, nil, nil, selfRegManager, uuidSvc, nil, nil, nil, nil, nil, nil, nil)

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
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

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
		Name             string
		TransactionerFn  func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		EventingSvcFn    func() *automock.EventingService
		SelfRegManagerFn func() *automock.SelfRegisterManager
		ExpectedOutput   *graphql.RuntimeEventingConfiguration
		ExpectedError    error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("GetForRuntime", txtest.CtxWithDBMatcher(), runtimeID).Return(modelRuntimeEventingCfg, nil).Once()

				return eventingSvc
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			ExpectedOutput:   gqlRuntimeEventingCfg,
			ExpectedError:    nil,
		}, {
			Name:            "Error when getting the configuration for runtime failed",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("GetForRuntime", txtest.CtxWithDBMatcher(), runtimeID).Return(nil, testErr).Once()

				return eventingSvc
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			ExpectedOutput:   nil,
			ExpectedError:    testErr,
		}, {
			Name:            "Error when beginning transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				return eventingSvc
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			ExpectedOutput:   nil,
			ExpectedError:    testErr,
		}, {
			Name:            "Error when committing transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			EventingSvcFn: func() *automock.EventingService {
				eventingSvc := &automock.EventingService{}
				eventingSvc.On("GetForRuntime", txtest.CtxWithDBMatcher(), runtimeID).Return(modelRuntimeEventingCfg, nil).Once()

				return eventingSvc
			},
			SelfRegManagerFn: rtmtest.NoopSelfRegManager,
			ExpectedOutput:   nil,
			ExpectedError:    testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			eventingSvc := testCase.EventingSvcFn()
			selfRegManager := testCase.SelfRegManagerFn()
			uuidSvc := &automock.UidService{}

			resolver := runtime.NewResolver(transact, nil, nil, nil, nil, nil, nil, eventingSvc, nil, selfRegManager, uuidSvc, nil, nil, nil, nil, nil, nil, nil)

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
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		result, err := resolver.EventingConfiguration(ctx, &graphql.Runtime{ID: "abc"})

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "while parsing runtime ID as UUID")
		assert.Nil(t, result)
	})

	t.Run("Error when parent object is nil", func(t *testing.T) {
		// GIVEN
		resolver := runtime.NewResolver(nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil)

		// WHEN
		result, err := resolver.EventingConfiguration(context.TODO(), nil)

		// THEN
		require.Error(t, err)
		assert.Contains(t, err.Error(), "Runtime cannot be empty")
		assert.Nil(t, result)
	})
}

func TestResolver_SubscribeTenant(t *testing.T) {
	// GIVEN
	providerID := "provider-id"
	subaccountID := "subaccount-id"
	providerSubaccountID := "provider-subaccount-id"
	region := "region"

	ctx := context.TODO()

	testErr := errors.New("this is a test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		SubscriptionSvcFn func() *automock.SubscriptionService
		ExpectedOutput    bool
		ExpectedError     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			SubscriptionSvcFn: func() *automock.SubscriptionService {
				subscriptionSvc := &automock.SubscriptionService{}
				subscriptionSvc.On("SubscribeTenant", txtest.CtxWithDBMatcher(), providerID, subaccountID, providerSubaccountID, region).Return(true, nil).Once()
				return subscriptionSvc
			},
			ExpectedOutput: true,
			ExpectedError:  nil,
		},
		{
			Name:            "Error when cannot start transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			SubscriptionSvcFn: func() *automock.SubscriptionService {
				return &automock.SubscriptionService{}
			},
			ExpectedOutput: false,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when cannot subscribe tenant to runtime",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			SubscriptionSvcFn: func() *automock.SubscriptionService {
				subscriptionSvc := &automock.SubscriptionService{}
				subscriptionSvc.On("SubscribeTenant", txtest.CtxWithDBMatcher(), providerID, subaccountID, providerSubaccountID, region).Return(false, testErr).Once()
				return subscriptionSvc
			},
			ExpectedOutput: false,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when cannot commit transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			SubscriptionSvcFn: func() *automock.SubscriptionService {
				subscriptionSvc := &automock.SubscriptionService{}
				subscriptionSvc.On("SubscribeTenant", txtest.CtxWithDBMatcher(), providerID, subaccountID, providerSubaccountID, region).Return(true, nil).Once()
				return subscriptionSvc
			},
			ExpectedOutput: false,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			subscriptionSvc := testCase.SubscriptionSvcFn()
			defer mock.AssertExpectationsForObjects(t, persist, transact, subscriptionSvc)

			resolver := runtime.NewResolver(transact, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, subscriptionSvc, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.SubscribeTenant(ctx, providerID, subaccountID, providerSubaccountID, region)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)
		})
	}
}

func TestResolver_UnsubscribeTenant(t *testing.T) {
	// GIVEN
	providerID := "provider-id"
	subaccountID := "subaccount-id"
	providerSubaccountID := "provider-subaccount-id"
	region := "region"

	ctx := context.TODO()

	testErr := errors.New("this is a test error")
	txGen := txtest.NewTransactionContextGenerator(testErr)

	testCases := []struct {
		Name              string
		TransactionerFn   func() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner)
		SubscriptionSvcFn func() *automock.SubscriptionService
		ExpectedOutput    bool
		ExpectedError     error
	}{
		{
			Name:            "Success",
			TransactionerFn: txGen.ThatSucceeds,
			SubscriptionSvcFn: func() *automock.SubscriptionService {
				subscriptionSvc := &automock.SubscriptionService{}
				subscriptionSvc.On("UnsubscribeTenant", txtest.CtxWithDBMatcher(), providerID, subaccountID, providerSubaccountID, region).Return(true, nil).Once()
				return subscriptionSvc
			},
			ExpectedOutput: true,
			ExpectedError:  nil,
		},
		{
			Name:            "Error when cannot start transaction",
			TransactionerFn: txGen.ThatFailsOnBegin,
			SubscriptionSvcFn: func() *automock.SubscriptionService {
				return &automock.SubscriptionService{}
			},
			ExpectedOutput: false,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when cannot unsubscribe tenant from runtime",
			TransactionerFn: txGen.ThatDoesntExpectCommit,
			SubscriptionSvcFn: func() *automock.SubscriptionService {
				subscriptionSvc := &automock.SubscriptionService{}
				subscriptionSvc.On("UnsubscribeTenant", txtest.CtxWithDBMatcher(), providerID, subaccountID, providerSubaccountID, region).Return(false, testErr).Once()
				return subscriptionSvc
			},
			ExpectedOutput: false,
			ExpectedError:  testErr,
		},
		{
			Name:            "Error when cannot commit transaction",
			TransactionerFn: txGen.ThatFailsOnCommit,
			SubscriptionSvcFn: func() *automock.SubscriptionService {
				subscriptionSvc := &automock.SubscriptionService{}
				subscriptionSvc.On("UnsubscribeTenant", txtest.CtxWithDBMatcher(), providerID, subaccountID, providerSubaccountID, region).Return(true, nil).Once()
				return subscriptionSvc
			},
			ExpectedOutput: false,
			ExpectedError:  testErr,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			persist, transact := testCase.TransactionerFn()
			subscriptionSvc := testCase.SubscriptionSvcFn()
			defer mock.AssertExpectationsForObjects(t, persist, transact, subscriptionSvc)

			resolver := runtime.NewResolver(transact, nil, nil, nil, nil, nil, nil, nil, nil, nil, nil, subscriptionSvc, nil, nil, nil, nil, nil, nil)

			// WHEN
			result, err := resolver.UnsubscribeTenant(ctx, providerID, subaccountID, providerSubaccountID, region)

			// THEN
			if testCase.ExpectedError != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.ExpectedOutput, result)
		})
	}
}

func fixOAuths() []pkgmodel.SystemAuth {
	return []pkgmodel.SystemAuth{
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

func UnusedRuntimeConverter() func() *automock.RuntimeConverter {
	return func() *automock.RuntimeConverter {
		return &automock.RuntimeConverter{}
	}
}

func UnusedScenarioAssignmentService() func() *automock.ScenarioAssignmentService {
	return func() *automock.ScenarioAssignmentService {
		return &automock.ScenarioAssignmentService{}
	}
}

func UnusedSysAuthService() func() *automock.SystemAuthService {
	return func() *automock.SystemAuthService {
		return &automock.SystemAuthService{}
	}
}

func UnusedOauth20Service() func() *automock.OAuth20Service {
	return func() *automock.OAuth20Service {
		return &automock.OAuth20Service{}
	}
}

func UnusedBundleInstanceAuthService() func() *automock.BundleInstanceAuthService {
	return func() *automock.BundleInstanceAuthService {
		return &automock.BundleInstanceAuthService{}
	}
}

func UnusedFormationService() func() *automock.FormationService {
	return func() *automock.FormationService {
		return &automock.FormationService{}
	}
}
