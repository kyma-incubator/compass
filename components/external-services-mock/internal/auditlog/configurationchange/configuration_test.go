package configurationchange_test

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"path"
	"sync"
	"testing"
	"time"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/external-services-mock/internal/auditlog/configurationchange"

	"github.com/gorilla/mux"
	"github.com/kyma-incubator/compass/components/gateway/pkg/auditlog/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	logID        = "879048d0-468e-49bb-b8b7-40e053218f0c"
	target       = "http://example.com"
	searchString = "test-msg"
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
			svc := configurationchange.NewService()
			input := testCase.input
			testCase.inputManipulation(&input)
			payload, err := json.Marshal(input)
			require.NoError(t, err)

			handler := configurationchange.NewConfigurationHandler(svc, nil)
			req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(payload))
			w := httptest.NewRecorder()

			//WHEN
			handler.Save(w, req)
			resp := w.Result()

			//THEN
			switch status := resp.StatusCode; status {
			case http.StatusCreated:
				{
					var response model.SuccessResponse
					err = unmarshallJson(resp.Body, &response)
					require.NoError(t, err)
					assert.Equal(t, logID, response.ID)
					saved := svc.Get(logID)
					assert.Equal(t, input, *saved)
				}
			case http.StatusBadRequest:
				{
					var response model.ErrorResponse
					err = unmarshallJson(resp.Body, &response)
					require.NoError(t, err)
					assert.Contains(t, response.Error, testCase.expectedErr)
				}
			default:
				t.Errorf("unexpected status code, got: %d", status)
			}

		})
	}
}

func TestConfigChangeHandler_Delete(t *testing.T) {
	//GIVEN
	input := fixConfigurationChange(logID)
	svc := configurationchange.NewService()
	_, err := svc.Save(input)
	require.NoError(t, err)

	handler := configurationchange.NewConfigurationHandler(svc, nil)

	endpointPath := path.Join("/", logID)
	req := httptest.NewRequest(http.MethodDelete, endpointPath, bytes.NewBuffer([]byte{}))
	vars := map[string]string{"id": logID}
	req = mux.SetURLVars(req, vars)

	//WHEN
	w := httptest.NewRecorder()
	handler.Delete(w, req)

	//THEN
	require.Equal(t, http.StatusOK, w.Code)
	require.Nil(t, svc.Get(logID))
}

func TestConfigChangeHandler_List(t *testing.T) {
	//GIVEN
	svc := configurationchange.NewService()
	input := fixConfigurationChange(logID)
	_, err := svc.Save(input)
	require.NoError(t, err)

	input = fixConfigurationChange("4020af04-7c7c-4b90-a410-967571e38bec")
	_, err = svc.Save(input)
	require.NoError(t, err)

	handler := configurationchange.NewConfigurationHandler(svc, nil)
	req := httptest.NewRequest(http.MethodGet, "http://localhost", bytes.NewBuffer([]byte{}))

	//WHEN
	w := httptest.NewRecorder()
	handler.List(w, req)
	//THEN
	require.Equal(t, http.StatusOK, w.Code)
	require.Len(t, svc.List(), 2)
}

func TestConfigChangeHandler_Get(t *testing.T) {
	//GIVEN
	svc := configurationchange.NewService()
	input := fixConfigurationChangeWithAttributes(logID)
	_, err := svc.Save(input)
	require.NoError(t, err)

	handler := configurationchange.NewConfigurationHandler(svc, nil)

	endpointPath := path.Join("/", logID)
	req := httptest.NewRequest(http.MethodGet, endpointPath, bytes.NewBuffer([]byte{}))
	vars := map[string]string{"id": logID}
	req = mux.SetURLVars(req, vars)

	//WHEN
	w := httptest.NewRecorder()
	handler.Get(w, req)

	//THEN
	require.Equal(t, w.Code, http.StatusOK)
	var response model.ConfigurationChange
	err = json.Unmarshal(w.Body.Bytes(), &response)
	require.NoError(t, err)
	assert.Equal(t, input, response)
}

func TestConfigChangeHandler_SearchByTimestamp_Success(t *testing.T) {
	//GIVEN
	svc := configurationchange.NewService()
	input := fixConfigurationChange(logID)
	_, err := svc.Save(input)
	require.NoError(t, err)

	startTime := time.Now()
	time.Sleep(10 * time.Millisecond)
	input = fixConfigurationChangeWithAttributes("4020af04-7c7c-4b90-a410-967571e38bec")
	_, err = svc.Save(input)
	require.NoError(t, err)

	requestURL := path.Join("/search")
	handler := configurationchange.NewConfigurationHandler(svc, nil)
	req := httptest.NewRequest(http.MethodGet, requestURL, bytes.NewBuffer([]byte{}))

	timeFrom := startTime.UTC()
	timeTo := startTime.Add(1 * time.Minute).UTC()

	timeFromStr := timeFrom.Format(time.RFC3339Nano)
	timeFromStr = timeFromStr[:len(timeFromStr)-1] // remove the 'Z' char from the time string

	timeToStr := timeTo.Format(time.RFC3339Nano)
	timeToStr = timeToStr[:len(timeToStr)-1] // remove the 'Z' char from the time string

	q := req.URL.Query()
	q.Add("time_from", timeFromStr)
	q.Add("time_to", timeToStr)
	req.URL.RawQuery = q.Encode()

	//WHEN
	w := httptest.NewRecorder()
	handler.SearchByTimestamp(w, req)

	//THEN
	require.Equal(t, http.StatusOK, w.Code)
	require.Len(t, svc.SearchByTimestamp(timeFrom, timeTo), 1)
}

