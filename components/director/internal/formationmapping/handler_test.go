package formationmapping_test

import (
	"bytes"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"

	"github.com/gorilla/mux"
	fm "github.com/kyma-incubator/compass/components/director/internal/formationmapping"
	"github.com/stretchr/testify/require"
)

func Test_StatusUpdate(t *testing.T) {
	testFormationID := "testFormationID"
	testFormationAssignmentID := "testFormationAsisgnmentID"
	url := fmt.Sprintf("/v1/businessIntegrations/{%s}/assignments/{%s}/status", fm.FormationIDParam, fm.FormationAssignmentIDParam)
	testValidConfig := `{"testK":"testV"}`
	urlVars := map[string]string{
		fm.FormationIDParam:           testFormationID,
		fm.FormationAssignmentIDParam: testFormationAssignmentID,
	}

	handler := fm.NewFormationMappingHandler()

	testCases := []struct {
		name               string
		reqBody            fm.RequestBody
		hasURLVars         bool
		headers            map[string][]string
		expectedStatusCode int
		expectedErrOutput  string
	}{
		// Request(+metadata) validation checks
		{
			name:               "Decode Error: Content-Type header is not application/json",
			headers:            map[string][]string{httputils.HeaderContentTypeKey: {"invalidContentType"}},
			expectedStatusCode: http.StatusUnsupportedMediaType,
			expectedErrOutput:  "Content-Type header is not application/json",
		},
		{
			name: "Error when one or more of the required path parameters are missing",
			reqBody: fm.RequestBody{
				State:         fm.ConfigurationStateReady,
				Configuration: json.RawMessage(testValidConfig),
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Not all of the required parameters are provided",
		},
		// Request body validation checks
		{
			name: "Success",
			reqBody: fm.RequestBody{
				State:         fm.ConfigurationStateReady,
				Configuration: json.RawMessage(testValidConfig),
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusOK,
			expectedErrOutput:  "",
		},
		{
			name: "Validate Error: error when we have ready state with config but also an error provided",
			reqBody: fm.RequestBody{
				State:         fm.ConfigurationStateReady,
				Configuration: json.RawMessage(testValidConfig),
				Error:         "testErrMsg",
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Request Body contains invalid input:",
		},
		{
			name: "Validate Error: error when configuration is provided but the state is incorrect",
			reqBody: fm.RequestBody{
				State:         fm.ConfigurationStateCreateError,
				Configuration: json.RawMessage(testValidConfig),
			},
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Request Body contains invalid input:",
		},
		{
			name: "Validate Error: error when request body contains only state",
			reqBody: fm.RequestBody{
				State: fm.ConfigurationStateReady,
			},
			hasURLVars:         true,
			expectedStatusCode: http.StatusBadRequest,
			expectedErrOutput:  "Request Body contains invalid input:",
		},
	}

	for _, tCase := range testCases {
		t.Run(tCase.name, func(t *testing.T) {
			marshalBody, err := json.Marshal(tCase.reqBody)
			require.NoError(t, err)

			httpReq := httptest.NewRequest(http.MethodPatch, url, bytes.NewBuffer(marshalBody))
			if tCase.hasURLVars {
				httpReq = mux.SetURLVars(httpReq, urlVars)
			}

			if tCase.headers != nil {
				httpReq.Header = tCase.headers
			}
			w := httptest.NewRecorder()

			// WHEN
			handler.UpdateStatus(w, httpReq)

			// THEN
			resp := w.Result()
			body, err := io.ReadAll(resp.Body)
			require.NoError(t, err)

			if tCase.expectedErrOutput == "" {
				require.NoError(t, err)
			} else {
				require.Contains(t, string(body), tCase.expectedErrOutput)
			}
			require.Equal(t, tCase.expectedStatusCode, resp.StatusCode)
		})
	}
}
