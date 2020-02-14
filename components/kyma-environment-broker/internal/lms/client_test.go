package lms

import (
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"crypto/x509/pkix"

	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	token      = "token-lms"
	samlTenant = "saml.tenant.io"
	tenantID   = "tenant-id-001"
	host       = "lms.io"
)

func TestClient_CreateTenant(t *testing.T) {
	// given
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, "/tenants", r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, r.Header.Get("X-LMS-Token"), token)
		called = true

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"id":"%s"}`, tenantID)))
	}))
	defer ts.Close()

	client := createClient(ts.URL)

	// when
	out, err := client.CreateTenant(CreateTenantInput{
		Name:   "testing-name",
		Region: "us",
	})

	// then
	require.NoError(t, err)
	assert.True(t, called)
	assert.Equal(t, tenantID, out.ID)
}

func TestClient_GetTenantInfo(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, fmt.Sprintf("/tenants/%s", tenantID), r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, r.Header.Get("X-LMS-Token"), token)
		called = true

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"id": "%s","dns": "%s"}`, tenantID, host)))
	}))
	defer ts.Close()

	client := createClient(ts.URL)

	// when
	info, err := client.GetTenantInfo(tenantID)

	// then
	require.NoError(t, err)
	assert.Equal(t, host, info.DNS)
	assert.True(t, called)
}

func TestClient_GetCACertificate(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, fmt.Sprintf("/tenants/%s/certs/ca", tenantID), r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, r.Header.Get("X-LMS-Token"), token)
		called = true

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"cert": "cert content"}`))
	}))
	defer ts.Close()

	client := createClient(ts.URL)

	// when
	cert, found, err := client.GetCACertificate(tenantID)

	// then
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "cert content", cert)
	assert.True(t, called)

}

func TestClient_GetSignedCertificate(t *testing.T) {
	const certID = "cert-id"
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, fmt.Sprintf("/tenants/%s/certs/%s", tenantID, certID), r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, r.Header.Get("X-LMS-Token"), token)
		called = true

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{"cert": "cert content"}`))
	}))
	defer ts.Close()

	client := createClient(ts.URL)

	// when
	cert, found, err := client.GetSignedCertificate(tenantID, certID)

	// then
	require.NoError(t, err)
	assert.True(t, found)
	assert.Equal(t, "cert content", cert)
	assert.True(t, called)
}

func TestClient_RequestCertificate(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, fmt.Sprintf("/tenants/%s/certs", tenantID), r.URL.Path)
		assert.Equal(t, http.MethodPost, r.Method)
		assert.Equal(t, r.Header.Get("X-LMS-Token"), token)
		called = true

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(fmt.Sprintf(`{"id": "%s","dns": "%s"}`, tenantID, host)))
	}))
	defer ts.Close()

	client := createClient(ts.URL)

	// when
	subj := pkix.Name{
		CommonName:         "fluentbit", // do not modify
		Organization:       []string{"global-account-id"},
		OrganizationalUnit: []string{"sub-account-id"},
	}
	certID, pkey, err := client.RequestCertificate(tenantID, subj)

	// then
	require.NoError(t, err)
	assert.NotEmpty(t, certID)
	assert.NotEmpty(t, pkey)
	assert.True(t, called)
}

func TestClient_GetTenantStatus(t *testing.T) {
	called := false
	ts := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		assert.Equal(t, fmt.Sprintf("/tenants/%s/status", tenantID), r.URL.Path)
		assert.Equal(t, http.MethodGet, r.Method)
		assert.Equal(t, r.Header.Get("X-LMS-Token"), token)
		called = true

		w.WriteHeader(http.StatusOK)
		w.Write([]byte(`{
						  "kibanaDNSResolves": true,
						  "elasticsearchDNSResolves": true,
						  "kibanaState": "green"
						}`))
	}))
	defer ts.Close()

	client := createClient(ts.URL)

	// when
	status, err := client.GetTenantStatus(tenantID)

	// then
	require.NoError(t, err)
	assert.Equal(t, TenantStatus{
		KibanaDNSResolves:        true,
		ElasticsearchDNSResolves: true,
		KibanaState:              "green",
	}, status)
	assert.True(t, called)
}

func createClient(url string) Client {
	return NewClient(Config{
		URL:         url,
		ClusterType: ClusterTypeSingleNode,
		Environment: EnvironmentProd,
		SamlTenant:  samlTenant,
		Token:       token,
	}, logrus.StandardLogger())
}
