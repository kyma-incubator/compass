package oathkeeper_test

import (
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/oathkeeper"

	"github.com/kyma-incubator/compass/components/connector/internal/certificates"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

var (
	csrSubjectConsts = certificates.CSRSubjectConsts{
		Country:            "DE",
		Organization:       "organization",
		OrganizationalUnit: "OrgUnit",
		Locality:           "Waldorf",
		Province:           "Waldorf",
	}
)

func TestParseCertHeader(t *testing.T) {

	t.Run("should return valid common name and hash", func(t *testing.T) {
		//given
		r, err := http.NewRequest("GET", "", nil)
		require.NoError(t, err)

		r.Header.Set(oathkeeper.ClientCertHeader, "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=test-application,OU=OrgUnit,O=organization,L=Waldorf,ST=Waldorf,C=DE\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;"+
			"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account")

		hp := oathkeeper.NewHeaderParser(csrSubjectConsts)

		//when
		commonName, hash, found := hp.GetCertificateData(r)

		//then
		require.True(t, found)
		assert.Equal(t, "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad", hash)
		assert.Equal(t, "test-application", commonName)
	})

	t.Run("should not found certificate data if non is matching", func(t *testing.T) {
		//given
		r, err := http.NewRequest("GET", "", nil)
		require.NoError(t, err)

		r.Header.Set(oathkeeper.ClientCertHeader, "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;"+
			"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account")

		hp := oathkeeper.NewHeaderParser(csrSubjectConsts)

		//when
		commonName, hash, found := hp.GetCertificateData(r)

		//then
		require.False(t, found)
		assert.Empty(t, commonName)
		assert.Empty(t, hash)
	})

	t.Run("should not found certificate data if header is invalid", func(t *testing.T) {
		//given
		r, err := http.NewRequest("GET", "", nil)
		require.NoError(t, err)

		r.Header.Set(oathkeeper.ClientCertHeader, "invalid header")

		hp := oathkeeper.NewHeaderParser(csrSubjectConsts)

		//when
		commonName, hash, found := hp.GetCertificateData(r)

		//then
		require.False(t, found)
		assert.Empty(t, commonName)
		assert.Empty(t, hash)
	})

	t.Run("should not found certificate data if header is empty", func(t *testing.T) {
		//given
		r, err := http.NewRequest("GET", "", nil)
		require.NoError(t, err)

		hp := oathkeeper.NewHeaderParser(csrSubjectConsts)

		//when
		commonName, hash, found := hp.GetCertificateData(r)

		// then
		require.False(t, found)
		assert.Empty(t, commonName)
		assert.Empty(t, hash)
	})
}
