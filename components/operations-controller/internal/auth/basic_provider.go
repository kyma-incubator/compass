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

package auth

import (
	"context"
	"encoding/base64"

	"github.com/pkg/errors"
)

// BasicAuthorizationProvider presents an AuthorizationProvider implementation which crafts Basic Authentication header values for the Authorization header
type basicAuthorizationProvider struct{}

// NewBasicAuthorizationProvider constructs a BasicAuthorizationProvider
func NewBasicAuthorizationProvider() *basicAuthorizationProvider {
	return &basicAuthorizationProvider{}
}

// Name specifies the name of the AuthorizationProvider
func (u basicAuthorizationProvider) Name() string {
	return "BasicAuthorizationProvider"
}

// Matches contains the logic for matching the AuthorizationProvider
func (u basicAuthorizationProvider) Matches(ctx context.Context) bool {
	credentials, err := LoadFromContext(ctx)
	if err != nil {
		return false
	}

	return credentials.Type() == BasicCredentialType
}

// GetAuthorization prepares the Authorization header Basic Authentication value for the executing request
func (u basicAuthorizationProvider) GetAuthorization(ctx context.Context) (string, error) {
	credentials, err := LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	basicCredentials, ok := credentials.Get().(*BasicCredentials)
	if !ok {
		return "", errors.New("failed to cast credentials to basic credentials type")
	}

	auth := basicCredentials.Username + ":" + basicCredentials.Password
	encodedAuth := base64.StdEncoding.EncodeToString([]byte(auth))

	return "Basic " + encodedAuth, nil
}
