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

type idHolder struct {
	id int64
}

func TestInternalEvaluationStep_Run(t *testing.T) {
	// given
	log := logrus.New()
	memoryStorage := storage.NewMemoryStorage()
	provisioningOperation := fixOperationCreateRuntime(t)
	err := memoryStorage.Operations().InsertProvisioningOperation(provisioningOperation)
	assert.NoError(t, err)

	idh := &idHolder{}
	mockOauthServer := newMockAvsOauthServer()
	defer mockOauthServer.Close()
	mockAvsServer := newMockAvsServer(t, idh, true)
	defer mockAvsServer.Close()
	avsConfig := avsConfig(mockOauthServer, mockAvsServer)
	avsDel := avs.NewDelegator(avsConfig, memoryStorage.Operations())

	ies := NewInternalEvaluationStep(avsConfig, avsDel)

	// when
	logger := log.WithFields(logrus.Fields{"step": "TEST"})
	provisioningOperation, repeat, err := ies.Run(provisioningOperation, logger)

	//then
	assert.NoError(t, err)
	assert.Equal(t, 0*time.Second, repeat)
	assert.Equal(t, idh.id, provisioningOperation.AvsEvaluationInternalId)

	inDB, err := memoryStorage.Operations().GetProvisioningOperationByID(provisioningOperation.ID)
	assert.NoError(t, err)
	assert.Equal(t, inDB.AvsEvaluationInternalId, idh.id)
}

func newMockAvsOauthServer() *httptest.Server {
	return httptest.NewServer(
		http.HandlerFunc(func(writer http.ResponseWriter, request *http.Request) {
			writer.Header().Set("Content-Type", "application/json")
			_, _ = writer.Write([]byte(`{"access_token": "90d64460d14870c08c81352a05dedd3465940a7c", "scope": "user", "token_type": "bearer", "expires_in": 86400}`))
		}))
}

func newMockAvsServer(t *testing.T, idh *idHolder, isInternal bool) *httptest.Server {
	return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, r.Header.Get("Content-Type"), "application/json")

		dec := json.NewDecoder(r.Body)
		var requestObj avs.BasicEvaluationCreateRequest
		err := dec.Decode(&requestObj)
		assert.NoError(t, err)

		if isInternal {
			assert.Empty(t, requestObj.URL)
		} else {
			assert.NotEmpty(t, requestObj.URL)
		}

		evalCreateResponse := createResponseObj(err, requestObj, t)
		idh.id = evalCreateResponse.Id
		responseObjAsBytes, _ := json.Marshal(evalCreateResponse)

		_, _ = w.Write(responseObjAsBytes)
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
