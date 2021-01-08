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

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"

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
type TokenVerifierProvider func(ctx context.Context, metadata OpenIDMetadata) TokenVerifier

// Handler is the base struct definition of the AuthenticationMappingHandler
type Handler struct {
	reqDataParser         tenantmapping.ReqDataParser
	httpClient            *http.Client
	tokenVerifierProvider TokenVerifierProvider
	verifiers             map[string]TokenVerifier
	verifiersMutex        sync.RWMutex
	authenticators        []authenticator.Config
}

// OpenIDMetadata contains basic metadata for OIDC provider needed during request authentication
type OpenIDMetadata struct {
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
func DefaultTokenVerifierProvider(ctx context.Context, metadata OpenIDMetadata) TokenVerifier {
	keySet := goidc.NewRemoteKeySet(ctx, metadata.JWKSURL)
	verifier := &oidcVerifier{
		IDTokenVerifier: goidc.NewVerifier(metadata.Issuer, keySet, &goidc.Config{SkipClientIDCheck: true}),
	}

	return verifier
}

// NewHandler constructs the AuthenticationMappingHandler
func NewHandler(reqDataParser tenantmapping.ReqDataParser, httpClient *http.Client, tokenVerifierProvider TokenVerifierProvider, authenticators []authenticator.Config) *Handler {
	return &Handler{
		reqDataParser:         reqDataParser,
		httpClient:            httpClient,
		tokenVerifierProvider: tokenVerifierProvider,
		verifiers:             make(map[string]TokenVerifier),
		verifiersMutex:        sync.RWMutex{},
		authenticators:        authenticators,
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
		http.Error(writer, "Unable to parse request data", http.StatusUnauthorized)
		return
	}

	claims, authCoordinates, err := h.verifyToken(ctx, reqData)
	if err != nil {
		h.logError(ctx, err, "An error has occurred while processing the request.")
		http.Error(writer, "Token validation failed", http.StatusUnauthorized)
		return
	}

	if err := claims.Claims(&reqData.Body.Extra); err != nil {
		h.logError(ctx, err, "An error has occurred while extracting claims to request body.extra")
		http.Error(writer, "Token claims extraction failed", http.StatusUnauthorized)
		return
	}
	reqData.Body.Extra[authenticator.CoordinatesKey] = authCoordinates

	h.respond(ctx, writer, reqData.Body)
}

func (h *Handler) verifyToken(ctx context.Context, reqData oathkeeper.ReqData) (TokenData, authenticator.Coordinates, error) {
	authorizationHeader := reqData.Header.Get("Authorization")
	if authorizationHeader == "" || !strings.HasPrefix(strings.ToLower(authorizationHeader), "bearer ") {
		return nil, authenticator.Coordinates{}, errors.New(fmt.Sprintf("unexpected or empty authorization header with length %d", len(authorizationHeader)))
	}

	token := strings.TrimSpace(authorizationHeader[len("Bearer "):])

	tokenPayload, err := getTokenPayload(token)
	if err != nil {
		return nil, authenticator.Coordinates{}, errors.Wrapf(err, "while getting token payload")
	}

	issuerURL, err := extractTokenIssuer(tokenPayload)
	if err != nil {
		return nil, authenticator.Coordinates{}, errors.Errorf("error while extracting token properties: %s", err)
	}

	if issuerURL == "" {
		return nil, authenticator.Coordinates{}, errors.New("invalid token: missing issuer URL")
	}

	coordinates, err := h.getAuthenticatorCoordinates(tokenPayload, issuerURL)
	if err != nil {
		return nil, authenticator.Coordinates{}, errors.Wrapf(err, "while getting authenticator coordinates")
	}

	h.verifiersMutex.RLock()
	verifier, found := h.verifiers[issuerURL]
	h.verifiersMutex.RUnlock()

	if !found {
		log.C(ctx).Infof("Verifier for issuer %q not found. Attempting to construct new verifier from well-known endpoint", issuerURL)
		resp, err := h.getOpenIDConfig(ctx, issuerURL)
		if err != nil {
			return nil, authenticator.Coordinates{}, err
		}

		if resp.StatusCode != http.StatusOK {
			return nil, authenticator.Coordinates{}, handleResponseError(resp)
		}

		buf, err := ioutil.ReadAll(resp.Body)
		if err != nil {
			return nil, authenticator.Coordinates{}, errors.Wrap(err, "failed read content from response")
		}

		var m OpenIDMetadata
		if err := json.Unmarshal(buf, &m); err != nil {
			return nil, authenticator.Coordinates{}, fmt.Errorf("while decoding body of response with status %s: %s", resp.Status, err.Error())
		}

		if issuerURL != m.Issuer {
			return nil, authenticator.Coordinates{}, errors.New(fmt.Sprintf("token issuer from token %q does not mismatch token issuer from well-known endpoint %q", issuerURL, m.Issuer))
		}

		verifier = h.tokenVerifierProvider(ctx, m)

		h.verifiersMutex.Lock()
		h.verifiers[issuerURL] = verifier
		h.verifiersMutex.Unlock()

		log.C(ctx).Infof("Successfully constructed verifier for issuer %q", issuerURL)
	} else {
		log.C(ctx).Infof("Verifier for issuer %q exists", issuerURL)
	}

	claims, err := verifier.Verify(ctx, token)
	if err != nil {
		return nil, authenticator.Coordinates{}, errors.Wrapf(err, "unable to verify token")
	}

	return claims, coordinates, nil
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

func (h *Handler) getAuthenticatorCoordinates(payload []byte, issuerURL string) (authenticator.Coordinates, error) {
	var authConfig *authenticator.Config
	for i, authn := range h.authenticators {
		uniqueAttribute := gjson.GetBytes(payload, authn.Attributes.UniqueAttribute.Key).String()
		if uniqueAttribute != "" || uniqueAttribute == authn.Attributes.UniqueAttribute.Value {
			authConfig = &h.authenticators[i]
			break
		}
	}
	if authConfig == nil {
		return authenticator.Coordinates{}, errors.New("could not find authenticator for token")
	}

	i, trusted := getTrustedIssuerIndex(authConfig, issuerURL)
	if !trusted {
		return authenticator.Coordinates{}, errors.New("could not find trusted issuer in given authenticator")
	}
	return authenticator.Coordinates{
		Name:  authConfig.Name,
		Index: i,
	}, nil
}

func getTokenPayload(token string) ([]byte, error) {
	// JWT format: <header>.<payload>.<signature>
	tokenParts := strings.Split(token, ".")
	if len(tokenParts) != 3 {
		return nil, errors.New("invalid token format")
	}
	payload := tokenParts[1]

	return base64.RawURLEncoding.DecodeString(payload)
}

func extractTokenIssuer(payload []byte) (string, error) {
	data := &struct {
		IssuerURL string `json:"iss"`
	}{}
	if err := json.Unmarshal(payload, data); err != nil {
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

func getTrustedIssuerIndex(authConfig *authenticator.Config, issuerURL string) (index int, found bool) {
	if !strings.Contains(issuerURL, ".") || !strings.Contains(issuerURL, "/") {
		return -1, false
	}
	stripedIssuerURL := issuerURL[strings.Index(issuerURL, ".")+1:] //strip the period as well
	stripedIssuerURL = stripedIssuerURL[:strings.Index(stripedIssuerURL, "/")]

	for i, iss := range authConfig.TrustedIssuers {
		if iss.DomainURL == stripedIssuerURL {
			index, found = i, true
			return
		}
	}
	return -1, false
}
