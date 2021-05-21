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

package director

import "github.com/pkg/errors"

// Settings type to be loaded from the environment
type Config struct {
	OperationEndpoint string `mapstructure:"operation_endpoint" description:"the operation endpoint of the Director component"`
}

// DefaultSettings returns the default values for configuring the System Broker
func DefaultConfig() *Config {
	return &Config{
		OperationEndpoint: "http://localhost:3002/operation",
	}
}

// Validate ensures that the director config properties have valid values
func (c *Config) Validate() error {
	if len(c.OperationEndpoint) == 0 {
		return errors.New("validate director settings: missing internal address")
	}
	return nil
}
