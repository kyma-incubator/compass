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

	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/pkg/errors"
)

// AuthorizationProvider missing godoc
//go:generate go run github.com/maxbrunsfeld/counterfeiter/v6 . AuthorizationProvider
type AuthorizationProvider interface {
	Name() string
	Matches(ctx context.Context) bool
	GetAuthorization(ctx context.Context) (string, error)
}

// SecuredTransport missing godoc
type SecuredTransport struct {
	roundTripper           HTTPRoundTripper
	authorizationProviders []AuthorizationProvider
}

// NewSecuredTransport missing godoc
func NewSecuredTransport(roundTripper HTTPRoundTripper, providers ...AuthorizationProvider) *SecuredTransport {
	return &SecuredTransport{
		roundTripper:           roundTripper,
		authorizationProviders: providers,
	}
}

// RoundTrip missing godoc
func (c *SecuredTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	logger := log.C(request.Context())

	authorization, err := c.getAuthorizationFromProvider(request.Context())
	if err != nil {
		logger.Errorf("Could not prepare authorization for request: %s", err.Error())
		return nil, err
	}

	logger.Debug("Successfully prepared authorization for request")
	request.Header.Set("Authorization", authorization)

	return c.roundTripper.RoundTrip(request)
}

// Clone clones the underlying transport
func (c *SecuredTransport) Clone() HTTPRoundTripper {
	return &SecuredTransport{
		roundTripper:           c.roundTripper.Clone(),
		authorizationProviders: c.authorizationProviders,
	}
}

// GetTransport returns the underlying transport.
func (c *SecuredTransport) GetTransport() *http.Transport {
	return c.roundTripper.GetTransport()
}

func (c *SecuredTransport) getAuthorizationFromProvider(ctx context.Context) (string, error) {
	for _, authProvider := range c.authorizationProviders {
		if !authProvider.Matches(ctx) {
			continue
		}
		log.C(ctx).Debugf("Successfully matched '%s' authorization provider", authProvider.Name())

		authorization, err := authProvider.GetAuthorization(ctx)
		if err != nil {
			return "", errors.Wrap(err, "error while obtaining authorization")
		}

		return authorization, nil
	}

	return "", errors.New("context did not match any authorization provider")
}
