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
	"os"
	"testing"

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
)

type config struct {
	DirectorURL                   string
	ORDServiceURL                 string
	ORDServiceStaticURL           string
	ORDServiceDefaultResponseType string
	DefaultScenarioEnabled        bool `envconfig:"default=true"`
}

var testConfig config

func TestMain(m *testing.M) {
	err := envconfig.Init(&testConfig)
	if err != nil {
		log.Fatal(errors.Wrap(err, "while initializing envconfig"))
	}

	tenant.TestTenants.Init()
	defer tenant.TestTenants.Cleanup()

	dexToken := server.Token()

	dexGraphQLClient = gql.NewAuthorizedGraphQLClient(dexToken)

	exitVal := m.Run()
	os.Exit(exitVal)

}
