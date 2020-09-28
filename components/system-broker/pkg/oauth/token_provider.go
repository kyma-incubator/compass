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
	"context"
	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"net/http"
)

const (
	contentTypeHeader                = "Content-Type"
	contentTypeApplicationURLEncoded = "application/x-www-form-urlencoded"

	grantTypeFieldName   = "grant_type"
	credentialsGrantType = "client_credentials"

	scopeFieldName = "scope"
	scopes         = "application:read application:write runtime:read runtime:write"

	clientIDKey       = "client_id"
	clientSecretKey   = "client_secret"
	tokensEndpointKey = "tokens_endpoint"
)

type RequestProvider interface {
	Provide(ctx context.Context, input httputils.RequestInput) (*http.Request, error)
}

func NewTokenProvider(config *Config, httpClient *http.Client, requestProvider RequestProvider) (httputils.TokenProvider, error) {
	if config.Local {
		return NewTokenProviderFromValue(config.TokenValue), nil
	}
	return NewTokenProviderFromSecret(config, httpClient, requestProvider)
}
