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

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"

	"golang.org/x/oauth2"

	"github.com/gorilla/mux"

	"github.com/kyma-incubator/compass/components/director/pkg/httputils"

	"github.com/tidwall/gjson"

	goidc "github.com/coreos/go-oidc/v3/oidc"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
	"github.com/pkg/errors"
)

// TokenData represents the authentication token
//go:generate mockery --name=TokenData --output=automock --outpkg=automock --case=underscore --disable-version-string
type TokenData interface {
	// Claims reads the Claims from the token into the specified struct
	Claims(v interface{}) error
}

// TokenVerifier attempts to verify a token and returns it or an error if the verification was not successful
//go:generate mockery --name=TokenVerifier --output=automock --outpkg=automock --case=underscore --disable-version-string
type TokenVerifier interface {
	// Verify verifies that the token is valid and returns a token if so, otherwise returns an error
	Verify(ctx context.Context, token string) (TokenData, error)
}

// ReqDataParser parses request data
//go:generate mockery --name=ReqDataParser --output=automock --outpkg=automock --case=underscore --disable-version-string
type ReqDataParser interface {
	Parse(req *http.Request) (oathkeeper.ReqData, error)
}

// TokenVerifierProvider defines different ways by which one can provide a TokenVerifier
type TokenVerifierProvider func(ctx context.Context, metadata OpenIDMetadata) TokenVerifier

// Handler is the base struct definition of the AuthenticationMappingHandler
type Handler struct {
	reqDataParser         ReqDataParser
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
func NewHandler(reqDataParser ReqDataParser, httpClient *http.Client, tokenVerifierProvider TokenVerifierProvider, authenticators []authenticator.Config) *Handler {
	return &Handler{
		reqDataParser:         reqDataParser,
		httpClient:            httpClient,
		tokenVerifierProvider: tokenVerifierProvider,
		verifiers:             make(map[string]TokenVerifier),
		verifiersMutex:        sync.RWMutex{},
		authenticators:        authenticators,
	}
}

type authenticationError struct {
	Message string `json:"message"`
}

// ServeHTTP missing godoc
func (h *Handler) ServeHTTP(writer http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		http.Error(writer, fmt.Sprintf("Bad request method. Got %s, expected POST", req.Method), http.StatusOK)
		return
	}

	ctx := context.WithValue(req.Context(), oauth2.HTTPClient, h.httpClient)

	reqData, err := h.reqDataParser.Parse(req)
	if err != nil {
		h.logError(ctx, err, "An error has occurred while parsing the request")
		http.Error(writer, "Unable to parse request data", http.StatusOK)
		return
	}

	vars := mux.Vars(req)
	matchedAuthenticator, ok := vars["authenticator"]
	if !ok {
		h.logError(ctx, errors.New("authenticator not found in path"), "An error has occurred while extracting authenticator name")
		reqData.Body.Extra["error"] = authenticationError{Message: "Missing authenticator"}
		h.respond(ctx, writer, reqData.Body)
		return
	}

	log.C(ctx).Infof("Matched authenticator is %s", matchedAuthenticator)

	claims, authCoordinates, err := h.verifyToken(ctx, reqData, matchedAuthenticator)
	if err != nil {
		h.logError(ctx, err, "An error has occurred while processing the request")
		reqData.Body.Extra["error"] = authenticationError{Message: "Token validation failed"}
		h.respond(ctx, writer, reqData.Body)
		return
	}

	if err := claims.Claims(&reqData.Body.Extra); err != nil {
		h.logError(ctx, err, "An error has occurred while extracting claims to request body.extra")
		reqData.Body.Extra["error"] = authenticationError{Message: "Token claims extraction failed"}
		h.respond(ctx, writer, reqData.Body)
		return
	}
	reqData.Body.Extra[authenticator.CoordinatesKey] = authCoordinates

	h.respond(ctx, writer, reqData.Body)
}

