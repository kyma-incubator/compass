package destinationfetchersvc_test

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/json"
	"encoding/pem"
	"fmt"
	"io"
	"math/big"
	"net"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/destinationfetchersvc"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/oauth"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	sensitiveEndpoint = "/destination-configuration/v1/destinations"
	syncEndpoint      = "/destination-configuration/v1/syncDestinations"
	subdomain         = "test"
	tokenPath         = "/test"
	noPageCountHeader = "noPageCount"
)

func TestClient_TenantEndpoint(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	mockServer := mockServerWithSyncEndpoint(t)
	defer mockServer.Close()

	apiConfig := destinationfetchersvc.DestinationServiceAPIConfig{
		GoroutineLimit:                10,
		RetryInterval:                 100 * time.Millisecond,
		RetryAttempts:                 3,
		EndpointGetTenantDestinations: mockServer.URL + syncEndpoint,
		EndpointFindDestination:       "",
		Timeout:                       100 * time.Millisecond,
		PageSize:                      100,
		PagingPageParam:               "$page",
		PagingSizeParam:               "$pageSize",
		PagingCountParam:              "$pageCount",
		PagingCountHeader:             "Page-Count",
		OAuthTokenPath:                tokenPath,
	}

	cert, key := generateTestCertAndKey(t, "test")
	instanceCfg := config.InstanceConfig{
		TokenURL: "https://subdomain.tokenurl",
	}
	instanceCfg.Cert = string(cert)
	instanceCfg.Key = string(key)
	client, err := destinationfetchersvc.NewClient(instanceCfg, apiConfig, subdomain)
	require.NoError(t, err)

	defer client.HTTPClient.CloseIdleConnections()
	setHTTPClientMockHost(client.HTTPClient, mockServer.URL)

	t.Run("Success fetching data page 3", func(t *testing.T) {
		// WHEN
		res, err := client.FetchTenantDestinationsPage(ctx, tenantID, "3")
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Success fetching data page but no Page-Count header is in response", func(t *testing.T) {
		// WHEN
		_, err := client.FetchTenantDestinationsPage(ctx, tenantID, noPageCountHeader)
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), fmt.Sprintf("missing '%s' header from destinations response",
			apiConfig.PagingCountHeader))
	})

	t.Run("Fetch should fail with status code 500, but do three attempts", func(t *testing.T) {
		// WHEN
		_, err := client.FetchTenantDestinationsPage(ctx, tenantID, "internalServerError")
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "status code 500")
	})

	t.Run("Fetch should fail with status code 4xx", func(t *testing.T) {
		// WHEN
		_, err := client.FetchTenantDestinationsPage(ctx, tenantID, "forbidden")
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "status code 403")
	})
}

func TestClient_SensitiveDataEndpoint(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	mockServer := mockServerWithDestinationEndpoint(t)
	defer mockServer.Close()

	apiConfig := destinationfetchersvc.DestinationServiceAPIConfig{}
	apiConfig.EndpointFindDestination = mockServer.URL + sensitiveEndpoint
	apiConfig.EndpointGetTenantDestinations = mockServer.URL + syncEndpoint
	apiConfig.RetryAttempts = 3
	apiConfig.RetryInterval = 100 * time.Millisecond
	apiConfig.OAuthTokenPath = tokenPath

	instanceCfg := config.InstanceConfig{
		TokenURL: "https://domain.tokenurl",
	}
	cert, key := generateTestCertAndKey(t, "test")
	instanceCfg.Cert = string(cert)
	instanceCfg.Key = string(key)
	client, err := destinationfetchersvc.NewClient(instanceCfg, apiConfig, subdomain)
	require.NoError(t, err)
	defer client.HTTPClient.CloseIdleConnections()
	setHTTPClientMockHost(client.HTTPClient, mockServer.URL)

	t.Run("Success fetching sensitive data", func(t *testing.T) {
		// WHEN
		res, err := client.FetchDestinationSensitiveData(ctx, "s4ext")
		// THEN
		require.NoError(t, err)
		assert.NotEmpty(t, res)
	})

	t.Run("Fetch should fail with status code 500, but do three attempts", func(t *testing.T) {
		// WHEN
		_, err := client.FetchDestinationSensitiveData(ctx, "internalServerError")
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "status code 500")
	})

	t.Run("NewNotFoundError should be returned for status 404", func(t *testing.T) {
		// WHEN
		_, err := client.FetchDestinationSensitiveData(ctx, "notFound")
		// THEN
		require.ErrorIs(t, err, apperrors.NewNotFoundError(resource.Destination, "notFound"))
	})

	t.Run("Error should be returned for status 400", func(t *testing.T) {
		// WHEN
		_, err := client.FetchDestinationSensitiveData(ctx, "badRequest")
		// THEN
		require.Error(t, err)
		require.Contains(t, err.Error(), "400")
	})
}

