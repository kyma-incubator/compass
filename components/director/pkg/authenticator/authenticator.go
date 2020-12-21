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

package authenticator

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"regexp"
	"strings"
)

// InitFromEnv loads authenticator configurations from environment if any exist
func InitFromEnv(envPrefix string) ([]Config, error) {
	authenticators := make(map[string]*Config, 0)
	attributesPattern := regexp.MustCompile(fmt.Sprintf("^%s_(.*)_AUTHENTICATOR_ATTRIBUTES$", envPrefix))
	scopePrefixPattern := regexp.MustCompile(fmt.Sprintf("^%s_(.*)_AUTHENTICATOR_SCOPE_PREFIXES$", envPrefix))

	for _, env := range os.Environ() {
		pair := strings.SplitN(env, "=", 2)
		key := pair[0]
		value := pair[1]

		matches := attributesPattern.FindStringSubmatch(key)
		if len(matches) > 0 {
			authenticatorName := matches[1]
			var attributes Attributes
			if err := json.Unmarshal([]byte(value), &attributes); err != nil {
				return nil, errors.New(fmt.Sprintf("unable to unmarshal environment variable with key %s and value %s: %s", key, value, err))
			}

			if authenticator, exists := authenticators[authenticatorName]; exists {
				authenticator.Attributes = attributes
			} else {
				authenticators[authenticatorName] = &Config{
					Name:       authenticatorName,
					Attributes: attributes,
				}
			}

			continue
		}

		matches = scopePrefixPattern.FindStringSubmatch(key)
		if len(matches) > 0 {
			authenticatorName := matches[1]

			prefixes := strings.Split(value, ",")
			if authenticator, exists := authenticators[authenticatorName]; exists {
				authenticator.ScopePrefixes = prefixes
			} else {
				authenticators[authenticatorName] = &Config{
					Name:          authenticatorName,
					ScopePrefixes: prefixes,
				}
			}
		}
	}

	result := make([]Config, 0, len(authenticators))
	for _, config := range authenticators {
		if err := config.Attributes.Validate(); err != nil {
			return nil, errors.New(fmt.Sprintf("insufficient configuration provided for authenticator %q: %s", config.Name, err.Error()))
		}

		result = append(result, *config)
	}

	return result, nil
}
