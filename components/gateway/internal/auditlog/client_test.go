package auditlog_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/automock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog"
	"github.com/kyma-incubator/compass/components/gateway/internal/auditlog/model"
)

const (
	configPath   = "/audit-log/v2/configuration-changes"
	securityPath = "/audit-log/v2/security-events"
)

func TestClient_LogConfigurationChange(t *testing.T) {
	//GIVEN
	configChangeMsg := fixFilledConfigChangeMsg()

	cfg := auditlog.Config{
		ConfigPath:   configPath,
		SecurityPath: securityPath,
	}

	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.Path, configPath)
			inputMsg := readConfigChangeRequestBody(t, r)
			assert.Equal(t, configChangeMsg, inputMsg)
			w.WriteHeader(http.StatusCreated)
		}))
		defer ts.Close()
		cfg.URL = ts.URL

		httpClient := &http.Client{}
		client, err := auditlog.NewClient(cfg, httpClient)
		require.NoError(t, err)

		//WHEN
		err = client.LogConfigurationChange(configChangeMsg)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Response Code different than 201", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.Path, configPath)
			inputMsg := readConfigChangeRequestBody(t, r)
			assert.Equal(t, configChangeMsg, inputMsg)
			w.WriteHeader(http.StatusForbidden)
		}))
		defer ts.Close()

		cfg.URL = ts.URL

		httpClient := &http.Client{}
		client, err := auditlog.NewClient(cfg, httpClient)
		require.NoError(t, err)

		//WHEN
		err = client.LogConfigurationChange(configChangeMsg)

		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Write to auditlog failed with status code: 403")
	})
}

func TestClient_LogSecurityEvent(t *testing.T) {
	//GIVEN
	securityEventMsg := fixFilledSecurityEventMsg()

	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.Path, securityPath)
			inputMsg := readSecurityEventRequestBody(t, r)
			assert.Equal(t, securityEventMsg, inputMsg)
			w.WriteHeader(http.StatusCreated)
		}))
		defer ts.Close()
		cfg := fixAuditlogConfig()
		cfg.URL = ts.URL

		httpClient := &http.Client{}
		client, err := auditlog.NewClient(cfg, httpClient)
		require.NoError(t, err)

		//WHEN
		err = client.LogSecurityEvent(securityEventMsg)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Success with tenant", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.Path, securityPath)
			inputMsg := readSecurityEventRequestBody(t, r)
			assert.Equal(t, securityEventMsg, inputMsg)
			w.WriteHeader(http.StatusCreated)
		}))
		defer ts.Close()
		cfg := fixAuditlogConfig()
		cfg.URL = ts.URL

		httpClient := &http.Client{}
		client, err := auditlog.NewClient(cfg, httpClient)
		require.NoError(t, err)

		//WHEN
		err = client.LogSecurityEvent(securityEventMsg)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Response Code different than 201", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.Path, securityPath)
			inputMsg := readSecurityEventRequestBody(t, r)
			assert.Equal(t, securityEventMsg, inputMsg)
			w.WriteHeader(http.StatusForbidden)
		}))
		defer ts.Close()

		cfg := fixAuditlogConfig()
		cfg.URL = ts.URL

		httpClient := &http.Client{}
		client, err := auditlog.NewClient(cfg, httpClient)
		require.NoError(t, err)

		//WHEN
		err = client.LogSecurityEvent(securityEventMsg)

		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Write to auditlog failed with status code: 403")
	})

	t.Run("http client return error", func(t *testing.T) {
		//GIVEN
		testErr := errors.New("test err")
		httpClient := &automock.HttpClient{}
		httpClient.On("Do", mock.Anything).Return(nil, testErr)

		cfg := fixAuditlogConfig()
		cfg.URL = "localhost:8080"
		client, err := auditlog.NewClient(cfg, httpClient)
		require.NoError(t, err)

		//WHEN
		err = client.LogSecurityEvent(securityEventMsg)

		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, fmt.Sprintf("while sending auditlog to: %s: %s", "localhost:8080", testErr.Error()))
	})
}

func TestDateFormat(t *testing.T) {
	//GIVEN
	expected := "2020-03-06T13:45:53.904Z"
	location, err := time.LoadLocation("UTC")
	require.NoError(t, err)
	timestamp := time.Date(2020, 03, 6, 13, 45, 53, 904000000, location)

	//WHEN
	formattedDate := timestamp.Format(auditlog.LogFormatDate)

	//THEN
	require.Equal(t, expected, formattedDate)
}

func readConfigChangeRequestBody(t *testing.T, r *http.Request) model.ConfigurationChange {
	output, err := ioutil.ReadAll(r.Body)
	require.NoError(t, err)
	var confChangeLog model.ConfigurationChange
	err = json.Unmarshal(output, &confChangeLog)
	require.NoError(t, err)
	return confChangeLog
}

func readSecurityEventRequestBody(t *testing.T, r *http.Request) model.SecurityEvent {
	output, err := ioutil.ReadAll(r.Body)
	require.NoError(t, err)
	var confChangeLog model.SecurityEvent
	err = json.Unmarshal(output, &confChangeLog)
	require.NoError(t, err)
	return confChangeLog
}

func fixAuditlogConfig() auditlog.Config {
	return auditlog.Config{
		ConfigPath:   configPath,
		SecurityPath: securityPath,
	}
}
