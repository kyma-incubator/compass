/*
 * Copyright 2020 The Compass Authors
 *
 * Licensed under the Apache License, Version 2.0 (the "License");
 * you may not use this file except in compliance with the License.
 * You may obtain a copy of the License at
 *
 *     http://www.apache.org/licenses/LICENSE-2.0
 *
 * Unless required by applicable law or agreed to in writing, software
 * distributed under the License is distributed on an "AS IS" BASIS,
 * WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
 * See the License for the specific language governing permissions and
 * limitations under the License.
 */

package tests

import (
	"context"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/tests/pkg/fixtures"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/model"
	directorSchema "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/tests/pkg/certs"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/stretchr/testify/require"
)

func TestAppRegistry(t *testing.T) {
	ctx := context.Background()

	defer fixtures.DeleteFormationWithinTenant(t, ctx, certSecuredGraphQLClient, testConfig.Tenant, testScenario)
	fixtures.CreateFormationWithinTenant(t, ctx, certSecuredGraphQLClient, testConfig.Tenant, testScenario)

	appInput := directorSchema.ApplicationRegisterInput{
		Name:           TestApp,
		ProviderName:   ptr.String("provider name"),
		Description:    ptr.String("my application"),
		HealthCheckURL: ptr.String("http://mywordpress.com/health"),
		Labels: directorSchema.Labels{
			"scenarios":                        []interface{}{testScenario},
			testConfig.ApplicationTypeLabelKey: string(util.ApplicationTypeC4C),
		},
	}

	descr := "test"
	runtimeInput := fixRuntimeInput(descr)

	appID, err := directorClient.CreateApplication(appInput)
	defer func() {
		err = directorClient.CleanupApplication(appID)
		require.NoError(t, err)
	}()
	require.NoError(t, err)

	runtime := fixtures.RegisterKymaRuntime(t, ctx, certSecuredGraphQLClient, testConfig.Tenant, runtimeInput, testConfig.GatewayOauth)
	defer fixtures.CleanupRuntime(t, ctx, certSecuredGraphQLClient, testConfig.Tenant, &runtime)

	require.NoError(t, err)

	err = directorClient.SetDefaultEventing(runtime.ID, appID, testConfig.EventsBaseURL)
	require.NoError(t, err)

	t.Run("App Registry Service flow for Application", func(t *testing.T) {
		client := clients.NewConnectorClient(directorClient, appID, testConfig.Tenant, testConfig.SkipSSLValidation)
		clientKey := certs.CreateKey(t)

		crtResponse, infoResponse := createCertificateChain(t, client, clientKey)
		require.NotEmpty(t, crtResponse.CRTChain)
		require.NotEmpty(t, infoResponse.Api.ManagementInfoURL)
		require.NotEmpty(t, infoResponse.Certificate)

		crtChainBytes := certs.DecodeBase64Cert(t, crtResponse.CRTChain)
		adapterClient, err := clients.NewSecuredClient(testConfig.SkipSSLValidation, clientKey, crtChainBytes, testConfig.Tenant)
		require.NoError(t, err)

		mgmInfoResponse, errorResponse := adapterClient.GetMgmInfo(t, infoResponse.Api.ManagementInfoURL)
		defer func() {
			errorResponse = adapterClient.RevokeCertificate(t, mgmInfoResponse.URLs.RevokeCertURL)
			require.Nil(t, errorResponse)
		}()
		require.Nil(t, errorResponse)
		require.NotEmpty(t, mgmInfoResponse.URLs.RenewCertURL)
		require.NotEmpty(t, mgmInfoResponse.Certificate)
		require.Equal(t, infoResponse.Certificate, mgmInfoResponse.Certificate)

		metadataURL := infoResponse.Api.MetadataURL

		services, errorResponse := adapterClient.ListServices(t, metadataURL)
		require.Nil(t, errorResponse)
		require.Len(t, services, 0)

		service := model.ServiceDetails{
			Name:        "test-service",
			Provider:    "provider",
			Description: "description",
			Api: &model.API{
				TargetUrl: "http://target.com",
				Credentials: &model.CredentialsWithCSRF{
					OauthWithCSRF: &model.OauthWithCSRF{
						Oauth: model.Oauth{
							URL:          "http://test.com/token",
							ClientID:     "client",
							ClientSecret: "secret",
						},
					},
				},
			},
			Labels: &map[string]string{},
			Events: &model.Events{
				Spec: ptrSpecResponse(`{"asyncapi":"1.2.0"}`),
			},
		}

		createServiceResponse, errorResponse := adapterClient.CreateService(t, metadataURL, service)
		defer adapterClient.CleanupService(t, metadataURL, createServiceResponse.ID)
		require.Nil(t, errorResponse)
		require.NotNil(t, createServiceResponse.ID)

		expectedService := service
		expectedService.Provider = ""

		serviceResponse, errorResponse := adapterClient.GetService(t, metadataURL, createServiceResponse.ID)
		require.Nil(t, errorResponse)
		require.Equal(t, &expectedService, serviceResponse)

		expectedService.Api.TargetUrl = service.Api.TargetUrl + "/test"

		updateServiceResponse, errorResponse := adapterClient.UpdateService(t, metadataURL, createServiceResponse.ID, service)
		require.Nil(t, errorResponse)
		require.Equal(t, &expectedService, updateServiceResponse)

		services, errorResponse = adapterClient.ListServices(t, metadataURL)
		require.Nil(t, errorResponse)
		require.Len(t, services, 1)
		require.Equal(t, expectedService.Name, services[0].Name)
		require.Equal(t, expectedService.Description, services[0].Description)

		errorResponse = adapterClient.DeleteService(t, metadataURL, createServiceResponse.ID)
		require.Nil(t, errorResponse)

		services, errorResponse = adapterClient.ListServices(t, metadataURL)
		require.Nil(t, errorResponse)
		require.Len(t, services, 0)
	})
}

func ptrSpecResponse(in model.SpecResponse) *model.SpecResponse {
	return &in
}

func fixRuntimeInput(descr string) directorSchema.RuntimeRegisterInput {
	return directorSchema.RuntimeRegisterInput{
		Name:        TestRuntime,
		Description: &descr,
		Labels: directorSchema.Labels{
			"scenarios": []interface{}{testScenario},
		},
	}
}
