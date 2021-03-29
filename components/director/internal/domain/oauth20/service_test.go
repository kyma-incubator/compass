package oauth20_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20"
	"github.com/kyma-incubator/compass/components/director/internal/domain/oauth20/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type hydraLinkMode string

const (
	firstAndNextAvailable  hydraLinkMode = "noClients"
	onlyFirstLinkAvailable hydraLinkMode = "singlePage"
	noLinkHeaderPresent    hydraLinkMode = "noLinkHeader"
	firstPage              hydraLinkMode = "firstPage"
	secondPage             hydraLinkMode = "secondPage"
)

const (
	publicEndpoint = "accessTokenURL"
	clientID       = "clientid"
	clientSecret   = "secret"
	objType        = model.IntegrationSystemReference
)

var scopes = []string{"foo", "bar", "baz"}

func TestService_CreateClient(t *testing.T) {
	// given
	successResult := &model.OAuthCredentialDataInput{
		ClientID:     clientID,
		ClientSecret: clientSecret,
		URL:          publicEndpoint,
	}
	expectedReqBody := map[string]interface{}{
		"grant_types": []interface{}{"client_credentials"},
		"client_id":   clientID,
		"scope":       strings.Join(scopes, " "),
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
				uidSvc.On("Generate").Return(clientID).Once()
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
				uidSvc.On("Generate").Return(clientID).Once()
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
				uidSvc.On("Generate").Return(clientID).Once()
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
				uidSvc.On("Generate").Return(clientID).Once()
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

func TestService_UpdateClient(t *testing.T) {
	// given
	testErr := errors.New("test err")
	expectedReqBody := map[string]interface{}{
		"scope": strings.Join(scopes, " "),
	}
	testCases := []struct {
		Name                 string
		ExpectedResult       *model.OAuthCredentialDataInput
		ExpectedResponseCode int
		ExpectedError        error
		ScopeCfgProviderFn   func() *automock.ScopeCfgProvider
		HTTPServerFn         func(t *testing.T) *httptest.Server
		Config               oauth20.Config
		Request              *http.Request
		Response             *http.Response
	}{
		{
			Name:                 "Success",
			ExpectedError:        nil,
			ExpectedResponseCode: http.StatusOK,
			ScopeCfgProviderFn: func() *automock.ScopeCfgProvider {
				scopeCfgProvider := &automock.ScopeCfgProvider{}
				scopeCfgProvider.On("GetRequiredScopes", "clientCredentialsRegistrationScopes.integration_system").Return(scopes, nil).Once()
				return scopeCfgProvider
			},
			HTTPServerFn: fixSuccessUpdateClientHTTPServer(expectedReqBody),
		},
		{
			Name:          "Error - Response Status Code",
			ExpectedError: errors.New("invalid HTTP status code: received: 500, expected 200"),
			ScopeCfgProviderFn: func() *automock.ScopeCfgProvider {
				scopeCfgProvider := &automock.ScopeCfgProvider{}
				scopeCfgProvider.On("GetRequiredScopes", "clientCredentialsRegistrationScopes.integration_system").Return(scopes, nil).Once()
				return scopeCfgProvider
			},
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
			ScopeCfgProviderFn: func() *automock.ScopeCfgProvider {
				scopeCfgProvider := &automock.ScopeCfgProvider{}
				scopeCfgProvider.On("GetRequiredScopes", "clientCredentialsRegistrationScopes.integration_system").Return(scopes, nil).Once()
				return scopeCfgProvider
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
			uidService := &automock.UIDService{}
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
			err := svc.UpdateClientScopes(ctx, clientID, objType)

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

			svc := oauth20.NewService(nil, nil, oauth20.Config{ClientEndpoint: httpServer.URL}, &http.Client{})

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

func TestService_ListClients(t *testing.T) {
	// given
	testCases := []struct {
		Name                 string
		ExpectedResponseCode int
		ExpectedError        error
		HTTPServerFn         func(t *testing.T) *httptest.Server
		Config               oauth20.Config
		Request              *http.Request
		Response             *http.Response
	}{
		{
			Name:                 "Success when response is in two pages",
			ExpectedError:        nil,
			ExpectedResponseCode: http.StatusOK,
			HTTPServerFn:         fixSuccessPageableListClientsHTTPServer(),
		},
		{
			Name:                 "Success when clients are in a single page and next link is available",
			ExpectedError:        nil,
			ExpectedResponseCode: http.StatusOK,
			HTTPServerFn:         fixSuccessSinglePageListClientsHTTPServer(firstAndNextAvailable),
		},
		{
			Name:                 "Success when clients are in a single page and next link is not available",
			ExpectedError:        nil,
			ExpectedResponseCode: http.StatusOK,
			HTTPServerFn:         fixSuccessSinglePageListClientsHTTPServer(onlyFirstLinkAvailable),
		},
		{
			Name:                 "Success when no link header is present",
			ExpectedError:        nil,
			ExpectedResponseCode: http.StatusOK,
			HTTPServerFn:         fixSuccessSinglePageListClientsHTTPServer(noLinkHeaderPresent),
		},
		{
			Name:          "Error - Response Status Code",
			ExpectedError: errors.New("invalid HTTP status code: received: 500, expected 200"),
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
			scopeCfgProvider := &automock.ScopeCfgProvider{}
			defer scopeCfgProvider.AssertExpectations(t)
			uidService := &automock.UIDService{}
			defer uidService.AssertExpectations(t)
			httpServer := testCase.HTTPServerFn(t)
			defer httpServer.Close()

			svc := oauth20.NewService(scopeCfgProvider, uidService, oauth20.Config{ClientEndpoint: httpServer.URL, PublicAccessTokenEndpoint: publicEndpoint}, &http.Client{})

			// when
			clients, err := svc.ListClients(ctx)

			// then
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
				require.Len(t, clients, 1)
			} else {
				require.Error(t, err)
				require.Len(t, clients, 0)
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
				"client_secret": clientSecret,
			}
			w.WriteHeader(http.StatusCreated)
			err = json.NewEncoder(w).Encode(&res)
			require.NoError(t, err)
		}))
		return tc
	}
}

func fixSuccessUpdateClientHTTPServer(expectedReqBody map[string]interface{}) func(t *testing.T) *httptest.Server {
	return func(t *testing.T) *httptest.Server {
		tc := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			defer func() {
				err := r.Body.Close()
				assert.NoError(t, err)
			}()
			assert.Equal(t, http.MethodPut, r.Method)
			assert.Equal(t, "application/json", r.Header.Get("Content-Type"))

			var reqBody map[string]interface{}
			err := json.NewDecoder(r.Body).Decode(&reqBody)
			require.NoError(t, err)
			assert.Equal(t, expectedReqBody, reqBody)

			w.WriteHeader(http.StatusOK)
		}))
		return tc
	}
}

