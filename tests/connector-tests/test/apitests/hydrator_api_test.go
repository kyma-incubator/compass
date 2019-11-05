package apitests

import (
	"net/http"
	"testing"

	"github.com/kyma-incubator/compass/components/connector/pkg/graphql/externalschema"
	"github.com/kyma-incubator/compass/components/connector/pkg/oathkeeper"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestHydrators(t *testing.T) {

	appID := "54f83a73-b340-418d-b653-d25b5ed47d75"
	runtimeID := "75f42q66-b340-418d-b653-d25b5ed47d75"

	hash := "df6ab69b34100a1808ddc6211010fa289518f14606d0c8eaa03a0f53ecba578a"

	for _, testCase := range []struct {
		clientType           string
		clientId             string
		tokenGenerationFunc  func(id string) (externalschema.Token, error)
		expectedTokenHeaders http.Header
		expectedCertsHeaders http.Header
	}{
		{
			clientType:          "Application",
			clientId:            appID,
			tokenGenerationFunc: internalClient.GenerateApplicationToken,
			expectedTokenHeaders: http.Header{
				oathkeeper.ClientIdFromTokenHeader: []string{appID},
			},
			expectedCertsHeaders: http.Header{
				oathkeeper.ClientIdFromCertificateHeader: []string{appID},
				oathkeeper.ClientCertificateHashHeader:   []string{hash},
			},
		},
		{
			clientType:          "Runtime",
			clientId:            runtimeID,
			tokenGenerationFunc: internalClient.GenerateRuntimeToken,
			expectedTokenHeaders: http.Header{
				oathkeeper.ClientIdFromTokenHeader: []string{runtimeID},
			},
			expectedCertsHeaders: http.Header{
				oathkeeper.ClientIdFromCertificateHeader: []string{runtimeID},
				oathkeeper.ClientCertificateHashHeader:   []string{hash},
			},
		},
	} {
		t.Run("should resolve one-time token for "+testCase.clientType, func(t *testing.T) {
			//given
			token, err := testCase.tokenGenerationFunc(testCase.clientId)
			require.NoError(t, err)
			require.NotEmpty(t, token.Token)

			headers := map[string][]string{
				oathkeeper.ConnectorTokenHeader: {token.Token},
			}

			//when
			authSession := hydratorClient.ResolveToken(t, headers)

			//then
			assert.Equal(t, testCase.expectedTokenHeaders, authSession.Header)
		})

		t.Run("should resolve certificate for "+testCase.clientType, func(t *testing.T) {
			//given
			token, err := testCase.tokenGenerationFunc(testCase.clientId)
			require.NoError(t, err)
			require.NotEmpty(t, token.Token)

			configuration, err := connectorClient.Configuration(token.Token)
			require.NoError(t, err)

			certDataHeader := createCertDataHeader(configuration.CertificateSigningRequestInfo.Subject, hash)

			headers := map[string][]string{
				config.CertificateDataHeader: {certDataHeader},
			}

			//when
			authSession := hydratorClient.ResolveCertificateData(t, headers)

			//then
			assert.Equal(t, testCase.expectedCertsHeaders, authSession.Header)
		})

		t.Run("should return empty Authentication Session when no valid headers found", func(t *testing.T) {
			//given
			token, err := testCase.tokenGenerationFunc(testCase.clientId)
			require.NoError(t, err)
			require.NotEmpty(t, token.Token)

			//when
			authSession := hydratorClient.ResolveToken(t, nil)

			//then
			assert.Empty(t, authSession)
		})
	}

}
