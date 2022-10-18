package resync_test

import (
	"context"
	"crypto/tls"
	"fmt"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/tenantfetchersvc/resync"
	"golang.org/x/oauth2"
	"golang.org/x/oauth2/clientcredentials"

	"github.com/kyma-incubator/compass/components/director/pkg/oauth"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestClient_FetchTenantEventsPage(t *testing.T) {
	// GIVEN
	ctx := context.TODO()
	mockClient, mockServerCloseFn, endpoint := fixHTTPClient(t)
	defer mockServerCloseFn()

	queryParams := resync.QueryParams{
		"pageSize":  "1",
		"pageNum":   "1",
		"timestamp": "1",
	}

	subaccountQueryParams := resync.QueryParams{
		"pageSize":  "1",
		"pageNum":   "1",
		"timestamp": "1",
		"region":    "test-region",
	}

	clientCfg := resync.ClientConfig{
		TenantProvider: "",
		APIConfig: resync.APIEndpointsConfig{
			EndpointTenantCreated:     endpoint + "/ga-created",
			EndpointTenantDeleted:     endpoint + "/ga-deleted",
			EndpointTenantUpdated:     endpoint + "/ga-updated",
			EndpointSubaccountCreated: endpoint + "/sub-created",
			EndpointSubaccountDeleted: endpoint + "/sub-deleted",
			EndpointSubaccountUpdated: endpoint + "/sub-updated",
			EndpointSubaccountMoved:   endpoint + "/sub-moved",
		},
		FieldMapping:        resync.TenantFieldMapping{},
		MovedSAFieldMapping: resync.MovedSubaccountsFieldMapping{},
	}

	client, err := resync.NewClient(resync.OAuth2Config{}, oauth.Standard, clientCfg, time.Second)
	require.NoError(t, err)

	client.SetHTTPClient(mockClient)

	t.Run("Success fetching account creation events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, resync.CreatedAccountType, queryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching account update events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, resync.UpdatedAccountType, queryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching account deletion events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, resync.DeletedAccountType, queryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching subaccount creation events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, resync.CreatedSubaccountType, subaccountQueryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching subaccount update events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, resync.UpdatedSubaccountType, subaccountQueryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching subaccount deletion events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, resync.DeletedSubaccountType, subaccountQueryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching moved subaccount events", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, resync.MovedSubaccountType, subaccountQueryParams)
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Error when unknown events type", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, -1, queryParams)
		// THEN
		require.EqualError(t, err, apperrors.NewInternalError("unknown events type").Error())
		assert.Empty(t, res)
	})

	// GIVEN
	clientCfg = resync.ClientConfig{
		TenantProvider: "",
		APIConfig: resync.APIEndpointsConfig{
			EndpointTenantCreated: "___ :// ___ ",
			EndpointTenantDeleted: "http://127.0.0.1:8111/badpath",
			EndpointTenantUpdated: endpoint + "/empty",
		},
		FieldMapping:        resync.TenantFieldMapping{},
		MovedSAFieldMapping: resync.MovedSubaccountsFieldMapping{},
	}

	client, err = resync.NewClient(resync.OAuth2Config{}, oauth.Standard, clientCfg, time.Second)
	require.NoError(t, err)

	client.SetHTTPClient(mockClient)

	t.Run("Success when no content", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, resync.UpdatedAccountType, queryParams)
		// THEN
		require.NoError(t, err)
		require.Empty(t, res)
	})

	t.Run("Error when endpoint not parsable", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, resync.CreatedAccountType, queryParams)
		// THEN
		require.EqualError(t, err, "parse \"___ :// ___ \": first path segment in URL cannot contain colon")
		assert.Empty(t, res)
	})

	t.Run("Error when bad path", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, resync.DeletedAccountType, queryParams)
		// THEN
		require.EqualError(t, err, "while sending get request: Get \"http://127.0.0.1:8111/badpath?pageNum=1&pageSize=1&timestamp=1\": dial tcp 127.0.0.1:8111: connect: connection refused")
		assert.Empty(t, res)
	})

	// GIVEN
	clientCfg = resync.ClientConfig{
		TenantProvider: "",
		APIConfig: resync.APIEndpointsConfig{
			EndpointTenantCreated: endpoint + "/created",
			EndpointTenantDeleted: endpoint + "/deleted",
			EndpointTenantUpdated: endpoint + "/badRequest",
		},
		FieldMapping:        resync.TenantFieldMapping{},
		MovedSAFieldMapping: resync.MovedSubaccountsFieldMapping{},
	}
	client, err = resync.NewClient(resync.OAuth2Config{}, oauth.Standard, clientCfg, time.Second)
	require.NoError(t, err)

	client.SetHTTPClient(mockClient)

	t.Run("Error when status code not equal to 200 OK and 204 No Content is returned", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, resync.UpdatedAccountType, queryParams)
		// THEN
		require.EqualError(t, err, fmt.Sprintf("request to \"%s/badRequest?pageNum=1&pageSize=1&timestamp=1\" returned status code 400 and body \"\"", endpoint))
		assert.Empty(t, res)
	})

	// GIVEN
	clientCfg = resync.ClientConfig{
		APIConfig: resync.APIEndpointsConfig{EndpointSubaccountMoved: ""},
	}
	client, err = resync.NewClient(resync.OAuth2Config{}, oauth.Standard, clientCfg, time.Second)
	require.NoError(t, err)

	client.SetHTTPClient(mockClient)

	t.Run("Skip fetching moved subaccount events when endpoint is not provided", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantEventsPage(ctx, resync.MovedSubaccountType, queryParams)
		// THEN
		require.NoError(t, err)
		require.Nil(t, res)
	})
}

