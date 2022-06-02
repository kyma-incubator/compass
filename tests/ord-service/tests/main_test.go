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
	"os"
	"testing"

	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/kyma-incubator/compass/tests/pkg/util"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	"github.com/vrischmann/envconfig"
)

var (
	certCache                certloader.Cache
	certSecuredGraphQLClient *graphql.Client
)

type TenantConfig struct {
	TenantFetcherURL             string
	RootAPI                      string
	RegionalHandlerEndpoint      string
	TenantPathParam              string
	RegionPathParam              string
	Region                       string
	TenantFetcherFullRegionalURL string `envconfig:"-"`
}

type SubscriptionConfig struct {
	URL                                string
	TokenURL                           string
	ClientID                           string
	ClientSecret                       string
	ProviderLabelKey                   string
	ProviderID                         string
	SelfRegisterLabelKey               string
	SelfRegisterLabelValuePrefix       string
	PropagatedProviderSubaccountHeader string
}

type config struct {
	TenantConfig
	CertLoaderConfig certloader.Config
	certprovider.ExternalCertProviderConfig
	SubscriptionConfig
	ExternalServicesMockBaseURL      string
	DirectorExternalCertSecuredURL   string
	ORDServiceURL                    string
	ORDExternalCertSecuredServiceURL string
	ORDServiceStaticPrefix           string
	ORDServiceDefaultResponseType    string
	DefaultScenarioEnabled           bool `envconfig:"default=true"`
	ConsumerTokenURL                 string
	TokenPath                        string
	ProviderClientID                 string
	ProviderClientSecret             string
	SkipSSLValidation                bool
	BasicUsername                    string
	BasicPassword                    string
	AccountTenantID                  string
	SubaccountTenantID               string
	TestConsumerAccountID            string
	TestProviderSubaccountID         string
	TestConsumerSubaccountID         string
	TestConsumerTenantID             string
	SelfRegDistinguishLabelKey       string
	SelfRegDistinguishLabelValue     string
	SelfRegRegion                    string
}

var testConfig config

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	ctx := context.Background()

	certCache, err = certloader.StartCertLoader(ctx, testConfig.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err := util.WaitForCache(certCache); err != nil {
		log.D().Fatal(err)
	}
	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, certCache.Get().PrivateKey, certCache.Get().Certificate, testConfig.SkipSSLValidation)

	testConfig.TenantFetcherFullRegionalURL = tenantfetcher.BuildTenantFetcherRegionalURL(testConfig.RegionalHandlerEndpoint, testConfig.TenantPathParam, testConfig.RegionPathParam, testConfig.TenantFetcherURL, testConfig.RootAPI)

	exitVal := m.Run()
	os.Exit(exitVal)

}
