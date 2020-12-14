package oauth20_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestService_CreateClient(t *testing.T) {
	// given
	publicEndpoint := "accessTokenURL"
	id := "foo"
	objType := model.IntegrationSystemReference

	scopes := []string{"foo", "bar", "baz"}
	successResult := &model.OAuthCredentialDataInput{
		ClientID:     "foo",
		ClientSecret: "c-secret",
		URL:          publicEndpoint,
	}
	expectedReqBody := map[string]interface{}{
		"grant_types": []interface{}{"client_credentials"},
		"client_id":   "foo",
		"scope":       "foo bar baz",
	}
	testErr := errors.New("test err")

	testCases := []struct {
		Name               string
		ExpectedResult     *model.OAuthCredentialDataInput
		ExpectedError      error
		ScopeCfgProviderFn func() *automock.ScopeCfgProvider
		UIDServiceFn       func() *automock.UIDService
		HTTPServerFn       func(t *testing.T) *httptest.Server
		Config             oauth20.Config
		Request            *http.Request
		Response           *http.Response
	}{
		{
			Name:           "Success",
			ExpectedError:  nil,
			ExpectedResult: successResult,
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(id).Once()
				return uidSvc
			},
			ScopeCfgProviderFn: func() *automock.ScopeCfgProvider {
				scopeCfgProvider := &automock.ScopeCfgProvider{}
				scopeCfgProvider.On("GetRequiredScopes", "clientCredentialsRegistrationScopes.integration_system").Return(scopes, nil).Once()
				return scopeCfgProvider
			},
			HTTPServerFn: fixSuccessCreateClientHTTPServer(expectedReqBody),
		},
		{
			Name:          "Error - Response Status Code",
			ExpectedError: errors.New("invalid HTTP status code: received: 500, expected 201"),
			ScopeCfgProviderFn: func() *automock.ScopeCfgProvider {
				scopeCfgProvider := &automock.ScopeCfgProvider{}
				scopeCfgProvider.On("GetRequiredScopes", "clientCredentialsRegistrationScopes.integration_system").Return(scopes, nil).Once()
				return scopeCfgProvider
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(id).Once()
				return uidSvc
			},
			HTTPServerFn: func(t *testing.T) *httptest.Server {
				tc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
				return tc
			},
		},
		{
			Name:          "Error - Invalid body",
			ExpectedError: errors.New("while decoding response body: invalid character 'D' looking for beginning of value"),
			ScopeCfgProviderFn: func() *automock.ScopeCfgProvider {
				scopeCfgProvider := &automock.ScopeCfgProvider{}
				scopeCfgProvider.On("GetRequiredScopes", "clientCredentialsRegistrationScopes.integration_system").Return(scopes, nil).Once()
				return scopeCfgProvider
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(id).Once()
				return uidSvc
			},
			HTTPServerFn: func(t *testing.T) *httptest.Server {
				tc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusCreated)
					_, err := w.Write([]byte("Dd"))
					assert.NoError(t, err)
				}))
				return tc
			},
		},
		{
			Name:          "Error - HTTP call error",
			ExpectedError: errors.New("connect: connection refused"),
			ScopeCfgProviderFn: func() *automock.ScopeCfgProvider {
				scopeCfgProvider := &automock.ScopeCfgProvider{}
				scopeCfgProvider.On("GetRequiredScopes", "clientCredentialsRegistrationScopes.integration_system").Return(scopes, nil).Once()
				return scopeCfgProvider
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				uidSvc.On("Generate").Return(id).Once()
				return uidSvc
			},
			HTTPServerFn: func(t *testing.T) *httptest.Server {
				tc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				tc.Close()
				return tc
			},
		},
		{
			Name:          "Error - Client Credential Scopes",
			ExpectedError: errors.New(fmt.Sprintf("while getting scopes for registering Client Credentials for %s: test err", objType)),
			ScopeCfgProviderFn: func() *automock.ScopeCfgProvider {
				scopeCfgProvider := &automock.ScopeCfgProvider{}
				scopeCfgProvider.On("GetRequiredScopes", "clientCredentialsRegistrationScopes.integration_system").Return(nil, testErr).Once()
				return scopeCfgProvider
			},
			UIDServiceFn: func() *automock.UIDService {
				uidSvc := &automock.UIDService{}
				return uidSvc
			},
			HTTPServerFn: func(t *testing.T) *httptest.Server {
				return nil
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			scopeCfgProvider := testCase.ScopeCfgProviderFn()
			defer scopeCfgProvider.AssertExpectations(t)
			uidService := testCase.UIDServiceFn()
			defer uidService.AssertExpectations(t)
			httpServer := testCase.HTTPServerFn(t)
			defer func() {
				if httpServer == nil {
					return
				}

				httpServer.Close()
			}()

			var url string
			if httpServer != nil {
				url = httpServer.URL
			}
			svc := oauth20.NewService(scopeCfgProvider, uidService, oauth20.Config{ClientEndpoint: url, PublicAccessTokenEndpoint: publicEndpoint}, &http.Client{})

			// when
			oauthData, err := svc.CreateClientCredentials(ctx, objType)

			// then
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				assert.Equal(t, testCase.ExpectedResult, oauthData)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}
		})
	}

}

func TestService_DeleteClientCredentials(t *testing.T) {
	// given
	id := "foo"
	testCases := []struct {
		Name          string
		ExpectedError error
		HTTPServerFn  func(t *testing.T) *httptest.Server
		Config        oauth20.Config
		Request       *http.Request
		Response      *http.Response
	}{
		{
			Name:          "Success",
			ExpectedError: nil,
			HTTPServerFn: func(t *testing.T) *httptest.Server {
				tc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					assert.Equal(t, "/foo", r.URL.Path)
					defer func() {
						err := r.Body.Close()
						assert.NoError(t, err)
					}()
					assert.Equal(t, http.MethodDelete, r.Method)
					assert.Equal(t, "application/json", r.Header.Get("Accept"))
					w.WriteHeader(http.StatusNoContent)
				}))
				return tc
			},
		},
		{
			Name:          "Error - Response Status Code",
			ExpectedError: errors.New("invalid HTTP status code: received: 500, expected 204"),
			HTTPServerFn: func(t *testing.T) *httptest.Server {
				tc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusInternalServerError)
				}))
				return tc
			},
		},
		{
			Name:          "Error - HTTP call error",
			ExpectedError: errors.New("connect: connection refused"),
			HTTPServerFn: func(t *testing.T) *httptest.Server {
				tc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {}))
				tc.Close()
				return tc
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			ctx := context.TODO()
			httpServer := testCase.HTTPServerFn(t)
			defer httpServer.Close()

			var url string
			if httpServer != nil {
				url = httpServer.URL
			}
			svc := oauth20.NewService(nil, nil, oauth20.Config{ClientEndpoint: url}, &http.Client{})

			// when
			err := svc.DeleteClientCredentials(ctx, id)

			// then
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedError.Error())
			}

		})
	}
}

func fixSuccessCreateClientHTTPServer(expectedReqBody map[string]interface{}) func(t *testing.T) *httptest.Server {
	return func(t *testing.T) *httptest.Server {
		tc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				err := r.Body.Close()
				assert.NoError(t, err)
			}()
			assert.Equal(t, http.MethodPost, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			assert.Equal(t, expectedReqBody, reqBody)

			res := map[string]interface{}{
				"client_secret": "c-secret",
			}
			w.WriteHeader(http.StatusCreated)
			err = json.NewEncoder(w).Encode(&res)
			require.NoError(t, err)
		}))
		return tc
	}
}
