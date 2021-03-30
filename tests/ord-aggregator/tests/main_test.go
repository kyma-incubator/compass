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
	"log"
	"os"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/tests/pkg/gql"
	"github.com/kyma-incubator/compass/tests/pkg/idtokenprovider"

	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
	c "github.com/robfig/cron/v3"
	"github.com/vrischmann/envconfig"
)

type config struct {
	DefaultTenant                 string
	DirectorURL                   string
	ORDServiceURL                 string
	AggregatorSchedule            string
	ExternalServicesMockBaseURL   string
	ORDServiceDefaultResponseType string
}

var (
	testConfig       config
	dexGraphQLClient *graphql.Client
)

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	dexToken, err := idtokenprovider.GetDexToken()
	if err != nil {
		log.Fatal(errors.Wrap(err, "while getting dex token"))
	}
	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

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
