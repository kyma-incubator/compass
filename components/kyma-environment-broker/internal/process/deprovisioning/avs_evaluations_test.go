package deprovisioning

import (
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/avs"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

var idsHolder []int64

const (
	internalEvalId = int64(1234)
	externalEvalId = int64(5678)
)

func TestAvsEvaluationsRemovalStep_Run(t *testing.T) {
	// given
	logger := logrus.New()
	memoryStorage := storage.NewMemoryStorage()

	deProvisioningOperation := fixDeprovisioningOperation()
	deProvisioningOperation.Avs.AvsEvaluationInternalId = internalEvalId
	deProvisioningOperation.Avs.AVSEvaluationExternalId = externalEvalId
	err := memoryStorage.Operations().InsertDeprovisioningOperation(deProvisioningOperation)
	assert.NoError(t, err)
	assert.False(t, deProvisioningOperation.Avs.AVSInternalEvaluationDeleted)
	assert.False(t, deProvisioningOperation.Avs.AVSExternalEvaluationDeleted)

	mockOauthServer := newMockAvsOauthServer()
	defer mockOauthServer.Close()
	mockAvsServer := newMockAvsServer(t)
	defer mockAvsServer.Close()
	avsConfig := avsConfig(mockOauthServer, mockAvsServer)
	avsDel := avs.NewDelegator(avsConfig, memoryStorage.Operations())
	internalEvalAssistant := avs.NewInternalEvalAssistant(avsConfig)
	externalEvalAssistant := avs.NewExternalEvalAssistant(avsConfig)
	step := NewAvsEvaluationsRemovalStep(avsDel, memoryStorage.Operations(), externalEvalAssistant, internalEvalAssistant)

	assert.Equal(t, 0, len(idsHolder))
	// when
	deProvisioningOperation, repeat, err := step.Run(deProvisioningOperation, logger)

	// then
	assert.NoError(t, err)
	assert.Equal(t, time.Duration(0), repeat)
	assert.Equal(t, 2, len(idsHolder))
	assert.Contains(t, idsHolder, internalEvalId)
	assert.Contains(t, idsHolder, externalEvalId)

	inDB, err := memoryStorage.Operations().GetDeprovisioningOperationByID(deProvisioningOperation.ID)
	assert.NoError(t, err)
	assert.True(t, inDB.Avs.AVSInternalEvaluationDeleted)
	assert.True(t, inDB.Avs.AVSExternalEvaluationDeleted)
	assert.Equal(t,internalEvalId, inDB.Avs.AvsEvaluationInternalId)
	assert.Equal(t,externalEvalId,inDB.Avs.AVSEvaluationExternalId)
}

func newMockAvsOauthServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`{"access_token": "90d64460d14870c08c81352a05dedd3465940a7c", "scope": "user", "token_type": "bearer", "expires_in": 86400}`))
		}))
}

func newMockAvsServer(t *testing.T) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Method, http.MethodDelete)
		uri := r.URL.RequestURI()
		id := uri[1:len(uri)]
		intVal, err := strconv.ParseInt(id, 10, 64)
		assert.NoError(t, err)
		idsHolder = append(idsHolder, intVal)
		w.WriteHeader(http.StatusOK)
	}))
}

func avsConfig(mockOauthServer *httptest.Server, mockAvsServer *httptest.Server) avs.Config {
	return avs.Config{
		OauthTokenEndpoint:     mockOauthServer.URL,
		OauthUsername:          "dummy",
		OauthPassword:          "dummy",
		OauthClientId:          "dummy",
		ApiEndpoint:            mockAvsServer.URL,
		DefinitionType:         avs.DefinitionType,
		InternalTesterAccessId: 1234,
		ExternalTesterAccessId: 5678,
		GroupId:                5555,
	}
}
