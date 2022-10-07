package tests

import (
	"context"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/stretchr/testify/require"
)

func TestTechnicalClient(t *testing.T) {
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

	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, pk, cert, conf.SkipSSLValidation)

	t.Log("Trying to list tenants")
	t.Log(externalCertProviderConfig.TestExternalCertSubject)
	tenants, err := fixtures.GetTenants(directorCertSecuredClient)
	require.NoError(t, err)
	require.NotEmpty(t, tenants)
}
