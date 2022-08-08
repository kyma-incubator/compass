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

package graphql

import (
	"context"
	"net/http"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	httpdirector "github.com/kyma-incubator/compass/components/director/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/director"
	httputil "github.com/kyma-incubator/compass/components/system-broker/pkg/http"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . GraphQLClient
type GraphQLClient interface {
	Run(ctx context.Context, req *graphql.Request, resp interface{}) error
}

type Client struct {
	gqlClient GraphQLClient
	logs      []string
	logging   bool
}

func PrepareGqlClient(cfg *Config, httpCfg *httputil.Config, providers ...httpdirector.AuthorizationProvider) (*director.GraphQLClient, error) {
	httpTransport := httpdirector.NewCorrelationIDTransport(httpdirector.NewServiceAccountTokenTransport(httpdirector.NewErrorHandlerTransport(httpdirector.NewHTTPTransportWrapper(httputil.NewHTTPTransport(httpCfg)))))

	securedTransport := httpdirector.NewSecuredTransport(httpTransport, providers...)
	securedClient := &http.Client{
		Transport: securedTransport,
		Timeout:   httpCfg.Timeout,
	}

	// prepare graphql client that uses secured http client as a basis
	// the provided endpoint in the graphClient will be changed in the secured transport based on the matching token provider
	return PrepareGqlClientWithHttpClient(cfg, securedClient)
}

func PrepareGqlClientWithHttpClient(cfg *Config, httpClient *http.Client) (*director.GraphQLClient, error) {
	graphClient := graphql.NewClient(cfg.GraphqlEndpoint, graphql.WithHTTPClient(httpClient))
	gqlClient := NewClient(cfg, graphClient)

	inputGraphqlizer := &graphqlizer.Graphqlizer{}
	outputGraphqlizer := &graphqlizer.GqlFieldsProvider{}

	// prepare director graphql client
	return director.NewGraphQLClient(gqlClient, inputGraphqlizer, outputGraphqlizer), nil
}

func NewClient(config *Config, gqlClient GraphQLClient) *Client {
	client := &Client{
		gqlClient: gqlClient,
		logging:   config.EnableLogging,
		logs:      []string{},
	}

	if c, ok := client.gqlClient.(*graphql.Client); ok {
		c.Log = client.addLog
	}

	return client
}

func (c *Client) Do(ctx context.Context, req *graphql.Request, res interface{}) error {
	c.clearLogs()

	if err := c.gqlClient.Run(ctx, req, res); err != nil {
		err = errors.Wrap(err, "while using gqlclient")
		for _, l := range c.logs {
			if l != "" {
				log.C(ctx).WithError(err).Error(l)
			}
		}
		return err
	}

	return nil
}

func (c *Client) addLog(log string) {
	if !c.logging {
		return
	}

	c.logs = append(c.logs, log)
}

func (c *Client) clearLogs() {
	c.logs = []string{}
}
