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
	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/tenant"
	"github.com/pkg/errors"
	"net/url"
)

const applicationReadScope = "application:read"

// UnsignedTokenAuthorizationProvider presents an AuthorizationProvider implementation which fabricates its own unsigned tokens for the Authorization header
type unsignedTokenAuthorizationProvider struct {
	targetURL *url.URL
}

// Claims defines the custom claims which will be placed inside tokens crafted by the unsignedTokenAuthorizationProvider
type Claims struct {
	Scopes string `json:"scopes"`
	Tenant string `json:"tenant"`
	jwt.StandardClaims
}

// NewUnsignedTokenAuthorizationProvider constructs an UnsignedTokenAuthorizationProvider
func NewUnsignedTokenAuthorizationProvider(targetURL string) (*unsignedTokenAuthorizationProvider, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	return &unsignedTokenAuthorizationProvider{
		targetURL: parsedURL,
	}, nil
}

// Name specifies the name of the AuthorizationProvider
func (u unsignedTokenAuthorizationProvider) Name() string {
	return "UnsignedTokenAuthorizationProvider"
}

// Matches contains the logic for matching the AuthorizationProvider
func (u unsignedTokenAuthorizationProvider) Matches(ctx context.Context) bool {
	_, err := LoadFromContext(ctx)
	if err != nil {
		return true
	}

	return false
}

// TargetURL returns the intented TargetURL for the executing request
func (u unsignedTokenAuthorizationProvider) TargetURL() *url.URL {
	return u.targetURL
}

// GetAuthorizationToken crafts an unsigned token to inject in the executing request
func (u unsignedTokenAuthorizationProvider) GetAuthorization(ctx context.Context) (string, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return "", err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, Claims{
		Tenant: tenantID,
		Scopes: applicationReadScope,
	})

	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return "", errors.Wrap(err, "while signing token")
	}

	return "Bearer " + signedToken, nil
}