func setHTTPClientMockHost(client *http.Client, testServerURL string) {
	client.Transport = &http.Transport{
		DialContext: func(_ context.Context, network, address string) (net.Conn, error) {
			return net.Dial(network, strings.TrimPrefix(testServerURL, "https://"))
		},
		TLSClientConfig: &tls.Config{
			InsecureSkipVerify: true,
		},
	}
}

func mockServerWithSyncEndpoint(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc(syncEndpoint, func(w http.ResponseWriter, r *http.Request) {
		pageCount := r.URL.Query().Get("$pageCount")
		page := r.URL.Query().Get("$page")
		pageSize := r.URL.Query().Get("$pageSize")

		if page == "forbidden" {
			http.Error(w, "forbidden", http.StatusForbidden)
			return
		}

		if page != "3" && page != noPageCountHeader {
			http.Error(w, "page number invalid", http.StatusInternalServerError)
			return
		}

		if pageSize != "100" {
			http.Error(w, "pageSize invalid", http.StatusInternalServerError)
			return
		}

		if pageCount != "true" {
			http.Error(w, "pageCount invalid", http.StatusInternalServerError)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		if page != noPageCountHeader {
			w.Header().Set("Page-Count", "3")
		}

		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixTenantDestinationsEndpoint())
		require.NoError(t, err)
	})

	mux.HandleFunc(tokenPath, tokenEndpointHandler)
	return httptest.NewTLSServer(mux)
}

func tokenEndpointHandler(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")
	if _, err := w.Write(newToken()); err != nil {
		panic(err)
	}
}

func newToken() []byte {
	data := map[string]interface{}{}
	data["access_token"] = "MTQ0NjJkZmQ5OTM2NDE1ZTZjNGZmZjI3"
	data["token_type"] = "Bearer"
	data["expires_in"] = 3600
	token, err := json.Marshal(data)
	if err != nil {
		panic(err)
	}
	return token
}

func mockServerWithDestinationEndpoint(t *testing.T) *httptest.Server {
	mux := http.NewServeMux()

	mux.HandleFunc(sensitiveEndpoint+"/s4ext", func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_, err := io.WriteString(w, fixSensitiveDataJSON())
		require.NoError(t, err)
	})
	mux.HandleFunc(sensitiveEndpoint+"/internalServerError", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
	})

	mux.HandleFunc(sensitiveEndpoint+"/badRequest", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
	})

	mux.HandleFunc(tokenPath, tokenEndpointHandler)

	return httptest.NewTLSServer(mux)
}

