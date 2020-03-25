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

	"github.com/google/uuid"
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
	msgID := uuid.New().String()
	timestamp := time.Date(2020, 3, 17, 12, 37, 44, 1093, time.FixedZone("test", 3600))
	configChangeLog := fixConfigChangeLog(msgID, timestamp)
	expectedLog := fixConfigChangeLog(msgID, timestamp)

	cfg := auditlog.Config{
		ConfigPath:   configPath,
		SecurityPath: securityPath,
	}

	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.Path, configPath)
			inputLog := readConfigChangeRequestBody(t, r)
			assert.Equal(t, expectedLog, inputLog)
			w.WriteHeader(http.StatusCreated)
		}))
		defer ts.Close()
		cfg.URL = ts.URL

		httpClient := &http.Client{}
		client, err := auditlog.NewClient(cfg, httpClient)
		require.NoError(t, err)

		//WHEN
		err = client.LogConfigurationChange(configChangeLog)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Response Code different than 201", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.Path, configPath)
			inputLog := readConfigChangeRequestBody(t, r)
			assert.Equal(t, expectedLog, inputLog)
			w.WriteHeader(http.StatusForbidden)
		}))
		defer ts.Close()

		cfg.URL = ts.URL

		httpClient := &http.Client{}
		client, err := auditlog.NewClient(cfg, httpClient)
		require.NoError(t, err)

		//WHEN
		err = client.LogConfigurationChange(configChangeLog)

		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Write to auditlog failed with status code: 403")
	})
}

func TestClient_LogSecurityEvent(t *testing.T) {
	//GIVEN
	msgID := uuid.New().String()
	timestamp := time.Now().UTC()
	securityEventLog := fixSecurityEventLog(msgID, timestamp)

	expectedLog := fixSecurityEventLog(msgID, timestamp)

	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.Path, securityPath)
			inputLog := readSecurityEventRequestBody(t, r)
			assert.Equal(t, expectedLog, inputLog)
			w.WriteHeader(http.StatusCreated)
		}))
		defer ts.Close()
		cfg := fixAuditlogConfig()
		cfg.URL = ts.URL

		httpClient := &http.Client{}
		client, err := auditlog.NewClient(cfg, httpClient)
		require.NoError(t, err)

		//WHEN
		err = client.LogSecurityEvent(securityEventLog)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Success with tenant", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.Path, securityPath)
			inputLog := readSecurityEventRequestBody(t, r)
			assert.Equal(t, expectedLog, inputLog)
			w.WriteHeader(http.StatusCreated)
		}))
		defer ts.Close()
		cfg := fixAuditlogConfig()
		cfg.URL = ts.URL

		httpClient := &http.Client{}
		client, err := auditlog.NewClient(cfg, httpClient)
		require.NoError(t, err)

		//WHEN
		err = client.LogSecurityEvent(securityEventLog)

		//THEN
		require.NoError(t, err)
	})

	t.Run("Response Code different than 201", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			assert.Equal(t, r.URL.Path, securityPath)
			inputLog := readSecurityEventRequestBody(t, r)
			assert.Equal(t, expectedLog, inputLog)
			w.WriteHeader(http.StatusForbidden)
		}))
		defer ts.Close()

		cfg := fixAuditlogConfig()
		cfg.URL = ts.URL

		httpClient := &http.Client{}
		client, err := auditlog.NewClient(cfg, httpClient)
		require.NoError(t, err)

		//WHEN
		err = client.LogSecurityEvent(securityEventLog)

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
		err = client.LogSecurityEvent(securityEventLog)

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

func fixConfigChangeLog(msgId string, timestamp time.Time) model.ConfigurationChange {
	return model.ConfigurationChange{
		User: "test-user",
		Object: model.Object{
			ID: map[string]string{
				"name":           "Config Change",
				"externalTenant": "external tenant",
				"apiConsumer":    "application",
				"consumerID":     "consumerID",
			},
			Type: "",
		},
		AuditlogMetadata: model.AuditlogMetadata{
			Time: timestamp.Format(auditlog.LogFormatDate),
			UUID: msgId,
		},
		Attributes: []model.Attribute{{Name: "name", Old: "", New: "new value"}},
	}
}

func fixSecurityEventLog(msgId string, timestamp time.Time) model.SecurityEvent {
	return model.SecurityEvent{
		AuditlogMetadata: model.AuditlogMetadata{
			Time: timestamp.Format(auditlog.LogFormatDate),
			UUID: msgId,
		},
		User: "test-user",
		Data: "test-data",
	}
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
