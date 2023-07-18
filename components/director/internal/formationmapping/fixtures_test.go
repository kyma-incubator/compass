package formationmapping_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	tenantpkg "github.com/kyma-incubator/compass/components/director/pkg/tenant"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"
	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	fm "github.com/kyma-incubator/compass/components/director/internal/formationmapping"
	"github.com/kyma-incubator/compass/components/director/internal/formationmapping/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	m "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

var (
	emptyCtx = context.Background()
	testErr  = errors.New("test error")
	txGen    = txtest.NewTransactionContextGenerator(testErr)

	// Tenant IDs variables
	internalTntID = "testInternalID"
	externalTntID = "testExternalID"

	// Formation Assignment variables
	faSourceID                = "testSourceID"
	faTargetID                = "testTargetID"
	testFormationAssignmentID = "testFormationAssignmentID"

	// Formation variables
	testFormationID         = "testFormationID"
	testFormationName       = "testFormationName"
	testFormationTemplateID = "testFormationTemplateID"
)

func fixTestHandler(t *testing.T) http.HandlerFunc {
	return func(writer http.ResponseWriter, request *http.Request) {
		writer.WriteHeader(http.StatusOK)
		_, err := writer.Write([]byte("OK"))
		require.NoError(t, err)
	}
}

func fixRequestWithContext(t *testing.T, ctx context.Context, httpMethod string) *http.Request {
	reqWithContext, err := http.NewRequest(httpMethod, "/", nil)
	require.NoError(t, err)
	reqWithContext = reqWithContext.WithContext(ctx)
	return reqWithContext
}

func fixGetConsumer(consumerID string, consumerType consumer.ConsumerType) consumer.Consumer {
	return consumer.Consumer{
		ConsumerID:   consumerID,
		ConsumerType: consumerType,
	}
}

func fixContextWithTenantAndConsumer(c consumer.Consumer, internalTntID, externalTntID string) context.Context {
	tenantCtx := tenant.SaveToContext(emptyCtx, internalTntID, externalTntID)
	consumerAndTenantCtx := consumer.SaveToContext(tenantCtx, c)

	return consumerAndTenantCtx
}

func fixContextWithConsumer(c consumer.Consumer) context.Context {
	return consumer.SaveToContext(emptyCtx, c)
}

func fixFormationWithState(state model.FormationState) *model.Formation {
	return &model.Formation{
		ID:                  testFormationID,
		TenantID:            internalTntID,
		FormationTemplateID: testFormationTemplateID,
		Name:                testFormationName,
		State:               state,
	}
}

func fixFormationAssignmentModel(testFormationID, testTenantID, sourceID, targetID string, sourceFAType, targetFAType model.FormationAssignmentType) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          "ID",
		FormationID: testFormationID,
		TenantID:    testTenantID,
		Source:      sourceID,
		SourceType:  sourceFAType,
		Target:      targetID,
		TargetType:  targetFAType,
	}
}

func fixFormationAssignmentModelWithStateAndConfig(testFormationAssignmentID, testFormationID, testTenantID, sourceID, targetID string, sourceFAType, targetFAType model.FormationAssignmentType, state model.FormationAssignmentState, config string) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          testFormationAssignmentID,
		FormationID: testFormationID,
		TenantID:    testTenantID,
		Source:      sourceID,
		SourceType:  sourceFAType,
		Target:      targetID,
		TargetType:  targetFAType,
		State:       string(state),
		Value:       json.RawMessage(config),
	}
}

func fixBusinessTenantMapping() *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             internalTntID,
		Name:           "tnt",
		ExternalTenant: externalTntID,
		Type:           tenantpkg.Account,
	}
}

func fixResourceGroupBusinessTenantMapping() *model.BusinessTenantMapping {
	return &model.BusinessTenantMapping{
		ID:             internalTntID,
		Name:           "tnt",
		ExternalTenant: externalTntID,
		Type:           tenantpkg.ResourceGroup,
	}
}

