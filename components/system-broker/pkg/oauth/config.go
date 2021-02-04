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

import (
	"time"

	"github.com/pkg/errors"
)

type Config struct {
	SecretName            string        `mapstructure:"secret_name"`
	SecretNamespace       string        `mapstructure:"secret_namespace"`
	WaitSecretTimeout     time.Duration `mapstructure:"wait_secret_timeout"`
	WaitKubeMapperTimeout time.Duration `mapstructure:"wait_kube_mapper_timeout"`
}

func DefaultConfig() *Config {
	return &Config{
		SecretName:            "compass-system-broker-credentials",
		SecretNamespace:       "compass-system",
		WaitSecretTimeout:     time.Minute,
		WaitKubeMapperTimeout: time.Minute,
	}
}

func (c *Config) Validate() error {
	if c.SecretName == "" {
		return errors.New("secret name cannot be empty")
	}

	if c.SecretNamespace == "" {
		return errors.New("secret namespace cannot be empty")
	}

	if c.WaitSecretTimeout <= 0 {
		return errors.New("wait secret timeout must be greater than zero")
	}

	if c.WaitKubeMapperTimeout <= 0 {
		return errors.New("wait kube mapper timeout must be greater than zero")
	}

	return nil
}
