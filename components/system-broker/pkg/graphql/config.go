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

import "github.com/pkg/errors"

type Config struct {
	GraphqlEndpoint string `mapstructure:"graphql_endpoint"`
	// used when system broker acts like an integration system
	GraphqlOauthEndpoint string `mapstructure:"graphql_oauth_endpoint"`
	EnableLogging        bool   `mapstructure:"enable_logging"`
}

func DefaultConfig() *Config {
	return &Config{
		GraphqlEndpoint:      "http://localhost:3000/graphql",
		GraphqlOauthEndpoint: "http://localhost:3000/graphql",
		EnableLogging:        true,
	}
}

func (c *Config) Validate() error {
	if c.GraphqlEndpoint == "" {
		return errors.New("graphql endpoint cannot be empty")
	}

	return nil
}
