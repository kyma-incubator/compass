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

package authnmappinghandler_test

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	goidc "github.com/coreos/go-oidc"
	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/internal/authnmappinghandler"
	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const wellKnownRespPattern = `{"issuer": %q, "jwks_uri": %q}`

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestHandler(t *testing.T) {
	target := "http://example.com/foo"
	issuer := "http://tenant.localhost:8080/oauth/token"
	jwksURL := "http://tenant.localhost:8080/keys"

	/*
		t.Run("success when cached verifier succeeds in verifying token", func(t *testing.T) {
			wellKnownResp := fmt.Sprintf(wellKnownRespPattern, issuer, jwksURL)

			httpClient := &http.Client{
				Transport: RoundTripFunc(func(req *http.Request) *http.Response {
					return &http.Response{
						StatusCode: http.StatusOK,
						Body:       ioutil.NopCloser(bytes.NewBufferString(wellKnownResp)),
					}
				}),
			}

			token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.StandardClaims{
				Issuer: issuer,
			})

			signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
			require.NoError(t, err)

			reqDataMock := oathkeeper.ReqData{
				Body: oathkeeper.ReqBody{
					Extra: map[string]interface{}{},
				},
				Header: map[string][]string{
					"Authorization": {"Bearer " + signedToken},
				},
			}

			reqDataParserMock := &automock.ReqDataParser{}
			reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

			req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
			w := httptest.NewRecorder()

			handler := authnmappinghandler.NewHandler(reqDataParserMock, httpClient)
			handler.ServeHTTP(w, req)

			resp := w.Result()
			require.Equal(t, http.StatusOK, resp.StatusCode)
			body, err := ioutil.ReadAll(resp.Body)
			require.NoError(t, err)
			fmt.Println(body)

			//require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

			//mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, userMockContextProvider)
		})

		t.Run("success when new verifier succeeds in verifying token", func(t *testing.T) {})

		t.Run("error when cached verifier fails to verify token", func(t *testing.T) {})

		t.Run("error when new verifier fails to verify token", func(t *testing.T) {

		})
	*/

	t.Run("error when well-known configuration issuer mismatches the issue value in the token", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		tokenIssuer := issuer
		wellKnownIssuer := "abc"

		wellKnownResp := fmt.Sprintf(wellKnownRespPattern, wellKnownIssuer, jwksURL)

		httpClient := &http.Client{
			Transport: RoundTripFunc(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusOK,
					Body:       ioutil.NopCloser(bytes.NewBufferString(wellKnownResp)),
				}
			}),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.StandardClaims{
			Issuer: tokenIssuer,
		})

		signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + signedToken},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, httpClient)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "does not mismatch token issuer from well-known endpoint")
	})

	t.Run("error when well-known configuration responds with different than 200 OK status code", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		httpClient := &http.Client{
			Transport: RoundTripFunc(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(bytes.NewBufferString("Server error")),
				}
			}),
		}

		token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.StandardClaims{
			Issuer: issuer,
		})

		signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + signedToken},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, httpClient)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "request failed: StatusCode: 500 Body: Server error")
	})

	t.Run("error when token JWT token contains issuer url which is not valid url", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.StandardClaims{
			Issuer: "abc",
		})

		signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + signedToken},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "invalid URI for request")
	})

	t.Run("error when token JWT token doesn't contain issuer url", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		token := jwt.NewWithClaims(jwt.SigningMethodNone, jwt.StandardClaims{})

		signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
		require.NoError(t, err)

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + signedToken},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "invalid token: missing issuer URL")
	})

	t.Run("error when token in authorization header isn't valid JWT token in terms of encoding", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer a.b.c"},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "error while extracting token properties: illegal base64 data")
	})

	t.Run("error when token in authorization header isn't valid JWT token", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer abc"},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "invalid token format")
	})

	t.Run("error when authorization header doesn't start with Bearer", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
			Header: map[string][]string{
				"Authorization": {"abc"},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "unexpected or empty authorization header with length")
	})

	t.Run("error when authorization header is empty", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "unexpected or empty authorization header with length")
	})

	t.Run("error when fails to parse request data", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		mockErr := errors.New("some-error")

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(oathkeeper.ReqData{}, mockErr).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "An error has occurred while parsing the request")
	})

	t.Run("error when request method is not POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		handler := authnmappinghandler.NewHandler(nil, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
	})
}

func mockTokenVerifier(ctx context.Context, token string, _ func(ctx context.Context, token string) (*goidc.IDToken, error)) (*goidc.IDToken, error) {
	claims := &jwt.StandardClaims{}
	_, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return jwt.UnsafeAllowNoneSignatureType, nil
	})
	if err != nil {
		return nil, err
	}

	idTokenClaims := &goidc.IDToken{
		Issuer:   claims.Issuer,
		Audience: claims.Audience,
		Subject:  claims.Subject,
		Expiry:   time.Unix(claims.ExpiresAt, 0),
		IssuedAt: time.Unix(claims.IssuedAt, 0),
	}

	// TODO: unmarshal claims into idTokenClaims.claims with reflection

	return idTokenClaims, nil
}
