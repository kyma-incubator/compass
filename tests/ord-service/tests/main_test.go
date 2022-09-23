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

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	cfg "github.com/kyma-incubator/compass/components/director/pkg/config"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/certs/certprovider"
	"github.com/kyma-incubator/compass/tests/pkg/clients"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/subscription"
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

type config struct {
	TFConfig           subscription.TenantFetcherConfig
	CertLoaderConfig   certloader.Config
	SubscriptionConfig subscription.Config
	certprovider.ExternalCertProviderConfig
	ExternalServicesMockBaseURL      string
	DirectorExternalCertSecuredURL   string
	ORDServiceURL                    string
	ORDExternalCertSecuredServiceURL string
	ORDServiceStaticPrefix           string
	ORDServiceDefaultResponseType    string
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
	ApplicationTypeLabelKey          string `envconfig:"APP_APPLICATION_TYPE_LABEL_KEY,default=applicationType"`
	SaaSAppNameLabelKey              string `envconfig:"APP_SELF_REGISTER_SAAS_APP_LABEL_KEY,default=CMPSaaSAppName"`
	DestinationAPIConfig             clients.DestinationServiceAPIConfig
	DestinationsConfig               cfg.DestinationsConfig
	DestinationConsumerSubdomain     string `envconfig:"APP_DESTINATION_CONSUMER_SUBDOMAIN"`
	ExternalClientCertSecretName string `envconfig:"EXTERNAL_CLIENT_CERT_SECRET_NAME"`

}

var conf config

func TestMain(m *testing.M) {
	err := envconfig.Init(&conf)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	err = conf.DestinationsConfig.MapInstanceConfigs()
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while loading destination instances config"))
	}

	tenant.TestTenants.Init()

	ctx := context.Background()

	certCache, err = certloader.StartCertLoader(ctx, conf.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err := util.WaitForCache(certCache); err != nil {
		log.D().Fatal(err)
	}
	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(conf.DirectorExternalCertSecuredURL, certCache.Get()[conf.ExternalClientCertSecretName].PrivateKey, certCache.Get()[conf.ExternalClientCertSecretName].Certificate, conf.SkipSSLValidation)

	conf.TFConfig.FullRegionalURL = tenantfetcher.BuildTenantFetcherRegionalURL(conf.TFConfig.RegionalHandlerEndpoint, conf.TFConfig.TenantPathParam, conf.TFConfig.RegionPathParam, conf.TFConfig.URL, conf.TFConfig.RootAPI)

	exitVal := m.Run()
	os.Exit(exitVal)

}
