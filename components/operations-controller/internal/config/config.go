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

package config

import (
	"github.com/kyma-incubator/compass/components/operations-controller/internal/director"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/server"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/webhook"
	"reflect"

	"github.com/kyma-incubator/compass/components/system-broker/pkg/env"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/graphql"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/pkg/errors"
	"github.com/spf13/pflag"
)

type Validatable interface {
	Validate() error
}

type Config struct {
	Server        *server.Config   `mapstructure:"server"`
	HttpClient    *http.Config     `mapstructure:"http_client"`
	GraphQLClient *graphql.Config  `mapstructure:"graphql_client"`
	Director      *director.Config `mapstructure:"director"`
	Webhook       *webhook.Config  `mapstructure:"webhook"`
}

func AddPFlags(set *pflag.FlagSet) {
	env.CreatePFlags(set, DefaultConfig())
	env.CreatePFlagsForConfigFile(set)
}

func DefaultConfig() *Config {
	return &Config{
		Server:        server.DefaultConfig(),
		HttpClient:    http.DefaultConfig(),
		GraphQLClient: graphql.DefaultConfig(),
		Webhook:       webhook.DefaultConfig(),
	}
}

func New(env env.Environment) (*Config, error) {
	config := DefaultConfig()
	if err := env.Unmarshal(config); err != nil {
		return nil, errors.Wrapf(err, "error loading cfg")
	}

	return config, nil
}

func (c *Config) Validate() error {
	validatableFields := make([]Validatable, 0, 0)

	v := reflect.ValueOf(*c)
	for i := 0; i < v.NumField(); i++ {
		field, ok := v.Field(i).Interface().(Validatable)
		if ok {
			validatableFields = append(validatableFields, field)
		}
	}

	for _, item := range validatableFields {
		if err := item.Validate(); err != nil {
			return err
		}
	}
	return nil
}
