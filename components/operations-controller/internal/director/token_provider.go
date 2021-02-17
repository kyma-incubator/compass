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

package director

import (
	"context"
	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/operations-controller/internal/tenant"
	httputils "github.com/kyma-incubator/compass/components/system-broker/pkg/http"
	"github.com/pkg/errors"
	"net/url"
)

const applicationReadScope = "application:read"

// UnsignedTokenProvider presents a TokenProvider implementation which fabricates it's own unsigned tokens
type UnsignedTokenProvider struct {
	targetURL *url.URL
}

type claims struct {
	Scopes string `json:"scopes"`
	Tenant string `json:"tenant"`
	jwt.StandardClaims
}

// NewUnsignedTokenProvider constructs an UnsignedTokenProvider
func NewUnsignedTokenProvider(targetURL string) (*UnsignedTokenProvider, error) {
	parsedURL, err := url.Parse(targetURL)
	if err != nil {
		return nil, err
	}

	return &UnsignedTokenProvider{
		targetURL: parsedURL,
	}, nil
}

// Name specifies the name of the TokenProvider
func (u UnsignedTokenProvider) Name() string {
	return "UnsignedTokenProvider"
}

// Matches contains the logic for matching the TokenProvider
func (u UnsignedTokenProvider) Matches(_ context.Context) bool {
	return true
}

// TargetURL returns the intented TargetURL for the executing request
func (u UnsignedTokenProvider) TargetURL() *url.URL {
	return u.targetURL
}

// GetAuthorizationToken crafts an unsigned token to inject in the executing request
func (u UnsignedTokenProvider) GetAuthorizationToken(ctx context.Context) (httputils.Token, error) {
	tenantID, err := tenant.LoadFromContext(ctx)
	if err != nil {
		return httputils.Token{}, err
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims{
		Tenant: tenantID,
		Scopes: applicationReadScope,
	})

	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	if err != nil {
		return httputils.Token{}, errors.Wrap(err, "while signing token")
	}

	return httputils.Token{
		AccessToken: signedToken,
		Expiration:  0,
	}, nil
}
