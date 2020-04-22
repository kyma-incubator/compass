// +build lms_integration

package lms

import (
	"crypto/x509/pkix"
	"os"
	"testing"

	"github.com/sirupsen/logrus"
)

// This file contains tests to test real calls to LMS. It must not be run in the `make verify` stage.
// The purpose of the test is to perform calls during development cycle
// Run the test with the following command:
//  go test -v -tags=lms_integration ./internal/lms/... -run TestCreateTenant
// before running, set the following envs:
// - URL - base URL to the LMS
// - TOKEN - lms Token
// In the log you can see Tenant ID, which should be saved as env to run next tests:
// export TOKEN_ID=<token>
func TestCreateTenant(t *testing.T) {
	url := os.Getenv("URL")
	token := os.Getenv("TOKEN")

	c := NewClient(Config{
		ClusterType: ClusterTypeSingleNode,
		Token:       token,
		Environment: EnvironmentDev,
		SamlTenant:  "ycloud.accounts.ondemand.com",
		URL:         url,
	}, logrus.StandardLogger())

	output, err := c.CreateTenant(CreateTenantInput{
		Region: "eu",
		Name:   "kymatest32109ooo87675abcdefgh",
	})

	t.Log(err)
	t.Logf("%+v", output)

	s, err := c.GetTenantStatus(output.ID)
	t.Log(s)
	t.Log(err)
}

// export TENANT_ID=<tenant id>
// go test -v -tags=lms_integration ./internal/lms/... -run TestTenantStatus
func TestTenantStatus(t *testing.T) {
	// set the correct Tenant ID
	tID := os.Getenv("TENANT_ID")

	url := os.Getenv("URL")
	token := os.Getenv("TOKEN")

	c := NewClient(Config{
		ClusterType: ClusterTypeSingleNode,
		Token:       token,
		Environment: EnvironmentDev,
		SamlTenant:  "ycloud.accounts.ondemand.com",
		URL:         url,
	}, logrus.StandardLogger())

	s, err := c.GetTenantStatus(tID)

	t.Logf("%+v\n%s", s, err)
}

// export TENANT_ID=<tenant id>
// go test -v -tags=lms_integration ./internal/lms/... -run TestGenerateCsr
func TestGenerateCsr(t *testing.T) {
	tID := os.Getenv("TENANT_ID")

	url := os.Getenv("URL")
	token := os.Getenv("TOKEN")

	c := NewClient(Config{
		ClusterType: ClusterTypeSingleNode,
		Token:       token,
		Environment: EnvironmentDev,
		SamlTenant:  "ycloud.accounts.ondemand.com",
		URL:         url,
	}, logrus.StandardLogger())

	subj := pkix.Name{
		CommonName:         "fluentbit", // do not modify
		Organization:       []string{"global-account-id"},
		OrganizationalUnit: []string{"sub-account-id"},
	}
	id, resp, err := c.RequestCertificate(tID, subj)
	t.Logf("CERT ID: %s", id)
	t.Log(string(resp))
	t.Log(err)
}

// export TENANT_ID=<tenant id>
// export CERT_ID=<cert_id>
// go test -v -tags=lms_integration ./internal/lms/... -run TestGetCert
func TestGetCert(t *testing.T) {
	tID := os.Getenv("TENANT_ID")
	certID := os.Getenv("CERT_ID")

	url := os.Getenv("URL")
	token := os.Getenv("TOKEN")

	c := NewClient(Config{
		ClusterType: ClusterTypeSingleNode,
		Token:       token,
		Environment: EnvironmentDev,
		SamlTenant:  "ycloud.accounts.ondemand.com",
		URL:         url,
	}, logrus.StandardLogger())

	signedCert, found, err := c.GetSignedCertificate(tID, certID)
	t.Logf("Found: %v", found)
	t.Logf(string(signedCert))
	t.Log(err)
}
