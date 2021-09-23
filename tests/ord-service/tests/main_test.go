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
	"crypto/tls"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/clients"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/server"
	"github.com/kyma-incubator/compass/tests/pkg/tenant"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	log "github.com/sirupsen/logrus"
	"github.com/vrischmann/envconfig"
)

var (
	dexGraphQLClient *graphql.Client
	httpClient *http.Client
)

type TenantConfig struct {
	TenantIDProperty               string `envconfig:"APP_TENANT_PROVIDER_TENANT_ID_PROPERTY"`
	SubaccountTenantIDProperty     string `envconfig:"APP_TENANT_PROVIDER_SUBACCOUNT_TENANT_ID_PROPERTY"`
	CustomerIDProperty             string `envconfig:"APP_TENANT_PROVIDER_CUSTOMER_ID_PROPERTY"`
	SubdomainProperty              string `envconfig:"APP_TENANT_PROVIDER_SUBDOMAIN_PROPERTY"`
	SubscriptionProviderIDProperty string `envconfig:"APP_TENANT_PROVIDER_SUBSCRIPTION_PROVIDER_ID_PROPERTY"`
	TenantFetcherURL          	   string `envconfig:"APP_TENANT_FETCHER_URL"`
	RootAPI                   	   string `envconfig:"APP_ROOT_API"`
	RegionalHandlerEndpoint   	   string `envconfig:"APP_REGIONAL_HANDLER_ENDPOINT"`
	TenantPathParam           	   string `envconfig:"APP_TENANT_PATH_PARAM"`
	RegionPathParam				   string `envconfig:"APP_REGION_PATH_PARAM"`
	TenantFetcherFullRegionalURL   string
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
	ORDServiceURL                    string
	ORDExternalCertSecuredServiceURL string
	ORDServiceStaticPrefix           string
	ORDServiceDefaultResponseType    string
	DefaultScenarioEnabled           bool `envconfig:"default=true"`
	ExternalServicesMockURL          string
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

	httpClient = &http.Client{
		Timeout: 10 * time.Second,
		Transport: &http.Transport{
			TLSClientConfig: &tls.Config{InsecureSkipVerify: true},
		},
	}

	k8sClientSet, err := clients.NewK8SClientSet(context.Background(), time.Second, time.Minute, time.Minute)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing k8s client"))
	}

	secret, err := k8sClientSet.CoreV1().Secrets(testConfig.CA.SecretNamespace).Get(context.Background(), testConfig.CA.SecretName, metav1.GetOptions{})
	if err != nil {
		log.Fatal(errors.Wrap(err, "while getting k8s secret"))
	}

	testConfig.CA.Certificate = secret.Data[testConfig.CA.SecretCertificateKey]
	testConfig.CA.Key = secret.Data[testConfig.CA.SecretKeyKey]

	regionalEndpoint := strings.Replace(testConfig.RegionalHandlerEndpoint, fmt.Sprintf("{%s}", testConfig.TenantPathParam), tenantPathParamValue, 1)
	regionalEndpoint = strings.Replace(regionalEndpoint, fmt.Sprintf("{%s}", testConfig.RegionPathParam), regionPathParamValue, 1)
	testConfig.TenantFetcherFullRegionalURL = testConfig.TenantFetcherURL + testConfig.RootAPI + regionalEndpoint

	dexToken := server.Token()

	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	exitVal := m.Run()
	os.Exit(exitVal)

}
