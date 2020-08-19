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

package server

import (
	"fmt"
	"time"
)

// Settings type to be loaded from the environment
type Config struct {
	Port            int           `mapstructure:"port" description:"port of the server"`
	RequestTimeout  time.Duration `mapstructure:"request_timeout" description:"read and write timeout duration for requests"`
	ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout" description:"time to wait for the server to shutdown"`
	RootAPI         string        `mapstructure:"root_api" description:"the root api used for all other subroutes"`
}

// DefaultSettings returns the default values for configuring the Service Manager
func DefaultConfig() *Config {
	return &Config{
		Port:            8080,
		RequestTimeout:  time.Second * 50000,
		ShutdownTimeout: time.Second * 50000,
		RootAPI:         "/broker",
	}
}

// Validate validates the server settings
func (s *Config) Validate() error {
	if s.Port == 0 {
		return fmt.Errorf("validate Settings: Port missing")
	}
	if s.RequestTimeout == 0 {
		return fmt.Errorf("validate Settings: RequestTimeout missing")
	}
	if s.ShutdownTimeout == 0 {
		return fmt.Errorf("validate Settings: ShutdownTimeout missing")
	}

	return nil
}
