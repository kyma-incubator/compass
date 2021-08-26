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

	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/kyma-incubator/compass/components/system-broker/pkg/oauth"
	"github.com/pkg/errors"
)

// TokenAuthorizationProvider presents a AuthorizationProvider implementation which crafts OAuth Bearer token values for the Authorization header
type tokenAuthorizationProvider struct {
	httpClient httputils.Client
}

// NewTokenAuthorizationProvider constructs an TokenAuthorizationProvider
func NewTokenAuthorizationProvider(httpClient httputils.Client) *tokenAuthorizationProvider {
	return &tokenAuthorizationProvider{
		httpClient: httpClient,
	}
}

// Name specifies the name of the AuthorizationProvider
func (u tokenAuthorizationProvider) Name() string {
	return "TokenAuthorizationProvider"
}

// Matches contains the logic for matching the AuthorizationProvider
func (u tokenAuthorizationProvider) Matches(ctx context.Context) bool {
	credentials, err := LoadFromContext(ctx)
	if err != nil {
		return false
	}

	return credentials.Type() == OAuthCredentialType
}

// GetAuthorization crafts an OAuth Bearer token to inject as part of the executing request
func (u tokenAuthorizationProvider) GetAuthorization(ctx context.Context) (string, error) {
	credentials, err := LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	oAuthCredentials, ok := credentials.Get().(*OAuthCredentials)
	if !ok {
		return "", errors.New("failed to cast credentials to oauth credentials type")
	}

	token, err := oauth.GetAuthorizationToken(ctx, u.httpClient, oauth.Credentials{
		ClientID:     oAuthCredentials.ClientID,
		ClientSecret: oAuthCredentials.ClientSecret,
		TokenURL:     oAuthCredentials.TokenURL,
	}, "")
	if err != nil {
		return "", err
	}

	return "Bearer " + token.AccessToken, nil
}
