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

package main

import (
	"context"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	"github.com/kyma-incubator/compass/components/system-broker/internal/config"
	"github.com/kyma-incubator/compass/components/system-broker/internal/director"
	"github.com/kyma-incubator/compass/components/system-broker/internal/osb"
	"github.com/kyma-incubator/compass/components/system-broker/internal/specs"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/env"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/graphql"
	httputil "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/oauth"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/server"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/signal"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/uid"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	signal.HandleInterrupts(ctx, cancel)

	env, err := env.Default(ctx, config.AddPFlags)
	fatalOnError(err)

	cfg, err := config.New(env)
	fatalOnError(err)

	err = cfg.Validate()
	fatalOnError(err)

	ctx, err = log.Configure(ctx, cfg.Log)
	fatalOnError(err)

	uuidSrv := uid.NewService()

	directorGraphQLClient, err := prepareGqlClient(cfg, uuidSrv)
	fatalOnError(err)

	systemBroker := osb.NewSystemBroker(directorGraphQLClient, cfg.Server.SelfURL)
	osbApi := osb.API(cfg.Server.RootAPI, systemBroker, log.NewDefaultLagerAdapter())
	specsApi := specs.API(cfg.Server.RootAPI, directorGraphQLClient)
	srv, err := server.New(cfg.Server, uuidSrv, osbApi, specsApi.Routes)
	fatalOnError(err)

	srv.Start(ctx)
}

func fatalOnError(err error) {
	if err != nil {
		log.D().Fatal(err.Error())
	}
}

func prepareGqlClient(cfg *config.Config, uudSrv httputil.UUIDService) (*director.GraphQLClient, error) {
	// prepare raw http transport and http client based on cfg
	httpTransport := httputil.NewCorrelationIDTransport(httputil.NewErrorHandlerTransport(httputil.NewHTTPTransport(cfg.HttpClient)), uudSrv)
	httpClient := httputil.NewClient(cfg.HttpClient.Timeout, httpTransport)

	//prepare secured http client with token provider picked from secret
	requestProvider := httputil.NewRequestProvider(uudSrv)

	oauthTokenProvider, err := oauth.NewTokenProvider(cfg.OAuthProvider, httpClient, requestProvider)
	if err != nil {
		return nil, err
	}
	securedClient, err := httputil.NewSecuredHTTPClient(cfg.HttpClient.Timeout, httpTransport, oauthTokenProvider)
	if err != nil {
		return nil, err
	}

	// prepare graphql client that uses secured http client as a basis
	gqlClient, err := graphql.NewClient(cfg.GraphQLClient, securedClient)
	if err != nil {
		return nil, err
	}

	inputGraphqlizer := &graphqlizer.Graphqlizer{}
	outputGraphqlizer := &graphqlizer.GqlFieldsProvider{}

	// prepare director graphql client
	return director.NewGraphQLClient(gqlClient, inputGraphqlizer, outputGraphqlizer), nil
}


