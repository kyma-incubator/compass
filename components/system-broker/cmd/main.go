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
	"net/http"
	"os"

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
	"github.com/kyma-incubator/compass/components/system-broker/pkg/uuid"
	gql "github.com/machinebox/graphql"
)

func main() {
	ctx, cancel := context.WithCancel(context.Background())

	defer cancel()

	term := make(chan os.Signal)
	signal.HandleInterrupts(ctx, cancel, term)

	env, err := env.Default(ctx, config.AddPFlags)
	fatalOnError(err)

	cfg, err := config.New(env)
	fatalOnError(err)

	err = cfg.Validate()
	fatalOnError(err)

	ctx, err = log.Configure(ctx, cfg.Log)
	fatalOnError(err)

	uuidSrv := uuid.NewService()

	directorGraphQLClient, err := prepareGqlClient(ctx, cfg, uuidSrv)
	fatalOnError(err)

	systemBroker := osb.NewSystemBroker(directorGraphQLClient, cfg.Server.SelfURL+cfg.Server.RootAPI)
	osbApi := osb.API(cfg.Server.RootAPI, systemBroker, log.NewDefaultLagerAdapter())
	specsApi := specs.API(cfg.Server.RootAPI, directorGraphQLClient)
	srv := server.New(cfg.Server, uuidSrv, cfg.HttpClient.ForwardHeaders, osbApi, specsApi)

	srv.Start(ctx)
}

func fatalOnError(err error) {
	if err != nil {
		log.D().Fatal(err.Error())
	}
}

func prepareGqlClient(ctx context.Context, cfg *config.Config, uudSrv uuid.Service) (*director.GraphQLClient, error) {
	// prepare raw http transport and http client based on cfg
	httpTransport := httputil.NewCorrelationIDTransport(httputil.NewErrorHandlerTransport(httputil.NewHTTPTransport(cfg.HttpClient)), uudSrv)
	httpClient := httputil.NewClient(cfg.HttpClient.Timeout, httpTransport)

	tokenProviders := make([]httputil.TokenProvider, 0)
	tokenProviderFromHeader := oauth.NewTokenProviderFromHeader()
	tokenProviders = append(tokenProviders, tokenProviderFromHeader)
	tokenProviderFromSecret, err := oauth.NewTokenProviderFromSecret(cfg.OAuthProvider, httpClient, cfg.HttpClient.Timeout, oauth.PrepareK8sClient)
	if err != nil {
		log.C(ctx).Warnf("Could not initialize token provider from secret: %s", err.Error())
	} else {
		tokenProviders = append(tokenProviders, tokenProviderFromSecret)
	}

	securedTransport := httputil.NewSecuredTransport(httpTransport, tokenProviders...)
	securedClient := &http.Client{
		Transport: securedTransport,
		Timeout:   cfg.HttpClient.Timeout,
	}

	// prepare graphql client that uses secured http client as a basis
	graphClient := gql.NewClient(cfg.GraphQLClient.GraphqlEndpoint, gql.WithHTTPClient(securedClient))
	gqlClient := graphql.NewClient(cfg.GraphQLClient, graphClient)

	inputGraphqlizer := &graphqlizer.Graphqlizer{}
	outputGraphqlizer := &graphqlizer.GqlFieldsProvider{}

	// prepare director graphql client
	return director.NewGraphQLClient(gqlClient, inputGraphqlizer, outputGraphqlizer), nil
}
