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

package oauth

import "github.com/pkg/errors"

type Config struct {
	Local           bool   `mapstructure:"local"`
	TokenValue      string `mapstructure:"token_value"`
	SecretName      string `mapstructure:"secret_name"`
	SecretNamespace string `mapstructure:"secret_namespace"`
}

func DefaultConfig() *Config {
	return &Config{
		Local:           false,
		TokenValue:      "eyJhbGciOiJub25lIiwidHlwIjoiSldUIn0.eyJzY29wZXMiOiJhcHBsaWNhdGlvbjpyZWFkIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OndyaXRlIGF1dG9tYXRpY19zY2VuYXJpb19hc3NpZ25tZW50OnJlYWQgaGVhbHRoX2NoZWNrczpyZWFkIGFwcGxpY2F0aW9uOndyaXRlIHJ1bnRpbWU6d3JpdGUgbGFiZWxfZGVmaW5pdGlvbjp3cml0ZSBsYWJlbF9kZWZpbml0aW9uOnJlYWQgcnVudGltZTpyZWFkIHRlbmFudDpyZWFkIiwidGVuYW50IjoiM2U2NGViYWUtMzhiNS00NmEwLWIxZWQtOWNjZWUxNTNhMGFlIn0.",
		SecretName:      "compass-system-broker-credentials",
		SecretNamespace: "compass-system",
	}
}

func (c *Config) Validate() error {
	if c.SecretName == "" {
		return errors.New("secret name cannot be empty")
	}

	if c.SecretNamespace == "" {
		return errors.New("secret namespace cannot be empty")
	}

	if c.Local && len(c.TokenValue) == 0 {
		return errors.New("token value cannot be empty when run locally")
	}

	return nil
}
