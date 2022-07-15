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

	"github.com/kyma-incubator/compass/tests/pkg/subscription"

	"github.com/kyma-incubator/compass/tests/pkg/util"

	"github.com/kyma-incubator/compass/components/director/pkg/certloader"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	c "github.com/robfig/cron/v3"
	"github.com/vrischmann/envconfig"
)

type config struct {
	DefaultTestTenant                                     string
	DirectorExternalCertSecuredURL                        string
	DirectorGraphqlOauthURL                               string
	ORDServiceURL                                         string
	AggregatorSchedule                                    string
	ExternalServicesMockBaseURL                           string
	ExternalServicesMockUnsecuredURL                      string
	ExternalServicesMockAbsoluteURL                       string
	ExternalServicesMockOrdCertSecuredURL                 string
	ExternalServicesMockBasicURL                          string
	ExternalServicesMockOauthURL                          string
	ExternalServicesMockUnsecuredWithAdditionalContentURL string
    ExternalServicesMockUnsecuredMultiTenantURL           string        `envconfig:"EXTERNAL_SERVICES_MOCK_UNSECURED_MULTI_TENANT_URL"`
	ClientID                                              string
	ClientSecret                                          string
	BasicUsername                                         string
	BasicPassword                                         string
	ORDServiceDefaultResponseType                         string
	GlobalRegistryURL                                     string
	TestConsumerSubaccountID                              string
	TestConsumerTenantID                                  string
	TestProviderSubaccountID                              string
	TokenPath                                             string
	CertLoaderConfig                                      certloader.Config
	ClientTimeout                                         time.Duration `envconfig:"default=60s"`
	SkipSSLValidation                                     bool          `envconfig:"default=false"`
	SubscriptionConfig                                    subscription.Config
}

var (
	testConfig config

	certSecuredGraphQLClient *graphql.Client
	certCache                certloader.Cache
)

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	ctx := context.Background()

	certCache, err = certloader.StartCertLoader(ctx, testConfig.CertLoaderConfig)
	if err != nil {
		log.D().Fatal(errors.Wrap(err, "while starting cert cache"))
	}

	if err := util.WaitForCache(certCache); err != nil {
		log.D().Fatal(err)
	}

	certSecuredGraphQLClient = gql.NewCertAuthorizedGraphQLClientWithCustomURL(testConfig.DirectorExternalCertSecuredURL, certCache.Get().PrivateKey, certCache.Get().Certificate, testConfig.SkipSSLValidation)

	exitVal := m.Run()
	os.Exit(exitVal)

}

func parseCronTime(cronTime string) (time.Duration, error) {
	parser := c.NewParser(c.Minute | c.Hour | c.Dom | c.Month | c.Dow)
	scheduleTime, err := parser.Parse(cronTime)
	if err != nil {
		return 0, errors.New("error while parsing cron time")
	}

	// This is the starting time that will be subtracted from the next activation cron time below. This way the cron time duration can be estimated.
	year, month, day := time.Now().Date()
	startingTime := time.Date(year, month, day, 0, 0, 0, 0, time.Now().Location())

	nextTime := scheduleTime.Next(startingTime)
	cronTimeDuration := nextTime.Sub(startingTime)

	return cronTimeDuration, nil
}
