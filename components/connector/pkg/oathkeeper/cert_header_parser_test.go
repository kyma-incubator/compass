package oathkeeper_test

import (
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/internal/certificates"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/cert"

	"github.com/stretchr/testify/require"
)

const (
	certHeader = "Certificate-Data"
)

func TestParseCertHeader(t *testing.T) {
	connectorSubjectConsts := certificates.CSRSubjectConsts{
		Country:            "DE",
		Organization:       "organization",
		OrganizationalUnit: "OrgUnit",
		Locality:           "Waldorf",
		Province:           "Waldorf",
	}

	externalSubjectConsts := certificates.ExternalIssuerSubjectConsts{
		Country:                   "DE",
		Organization:              "organization",
		OrganizationalUnitPattern: "(?i)[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}|Region|SAP Cloud Platform Clients",
	}

	//expectedAuthSessionExtra := map[string]interface{}{
	//	cert.ConsumerTypeExtraField:  "test_consumer_type",
	//	cert.InternalConsumerIDField: "test_internal_consumer_id",
	//	cert.AccessLevelExtraField:   "test_access_level",
	//}

	for _, testCase := range []struct {
		name                            string
		certHeader                      string
		subjectConsts                   certificates.CSRSubjectConsts
		issuer                          string
		subjectMatcher                  func(string) bool
		clientIDFunc                    func(string) string
		authSessionExtraFromSubjectFunc func(string) map[string]interface{}
		found                           bool
		expectedHash                    string
		expectedClientID                string
		expectedAuthSessionExtra        map[string]interface{}
	}{
		{
			name: "connector header parser should return common name and hash",
			certHeader: "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=test-application,OU=OrgUnit,O=organization,L=Waldorf,ST=Waldorf,C=DE\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;" +
				"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			issuer:           oathkeeper.ConnectorIssuer,
			subjectMatcher:   oathkeeper.ConnectorCertificateSubjectMatcher(connectorSubjectConsts),
			clientIDFunc:     cert.GetCommonName,
			found:            true,
			expectedHash:     "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad",
			expectedClientID: "test-application",
		},
		{
			name: "external header parser should return common name and hash when subject use , for same values separation",
			certHeader: "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=test-application,OU=2d149cda-a4fe-45c9-a21d-915c52fb56a1,OU=Region,OU=SAP Cloud Platform Clients,O=organization,L=Waldorf,ST=Waldorf,C=DE\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;" +
				"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			issuer:           oathkeeper.ExternalIssuer,
			subjectMatcher:   oathkeeper.ExternalCertIssuerSubjectMatcher(externalSubjectConsts),
			clientIDFunc:     cert.GetUUIDOrganizationalUnit,
			found:            true,
			expectedHash:     "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad",
			expectedClientID: "2d149cda-a4fe-45c9-a21d-915c52fb56a1",
		},
		{
			name: "external header parser should return common name and hash when subject use + for same values separation",
			certHeader: "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=common-name,OU=123e4567-e89b-12d3-a456-426614174001+OU=SAP Cloud Platform Clients+OU=Region,O=organization,L=locality,C=DE\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;" +
				"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			issuer:           oathkeeper.ExternalIssuer,
			subjectMatcher:   oathkeeper.ExternalCertIssuerSubjectMatcher(externalSubjectConsts),
			clientIDFunc:     cert.GetUUIDOrganizationalUnit,
			found:            true,
			expectedHash:     "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad",
			expectedClientID: "123e4567-e89b-12d3-a456-426614174001",
		},
		{
			name: "should not found certificate data if non is matching",
			certHeader: "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;" +
				"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			issuer:         oathkeeper.ConnectorIssuer,
			subjectMatcher: oathkeeper.ConnectorCertificateSubjectMatcher(connectorSubjectConsts),
			clientIDFunc:   cert.GetCommonName,
			found:          false,
		},
		{
			name:           "should not found certificate data if header is invalid",
			certHeader:     "invalid header",
			issuer:         oathkeeper.ConnectorIssuer,
			subjectMatcher: oathkeeper.ConnectorCertificateSubjectMatcher(connectorSubjectConsts),
			clientIDFunc:   cert.GetCommonName,
			found:          false,
		},
		{
			name:           "should not found certificate data if header is empty",
			certHeader:     "",
			issuer:         oathkeeper.ConnectorIssuer,
			subjectMatcher: oathkeeper.ConnectorCertificateSubjectMatcher(connectorSubjectConsts),
			clientIDFunc:   cert.GetCommonName,
			found:          false,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			r, err := http.NewRequest("GET", "", nil)
			require.NoError(t, err)

			r.Header.Set(certHeader, testCase.certHeader)

			hp := oathkeeper.NewHeaderParser(certHeader, testCase.issuer, testCase.subjectMatcher, testCase.clientIDFunc, testCase.authSessionExtraFromSubjectFunc)

			//when
			certificateData := hp.GetCertificateData(r)

			//then
			if testCase.found {
				require.Equal(t, testCase.expectedAuthSessionExtra, certificateData.AuthSessionExtra)
				require.Equal(t, testCase.expectedHash, certificateData.CertificateHash)
				require.Equal(t, testCase.expectedClientID, certificateData.ClientID)
				require.Equal(t, testCase.issuer, hp.GetIssuer())
			} else {
				require.Nil(t, certificateData)
			}
		})
	}
}
