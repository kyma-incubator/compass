package certresolver_test

import (
	"context"
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/hydrator/internal/certresolver"
	"github.com/kyma-incubator/compass/components/hydrator/internal/subject"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/pkg/cert"

	"github.com/stretchr/testify/require"
)

const (
	certHeader = "Certificate-Data"
)

func TestParseCertHeader(t *testing.T) {
	connectorSubjectConsts := subject.CSRSubjectConfig{
		Country:            "DE",
		Organization:       "organization",
		OrganizationalUnit: "OrgUnit",
		Locality:           "Waldorf",
		Province:           "Waldorf",
	}

	externalSubjectConsts := subject.ExternalIssuerSubjectConfig{
		Country:                   "DE",
		Organization:              "organization",
		OrganizationalUnitPattern: "(?i)[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12}|Region|SAP Cloud Platform Clients",
	}

	expectedAuthSessionExtra := map[string]interface{}{
		cert.ConsumerTypeExtraField:  "test_consumer_type",
		cert.InternalConsumerIDField: "test_internal_consumer_id",
		cert.AccessLevelsExtraField:  []string{"test_access_level"},
	}

	noopAuthSessionExtra := func(ctx context.Context, s string) map[string]interface{} { return nil }

	for _, testCase := range []struct {
		name                            string
		certHeader                      string
		subjectConsts                   subject.CSRSubjectConfig
		issuer                          string
		subjectMatcher                  func(string) bool
		clientIDFunc                    func(string) string
		authSessionExtraFromSubjectFunc func(context.Context, string) map[string]interface{}
		found                           bool
		expectedHash                    string
		expectedClientID                string
		expectedAuthSessionExtra        map[string]interface{}
	}{
		{
			name: "connector header parser should return common name and hash",
			certHeader: "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=test-application,OU=OrgUnit,O=organization,L=Waldorf,ST=Waldorf,C=DE\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;" +
				"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			issuer:                          oathkeeper.ConnectorIssuer,
			subjectMatcher:                  subject.ConnectorCertificateSubjectMatcher(connectorSubjectConsts),
			clientIDFunc:                    cert.GetCommonName,
			found:                           true,
			expectedHash:                    "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad",
			expectedClientID:                "test-application",
			authSessionExtraFromSubjectFunc: noopAuthSessionExtra,
		},
		{
			name: "connector header parser should append extra",
			certHeader: "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=test-application,OU=OrgUnit,O=organization,L=Waldorf,ST=Waldorf,C=DE\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;" +
				"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			issuer:         oathkeeper.ConnectorIssuer,
			subjectMatcher: subject.ConnectorCertificateSubjectMatcher(connectorSubjectConsts),
			clientIDFunc:   cert.GetCommonName,
			authSessionExtraFromSubjectFunc: func(ctx context.Context, s string) map[string]interface{} {
				return expectedAuthSessionExtra
			},
			found:                    true,
			expectedHash:             "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad",
			expectedClientID:         "test-application",
			expectedAuthSessionExtra: expectedAuthSessionExtra,
		},
		{
			name: "external header parser should return common name and hash when subject use , for same values separation",
			certHeader: "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=test-application,OU=2d149cda-a4fe-45c9-a21d-915c52fb56a1,OU=Region,OU=SAP Cloud Platform Clients,O=organization,L=Waldorf,ST=Waldorf,C=DE\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;" +
				"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			issuer:                          oathkeeper.ExternalIssuer,
			subjectMatcher:                  subject.ExternalCertIssuerSubjectMatcher(externalSubjectConsts),
			clientIDFunc:                    cert.GetUUIDOrganizationalUnit,
			found:                           true,
			expectedHash:                    "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad",
			expectedClientID:                "2d149cda-a4fe-45c9-a21d-915c52fb56a1",
			authSessionExtraFromSubjectFunc: noopAuthSessionExtra,
		},
		{
			name: "external header parser should return common name and hash when subject use + for same values separation",
			certHeader: "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=common-name,OU=123e4567-e89b-12d3-a456-426614174001+OU=SAP Cloud Platform Clients+OU=Region,O=organization,L=locality,C=DE\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;" +
				"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			issuer:                          oathkeeper.ExternalIssuer,
			subjectMatcher:                  subject.ExternalCertIssuerSubjectMatcher(externalSubjectConsts),
			clientIDFunc:                    cert.GetUUIDOrganizationalUnit,
			found:                           true,
			expectedHash:                    "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad",
			expectedClientID:                "123e4567-e89b-12d3-a456-426614174001",
			authSessionExtraFromSubjectFunc: noopAuthSessionExtra,
		},
		{
			name: "should not found certificate data if non is matching",
			certHeader: "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;" +
				"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			issuer:                          oathkeeper.ConnectorIssuer,
			subjectMatcher:                  subject.ConnectorCertificateSubjectMatcher(connectorSubjectConsts),
			clientIDFunc:                    cert.GetCommonName,
			found:                           false,
			authSessionExtraFromSubjectFunc: noopAuthSessionExtra,
		},
		{
			name:                            "should not found certificate data if header is invalid",
			certHeader:                      "invalid header",
			issuer:                          oathkeeper.ConnectorIssuer,
			subjectMatcher:                  subject.ConnectorCertificateSubjectMatcher(connectorSubjectConsts),
			clientIDFunc:                    cert.GetCommonName,
			found:                           false,
			authSessionExtraFromSubjectFunc: noopAuthSessionExtra,
		},
		{
			name:                            "should not found certificate data if header is empty",
			certHeader:                      "",
			issuer:                          oathkeeper.ConnectorIssuer,
			subjectMatcher:                  subject.ConnectorCertificateSubjectMatcher(connectorSubjectConsts),
			clientIDFunc:                    cert.GetCommonName,
			found:                           false,
			authSessionExtraFromSubjectFunc: noopAuthSessionExtra,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			r, err := http.NewRequest("GET", "", nil)
			require.NoError(t, err)

			r.Header.Set(certHeader, testCase.certHeader)

			hp := certresolver.NewHeaderParser(certHeader, testCase.issuer, testCase.subjectMatcher, testCase.clientIDFunc, testCase.authSessionExtraFromSubjectFunc)

			// when
			certificateData := hp.GetCertificateData(r)

			// then
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
