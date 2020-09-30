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
	"github.com/sirupsen/logrus"
	"os"
)

const (
	// FieldComponentName is the key of the component field in the log message.
	FieldComponentName = "component"
)

// Config type to be loaded from the environment
type Config struct {
	Level                  string `description:"minimum level for log messages" json:"level,omitempty" mapstructure:"level"`
	Format                 string `description:"format of log messages. Allowed values - text, json" json:"format,omitempty" mapstructure:"format"`
	Output                 string `description:"output for the logs. Allowed values - /dev/stdout, /dev/stderr" json:"output,omitempty" mapstructure:"output"`
	BootstrapCorrelationID string `description:"the value of the bootstrap correlation id for this component" json:"bootstrap_correlation_id,omitempty" mapstructure:"bootstrap_correlation_id"`
}

// DefaultConfig returns default values for Log settings
func DefaultConfig() *Config {
	return &Config{
		Level:                  "info",
		Format:                 "text",
		Output:                 os.Stdout.Name(),
		BootstrapCorrelationID: "system-broker-bootstrap",
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