func TestNewClient(t *testing.T) {
	const clientID = "client"
	const clientSecret = "secret"

	cert, key := generateTestCertAndKey(t, "test")

	t.Run("mtls+client-secret mode", func(t *testing.T) {
		instanceCfg := config.InstanceConfig{
			ClientID:     clientID,
			ClientSecret: clientSecret,
			URL:          "url",
			TokenURL:     "https://subdomain.tokenurl",
			Cert:         string(cert),
			Key:          string(key),
		}

		client, err := destinationfetchersvc.NewClient(instanceCfg,
			destinationfetchersvc.DestinationServiceAPIConfig{OAuthTokenPath: "/oauth/token"}, "subdomain")
		require.NoError(t, err)

		certCfg := oauth.X509Config{
			Cert: string(cert),
			Key:  string(key),
		}

		tlsCert, err := certCfg.ParseCertificate()
		require.NoError(t, err)

		expectedTransport := &http.Transport{
			TLSClientConfig: &tls.Config{
				Certificates: []tls.Certificate{*tlsCert},
			},
		}
		require.Equal(t, client.HTTPClient.Transport, expectedTransport)
	})

	t.Run("token url with no subdomain", func(t *testing.T) {
		instanceCfg := config.InstanceConfig{
			TokenURL: "https://nosubdomaintokenurl",
		}
		_, err := destinationfetchersvc.NewClient(instanceCfg,
			destinationfetchersvc.DestinationServiceAPIConfig{OAuthTokenPath: "/oauth/token"}, "subdomain")
		require.Error(t, err, fmt.Sprintf("auth url '%s' should have a subdomain", instanceCfg.TokenURL))
	})

	t.Run("invalid token url", func(t *testing.T) {
		instanceCfg := config.InstanceConfig{
			TokenURL: ":invalid",
		}
		_, err := destinationfetchersvc.NewClient(instanceCfg,
			destinationfetchersvc.DestinationServiceAPIConfig{OAuthTokenPath: "/oauth/token"}, "subdomain")
		require.Error(t, err)
		require.Contains(t, err.Error(), "failed to parse auth url")
	})
}

func fixSensitiveDataJSON() string {
	return `{
		"s4ext": {
      "owner": {
        "SubaccountId": "8fb6ac72-124e-11ed-861d-0242ac120002",
        "InstanceId": null
      },
      "destinationConfiguration": {
        "Name": "s4ext",
        "Type": "HTTP",
        "URL": "https://kaladin.bg",
        "Authentication": "BasicAuthentication",
        "ProxyType": "Internet",
        "XFSystemName": "Rock",
        "HTML5.DynamicDestination": "true",
        "User": "Kaladin",
        "product.name": "SAP S/4HANA Cloud",
        "Password": "securePass",
      },
      "authTokens": [
        {
          "type": "Basic",
          "value": "blJhbHQ1==",
          "http_header": {
            "key": "Authorization",
            "value": "Basic blJhbHQ1=="
          }
        }
      ]
    }
  }`
}

func fixTenantDestinationsEndpoint() string {
	return `
  [
    {
      "Name": "string",
      "Type": "HTTP",
      "PropertyName": "string"
    },
    {
      "Name": "string",
      "Type": "HTTP",
      "PropertyName": "string"
    }
  ]`
}

func generateTestCertAndKey(t *testing.T, commonName string) (crtPem, keyPem []byte) {
	clientKey, err := rsa.GenerateKey(rand.Reader, 2048)
	require.NoError(t, err)

	template := &x509.Certificate{
		IsCA:         true,
		SerialNumber: big.NewInt(1234),
		Subject: pkix.Name{
			CommonName: commonName,
		},
		NotBefore:   time.Now(),
		NotAfter:    time.Now().Add(time.Hour),
		KeyUsage:    x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		ExtKeyUsage: []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth},
	}

	parent := template
	certRaw, err := x509.CreateCertificate(rand.Reader, template, parent, &clientKey.PublicKey, clientKey)
	require.NoError(t, err)

	crtPem = pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: certRaw})
	keyPem = pem.EncodeToMemory(&pem.Block{Type: "RSA PRIVATE KEY", Bytes: x509.MarshalPKCS1PrivateKey(clientKey)})

	return
}
