package configuration_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"testing"
	"time"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/auditlog/configuration"
	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	logID      = "879048d0-468e-49bb-b8b7-40e053218f0c"
	target     = "http://example.com/"
	configPath = "/auditlog/v2/configuration-changes"
)

func TestConfigChangeHandler_Save(t *testing.T) {
	testCases := []struct {
		Name              string
		input             model.ConfigurationChange
		inputManipulation func(change *model.ConfigurationChange)
		expectedErr       string
	}{
		{
			Name:              "Success",
			input:             fixConfigurationChange(logID),
			inputManipulation: func(change *model.ConfigurationChange) {},
		}, {
			Name:  "Invalid user",
			input: fixConfigurationChange(logID),
			inputManipulation: func(change *model.ConfigurationChange) {
				change.User = "Invalid"
			},
			expectedErr: "User is not valid",
		}, {
			Name:  "Invalid tenant",
			input: fixConfigurationChange(logID),
			inputManipulation: func(change *model.ConfigurationChange) {
				change.Tenant = "Invalid"
			},
			expectedErr: "Tenant is not valid",
		}, {
			Name:  "Invalid date",
			input: fixConfigurationChange(logID),
			inputManipulation: func(change *model.ConfigurationChange) {
				change.Metadata.Time = time.Now().Format(time.Kitchen)
			},
			expectedErr: "Time is not valid",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			svc := configuration.NewService()
			input := testCase.input
			testCase.inputManipulation(&input)
			payload, err := json.Marshal(input)
			require.NoError(t, err)

			handler := configuration.NewConfigurationHandler(svc, nil)
			req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(payload))
			w := httptest.NewRecorder()

			//WHEN
			handler.Save(w, req)
			resp := w.Result()

			//THEN
			if testCase.expectedErr == "" {
				assert.Equal(t, resp.StatusCode, http.StatusCreated)
				var response model.SuccessResponse
				err = unmarshallJson(resp.Body, &response)
				require.NoError(t, err)
				assert.Equal(t, logID, response.ID)
				saved := svc.Get(logID)
				assert.Equal(t, input, *saved)
			} else {
				assert.Equal(t, http.StatusBadRequest, resp.StatusCode)
				var response model.ErrorResponse
				err = unmarshallJson(resp.Body, &response)
				require.NoError(t, err)
				assert.Contains(t, response.Error, testCase.expectedErr)
			}
		})
	}
}

func TestConfigChangeHandler_Delete(t *testing.T) {
	//GIVEN
	input := fixConfigurationChange(logID)
	svc := configuration.NewService()
	_, err := svc.Save(input)
	require.NoError(t, err)

	handler := configuration.NewConfigurationHandler(svc, nil)
	router := mux.NewRouter()
	router.HandleFunc(configPath+"/{id}", handler.Delete).Methods(http.MethodDelete)

	endpointPath := path.Join(configPath, logID)
	req := httptest.NewRequest(http.MethodDelete, endpointPath, bytes.NewBuffer([]byte{}))
	//WHEN
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	//THEN
	require.Equal(t, w.Code, http.StatusOK)
	require.Nil(t, svc.Get(logID))
}

func TestConfigChangeHandler_List(t *testing.T) {
	//GIVEN
	svc := configuration.NewService()
	input := fixConfigurationChange(logID)
	_, err := svc.Save(input)
	require.NoError(t, err)

	input = fixConfigurationChange("4020af04-7c7c-4b90-a410-967571e38bec")
	_, err = svc.Save(input)
	require.NoError(t, err)

	handler := configuration.NewConfigurationHandler(svc, nil)
	router := mux.NewRouter()
	router.HandleFunc(configPath, handler.List).Methods(http.MethodGet)

	req := httptest.NewRequest(http.MethodGet, configPath, bytes.NewBuffer([]byte{}))
	//WHEN
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	//THEN
	require.Equal(t, w.Code, http.StatusOK)
	require.Len(t, svc.List(), 2)
}

func TestConfigChangeHandler_Get(t *testing.T) {
	//GIVEN
	svc := configuration.NewService()
	input := fixConfigurationChange(logID)
	_, err := svc.Save(input)
	require.NoError(t, err)

	handler := configuration.NewConfigurationHandler(svc, nil)
	router := mux.NewRouter()
	router.HandleFunc(configPath+"/{id}", handler.Get).Methods(http.MethodGet)

	endpointPath := path.Join(configPath, logID)
	req := httptest.NewRequest(http.MethodGet, endpointPath, bytes.NewBuffer([]byte{}))
	//WHEN
	w := httptest.NewRecorder()
	router.ServeHTTP(w, req)

	//THEN
	require.Equal(t, w.Code, http.StatusOK)
	var response model.ConfigurationChange
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, input, response)
}

func unmarshallJson(r io.Reader, target interface{}) error {
	e := json.NewDecoder(r)
	e.DisallowUnknownFields()
	return e.Decode(&target)
}

func fixConfigurationChange(id string) model.ConfigurationChange {
	return model.ConfigurationChange{
		User:       "$USER",
		Object:     model.Object{},
		Attributes: nil,
		Success:    nil,
		Metadata: model.Metadata{
			Time:   time.Now().Format(model.LogFormatDate),
			Tenant: "$PROVIDER",
			UUID:   id,
		},
	}
}
