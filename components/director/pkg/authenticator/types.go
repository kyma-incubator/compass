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
	"errors"
)

const CoordinatesKey = "authenticator_coordinates"

// Config holds all configuration related to an additional authenticator provided to the Director
type Config struct {
	Name           string          `json:"name"`
	TrustedIssuers []TrustedIssuer `json:"trusted_issuers"`
	Attributes     Attributes      `json:"attributes"`
}

type TrustedIssuer struct {
	DomainURL   string `json:"domain_url"`
	ScopePrefix string `json:"scope_prefix"`
}

type Coordinates struct {
	Name  string `json:"name"`
	Index int    `json:"index"`
}

// Attributes holds all attribute properties and values related to an authenticator
type Attributes struct {
	UniqueAttribute   Attribute `json:"uniqueAttribute"`
	IdentityAttribute Attribute `json:"identity"`
	TenantAttribute   Attribute `json:"tenant"`
}

// Validate validates all attributes
func (a *Attributes) Validate() error {
	for _, attr := range []Attribute{a.UniqueAttribute, a.IdentityAttribute, a.TenantAttribute} {
		if err := attr.Validate(); err != nil {
			return err
		}
	}

	if a.UniqueAttribute.Value == "" {
		return errors.New("unique attribute value cannot be empty")
	}

	return nil
}

// Attribute represents a single attribute associated with an authenticator
type Attribute struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

// Validate validates the attribute
func (a *Attribute) Validate() error {
	if a.Key == "" {
		return errors.New("attribute key cannot be empty")
	}

	return nil
}
