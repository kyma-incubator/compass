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
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/clients"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/kyma-incubator/compass/tests/pkg/tenantfetcher"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

var (
	dexGraphQLClient *graphql.Client
)

type TenantConfig struct {
	TenantIDProperty               string
	SubaccountTenantIDProperty     string
	CustomerIDProperty             string
	SubdomainProperty              string
	SubscriptionProviderIDProperty string
	TenantFetcherURL               string
	RootAPI                        string
	RegionalHandlerEndpoint        string
	TenantPathParam                string
	RegionPathParam                string
	TenantFetcherFullRegionalURL   string `envconfig:"-"`
}

type ConnectorCAConfig struct {
	Certificate          []byte `envconfig:"-"`
	Key                  []byte `envconfig:"-"`
	SecretName           string
	SecretNamespace      string
	SecretCertificateKey string
	SecretKeyKey         string
}

type config struct {
	TenantConfig
	CA ConnectorCAConfig

	DirectorURL                      string
	DirectorExternalCertSecuredURL   string
	ORDServiceURL                    string
	ORDExternalCertSecuredServiceURL string
	ORDServiceStaticPrefix           string
	ORDServiceDefaultResponseType    string
	DefaultScenarioEnabled           bool `envconfig:"default=true"`
	ExternalServicesMockURL          string
	ClientID                         string
	ClientSecret                     string
	SubscriptionProviderLabelKey     string
	ConsumerSubaccountIdsLabelKey    string
}

var testConfig config

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	ctx := context.Background()

	k8sClientSet, err := clients.NewK8SClientSet(ctx, time.Second, time.Minute, time.Minute)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing k8s client"))
	}

	secret, err := k8sClientSet.CoreV1().Secrets(testConfig.CA.SecretNamespace).Get(ctx, testConfig.CA.SecretName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(errors.Wrap(err, "while getting k8s secret"))
	}

	testConfig.CA.Certificate = secret.Data[testConfig.CA.SecretCertificateKey]
	testConfig.CA.Key = secret.Data[testConfig.CA.SecretKeyKey]

	testConfig.TenantFetcherFullRegionalURL = tenantfetcher.BuildTenantFetcherRegionalURL(testConfig.RegionalHandlerEndpoint, testConfig.TenantPathParam, testConfig.RegionPathParam, testConfig.TenantFetcherURL, testConfig.RootAPI)

	dexToken := server.Token()

	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	exitVal := m.Run()
	os.Exit(exitVal)

}
