package oathkeeper_test

import (
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"

	"github.com/kyma-incubator/compass/components/connector/internal/certificates"

	"github.com/stretchr/testify/require"
)

const (
	certHeader = "Certificate-Data"
)

func TestParseCertHeader(t *testing.T) {
	connectorSubjectConsts := certificates.SubjectConsts{
		Country:            "DE",
		Organization:       "organization",
		OrganizationalUnit: "OrgUnit",
		Locality:           "Waldorf",
		Province:           "Waldorf",
	}

	externalSubjectConsts := certificates.SubjectConsts{
		Country:            "DE",
		Organization:       "organization",
		OrganizationalUnit: "(?i)[0-9a-f]{8}-[0-9a-f]{4}-[1-5][0-9a-f]{3}-[89ab][0-9a-f]{3}-[0-9a-f]{12},Region,SAP Cloud Platform Clients",
	}

	for _, testCase := range []struct {
		name             string
		certHeader       string
		subjectConsts    certificates.SubjectConsts
		issuer           string
		subjectMatcher   func(consts certificates.SubjectConsts) func(string) bool
		clientIDFunc     func(string) string
		found            bool
		expectedHash     string
		expectedClientID string
	}{
		{
			name: "connector header parser should return common name and hash",
			certHeader: "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=test-application,OU=OrgUnit,O=organization,L=Waldorf,ST=Waldorf,C=DE\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;" +
				"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			subjectConsts:    connectorSubjectConsts,
			issuer:           oathkeeper.ConnectorIssuer,
			subjectMatcher:   oathkeeper.ConnectorCertificateSubjectMatcher,
			clientIDFunc:     oathkeeper.GetCommonName,
			found:            true,
			expectedHash:     "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad",
			expectedClientID: "test-application",
		},
		{
			name: "external header parser should return common name and hash",
			certHeader: "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"CN=test-application,OU=2d149cda-a4fe-45c9-a21d-915c52fb56a1,OU=Region,OU=SAP Cloud Platform Clients,O=organization,L=Waldorf,ST=Waldorf,C=DE\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;" +
				"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			subjectConsts:    externalSubjectConsts,
			issuer:           oathkeeper.ExternalIssuer,
			subjectMatcher:   oathkeeper.ExternalCertIssuerSubjectMatcher,
			clientIDFunc:     oathkeeper.GetOrganizationalUnit,
			found:            true,
			expectedHash:     "f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad",
			expectedClientID: "2d149cda-a4fe-45c9-a21d-915c52fb56a1",
		},
		{
			name: "should not found certificate data if non is matching",
			certHeader: "Hash=f4cf22fb633d4df500e371daf703d4b4d14a0ea9d69cd631f95f9e6ba840f8ad;Subject=\"\";URI=spiffe://cluster.local/ns/kyma-integration/sa/default;" +
				"Hash=6d1f9f3a6ac94ff925841aeb9c15bb3323014e3da2c224ea7697698acf413226;Subject=\"\";URI=spiffe://cluster.local/ns/istio-system/sa/istio-ingressgateway-service-account",
			subjectConsts:  connectorSubjectConsts,
			issuer:         oathkeeper.ConnectorIssuer,
			subjectMatcher: oathkeeper.ConnectorCertificateSubjectMatcher,
			clientIDFunc:   oathkeeper.GetCommonName,
			found:          false,
		},
		{
			name:           "should not found certificate data if header is invalid",
			certHeader:     "invalid header",
			subjectConsts:  connectorSubjectConsts,
			issuer:         oathkeeper.ConnectorIssuer,
			subjectMatcher: oathkeeper.ConnectorCertificateSubjectMatcher,
			clientIDFunc:   oathkeeper.GetCommonName,
			found:          false,
		},
		{
			name:           "should not found certificate data if header is empty",
			certHeader:     "",
			subjectConsts:  connectorSubjectConsts,
			issuer:         oathkeeper.ConnectorIssuer,
			subjectMatcher: oathkeeper.ConnectorCertificateSubjectMatcher,
			clientIDFunc:   oathkeeper.GetCommonName,
			found:          false,
		},
	} {
		t.Run(testCase.name, func(t *testing.T) {
			r, err := http.NewRequest("GET", "", nil)
			require.NoError(t, err)

			r.Header.Set(certHeader, testCase.certHeader)

			hp := oathkeeper.NewHeaderParser(certHeader, testCase.issuer, testCase.subjectMatcher(testCase.subjectConsts), testCase.clientIDFunc)

			//when
			clientID, hash, found := hp.GetCertificateData(r)

			//then
			require.Equal(t, testCase.found, found)
			require.Equal(t, testCase.expectedHash, hash)
			require.Equal(t, testCase.expectedClientID, clientID)
			require.Equal(t, testCase.issuer, hp.GetIssuer())
		})
	}
}
