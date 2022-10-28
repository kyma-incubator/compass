package formationmapping_test

import (
	"context"
	"encoding/json"
	"net/http"
	"testing"

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

func fixFormationAssignmentModel(testFormationID, sourceID, targetID string, sourceFAType, targetFAType model.FormationAssignmentType) *model.FormationAssignment {
	return &model.FormationAssignment{
		ID:          "ID",
		FormationID: testFormationID,
		TenantID:    "tenant",
		Source:      sourceID,
		SourceType:  sourceFAType,
		Target:      targetID,
		TargetType:  targetFAType,
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
