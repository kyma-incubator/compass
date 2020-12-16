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

package authnmappinghandler

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"

	goidc "github.com/coreos/go-oidc"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/pkg/errors"
)

// TokenData represents the authentication token
//go:generate mockery -name=TokenData -output=automock -outpkg=automock -case=underscore
type TokenData interface {
	// Claims reads the Claims from the token into the specified struct
	Claims(v interface{}) error
}

// TokenVerifier attempts to verify a token and returns it or an error if the verification was not successful
//go:generate mockery -name=TokenVerifier -output=automock -outpkg=automock -case=underscore
type TokenVerifier interface {
	// Verify verifies that the token is valid and returns a token if so, otherwise returns an error
	Verify(ctx context.Context, token string) (TokenData, error)
}

// TokenVerifierProvider defines different ways by which one can provide a TokenVerifier
type TokenVerifierProvider func(ctx context.Context, claims Claims) TokenVerifier

// Handler is the base struct definition of the AuthenticationMappingHandler
type Handler struct {
	reqDataParser         tenantmapping.ReqDataParser
	httpClient            *http.Client
	tokenVerifierProvider TokenVerifierProvider
	verifiers             map[string]TokenVerifier
	verifiersMutex        sync.RWMutex
}

// Claims contains basic Claims needed during request authenticcation
type Claims struct {
	Issuer  string `json:"issuer"`
	JWKSURL string `json:"jwks_uri"`
}

// oidcVerifier wraps the default goidc.IDTokenVerifier
type oidcVerifier struct {
	*goidc.IDTokenVerifier
}

// Verify implements security.TokenVerifier and delegates to oidc.IDTokenVerifier
func (v *oidcVerifier) Verify(ctx context.Context, idToken string) (TokenData, error) {
	return v.IDTokenVerifier.Verify(ctx, idToken)
}

// DefaultTokenVerifierProvider is the default TokenVerifierProvider which leverages goidc liberay
func DefaultTokenVerifierProvider(ctx context.Context, claims Claims) TokenVerifier {
	keySet := goidc.NewRemoteKeySet(ctx, claims.JWKSURL)
	verifier := &oidcVerifier{
		IDTokenVerifier: goidc.NewVerifier(claims.Issuer, keySet, &goidc.Config{SkipClientIDCheck: true}),
	}

	return verifier
}

// NewHandler constructs the AuthenticationMappingHandler
func NewHandler(reqDataParser tenantmapping.ReqDataParser, httpClient *http.Client, tokenVerifierProvider TokenVerifierProvider) *Handler {
	return &Handler{
		reqDataParser:         reqDataParser,
		httpClient:            httpClient,
		tokenVerifierProvider: tokenVerifierProvider,
		verifiers:             make(map[string]TokenVerifier),
		verifiersMutex:        sync.RWMutex{},
	}
}

func (h *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(writer, fmt.Sprintf("Bad request method. Got %s, expected POST", req.Method), http.StatusBadRequest)
		return
	}

	ctx := req.Context()

	reqData, err := h.reqDataParser.Parse(req)
	if err != nil {
		h.logError(ctx, err, "An error has occurred while parsing the request.")
		http.Error(writer, "Unable to parse request data", http.StatusBadRequest)
		return
	}

	claims, err := h.verifyToken(ctx, reqData)
	if err != nil {
		h.logError(ctx, err, "An error has occurred while processing the request.")
		http.Error(writer, "Token validation failed", http.StatusBadRequest)
		return
	}

	if err := claims.Claims(&reqData.Body.Extra); err != nil {
		h.logError(ctx, err, "An error has occurred while extracting claims from body.extra.")
		http.Error(writer, "Token claims extraction failed", http.StatusBadRequest)
		return
	}

	h.respond(ctx, writer, reqData.Body)
}