func fixHTTPClient(t *testing.T) (*http.Client, func(), string) {
	mux := http.NewServeMux()

	mux.HandleFunc("/ga-created", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixCreatedTenantsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/ga-deleted", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixDeletedAccountsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/ga-updated", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixUpdatedAccountsJSON())
		require.NoError(t, err)
	})

	mux.HandleFunc("/sub-created", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixCreatedSubaccountsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/sub-deleted", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixDeletedSubaccountsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/sub-updated", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixUpdatedSubaccountsJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc("/sub-moved", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixMovedSubaccountsJSON())
		require.NoError(t, err)
	})

	mux.HandleFunc("/empty", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNoContent)
	})
	mux.HandleFunc("/badRequest", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	ts := httptest.NewServer(mux)

	return ts.Client(), ts.Close, ts.URL
}

func fixCreatedTenantsJSON() string {
	return `{
"events": [
  {
    "id": 5,
    "type": "GLOBALACCOUNT_CREATION",
    "timestamp": "1579771215736",
    "eventData": "{\"id\":\"55\",\"displayName\":\"TEN5\",\"model\":\"default\"}"
  },
  {
    "id": 4,
    "type": "GLOBALACCOUNT_CREATION",
    "timestamp": "1579771215636",
    "eventData": "{\"id\":\"44\",\"displayName\":\"TEN4\",\"model\":\"default\"}"
  },
	{
    "id": 3,
    "type": "GLOBALACCOUNT_CREATION",
    "timestamp": "1579771215536",
    "eventData": "{\"id\":\"33\",\"displayName\":\"TEN3\",\"model\":\"default\"}"
  },
	{
    "id": 2,
    "type": "GLOBALACCOUNT_CREATION",
    "timestamp": "1579771215436",
    "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
  },
	{
    "id": 1,
    "type": "GLOBALACCOUNT_CREATION",
    "timestamp": "1579771215336",
    "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
  }
],
"totalResults": 5,
"totalPages": 1
}`
}

func fixUpdatedAccountsJSON() string {
	return `{
"events": [
	{
    "id": 2,
    "type": "GLOBALACCOUNT_UPDATE",
    "timestamp": "1579771215436",
    "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
  },
	{
    "id": 1,
    "type": "GLOBALACCOUNT_UPDATE",
    "timestamp": "1579771215336",
    "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
  }
],
"totalResults": 2,
"totalPages": 1
}`
}

