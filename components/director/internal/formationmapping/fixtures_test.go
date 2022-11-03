package formationmapping_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tenant"
	fm "github.com/kyma-incubator/compass/components/director/internal/formationmapping"
	"github.com/kyma-incubator/compass/components/director/internal/formationmapping/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/consumer"
	persistenceautomock "github.com/kyma-incubator/compass/components/director/pkg/persistence/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
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
	emptyCtx := context.Background()
	tenantCtx := tenant.SaveToContext(emptyCtx, internalTntID, externalTntID)
	consumerAndTenantCtx := consumer.SaveToContext(tenantCtx, c)

	return consumerAndTenantCtx
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

func fixFormationAssignmentModelWithStateAndConfig(testFormationID, testTenantID, sourceID, targetID string, sourceFAType, targetFAType model.FormationAssignmentType, state model.FormationAssignmentState, config string) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          "ID",
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

func fixFormationAssignmentInput(testFormationID, sourceID, targetID string, sourceFAType, targetFAType model.FormationAssignmentType, state model.FormationAssignmentState, config string) *model.FormationAssignmentInput {
	return &model.FormationAssignmentInput{
		FormationID: testFormationID,
		Source:      sourceID,
		SourceType:  sourceFAType,
		Target:      targetID,
		TargetType:  targetFAType,
		State:       string(state),
		Value:       json.RawMessage(config),
	}
}

func fixEmptyNotificationRequest() *webhookclient.NotificationRequest {
	return &webhookclient.NotificationRequest{
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

func fixUnusedFormationAssignmentConverter() *automock.FormationAssignmentConverter {
	return &automock.FormationAssignmentConverter{}
}

func fixUnusedFormationAssignmentNotificationSvc() *automock.FormationAssignmentNotificationService {
	return &automock.FormationAssignmentNotificationService{}
}

func fixUnusedFormationRepo() *automock.FormationRepository {
	return &automock.FormationRepository{}
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
