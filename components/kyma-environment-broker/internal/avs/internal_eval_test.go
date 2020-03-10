package avs

import (
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strconv"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input"
	inputAutomock "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input/automock"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
)

const (
	kymaVersion            = "1.10"
	instanceID             = "58f8c703-1756-48ab-9299-a847974d1fee"
	operationID            = "fd5cee4d-0eeb-40d0-a7a7-0708e5eba470"
	globalAccountID        = "80ac17bd-33e8-4ffa-8d56-1d5367755723"
	subAccountID           = "12df5747-3efb-4df6-ad6f-4414bb661ce3"
	serviceManagerURL      = "http://sm.com"
	serviceManagerUser     = "admin"
	serviceManagerPassword = "admin123"
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
		var requestObj basicEvaluationCreateRequest
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

func avsConfig(mockOauthServer *httptest.Server, mockAvsServer *httptest.Server) Config {
	return Config{
		OauthTokenEndpoint: mockOauthServer.URL,
		OauthUsername:      "dummy",
		OauthPassword:      "dummy",
		OauthClientId:      "dummy",
		ApiEndpoint:        mockAvsServer.URL,
		DefinitionType:     definitionType,
	}
}

func createResponseObj(err error, requestObj basicEvaluationCreateRequest, t *testing.T) *basicEvaluationCreateResponse {
	parseInt, err := strconv.ParseInt(requestObj.Threshold, 10, 64)
	assert.NoError(t, err)

	timeUnixEpoch := time.Now().Unix()
	id := time.Now().Unix()

	evalCreateResponse := &basicEvaluationCreateResponse{
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

func fixOperationCreateRuntime(t *testing.T) internal.ProvisioningOperation {
	return internal.ProvisioningOperation{
		Operation: internal.Operation{
			ID:          operationID,
			InstanceID:  instanceID,
			Description: "",
			UpdatedAt:   time.Now(),
		},
		ProvisioningParameters: fixProvisioningParameters(t),
		InputCreator:           fixInputCreator(t),
	}
}

func fixProvisioningParameters(t *testing.T) string {
	parameters := internal.ProvisioningParameters{
		PlanID:    broker.GcpPlanID,
		ServiceID: "",
		ErsContext: internal.ERSContext{
			GlobalAccountID: globalAccountID,
			SubAccountID:    subAccountID,
			ServiceManager: internal.ServiceManagerEntryDTO{
				Credentials: internal.ServiceManagerCredentials{
					BasicAuth: internal.ServiceManagerBasicAuth{
						Username: serviceManagerUser,
						Password: serviceManagerPassword,
					},
				},
				URL: serviceManagerURL,
			},
		},
		Parameters: internal.ProvisioningParametersDTO{
			NodeCount: ptr.Integer(2),
			Region:    ptr.String("europe-west4-a"),
		},
	}

	rawParameters, err := json.Marshal(parameters)
	if err != nil {
		t.Errorf("cannot marshal provisioning parameters: %s", err)
	}

	return string(rawParameters)
}

func fixInputCreator(t *testing.T) internal.ProvisionInputCreator {
	optComponentsSvc := &inputAutomock.OptionalComponentService{}

	optComponentsSvc.On("ComputeComponentsToDisable", []string(nil)).Return([]string{})
	optComponentsSvc.On("ExecuteDisablers", internal.ComponentConfigurationInputList{
		{
			Component:     "to-remove-component",
			Namespace:     "kyma-system",
			Configuration: nil,
		},
		{
			Component:     "keb",
			Namespace:     "kyma-system",
			Configuration: nil,
		},
		{
			Component:     input.ServiceManagerComponentName,
			Namespace:     "kyma-system",
			Configuration: nil,
		},
	}).Return(internal.ComponentConfigurationInputList{
		{
			Component:     "keb",
			Namespace:     "kyma-system",
			Configuration: nil,
		},
		{
			Component:     input.ServiceManagerComponentName,
			Namespace:     "kyma-system",
			Configuration: nil,
		},
	}, nil)

	kymaComponentList := []v1alpha1.KymaComponent{
		{
			Name:      "to-remove-component",
			Namespace: "kyma-system",
		},
		{
			Name:      "keb",
			Namespace: "kyma-system",
		},
		{
			Name:      input.ServiceManagerComponentName,
			Namespace: "kyma-system",
		},
	}
	ibf := input.NewInputBuilderFactory(optComponentsSvc, kymaComponentList, input.Config{}, kymaVersion)

	creator, found := ibf.ForPlan(broker.GcpPlanID)
	if !found {
		t.Errorf("input creator for %q plan does not exist", broker.GcpPlanID)
	}

	return creator
}
