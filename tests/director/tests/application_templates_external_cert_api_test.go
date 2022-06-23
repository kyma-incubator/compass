package tests

import (
	"context"
	"fmt"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/assertions"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/stretchr/testify/require"
)

func TestApplicationTemplateWithExternalCertificate(t *testing.T) {
	// GIVEN
	ctx := context.Background()
	// Build graphql director client configured with certificate
	externalCertProviderConfig := certprovider.ExternalCertProviderConfig{
		ExternalClientCertTestSecretName:      conf.ExternalCertProviderConfig.ExternalClientCertTestSecretName,
		ExternalClientCertTestSecretNamespace: conf.ExternalCertProviderConfig.ExternalClientCertTestSecretNamespace,
		CertSvcInstanceTestSecretName:         conf.ExternalCertProviderConfig.CertSvcInstanceTestSecretName,
		ExternalCertCronjobContainerName:      conf.ExternalCertProviderConfig.ExternalCertCronjobContainerName,
		ExternalCertTestJobName:               conf.ExternalCertProviderConfig.ExternalCertTestJobName,
		TestExternalCertSubject:               strings.Replace(conf.ExternalCertProviderConfig.TestExternalCertSubject, "integration-system-test", "external-cert", -1),
		ExternalClientCertCertKey:             conf.ExternalCertProviderConfig.ExternalClientCertCertKey,
		ExternalClientCertKeyKey:              conf.ExternalCertProviderConfig.ExternalClientCertKeyKey,
	}

	pk, cert := certprovider.NewExternalCertFromConfig(t, ctx, externalCertProviderConfig)
	directorCertSecuredClient := gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, pk, cert, conf.SkipSSLValidation)
	tenantId := tenant.TestTenants.GetDefaultTenantID()

	name := createAppTemplateName("app-template-with-external-cert-name")

	t.Run("Create Application Template with external certificate", func(t *testing.T) {
		// WHEN
		t.Log("Create application template")
		appTemplateInput := fixAppTemplateInput(name)
		appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertSecuredClient, tenantId, appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, directorCertSecuredClient, tenantId, &appTemplate)

		//THEN
		require.NoError(t, err)
		require.NotEmpty(t, appTemplate.ID)

		t.Log("Check if application template was created")
		appTemplateOutput := fixtures.GetApplicationTemplate(t, ctx, directorCertSecuredClient, tenantId, appTemplate.ID)
		require.NotEmpty(t, appTemplateOutput)

		// Enhance input to match the newly created labels
		appTemplateInput.Labels[conf.SelfRegLabelKey] = appTemplateOutput.Labels[conf.SelfRegLabelKey]
		appTemplateInput.Labels["global_subaccount_id"] = conf.ConsumerID
		appTemplateInput.ApplicationInput.Labels["applicationType"] = fmt.Sprintf("%s (%s)", name, conf.SelfRegRegion)
		assertions.AssertApplicationTemplate(t, appTemplateInput, appTemplateOutput)
	})

	t.Run("Delete Application Template with external certificate", func(t *testing.T) {
		t.Log("Create application template")
		appTemplateInput := fixAppTemplateInput(name)
		appTemplate, err := fixtures.CreateApplicationTemplateFromInput(t, ctx, directorCertSecuredClient, tenantId, appTemplateInput)
		defer fixtures.CleanupApplicationTemplate(t, ctx, directorCertSecuredClient, tenantId, &appTemplate)

		require.NoError(t, err)
		require.NotEmpty(t, appTemplate.ID)

		// WHEN
		t.Log("Delete application template")
		fixtures.DeleteApplicationTemplate(t, ctx, directorCertSecuredClient, tenantId, appTemplate.ID)

		//THEN
		t.Log("Check if application template was deleted")
		out := fixtures.GetApplicationTemplate(t, ctx, directorCertSecuredClient, tenantId, appTemplate.ID)

		require.Empty(t, out)
	})
}
