package authnmappinghandler_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"

	"github.com/kyma-incubator/compass/components/hydrator/internal/authnmappinghandler"
	"github.com/kyma-incubator/compass/components/hydrator/internal/authnmappinghandler/automock"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"github.com/gorilla/mux"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/form3tech-oss/jwt-go"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const wellKnownRespPattern = `{"issuer": %q, "jwks_uri": %q}`

type RoundTripFunc func(req *http.Request) *http.Response

func (f RoundTripFunc) RoundTrip(req *http.Request) (*http.Response, error) {
	return f(req), nil
}

func TestHandler(t *testing.T) {
	target := "http://example.com/authn-mapping/authenticatorName"
	domainURL := "localhost:8080"
	issuer := "http://tenant." + domainURL + "/oauth/token"
	jwksURL := "http://tenant." + domainURL + "/keys"
	uniqueAttributeKey := "unique_key"
	uniqueAttributeValue := "testUniqueValue"
	authenticatorName := "authenticatorName"
	mockedAttributes := map[string]interface{}{uniqueAttributeKey: uniqueAttributeValue}
	authenticators := []authenticator.Config{
		{
			Name: authenticatorName,
			Attributes: authenticator.Attributes{
				UniqueAttribute: authenticator.Attribute{
					Key:   "custom_attributes." + uniqueAttributeKey,
					Value: uniqueAttributeValue,
				},
			},
			TrustedIssuers: []authenticator.TrustedIssuer{
				{DomainURL: domainURL},
			},
		},
	}

	wellKnownResp := fmt.Sprintf(wellKnownRespPattern, issuer, jwksURL)

	mockedWellKnownConfigClient := &http.Client{
		Transport: RoundTripFunc(func(req *http.Request) *http.Response {
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       ioutil.NopCloser(bytes.NewBufferString(wellKnownResp)),
			}
		}),
	}

	mockedToken := generateToken(t, issuer, mockedAttributes)

	t.Run("success when new verifier succeeds in verifying token", func(t *testing.T) {
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					"tenant": "test-tenant",
				},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + mockedToken},
			},
		}
		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))

		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})

		w := httptest.NewRecorder()

		tokenDataMock := &automock.TokenData{}
		tokenDataMock.On("Claims", &reqDataMock.Body.Extra).Return(nil).Once()

		verifierMock := &automock.TokenVerifier{}
		verifierMock.On("Verify", mock.Anything, mockedToken).Return(tokenDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, mockedWellKnownConfigClient, func(_ context.Context, _ authnmappinghandler.OpenIDMetadata) authnmappinghandler.TokenVerifier {
			return verifierMock
		}, authenticators)
		handler.ServeHTTP(w, reqWithVars)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		expectedPayload, err := json.Marshal(reqDataMock.Body)
		require.NoError(t, err)

		require.Contains(t, strings.TrimSpace(string(body)), string(expectedPayload))
		mock.AssertExpectationsForObjects(t, reqDataParserMock, verifierMock, tokenDataMock)
	})

	t.Run("success when cached verifier succeeds in verifying token", func(t *testing.T) {
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					"tenant": "test-tenant",
				},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + mockedToken},
			},
		}
		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Twice()

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})
		w := httptest.NewRecorder()

		tokenDataMock := &automock.TokenData{}
		tokenDataMock.On("Claims", &reqDataMock.Body.Extra).Return(nil).Twice()

		verifierMock := &automock.TokenVerifier{}
		verifierMock.On("Verify", mock.Anything, mockedToken).Return(tokenDataMock, nil).Twice()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, mockedWellKnownConfigClient, func(_ context.Context, _ authnmappinghandler.OpenIDMetadata) authnmappinghandler.TokenVerifier {
			return verifierMock
		}, authenticators)
		handler.ServeHTTP(w, reqWithVars)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		expectedPayload, err := json.Marshal(reqDataMock.Body)
		require.NoError(t, err)

		require.Contains(t, strings.TrimSpace(string(body)), string(expectedPayload))

		req2 := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		reqWithVars2 := mux.SetURLVars(req2, map[string]string{
			"authenticator": authenticatorName,
		})
		w2 := httptest.NewRecorder()

		handler.ServeHTTP(w2, reqWithVars2)

		resp2 := w2.Result()
		require.Equal(t, http.StatusOK, resp2.StatusCode)
		body2, err := ioutil.ReadAll(resp2.Body)
		require.NoError(t, err)

		require.Contains(t, strings.TrimSpace(string(body2)), string(expectedPayload))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, verifierMock, tokenDataMock)
	})

	t.Run("error when new verifier fails to verify token", func(t *testing.T) {
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					"tenant": "test-tenant",
				},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + mockedToken},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})
		w := httptest.NewRecorder()

		mockErr := errors.New("some-error")

		verifierMock := &automock.TokenVerifier{}
		verifierMock.On("Verify", mock.Anything, mockedToken).Return(nil, mockErr).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, mockedWellKnownConfigClient, func(_ context.Context, _ authnmappinghandler.OpenIDMetadata) authnmappinghandler.TokenVerifier {
			return verifierMock
		}, authenticators)
		handler.ServeHTTP(w, reqWithVars)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		require.Contains(t, logsBuffer.String(), mockErr.Error())

		mock.AssertExpectationsForObjects(t, reqDataParserMock, verifierMock)
	})

	t.Run("error when cached verifier fails to verify token", func(t *testing.T) {
		mockedToken1 := generateToken(t, issuer, mockedAttributes)
		mockedToken2 := generateToken(t, issuer, mockedAttributes)

		reqDataMock1 := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					"tenant": "test-tenant",
				},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + mockedToken1},
			},
		}

		reqDataMock2 := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					"tenant": "test-tenant",
				},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + mockedToken2},
			},
		}

		req1 := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		reqWithVars1 := mux.SetURLVars(req1, map[string]string{
			"authenticator": authenticatorName,
		})
		w1 := httptest.NewRecorder()

		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req2 := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req2 = req2.WithContext(ctx)
		reqWithVars2 := mux.SetURLVars(req2, map[string]string{
			"authenticator": authenticatorName,
		})
		w2 := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", reqWithVars1).Return(reqDataMock1, nil).Once()
		reqDataParserMock.On("Parse", reqWithVars2).Return(reqDataMock2, nil).Once()

		tokenDataMock := &automock.TokenData{}
		tokenDataMock.On("Claims", &reqDataMock1.Body.Extra).Return(nil).Once()

		mockErr := errors.New("some-error")

		verifierMock := &automock.TokenVerifier{}
		verifierMock.On("Verify", mock.Anything, mockedToken1).Return(tokenDataMock, nil).Once()
		verifierMock.On("Verify", mock.Anything, mockedToken2).Return(nil, mockErr).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, mockedWellKnownConfigClient, func(_ context.Context, _ authnmappinghandler.OpenIDMetadata) authnmappinghandler.TokenVerifier {
			return verifierMock
		}, authenticators)
		handler.ServeHTTP(w1, reqWithVars1)

		resp1 := w1.Result()
		require.Equal(t, http.StatusOK, resp1.StatusCode)
		body1, err := ioutil.ReadAll(resp1.Body)
		require.NoError(t, err)

		expectedPayload, err := json.Marshal(reqDataMock1.Body)
		require.NoError(t, err)

		require.Contains(t, strings.TrimSpace(string(body1)), string(expectedPayload))

		handler.ServeHTTP(w2, reqWithVars2)

		resp2 := w2.Result()
		require.Equal(t, http.StatusOK, resp2.StatusCode)
		require.Contains(t, logsBuffer.String(), mockErr.Error())

		mock.AssertExpectationsForObjects(t, reqDataParserMock, verifierMock, tokenDataMock)
	})

	t.Run("error when well-known configuration responds with different than 200 OK status code", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})
		w := httptest.NewRecorder()

		mockedWellKnownClient := &http.Client{
			Transport: RoundTripFunc(func(req *http.Request) *http.Response {
				return &http.Response{
					StatusCode: http.StatusInternalServerError,
					Body:       ioutil.NopCloser(bytes.NewBufferString("Server error")),
				}
			}),
		}

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + mockedToken},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, mockedWellKnownClient, nil, authenticators)
		handler.ServeHTTP(w, reqWithVars)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "request failed: StatusCode: 500 Body: Server error")
	})

	t.Run("error when authenticator unique attribute missmatch", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})
		w := httptest.NewRecorder()

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + generateToken(t, issuer, map[string]interface{}{uniqueAttributeKey: "unexpected-unique-attribute-value"})},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil, nil, authenticators)
		handler.ServeHTTP(w, reqWithVars)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "unique attribute mismatch")
	})

	t.Run("error when JWT token contains issuer url which is not valid url", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})
		w := httptest.NewRecorder()

		signedToken := generateToken(t, "abc", mockedAttributes)

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

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil, nil, authenticators)
		handler.ServeHTTP(w, reqWithVars)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "error while extracting token issuer subdomain")
	})

	t.Run("error when token JWT token doesn't contain issuer url", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})
		w := httptest.NewRecorder()

		signedToken := generateToken(t, "", mockedAttributes)

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

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil, nil, authenticators)
		handler.ServeHTTP(w, reqWithVars)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "error while extracting token issuer subdomain")
	})

	t.Run("error when token in authorization header isn't valid JWT token in terms of encoding", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})
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

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil, nil, nil)
		handler.ServeHTTP(w, reqWithVars)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "illegal base64 data")
	})

	t.Run("error when token in authorization header isn't valid JWT token", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)

		ctx := log.ContextWithLogger(context.Background(), entry)
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})
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

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil, nil, nil)
		handler.ServeHTTP(w, reqWithVars)

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
		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})
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

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil, nil, nil)
		handler.ServeHTTP(w, reqWithVars)

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
		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})
		w := httptest.NewRecorder()

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil, nil, nil)
		handler.ServeHTTP(w, reqWithVars)

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

		handler := authnmappinghandler.NewHandler(reqDataParserMock, nil, nil, nil)
		req = mux.SetURLVars(req, map[string]string{"authenticator": "auth"})
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "An error has occurred while parsing the request")
	})

	t.Run("error when request method is not POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		handler := authnmappinghandler.NewHandler(nil, nil, nil, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
	})

	t.Run("error when extracting token claims", func(t *testing.T) {
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					"consumerTenant": "test-tenant",
				},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + mockedToken},
			},
		}
		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})
		w := httptest.NewRecorder()
		tokenDataMock := &automock.TokenData{}
		tokenDataMock.On("Claims", &reqDataMock.Body.Extra).Return(errors.New("expected")).Once()

		verifierMock := &automock.TokenVerifier{}
		verifierMock.On("Verify", mock.Anything, mockedToken).Return(tokenDataMock, nil).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, mockedWellKnownConfigClient, func(_ context.Context, _ authnmappinghandler.OpenIDMetadata) authnmappinghandler.TokenVerifier {
			return verifierMock
		}, authenticators)

		handler.ServeHTTP(w, reqWithVars)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		expectedResponse := "Token claims extraction failed"
		require.NoError(t, err)

		require.Contains(t, strings.TrimSpace(string(body)), expectedResponse)
		mock.AssertExpectationsForObjects(t, reqDataParserMock, verifierMock, tokenDataMock)
	})

	t.Run("error when authenticator is found but it does not contain matching trusted issuer", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)
		ctx := log.ContextWithLogger(context.Background(), entry)

		unknownIssuer := "http://tenant." + "unknown.com" + "/oauth/token"
		invalidToken := generateToken(t, unknownIssuer, mockedAttributes)

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					"tenant": "test-tenant",
				},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + invalidToken},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		req = req.WithContext(ctx)
		reqWithVars := mux.SetURLVars(req, map[string]string{
			"authenticator": authenticatorName,
		})
		w := httptest.NewRecorder()

		mockErr := errors.New("invalid token")

		verifierMock := &automock.TokenVerifier{}
		verifierMock.On("Verify", mock.Anything, invalidToken).Return(nil, mockErr).Once()

		handler := authnmappinghandler.NewHandler(reqDataParserMock, mockedWellKnownConfigClient, func(_ context.Context, _ authnmappinghandler.OpenIDMetadata) authnmappinghandler.TokenVerifier {
			return verifierMock
		}, authenticators)
		handler.ServeHTTP(w, reqWithVars)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), mockErr.Error())
	})

	t.Run("error when no matching authenticator is found in path", func(t *testing.T) {
		logsBuffer := &bytes.Buffer{}
		entry := log.DefaultLogger()
		entry.Logger.SetOutput(logsBuffer)
		ctx := log.ContextWithLogger(context.Background(), entry)

		unknownIssuer := "http://tenant." + "unknown.com" + "/oauth/token"
		invalidToken := generateToken(t, unknownIssuer, mockedAttributes)

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					"tenant": "test-tenant",
				},
			},
			Header: map[string][]string{
				"Authorization": {"Bearer " + invalidToken},
			},
		}

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		req := httptest.NewRequest(http.MethodPost, "http://example.com/authn-mapping/", strings.NewReader(""))
		req = req.WithContext(ctx)
		w := httptest.NewRecorder()

		verifierMock := &automock.TokenVerifier{}
		handler := authnmappinghandler.NewHandler(reqDataParserMock, mockedWellKnownConfigClient, func(_ context.Context, _ authnmappinghandler.OpenIDMetadata) authnmappinghandler.TokenVerifier {
			return verifierMock
		}, authenticators)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		require.Contains(t, logsBuffer.String(), "authenticator not found in path")
	})
}

func generateToken(t *testing.T, issuer string, attributes map[string]interface{}) string {
	claims := MockedClaims{
		CustomAttributes: attributes,
	}
	if issuer != "" {
		claims.Issuer = issuer
	}

	token := jwt.NewWithClaims(jwt.SigningMethodNone, claims)

	signedToken, err := token.SignedString(jwt.UnsafeAllowNoneSignatureType)
	require.NoError(t, err)

	return signedToken
}

type MockedClaims struct {
	jwt.StandardClaims
	CustomAttributes map[string]interface{} `json:"custom_attributes"`
}