func linkHeaderForMode(mode hydraLinkMode) string {
	var linkHeader string
	switch mode {
	case firstAndNextAvailable:
		linkHeader = `</clients?limit=100&offset=0>; rel="first",</clients?limit=100&offset=100>; rel="next",</clients?limit=100&offset=-100>; rel="prev"`
	case onlyFirstLinkAvailable:
		linkHeader = `</clients?limit=3&offset=0>; rel="first"`
	case firstPage:
		linkHeader = `</clients?limit=1&offset=1>; rel="next",</clients?limit=1&offset=33>; rel="last"`
	case secondPage:
		linkHeader = `</clients?limit=1&offset=0>; rel="first",</clients?limit=1&offset=0>; rel="prev"`
	}

	return linkHeader
}

func fixSuccessSinglePageListClientsHTTPServer(mode hydraLinkMode) func(t *testing.T) *httptest.Server {
	return func(t *testing.T) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			res := []map[string]interface{}{
				{
					"client_id": clientID,
					"scope":     strings.Join(scopes, " "),
				},
			}

			if linkHeader := linkHeaderForMode(mode); linkHeader != "" {
				w.Header().Add("link", linkHeader)
			}
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(&res)
			require.NoError(t, err)
		}))
	}
}

var listCalls = 0

func fixSuccessPageableListClientsHTTPServer() func(t *testing.T) *httptest.Server {
	return func(t *testing.T) *httptest.Server {
		return httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			var linkHeader string
			var res []map[string]interface{}
			if listCalls >= 1 {
				res = []map[string]interface{}{}
				linkHeader = linkHeaderForMode(secondPage)
			} else {
				res = []map[string]interface{}{
					{
						"client_id": clientID,
						"scope":     strings.Join(scopes, " "),
					},
				}
				linkHeader = linkHeaderForMode(firstPage)
				listCalls++
			}

			w.Header().Add("link", linkHeader)
			w.WriteHeader(http.StatusOK)
			err := json.NewEncoder(w).Encode(&res)
			require.NoError(t, err)
		}))
	}
}