func fixDeletedAccountsJSON() string {
	return `{
"events": [
	{
    "id": 2,
    "type": "GLOBALACCOUNT_DELETION",
    "timestamp": "1579771215436",
    "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
  },
	{
    "id": 1,
    "type": "GLOBALACCOUNT_DELETION",
    "timestamp": "1579771215336",
    "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
  }
],
"totalResults": 2,
"totalPages": 1
}`
}

func fixCreatedSubaccountsJSON() string {
	return `{
"events": [
  {
    "id": 5,
    "type": "SUBACCOUNT_CREATION",
	 "region": "test-region",
    "timestamp": "1579771215736",
    "eventData": "{\"id\":\"55\",\"displayName\":\"TEN5\",\"model\":\"default\"}"
  },
  {
    "id": 4,
    "type": "SUBACCOUNT_CREATION",
	 "region": "test-region",
    "timestamp": "1579771215636",
    "eventData": "{\"id\":\"44\",\"displayName\":\"TEN4\",\"model\":\"default\"}"
  },
	{
    "id": 3,
    "type": "SUBACCOUNT_CREATION",
	 "region": "test-region",
    "timestamp": "1579771215536",
    "eventData": "{\"id\":\"33\",\"displayName\":\"TEN3\",\"model\":\"default\"}"
  },
	{
    "id": 2,
    "type": "SUBACCOUNT_CREATION",
	 "region": "test-region",
    "timestamp": "1579771215436",
    "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
  },
	{
    "id": 1,
    "type": "SUBACCOUNT_CREATION",
	 "region": "test-region",
    "timestamp": "1579771215336",
    "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
  }
],
"totalResults": 5,
"totalPages": 1
}`
}

func fixUpdatedSubaccountsJSON() string {
	return `{
"events": [
	{
    "id": 2,
    "type": "SUBACCOUNT_UPDATE",
	 "region": "test-region",
    "timestamp": "1579771215436",
    "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
  },
	{
    "id": 1,
    "type": "SUBACCOUNT_UPDATE",
	 "region": "test-region",
    "timestamp": "1579771215336",
    "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
  }
],
"totalResults": 2,
"totalPages": 1
}`
}

func fixDeletedSubaccountsJSON() string {
	return `{
"events": [
	{
    "id": 2,
    "type": "SUBACCOUNT_DELETION",
	 "region": "test-region",
    "timestamp": "1579771215436",
    "eventData": "{\"id\":\"22\",\"displayName\":\"TEN2\",\"model\":\"default\"}"
  },
	{
    "id": 1,
    "type": "SUBACCOUNT_DELETION",
	 "region": "test-region",
    "timestamp": "1579771215336",
    "eventData": "{\"id\":\"11\",\"displayName\":\"TEN1\",\"model\":\"default\"}"
  }
],
"totalResults": 2,
"totalPages": 1
}`
}

func fixMovedSubaccountsJSON() string {
	return `{
"events": [
	{
    "id": 2,
    "type": "SUBACCOUNT_MOVED",
	 "region": "test-region",
    "timestamp": "1579771215436",
    "eventData": "{\"id\":\"22\",\"source\":\"TEN1\",\"target\":\"TEN2\"}"
  },
	{
    "id": 1,
    "type": "SUBACCOUNT_MOVED",
	 "region": "test-region",
    "timestamp": "1579771215336",
    "eventData": "{\"id\":\"11\",\"source\":\"TEN3\",\"target\":\"TEN4\"}"
  }
],
"totalResults": 2,
"totalPages": 1
}`
}

