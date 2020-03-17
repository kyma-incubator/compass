package provisioning

import (
	"encoding/json"
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

func TestInternalEvaluationStep_Run(t *testing.T) {
	// given
	log := logrus.New()
	memoryStorage := storage.NewMemoryStorage()
	var id int64
	provisioningOperation := fixOperationCreateRuntime(t)
	err := memoryStorage.Operations().InsertProvisioningOperation(provisioningOperation)
	assert.NoError(t, err)

	mockOauthServer := httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`{"access_token": "90d64460d14870c08c81352a05dedd3465940a7c", "scope": "user", "token_type": "bearer", "expires_in": 86400}`))
		}))
	defer mockOauthServer.Close()

	mockAvsServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("Content-Type"), "application/json")

		dec := json.NewDecoder(r.Body)
		var requestObj avs.BasicEvaluationCreateRequest
		err := dec.Decode(&requestObj)
		assert.NoError(t, err)

		assert.Empty(t, requestObj.URL)

		evalCreateResponse := createResponseObj(err, requestObj, t)
		id = evalCreateResponse.Id
		responseObjAsBytes, _ := json.Marshal(evalCreateResponse)

		_, _ = w.Write(responseObjAsBytes)
		w.WriteHeader(http.StatusOK)
	}))
	defer mockAvsServer.Close()
	ies := NewInternalEvaluationStep(avsConfig(mockOauthServer, mockAvsServer), memoryStorage.Operations())

	// when
	logger := log.WithFields(logrus.Fields{"step": "TEST"})
	provisioningOperation, repeat, err := ies.Run(provisioningOperation, logger)

	//then
	assert.NoError(t, err)
	assert.Equal(t, 0*time.Second, repeat)
	assert.Equal(t, id, provisioningOperation.AvsEvaluationInternalId)

	inDB, err := memoryStorage.Operations().GetProvisioningOperationByID(provisioningOperation.ID)
	assert.NoError(t, err)
	assert.Equal(t, inDB.AvsEvaluationInternalId, id)
}

func avsConfig(mockOauthServer *httptest.Server, mockAvsServer *httptest.Server) avs.Config {
	return avs.Config{
		OauthTokenEndpoint: mockOauthServer.URL,
		OauthUsername:      "dummy",
		OauthPassword:      "dummy",
		OauthClientId:      "dummy",
		ApiEndpoint:        mockAvsServer.URL,
		DefinitionType:     avs.DefinitionType,
	}
}

func createResponseObj(err error, requestObj avs.BasicEvaluationCreateRequest, t *testing.T) *avs.BasicEvaluationCreateResponse {
	parseInt, err := strconv.ParseInt(requestObj.Threshold, 10, 64)
	assert.NoError(t, err)

	timeUnixEpoch := time.Now().Unix()
	id := time.Now().Unix()

	evalCreateResponse := &avs.BasicEvaluationCreateResponse{
		DefinitionType:             requestObj.DefinitionType,
		Name:                       requestObj.Name,
		Description:                requestObj.Description,
		Service:                    requestObj.Service,
		URL:                        requestObj.URL,
		CheckType:                  requestObj.CheckType,
		Interval:                   requestObj.Interval,
		TesterAccessId:             requestObj.TesterAccessId,
		Timeout:                    requestObj.Timeout,
		ReadOnly:                   requestObj.ReadOnly,
		ContentCheck:               requestObj.ContentCheck,
		ContentCheckType:           requestObj.ContentCheck,
		Threshold:                  parseInt,
		GroupId:                    requestObj.GroupId,
		Visibility:                 requestObj.Visibility,
		DateCreated:                timeUnixEpoch,
		DateChanged:                timeUnixEpoch,
		Owner:                      "abc@xyz.corp",
		Status:                     "ACTIVE",
		Alerts:                     nil,
		Tags:                       nil,
		Id:                         id,
		LegacyCheckId:              id,
		InternalInterval:           60,
		AuthType:                   "AUTH_NONE",
		IndividualOutageEventsOnly: false,
		IdOnTester:                 "",
	}
	return evalCreateResponse
}