func TestConfigChangeHandler_SearchByTimestamp_WhenMissingTimeFrom_Fails(t *testing.T) {
	//GIVEN
	svc := configurationchange.NewService()
	input := fixConfigurationChange(logID)
	_, err := svc.Save(input)
	require.NoError(t, err)

	startTime := time.Now()
	input = fixConfigurationChangeWithAttributes("4020af04-7c7c-4b90-a410-967571e38bec")
	_, err = svc.Save(input)
	require.NoError(t, err)

	requestURL := path.Join("/search")
	handler := configurationchange.NewConfigurationHandler(svc, nil)
	req := httptest.NewRequest(http.MethodGet, requestURL, bytes.NewBuffer([]byte{}))

	timeTo := startTime.Add(1 * time.Minute).UTC()

	timeToStr := timeTo.Format(time.RFC3339Nano)
	timeToStr = timeToStr[:len(timeToStr)-1] // remove the 'Z' char from the time string

	q := req.URL.Query()
	q.Add("time_to", timeToStr)
	req.URL.RawQuery = q.Encode()

	//WHEN
	w := httptest.NewRecorder()
	handler.SearchByTimestamp(w, req)

	//THEN
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConfigChangeHandler_SearchByTimestamp_WhenMissingTimeTo_Fails(t *testing.T) {
	//GIVEN
	svc := configurationchange.NewService()
	input := fixConfigurationChange(logID)
	_, err := svc.Save(input)
	require.NoError(t, err)

	startTime := time.Now()
	input = fixConfigurationChangeWithAttributes("4020af04-7c7c-4b90-a410-967571e38bec")
	_, err = svc.Save(input)
	require.NoError(t, err)

	requestURL := path.Join("/search")
	handler := configurationchange.NewConfigurationHandler(svc, nil)
	req := httptest.NewRequest(http.MethodGet, requestURL, bytes.NewBuffer([]byte{}))

	timeFrom := startTime.UTC()

	timeFromStr := timeFrom.Format(time.RFC3339Nano)
	timeFromStr = timeFromStr[:len(timeFromStr)-1] // remove the 'Z' char from the time string

	q := req.URL.Query()
	q.Add("time_from", timeFromStr)
	req.URL.RawQuery = q.Encode()

	//WHEN
	w := httptest.NewRecorder()
	handler.SearchByTimestamp(w, req)

	//THEN
	require.Equal(t, http.StatusBadRequest, w.Code)
}

func TestConcurrentCalls(t *testing.T) {
	//GIVEN
	var wg sync.WaitGroup
	var goroutinesCnt = 10

	svc := configurationchange.NewService()
	handler := configurationchange.NewConfigurationHandler(svc, nil)

	for i := 0; i < goroutinesCnt; i++ {
		wg.Add(1)
		go func(wg *sync.WaitGroup) {
			defer wg.Done()

			logID := uuid.New().String()
			input := fixConfigurationChange(logID)
			payload, err := json.Marshal(input)
			require.NoError(t, err)

			req := httptest.NewRequest(http.MethodPost, target, bytes.NewBuffer(payload))
			w := httptest.NewRecorder()

			//WHEN
			handler.Save(w, req)

			//THEN
			require.Equal(t, http.StatusCreated, w.Code)
			require.NotNil(t, svc.Get(logID))

			//WHEN
			endpointPath := path.Join("/", logID)
			req = httptest.NewRequest(http.MethodDelete, endpointPath, bytes.NewBuffer([]byte{}))
			vars := map[string]string{"id": logID}
			req = mux.SetURLVars(req, vars)

			w = httptest.NewRecorder()
			handler.Delete(w, req)

			//THEN
			require.Equal(t, http.StatusOK, w.Code)
			require.Nil(t, svc.Get(logID))

		}(&wg)
	}
	wg.Wait()
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

func fixConfigurationChangeWithAttributes(id string) model.ConfigurationChange {
	return model.ConfigurationChange{
		User:   "$USER",
		Object: model.Object{},
		Attributes: []model.Attribute{
			{
				Name: "test",
				Old:  "",
				New:  searchString,
			},
		},
		Success: nil,
		Metadata: model.Metadata{
			Time:   time.Now().Format(model.LogFormatDate),
			Tenant: "$PROVIDER",
			UUID:   id,
		},
	}
}