func TestNewClient(t *testing.T) {
	const clientID = "client"
	const clientSecret = "secret"

	t.Run("expect error on invalid auth mode", func(t *testing.T) {
		_, err := resync.NewClient(resync.OAuth2Config{}, "invalid-auth-mode", resync.ClientConfig{}, 1)
		require.Error(t, err)
	})

	t.Run("standard client-credentials mode", func(t *testing.T) {
		client, err := resync.NewClient(resync.OAuth2Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}, oauth.Standard, resync.ClientConfig{}, 1)
		require.NoError(t, err)

		httpClient := client.GetHTTPClient()
		tr, ok := httpClient.Transport.(*oauth2.Transport)
		require.True(t, ok, "expected *oauth2.Transport")
		require.Equal(t, http.DefaultClient.Transport, tr.Base)

		cfg := clientcredentials.Config{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}
		expectedTokenSrc := cfg.TokenSource(context.Background())
		require.Equal(t, expectedTokenSrc, tr.Source)
	})

	t.Run("mtls+client-secret mode", func(t *testing.T) {
		const certificate = "-----BEGIN CERTIFICATE-----\nMIIDbjCCAlYCCQDg7pmtw8dIVTANBgkqhkiG9w0BAQsFADB5MQswCQYDVQQGEwJC\nRzENMAsGA1UECAwEVGVzdDENMAsGA1UEBwwEVGVzdDENMAsGA1UECgwEVGVzdDEN\nMAsGA1UECwwEVGVzdDENMAsGA1UEAwwEVGVzdDEfMB0GCSqGSIb3DQEJARYQdGVz\ndEBleGFtcGxlLmNvbTAeFw0yMjAxMjQxMTM4MDFaFw0zMjAxMjIxMTM4MDFaMHkx\nCzAJBgNVBAYTAkJHMQ0wCwYDVQQIDARUZXN0MQ0wCwYDVQQHDARUZXN0MQ0wCwYD\nVQQKDARUZXN0MQ0wCwYDVQQLDARUZXN0MQ0wCwYDVQQDDARUZXN0MR8wHQYJKoZI\nhvcNAQkBFhB0ZXN0QGV4YW1wbGUuY29tMIIBIjANBgkqhkiG9w0BAQEFAAOCAQ8A\nMIIBCgKCAQEAuiFt98GUVTDSCHsOlBcblvUB/02uEmsalsG+DKEufzIVrp4DCxsA\nEsIN85Ywkd1Fsl0vwg9+3ibQlf1XtyXqJ6/jwm2zFdJPM3u2JfGGiiQpscHYp5hS\nlVscBjxZh1CQMKeBXltDsD64EV+XgHGN1aaw9mWKb6iSKsHLhBz594jYMFCnP3wH\nw9/hm6zBAhoF4Xr6UMOp4ZzzY8nzLCGPQuQ9UGp4lyAethrBpsqI6zAxjPKlqhmx\nL3591wkQgTzuL9th54yLEmyEvPTE26ONJBKylH2BqbAFiZPrwet0+PRJSflAfMU8\nYHqqo2AkaY1lmMAZiKDhj1RxMe/jt3HmVQIDAQABMA0GCSqGSIb3DQEBCwUAA4IB\nAQBx8BRhJ59UA3JDL+FHNKwIpxFewxjJwIGWqJTsOh4+rjPK3QeSnF0vt4cnLrCY\n+FLuhhUdFxjeFqJtWN7tHDK3ywSn/yZQTD5Nwcy/F1RmLjl91hjudxO/VewznOlq\nHJlDoM7kW9kOG6xS2HbbSaC1CzU33E90QOwcyCoeVXJ8aMDe6v/kWC65RoI9evg5\n2OxoARA8fpjyUphMTXuVNVI1kd2Uskpo8PePbc1h3OJVzYPIQ4+qMGsu7n3ZdwzI\nqDs2kdBD77k6cBQS+n7g5ETwv5OAgl5q1O17ye/YFNA/T3FhL9to6Nmrkqt7rlnF\nL8uAkeTGuHEATjmosQWUmbYi\n-----END CERTIFICATE-----\n"
		const key = "-----BEGIN RSA PRIVATE KEY-----\nMIIEowIBAAKCAQEAuiFt98GUVTDSCHsOlBcblvUB/02uEmsalsG+DKEufzIVrp4D\nCxsAEsIN85Ywkd1Fsl0vwg9+3ibQlf1XtyXqJ6/jwm2zFdJPM3u2JfGGiiQpscHY\np5hSlVscBjxZh1CQMKeBXltDsD64EV+XgHGN1aaw9mWKb6iSKsHLhBz594jYMFCn\nP3wHw9/hm6zBAhoF4Xr6UMOp4ZzzY8nzLCGPQuQ9UGp4lyAethrBpsqI6zAxjPKl\nqhmxL3591wkQgTzuL9th54yLEmyEvPTE26ONJBKylH2BqbAFiZPrwet0+PRJSflA\nfMU8YHqqo2AkaY1lmMAZiKDhj1RxMe/jt3HmVQIDAQABAoIBAH+9xa0N6/FzqhIr\n8ltsaID38cD33QnC++KPYRFl5XViOEM5KrmKdEhragvM/dR92gGJtucmn1lzph/q\nWTLXEJbgPh4ID6pgRf79Xos38bAJFZxrf3e2MKdUei1FaeRWRD9AFqddV100DjvO\nMTnztPX2iujv00zCkl5J1pT7FgrtcYgDPxXQK7dIcHrc9bV9fdTQUnpbVIs/9U7a\n7Qk/eJnEkezbjQCk7+Pgt3ymR29s4vJvyPen3jek0FKhQCxAg6iA5ZOtY+J5AS9e\n3ozZLUEa3b0eOABMw8QnKMtGTmIhLbf9JhISK2Ltsisc/yHHH3KfFE2nayqjvLZf\n5GR62hkCgYEA612EgoRHg4+BSfPfLNG3xsSnM+a98nZOmyxgZ3eNFWpSvi+7MemL\nCJHpwwje412OU1wCc2MtWYvGFY+heL62FxT8+JJLntykZcTQzQoHX3wvaMwopWRi\nJdrv3tEDtSJo9za54kfrNqnVyaxu82r7zgxVbcNiAVR+n7cRXuov288CgYEAynLm\nVI7cIKBOM6U44unkKyIS99Bh57FPjE1QAIsEOiNCWZay4qmzdEboOXjtC95Qyyxn\nTb+MONybwXKkGiLZQZQ2SlgjtEMBDQ+ofk2fK+yHWf4VeLtYWJdBESaAz85xGCCY\nYqlqbFEQd8cl86gTne+emLXp8KrDMuXhbbPvMJsCgYEAgBISAacS9t6GfoQqA0xW\nkNz/EnnTD/UaTst15bci2O1S+tQkK0OmeNJU/eB80AFfabKeTsU/rwMklSTjuz0i\n/ipYgLWyWk47UnknGPsFCgscDQ1SbLTTxz972KWpO83uid6IhT2XGtaNU0D12pRz\nUipZ7fEsCgc9I5FM7XXG9vcCgYBp6xN2ygeBSl2fx6GrlpM5veoOnYeboLjtvsVM\ng28Cu8/K731H+WFaRH7bEtlyjC3ZHrItiznhxgn3e/M/eVwRY2nEG7kSZrv2CWsu\nKY5NfMKT4st5Dwt5zijMwEhEcM3awbL4a4qygPcMs7S3dghNaUCgxQxQTgcyafM3\nYhySYQKBgF7pqQW7ESo1Mp9by+HzJBJsSju5zPBrCZrx8rFAMLCk1uDAIRcUuQtq\n+YwKU8ViemkOHWfN6bePap3/kdVHUxj2xJ6xTAUYHpVOQVMhTw1UmOikiV4FwUo+\nGb5Nk5evWBGhsl2LFqoOqhvFpjftv8+qgRHxmWtj4EoJYWng+hRz\n-----END RSA PRIVATE KEY-----\n"

		certCfg := resync.X509Config{
			Cert: certificate,
			Key:  key,
		}

		tlsCert, err := certCfg.ParseCertificate()
		require.NoError(t, err)

		oauthCfg := resync.OAuth2Config{
			X509Config: certCfg,
			ClientID:   clientID,
		}
		client, err := resync.NewClient(oauthCfg, oauth.Mtls, resync.ClientConfig{}, 1)
		require.NoError(t, err)

		httpClient := client.GetHTTPClient()
		tr, ok := httpClient.Transport.(*oauth2.Transport)
		require.True(t, ok, "expected *oauth2.Transport")

		expectedTransport := &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates:       []tls.Certificate{*tlsCert},
				InsecureSkipVerify: oauthCfg.SkipSSLValidation,
			},
		}
		require.Equal(t, tr.Base, expectedTransport)
	})
}
