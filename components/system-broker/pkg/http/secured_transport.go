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

package http

import (
	"context"
	"net/http"
	"net/url"

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . TokenProvider
type TokenProvider interface {
	Name() string
	Matches(ctx context.Context) bool
	TargetURL() *url.URL
	GetAuthorizationToken(ctx context.Context) (Token, error)
}

type SecuredTransport struct {
	roundTripper   HTTPRoundTripper
	tokenProviders []TokenProvider
}

func NewSecuredTransport(roundTripper HTTPRoundTripper, providers ...TokenProvider) *SecuredTransport {
	return &SecuredTransport{
		roundTripper:   roundTripper,
		tokenProviders: providers,
	}
}

func (c *SecuredTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	logger := log.C(request.Context())

	targetURL, token, err := c.getCredentialsFromProvider(request.Context())
	if err != nil {
		logger.Errorf("Could not get token for request: %s", err.Error())
		return nil, err
	}

	logger.Debug("Successfully got token for request")
	request.Header.Set("Authorization", "Bearer "+token.AccessToken)

	request.URL = targetURL

	return c.roundTripper.RoundTrip(request)
}

func (c *SecuredTransport) getCredentialsFromProvider(ctx context.Context) (*url.URL, Token, error) {
	for _, tokenProvider := range c.tokenProviders {
		if !tokenProvider.Matches(ctx) {
			continue
		}
		log.C(ctx).Debugf("Successfully matched '%s' token provider", tokenProvider.Name())

		token, err := tokenProvider.GetAuthorizationToken(ctx)
		if err != nil {
			return nil, Token{}, errors.Wrap(err, "error while obtaining token")
		}

		return tokenProvider.TargetURL(), token, nil
	}

	return nil, Token{}, errors.New("context did not match any token provider")
}
