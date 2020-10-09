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

	"github.com/kyma-incubator/compass/components/system-broker/pkg/log"
	"github.com/machinebox/graphql"
	"github.com/pkg/errors"
)

//go:generate counterfeiter . GraphQLClient
type GraphQLClient interface {
	Run(ctx context.Context, req *graphql.Request, resp interface{}) error
}

type Client struct {
	gqlClient GraphQLClient
	logs      []string
	logging   bool
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
