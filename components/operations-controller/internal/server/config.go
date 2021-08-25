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
	"github.com/kyma-incubator/compass/components/system-broker/pkg/server"
)

// Settings type to be loaded from the environment
type Config struct {
	*server.Config
	MetricAddress        string `mapstructure:"metric_address" description:"the address the metric endpoint binds to"`
	HealthAddress        string `mapstructure:"health_address" description:"the address the health endpoint binds to"`
	EnableLeaderElection bool   `mapstructure:"enable_leader_election" description:"enable leader election for controller manager"`
}

// DefaultSettings returns the default values for configuring the System Broker
func DefaultConfig() *Config {
	return &Config{
		Config:               server.DefaultConfig(),
		MetricAddress:        ":8080",
		EnableLeaderElection: false,
	}
}