func (h *Handler) verifyToken(ctx context.Context, reqData oathkeeper.ReqData) (TokenData, error) {
	authorizationHeader := reqData.Header.Get("Authorization")
	if authorizationHeader == "" || !strings.HasPrefix(strings.ToLower(authorizationHeader), "bearer ") {
		return nil, errors.New(fmt.Sprintf("unexpected or empty authorization header with length %d", len(authorizationHeader)))
	}

	token := strings.TrimSpace(authorizationHeader[len("Bearer "):])

	issuerURL, err := extractTokenIssuer(token)
	if err != nil {
		return nil, fmt.Errorf("error while extracting token properties: %s", err)
	}

	if issuerURL == "" {
		return nil, errors.New("invalid token: missing issuer URL")
	}

	h.verifiersMutex.RLock()
	verifier, found := h.verifiers[issuerURL]
	h.verifiersMutex.RUnlock()

	if !found {
		log.C(ctx).Infof("Verifier for issuer %q not found. Attempting to construct new verifier from well-known endpoint", issuerURL)
		resp, err := h.getOpenIDConfig(ctx, issuerURL)
		if err != nil {
			return nil, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, handleResponseError(resp)
		}

		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, errors.Wrap(err, "failed read content from response")
		}

		var c Claims
		if err := json.Unmarshal(buf, &c); err != nil {
			return nil, fmt.Errorf("error decoding body of response with status %s: %s", resp.Status, err.Error())
		}

		if issuerURL != c.Issuer {
			return nil, errors.New(fmt.Sprintf("token issuer from token %q does not mismatch token issuer from well-known endpoint %q", issuerURL, c.Issuer))
		}

		verifier = h.tokenVerifierProvider(ctx, c)

		h.verifiersMutex.Lock()
		h.verifiers[issuerURL] = verifier
		h.verifiersMutex.Unlock()

		log.C(ctx).Infof("Successfully constructed verifier for issuer %q", issuerURL)
	} else {
		log.C(ctx).Infof("Verifier for issuer %q exists", issuerURL)
	}

	claims, err := verifier.Verify(ctx, token)
	if err != nil {
		return nil, errors.Wrapf(err, "unable to verify token")
	}

	return claims, nil
}

func (h *Handler) logError(ctx context.Context, err error, message string) {
	log.C(ctx).WithError(err).Error(message)
}

func (h *Handler) respond(ctx context.Context, writer http.ResponseWriter, body oathkeeper.ReqBody) {
	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(body)
	if err != nil {
		h.logError(ctx, err, "An error has occurred while encoding data.")
	}
}

func (h *Handler) getOpenIDConfig(ctx context.Context, issuerURL string) (*http.Response, error) {
	// Work around for UAA until https://github.com/cloudfoundry/uaa/issues/805 is fixed
	// Then goidc.NewProvider(ctx, options.IssuerURL) should be used
	if _, err := url.ParseRequestURI(issuerURL); err != nil {
		return nil, err
	}

	wellKnown := strings.TrimSuffix(strings.TrimSuffix(issuerURL, "/"), "/oauth/token") + "/.well-known/openid-configuration"
	req, err := http.NewRequest(http.MethodGet, wellKnown, nil)
	if err != nil {
		return nil, err
	}

	resp, err := h.httpClient.Do(req.WithContext(ctx))
	if err != nil {
		return nil, err
	}

	return resp, nil
}

func extractTokenIssuer(token string) (string, error) {
	// JWT format: <header>.<payload>.<signature>
	tokenParts := strings.Split(token, ".")
	if len(tokenParts) != 3 {
		return "", errors.New("invalid token format")
	}
	payload := tokenParts[1]

	decoded, err := base64.RawURLEncoding.DecodeString(payload)
	if err != nil {
		return "", err
	}

	data := &struct {
		IssuerURL string `json:"iss"`
	}{}
	if err = json.Unmarshal(decoded, data); err != nil {
		return "", err
	}

	return data.IssuerURL, nil
}

// handleResponseError builds an error from the given response
func handleResponseError(response *http.Response) error {
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.D().Errorf("ReadCloser couldn't be closed: %v", err)
		}
	}()

	body, err := ioutil.ReadAll(response.Body)
	if err != nil {
		body = []byte(fmt.Sprintf("error reading response body: %s", err))
	}

	err = fmt.Errorf("StatusCode: %d Body: %s", response.StatusCode, body)
	if response.Request != nil {
		return fmt.Errorf("request %s %s failed: %s", response.Request.Method, response.Request.URL, err)
	}
	return fmt.Errorf("request failed: %s", err)
}
