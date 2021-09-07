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

package log

import (
	"fmt"
	"os"

	"github.com/sirupsen/logrus"
)

// Config missing godoc
type Config struct {
	Level                  string `envconfig:"APP_LOG_LEVEL,default=info"`
	Format                 string `envconfig:"APP_LOG_FORMAT,default=text"`
	Output                 string `envconfig:"APP_LOG_OUTPUT,default=/dev/stdout"`
	BootstrapCorrelationID string `envconfig:"APP_LOG_BOOTSTRAP_CORRELATION_ID,default=bootstrap"`
}

// DefaultConfig returns default values for Log settings
func DefaultConfig() *Config {
	return &Config{
		Level:                  "info",
		Format:                 "text",
		Output:                 os.Stdout.Name(),
		BootstrapCorrelationID: "bootstrap",
	}
}

// Validate validates the logging settings
func (s *Config) Validate() error {
	if _, err := logrus.ParseLevel(s.Level); err != nil {
		return fmt.Errorf("validate Config: log level %s is invalid: %s", s.Level, err)
	}

	if len(s.Format) == 0 {
		return fmt.Errorf("validate Config: log format missing")
	}

	if _, ok := supportedFormatters[s.Format]; !ok {
		return fmt.Errorf("validate Config: log format %s is invalid", s.Format)
	}

	if len(s.Output) == 0 {
		return fmt.Errorf("validate Config: log output missing")
	}

	if _, ok := supportedOutputs[s.Output]; !ok {
		return fmt.Errorf("validate Config: log output %s is invalid", s.Output)
	}

	if s.BootstrapCorrelationID == "" {
		return fmt.Errorf("validate Config: bootstrap correlation id cannot be empty")
	}

	return nil
}
