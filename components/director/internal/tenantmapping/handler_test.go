package tenantmapping_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	"github.com/kyma-incubator/compass/components/director/internal/oathkeeper"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/director/internal/tenantmapping/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestHandler(t *testing.T) {
	username := "admin"
	target := "http://example.com/foo"
	externalTenantID := "external-" + uuid.New().String()
	tenantID := uuid.New()
	systemAuthID := uuid.New()
	objID := uuid.New()
	testError := errors.New("some error")
	txGen := txtest.NewTransactionContextGenerator(testError)

	jwtAuthDetails := oathkeeper.AuthDetails{AuthID: username, AuthFlow: oathkeeper.JWTAuthFlow}
	oAuthAuthDetails := oathkeeper.AuthDetails{AuthID: systemAuthID.String(), AuthFlow: oathkeeper.OAuth2Flow}
	certAuthDetails := oathkeeper.AuthDetails{AuthID: systemAuthID.String(), AuthFlow: oathkeeper.CertificateFlow}
	oneTimeTokenAuthDetails := oathkeeper.AuthDetails{AuthID: systemAuthID.String(), AuthFlow: oathkeeper.OneTimeTokenFlow}

	t.Run("success for the request parsed as JWT flow", func(t *testing.T) {
		scopes := "application:read"
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.UsernameKey: username,
				},
			},
		}
		objCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			Scopes:       scopes,
			ConsumerID:   username,
			ConsumerType: "Static User",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + username + `","consumerType":"Static User","externalTenant":"` + externalTenantID + `","name":"` + username + `","scope":"` + scopes + `","tenant":"` + tenantID.String() + `"},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetails).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.UserObjectContextProvider: userMockContextProvider,
		}

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, userMockContextProvider)
	})

	t.Run("success for the request parsed as JWT flow with custom authenticator", func(t *testing.T) {
		uniqueAttributeKey := "uniqueAttribute"
		uniqueAttributeValue := "uniqueAttributeValue"
		identityAttributeKey := "identity"
		scopes := "application:read"
		authenticatorName := "testAuthenticator"
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					uniqueAttributeKey:   uniqueAttributeValue,
					identityAttributeKey: username,
					authenticator.CoordinatesKey: authenticator.Coordinates{
						Name:  authenticatorName,
						Index: 0,
					},
				},
			},
		}
		objCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			Scopes:       scopes,
			ConsumerID:   username,
			ConsumerType: "Static User",
		}
		authn := []authenticator.Config{
			{
				Name: authenticatorName,
				Attributes: authenticator.Attributes{
					UniqueAttribute: authenticator.Attribute{
						Key:   uniqueAttributeKey,
						Value: uniqueAttributeValue,
					},
					IdentityAttribute: authenticator.Attribute{
						Key: identityAttributeKey,
					},
				},
				TrustedIssuers: []authenticator.TrustedIssuer{{ScopePrefix: ""}},
			},
		}

		jwtAuthDetailsWithAuthenticator := oathkeeper.AuthDetails{AuthID: username, AuthFlow: oathkeeper.JWTAuthFlow, Authenticator: &authn[0], ScopePrefix: ""}
		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + username + `","consumerType":"Static User","externalTenant":"` + externalTenantID + `","identity":"` + username + `","scope":"` + scopes + `","tenant":"` + tenantID.String() + `","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		authenticatorMockContextProvider := getMockContextProvider()
		authenticatorMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.AuthenticatorObjectContextProvider: authenticatorMockContextProvider,
		}

		handler := tenantmapping.NewHandler(authn, reqDataParserMock, transact, objectContextProviders)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, authenticatorMockContextProvider)
	})

	t.Run("success for the request parsed as JWT flow when both normal user is present and custom authenticator are present", func(t *testing.T) {
		uniqueAttributeKey := "uniqueAttribute"
		uniqueAttributeValue := "uniqueAttributeValue"
		identityAttributeKey := "identity"
		identityUsername := "identityAdmin"
		scopes := "application:read"
		authenticatorName := "testAuthenticator"
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.UsernameKey: username,
					uniqueAttributeKey:     uniqueAttributeValue,
					identityAttributeKey:   identityUsername,
					authenticator.CoordinatesKey: authenticator.Coordinates{
						Name:  authenticatorName,
						Index: 0,
					},
				},
			},
		}
		objCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			Scopes:       scopes,
			ConsumerID:   username,
			ConsumerType: "Static User",
		}
		authn := []authenticator.Config{
			{
				Name: authenticatorName,
				Attributes: authenticator.Attributes{
					UniqueAttribute: authenticator.Attribute{
						Key:   uniqueAttributeKey,
						Value: uniqueAttributeValue,
					},
					IdentityAttribute: authenticator.Attribute{
						Key: identityAttributeKey,
					},
				},
				TrustedIssuers: []authenticator.TrustedIssuer{{}},
			},
		}

		jwtAuthDetailsWithAuthenticator := oathkeeper.AuthDetails{AuthID: identityUsername, AuthFlow: oathkeeper.JWTAuthFlow, Authenticator: &authn[0]}
		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + username + `","consumerType":"Static User","externalTenant":"` + externalTenantID + `","identity":"` + identityUsername + `","name":"` + username + `","scope":"` + scopes + `","tenant":"` + tenantID.String() + `","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.AuthenticatorObjectContextProvider: userMockContextProvider,
		}

		handler := tenantmapping.NewHandler(authn, reqDataParserMock, transact, objectContextProviders)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, userMockContextProvider)
	})

	t.Run("success for the request parsed as JWT flow when both normal user is present and custom authenticator are present but no authenticator matches", func(t *testing.T) {
		uniqueAttributeKey := "uniqueAttribute"
		uniqueAttributeValue := "uniqueAttributeValue"
		identityAttributeKey := "identity"
		scopes := "application:read"
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.UsernameKey: username,
				},
			},
		}
		objCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			Scopes:       scopes,
			ConsumerID:   username,
			ConsumerType: "Static User",
		}
		authn := []authenticator.Config{
			{
				Attributes: authenticator.Attributes{
					UniqueAttribute: authenticator.Attribute{
						Key:   uniqueAttributeKey,
						Value: uniqueAttributeValue,
					},
					IdentityAttribute: authenticator.Attribute{
						Key: identityAttributeKey,
					},
				},
			},
		}

		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + username + `","consumerType":"Static User","externalTenant":"` + externalTenantID + `","name":"` + username + `","scope":"` + scopes + `","tenant":"` + tenantID.String() + `"},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetails).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.UserObjectContextProvider: userMockContextProvider,
		}

		handler := tenantmapping.NewHandler(authn, reqDataParserMock, transact, objectContextProviders)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, userMockContextProvider)
	})

	t.Run("success for the request parsed as OAuth2 flow", func(t *testing.T) {
		scopes := "application:read"
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.ClientIDKey: systemAuthID.String(),
				},
			},
		}
		objCtx := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			Scopes:       scopes,
			ConsumerID:   objID.String(),
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"client_id":"` + systemAuthID.String() + `","consumerID":"` + objID.String() + `","consumerType":"Integration System","externalTenant":"` + externalTenantID + `","scope":"` + scopes + `","tenant":"` + tenantID.String() + `"},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		systemAuthMockContextProvider := getMockContextProvider()
		systemAuthMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, oAuthAuthDetails).Return(objCtx, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.SystemAuthObjectContextProvider: systemAuthMockContextProvider,
		}

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, systemAuthMockContextProvider)
	})

	t.Run("success for the request parsed as Certificate flow", func(t *testing.T) {
		scopes := "application:read"
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: make(map[string]interface{}),
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey): []string{systemAuthID.String()},
				},
			},
		}
		objCtx := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			Scopes:       scopes,
			ConsumerID:   objID.String(),
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + objID.String() + `","consumerType":"Integration System","externalTenant":"` + externalTenantID + `","scope":"` + scopes + `","tenant":"` + tenantID.String() + `"},"header":{"Client-Id-From-Certificate":["` + systemAuthID.String() + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		systemAuthMockContextProvider := getMockContextProvider()
		systemAuthMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, certAuthDetails).Return(objCtx, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.SystemAuthObjectContextProvider: systemAuthMockContextProvider,
		}

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, systemAuthMockContextProvider)
	})

	t.Run("success for the request parsed as OneTimeToken flow", func(t *testing.T) {
		scopes := "application:read"
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: make(map[string]interface{}),
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDTokenKey): []string{systemAuthID.String()},
				},
			},
		}
		objCtx := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			Scopes:       scopes,
			ConsumerID:   objID.String(),
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + objID.String() + `","consumerType":"Integration System","externalTenant":"` + externalTenantID + `","scope":"` + scopes + `","tenant":"` + tenantID.String() + `"},"header":{"Client-Id-From-Token":["` + systemAuthID.String() + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		systemAuthMockContextProvider := getMockContextProvider()
		systemAuthMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, oneTimeTokenAuthDetails).Return(objCtx, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.SystemAuthObjectContextProvider: systemAuthMockContextProvider,
		}

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, systemAuthMockContextProvider)
	})

	t.Run("error when sending different HTTP verb than POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		handler := tenantmapping.NewHandler(nil, nil, nil, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, "Bad request method. Got GET, expected POST", strings.TrimSpace(string(body)))
	})

	t.Run("error when body parser returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(oathkeeper.ReqData{}, errors.New("some error")).Once()

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, nil, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := oathkeeper.ReqData{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, oathkeeper.ReqData{}, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock)
	})

	t.Run("error when transaction commit fails", func(t *testing.T) {
		username := "admin"
		scopes := "application:read"
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.UsernameKey: username,
				},
			},
		}
		objCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			Scopes:       scopes,
			ConsumerID:   username,
			ConsumerType: "Static User",
		}

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqData, nil).Once()

		persist, transact := txGen.ThatFailsOnCommit()

		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqData, jwtAuthDetails).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.UserObjectContextProvider: userMockContextProvider,
		}

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := oathkeeper.ReqBody{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, reqData.Body, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, userMockContextProvider)
	})

	t.Run("error when transaction begin fails", func(t *testing.T) {
		reqData := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.UsernameKey: "test",
				},
			},
		}

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader("{}"))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqData, nil).Once()

		persist, transact := txGen.ThatFailsOnBegin()

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := oathkeeper.ReqBody{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, reqData.Body, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact)
	})

	t.Run("error when GetAuthID returns error when looking for Auth ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(oathkeeper.ReqData{}, nil).Once()

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, nil, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := oathkeeper.ReqBody{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, oathkeeper.ReqBody{}, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock)
	})
}

func getMockContextProvider() *automock.ObjectContextProvider {
	provider := &automock.ObjectContextProvider{}
	return provider
}
