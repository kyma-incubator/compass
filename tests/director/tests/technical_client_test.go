package tests

import (
	"context"
	"strings"
	"testing"

	testingx "github.com/kyma-incubator/compass/tests/pkg/testing"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/stretchr/testify/require"
)

func TestTechnicalClient(stdT *testing.T) {
	t := testingx.NewT(stdT)
	ctx := context.Background()

	replacer := strings.NewReplacer(conf.TestProviderSubaccountID, conf.TestConsumerSubaccountID, conf.ExternalCertCommonName, "technical-client-test")

	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.CertSvcInstanceTestIntSystemSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestAtomExternalCertSubject:           replacer.Replace(conf.ExternalCertProviderConfig.TestAtomExternalCertSubject),
		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
		ExternalCertProvider:                  certprovider.Atom,
	}

	pk, cert := certprovider.NewExternalCertFromConfig(stdT, ctx, externalCertProviderConfig)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, pk, cert, conf.SkipSSLValidation)

	t.Run("Successfully list tenants", func(stdT *testing.T) {
		tenants, err := fixtures.GetTenants(directorCertSecuredClient)
		require.NoError(t, err)
		require.NotEmpty(t, tenants)
	})

}
