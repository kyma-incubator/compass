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

package webhook

import (
	"fmt"
	"time"

	"github.com/pkg/errors"
)

// Settings type to be loaded from the environment
type Config struct {
	TimeoutFactor   int           `mapstructure:"timeout_factor" description:"the factor by which to multiple the reconciliation timeout"`
	WebhookTimeout  time.Duration `mapstructure:"webhook_timeout" description:"defines the maximum time to process a webhook"`
	RequeueInterval time.Duration `mapstructure:"requeue_interval" description:"defines the default requeue interval"`
	TimeLayout      string        `mapstructure:"time_layout" description:"defines the default timestamp time layout"`
}

// DefaultSettings returns the default values for configuring the System Broker
func DefaultConfig() *Config {
	return &Config{
		TimeoutFactor:   2,
		WebhookTimeout:  2 * time.Hour,
		RequeueInterval: 2 * time.Minute,
		TimeLayout:      time.RFC3339Nano,
	}
}

func (s *Config) Validate() error {
	if s.TimeoutFactor <= 0 {
		return errors.New("validate webhook settings: timeout factor should be > 0")
	}
	if s.WebhookTimeout < 0 {
		return errors.New("validate webhook settings: webhook timeout should be >= 0")
	}
	if s.RequeueInterval < 0 {
		return errors.New("validate webhook settings: requeue interval should be >= 0")
	}
	if s.TimeLayout != time.RFC3339Nano {
		return fmt.Errorf("validate webhook settings: time layout should be %s", time.RFC3339Nano)
	}
	return nil
}
