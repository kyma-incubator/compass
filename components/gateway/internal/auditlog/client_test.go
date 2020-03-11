package auditlog_test

import (
	"encoding/json"
	"io"
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

//TODO:
//{"attributes":"The 'attributes' property cannot be null or empty"}<nil>+--- PASS: TestLogConfigurationChangeToAuditlog (1.11s)

const testTenant = "bfd679c3-aada-4af3-b8e2-74d710c4ed2e"

func TestLogConfigurationChangeToAuditlog(t *testing.T) {
	//GIVEN
	msgID := uuid.New().String()
	timestamp := time.Now().UTC()
	configChangeLog := fixConfigChangeLog()

	expectedLog := fixConfigChangeLog()
	expectedLog.UUID = msgID
	expectedLog.Time = timestamp.Format(auditlog.LogFormatDate)

	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer closeBody(t, r.Body)
			assert.Equal(t, r.URL.Path, "/audit-log/v2/configuration-changes")
			inputLog := readConfigChangeRequestBody(t, r)
			assert.Equal(t, expectedLog, inputLog)
			w.WriteHeader(http.StatusCreated)
		}))
		defer ts.Close()

		cfg := auditlog.AuditlogConfig{User: "user", Password: "password", AuditLogURL: ts.URL, Tenant: testTenant}

		uuidSvc := &automock.UUIDService{}
		uuidSvc.On("Generate").Return(msgID).Once()

		timeSvc := &automock.TimeService{}
		timeSvc.On("Now").Return(timestamp).Once()

		client := auditlog.NewClient(cfg, uuidSvc, timeSvc)

		//WHEN
		err := client.LogConfigurationChange(configChangeLog)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, uuidSvc, timeSvc)
	})

	t.Run("Response Code different than 201", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer closeBody(t, r.Body)
			assert.Equal(t, r.URL.Path, "/audit-log/v2/configuration-changes")
			inputLog := readConfigChangeRequestBody(t, r)
			assert.Equal(t, expectedLog, inputLog)
			w.WriteHeader(http.StatusForbidden)
		}))
		defer ts.Close()

		cfg := auditlog.AuditlogConfig{User: "user", Password: "password", AuditLogURL: ts.URL, Tenant: testTenant}

		uuidSvc := &automock.UUIDService{}
		uuidSvc.On("Generate").Return(msgID).Once()

		timeSvc := &automock.TimeService{}
		timeSvc.On("Now").Return(timestamp).Once()

		client := auditlog.NewClient(cfg, uuidSvc, timeSvc)

		//WHEN
		err := client.LogConfigurationChange(configChangeLog)

		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Write to auditlog failed with status code: 403")
		mock.AssertExpectationsForObjects(t, uuidSvc, timeSvc)
	})
}

func TestClient_LogSecurityEvent(t *testing.T) {
	//GIVEN
	msgID := uuid.New().String()
	timestamp := time.Now().UTC()
	securityEventLog := fixSecurityEventLog()

	expectedLog := fixSecurityEventLog()
	expectedLog.UUID = msgID
	expectedLog.Time = timestamp.Format(auditlog.LogFormatDate)
	expectedLog.Tenant = testTenant

	t.Run("Success", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer closeBody(t, r.Body)
			assert.Equal(t, r.URL.Path, "/audit-log/v2/security-events")
			inputLog := readSecurityEventRequestBody(t, r)
			assert.Equal(t, expectedLog, inputLog)
			w.WriteHeader(http.StatusCreated)
		}))
		defer ts.Close()

		cfg := auditlog.AuditlogConfig{User: "user", Password: "password", AuditLogURL: ts.URL, Tenant: testTenant}

		uuidSvc := &automock.UUIDService{}
		uuidSvc.On("Generate").Return(msgID).Once()

		timeSvc := &automock.TimeService{}
		timeSvc.On("Now").Return(timestamp).Once()

		client := auditlog.NewClient(cfg, uuidSvc, timeSvc)

		//WHEN
		err := client.LogSecurityEvent(securityEventLog)

		//THEN
		require.NoError(t, err)
		mock.AssertExpectationsForObjects(t, uuidSvc, timeSvc)
	})

	t.Run("Response Code different than 201", func(t *testing.T) {
		ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer closeBody(t, r.Body)
			assert.Equal(t, r.URL.Path, "/audit-log/v2/security-events")
			inputLog := readSecurityEventRequestBody(t, r)
			assert.Equal(t, expectedLog, inputLog)
			w.WriteHeader(http.StatusForbidden)
		}))
		defer ts.Close()

		cfg := auditlog.AuditlogConfig{User: "user", Password: "password", AuditLogURL: ts.URL, Tenant: testTenant}

		uuidSvc := &automock.UUIDService{}
		uuidSvc.On("Generate").Return(msgID).Once()

		timeSvc := &automock.TimeService{}
		timeSvc.On("Now").Return(timestamp).Once()

		client := auditlog.NewClient(cfg, uuidSvc, timeSvc)

		//WHEN
		err := client.LogSecurityEvent(securityEventLog)

		//THEN
		require.Error(t, err)
		assert.EqualError(t, err, "Write to auditlog failed with status code: 403")
		mock.AssertExpectationsForObjects(t, uuidSvc, timeSvc)
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

func fixConfigChangeLog() model.ConfigurationChange {
	return model.ConfigurationChange{
		User: "test-user",
		Object: model.Object{
			ID: map[string]string{
				"name":           "Config Change",
				"externalTenant": "external tenant",
				"apiConsumer":    "applicaiton",
				"consumerID":     "consumerID",
			},
			Type: "",
		},
		Attributes: []model.Attribute{{Name: "name", Old: "", New: "new value"}},
		Tenant:     testTenant,
	}
}

func fixSecurityEventLog() model.SecurityEvent {
	return model.SecurityEvent{
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

func closeBody(t *testing.T, body io.ReadCloser) {
	err := body.Close()
	require.NoError(t, err)
}