func fixEmptyNotificationRequest() *webhookclient.FormationAssignmentNotificationRequest {
	return &webhookclient.FormationAssignmentNotificationRequest{
		Webhook:       graphql.Webhook{},
		Object:        nil,
		CorrelationID: "",
	}
}

func contextThatHasTenant(expectedTenant string) interface{} {
	return mock.MatchedBy(func(actual context.Context) bool {
		actualTenant, err := tenant.LoadFromContext(actual)
		if err != nil {
			return false
		}
		return actualTenant == expectedTenant
	})
}

func contextThatHasConsumer(expectedConsumerID string) interface{} {
	return mock.MatchedBy(func(actual context.Context) bool {
		consumer, err := consumer.LoadFromContext(actual)
		if err != nil {
			return false
		}
		return consumer.ConsumerID == expectedConsumerID
	})
}

func fixBuildExpectedErrResponse(t *testing.T, errMsg string) string {
	errorResponse := fm.ErrorResponse{Message: errMsg}
	encodingErr, err := json.Marshal(errorResponse)
	require.NoError(t, err)
	return string(encodingErr)
}

func fixUnusedTransactioner() (*persistenceautomock.PersistenceTx, *persistenceautomock.Transactioner) {
	return &persistenceautomock.PersistenceTx{}, &persistenceautomock.Transactioner{}
}

func fixUnusedFormationAssignmentSvc() *automock.FormationAssignmentService {
	return &automock.FormationAssignmentService{}
}

func fixUnusedFormationAssignmentNotificationSvc() *automock.FormationAssignmentNotificationService {
	return &automock.FormationAssignmentNotificationService{}
}

func fixUnusedFormationSvc() *automock.FormationService {
	return &automock.FormationService{}
}

func fixUnusedRuntimeRepo() *automock.RuntimeRepository {
	return &automock.RuntimeRepository{}
}

func fixUnusedRuntimeContextRepo() *automock.RuntimeContextRepository {
	return &automock.RuntimeContextRepository{}
}

func fixUnusedAppRepo() *automock.ApplicationRepository {
	return &automock.ApplicationRepository{}
}

func fixUnusedAppTemplateRepo() *automock.ApplicationTemplateRepository {
	return &automock.ApplicationTemplateRepository{}
}

func fixUnusedLabelRepo() *automock.LabelRepository {
	return &automock.LabelRepository{}
}

func fixUnusedTenantRepo() *automock.TenantRepository {
	return &automock.TenantRepository{}
}

func fixUnusedFormationRepo() *automock.FormationRepository {
	return &automock.FormationRepository{}
}

func fixUnusedFormationTemplateRepo() *automock.FormationTemplateRepository {
	return &automock.FormationTemplateRepository{}
}

func ThatDoesNotCommitInGoRoutine() (*m.PersistenceTx, *m.Transactioner) {
	persistTx := &m.PersistenceTx{}
	persistTx.On("Commit").Return(nil).Once()

	transact := &m.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Twice()
	transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
	transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

	return persistTx, transact
}

func ThatFailsOnCommitInGoRoutine() (*m.PersistenceTx, *m.Transactioner) {
	persistTx := &m.PersistenceTx{}
	persistTx.On("Commit").Return(nil).Once()
	persistTx.On("Commit").Return(testErr).Once()

	transact := &m.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Twice()
	transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(true).Once()
	transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

	return persistTx, transact
}

func ThatFailsOnBeginInGoRoutine() (*m.PersistenceTx, *m.Transactioner) {
	persistTx := &m.PersistenceTx{}
	persistTx.On("Commit").Return(nil).Once()

	transact := &m.Transactioner{}
	transact.On("Begin").Return(persistTx, nil).Once()
	transact.On("Begin").Return(persistTx, testErr).Once()
	transact.On("RollbackUnlessCommitted", mock.Anything, persistTx).Return(false).Once()

	return persistTx, transact
}