func (h *Handler) verifyToken(ctx context.Context, reqData oathkeeper.ReqData, authenticatorName string) (TokenData, authenticator.Coordinates, error) {
	authorizationHeader := reqData.Header.Get("Authorization")
	if authorizationHeader == "" || !strings.HasPrefix(strings.ToLower(authorizationHeader), "bearer ") {
		return nil, authenticator.Coordinates{}, errors.Errorf("unexpected or empty authorization header with length %d", len(authorizationHeader))
	}

	token := strings.TrimSpace(authorizationHeader[len("Bearer "):])

	tokenPayload, err := getTokenPayload(token)
	if err != nil {
		return nil, authenticator.Coordinates{}, errors.Wrapf(err, "while getting token payload")
	}

	issuerSubdomain, err := extractTokenIssuerSubdomain(tokenPayload)
	if err != nil {
		return nil, authenticator.Coordinates{}, errors.Wrap(err, "error while extracting token issuer subdomain")
	}

	config, err := h.getAuthenticatorConfig(authenticatorName, tokenPayload)
	if err != nil {
		return nil, authenticator.Coordinates{}, errors.Wrapf(err, "while getting matched authenticator config")
	}

	index := -1
	var claims TokenData
	aggregatedErr := errors.New("aggregated error for all issuers")

	for i, issuer := range config.TrustedIssuers {
		protocol := "https"
		if len(issuer.Protocol) > 0 {
			protocol = issuer.Protocol
		}
		issuerURL := fmt.Sprintf("%s://%s.%s%s", protocol, issuerSubdomain, issuer.DomainURL, "/oauth/token")

		h.verifiersMutex.RLock()
		verifier, found := h.verifiers[issuerURL]
		h.verifiersMutex.RUnlock()

		if !found {
			log.C(ctx).Infof("Verifier for issuer %q not found. Attempting to construct new verifier from well-known endpoint", issuerURL)
			resp, err := h.getOpenIDConfig(ctx, issuerURL)
			if err != nil {
				aggregatedErr = errors.Wrapf(aggregatedErr, "error while getting OpenIDCOnfig for issuer %q: %s", issuerURL, err)
				continue
			}

			if resp.StatusCode != http.StatusOK {
				aggregatedErr = errors.Wrapf(aggregatedErr, "error for issuer %q: %s", issuerURL, handleResponseError(ctx, resp))
				continue
			}

			var m OpenIDMetadata
			if err := json.NewDecoder(resp.Body).Decode(&m); err != nil {
				aggregatedErr = errors.Wrapf(aggregatedErr, "while decoding body of response for issuer %q: %s", issuerURL, err)
				continue
			}

			defer httputils.Close(ctx, resp.Body)

			verifier = h.tokenVerifierProvider(ctx, m)

			h.verifiersMutex.Lock()
			h.verifiers[issuerURL] = verifier
			h.verifiersMutex.Unlock()

			log.C(ctx).Infof("Successfully constructed verifier for issuer %q", issuerURL)
		} else {
			log.C(ctx).Infof("Verifier for issuer %q exists", issuerURL)
		}

		claims, err = verifier.Verify(ctx, token)
		if err != nil {
			aggregatedErr = errors.Wrapf(aggregatedErr, "unable to verify token with issuer %q: %s", issuerURL, err)
			continue
		}
		index = i
		break
	}

	if index == -1 {
		return nil, authenticator.Coordinates{}, aggregatedErr
	}

	if config.CheckSuffix {
		c := make(map[string]interface{})
		if err = claims.Claims(&c); err != nil {
			return nil, authenticator.Coordinates{}, err
		}
		for _, suffix := range config.ClientIDSuffixes {
			if strings.HasSuffix(c[config.Attributes.ClientID.Key].(string), suffix) {
				return claims, authenticator.Coordinates{
					Name:  config.Name,
					Index: index,
				}, nil
			}
		}
		return nil, authenticator.Coordinates{}, errors.Wrapf(aggregatedErr, "client suffix mismatch")
	}
	return claims, authenticator.Coordinates{
		Name:  config.Name,
		Index: index,
	}, nil
}

func (h *Handler) logError(ctx context.Context, err error, message string) {
	log.C(ctx).WithError(err).Error(message)
}

func (h *Handler) respond(ctx context.Context, writer http.ResponseWriter, body oathkeeper.ReqBody) {
	writer.Header().Set("Content-Type", "application/json")
	err := json.NewEncoder(writer).Encode(body)
	if err != nil {
		h.logError(ctx, err, "An error has occurred while encoding data")
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

func (h *Handler) getAuthenticatorConfig(matchedAuthenticatorName string, payload []byte) (*authenticator.Config, error) {
	var authConfig *authenticator.Config
	for i, authn := range h.authenticators {
		if authn.Name == matchedAuthenticatorName {
			uniqueAttribute := gjson.GetBytes(payload, authn.Attributes.UniqueAttribute.Key).String()
			if uniqueAttribute == "" || uniqueAttribute != authn.Attributes.UniqueAttribute.Value {
				return nil, errors.New("unique attribute mismatch")
			}
			authConfig = &h.authenticators[i]
			break
		}
	}
	if authConfig == nil {
		return nil, errors.New("could not find authenticator for token")
	}

	return authConfig, nil
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

func extractTokenIssuerSubdomain(payload []byte) (string, error) {
	data := &struct {
		IssuerURL string `json:"iss"`
	}{}
	if err := json.Unmarshal(payload, data); err != nil {
		return "", err
	}

	parsedIssuer, err := url.Parse(data.IssuerURL)
	if err != nil {
		return "", err
	}

	s := strings.Split(parsedIssuer.Hostname(), ".")
	if len(s) < 2 || s[0] == "" {
		return "", fmt.Errorf("could not extract subdomain from issuer URL %s", data.IssuerURL)
	}

	return s[0], nil
}

// handleResponseError builds an error from the given response
func handleResponseError(ctx context.Context, response *http.Response) error {
	defer func() {
		if err := response.Body.Close(); err != nil {
			log.C(ctx).Errorf("ReadCloser couldn't be closed: %v", err)
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
