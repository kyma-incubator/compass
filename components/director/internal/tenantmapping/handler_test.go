package tenantmapping_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	oathkeeper2 "github.com/kyma-incubator/compass/components/director/pkg/oathkeeper"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"

	"github.com/kyma-incubator/compass/components/director/pkg/persistence/txtest"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
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

	jwtAuthDetails := oathkeeper2.AuthDetails{AuthID: username, AuthFlow: oathkeeper2.JWTAuthFlow}
	oAuthAuthDetails := oathkeeper2.AuthDetails{AuthID: systemAuthID.String(), AuthFlow: oathkeeper2.OAuth2Flow}
	certAuthDetails := oathkeeper2.AuthDetails{AuthID: systemAuthID.String(), AuthFlow: oathkeeper2.CertificateFlow, CertIssuer: oathkeeper2.ConnectorIssuer}
	externalCertAuthDetails := oathkeeper2.AuthDetails{AuthID: externalTenantID, AuthFlow: oathkeeper2.CertificateFlow, CertIssuer: oathkeeper2.ExternalIssuer}
	oneTimeTokenAuthDetails := oathkeeper2.AuthDetails{AuthID: systemAuthID.String(), AuthFlow: oathkeeper2.OneTimeTokenFlow}

	t.Run("success for the request parsed as JWT flow", func(t *testing.T) {
		scopes := "application:read"

		keys := tenantmapping.KeysExtra{
			TenantKey:         "consumerTenant",
			ExternalTenantKey: "externalTenant",
		}

		reqDataMock := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper2.UsernameKey: username,
				},
			},
		}

		objCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			KeysExtra:    keys,
			Scopes:       scopes,
			ConsumerID:   username,
			AuthFlow:     oathkeeper2.JWTAuthFlow,
			ConsumerType: "Static User",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper2.JWTAuthFlow) + `","name":"` + username + `","onBehalfOf":"","region":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":null}`
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetails, nil).Once()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetails).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.UserObjectContextProvider: userMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", username, string(oathkeeper2.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, userMockContextProvider)
	})

	t.Run("success for the request parsed as JWT flow with custom authenticator", func(t *testing.T) {
		keys := tenantmapping.KeysExtra{
			TenantKey:         "consumerTenant",
			ExternalTenantKey: "externalTenant",
		}
		uniqueAttributeKey := "uniqueAttribute"
		uniqueAttributeValue := "uniqueAttributeValue"
		identityAttributeKey := "identity"
		clientIDAttributeKey := "clientid"
		scopes := "application:read"
		authenticatorName := "testAuthenticator"
		reqDataMock := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Extra: map[string]interface{}{
					uniqueAttributeKey:   uniqueAttributeValue,
					identityAttributeKey: username,
					clientIDAttributeKey: "client_id",
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
			KeysExtra:       keys,
			Scopes:          scopes,
			Region:          "region",
			OauthClientID:   "client_id",
			ConsumerID:      username,
			AuthFlow:        oathkeeper2.JWTAuthFlow,
			ConsumerType:    "Static User",
			ContextProvider: tenantmapping.AuthenticatorObjectContextProvider,
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
					ClientID: authenticator.Attribute{
						Key: clientIDAttributeKey,
					},
				},
				TrustedIssuers: []authenticator.TrustedIssuer{{ScopePrefix: "", Region: "region"}},
			},
		}

		jwtAuthDetailsWithAuthenticator := oathkeeper2.AuthDetails{AuthID: username, AuthFlow: oathkeeper2.JWTAuthFlow, Authenticator: &authn[0], ScopePrefix: "", Region: "region"}
		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"` + clientIDAttributeKey + `":"client_id","consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper2.JWTAuthFlow) + `","identity":"` + username + `","onBehalfOf":"","region":"region","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"client_id","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		authenticatorMockContextProvider := getMockContextProvider()
		authenticatorMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		authenticatorMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.AuthenticatorObjectContextProvider: authenticatorMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", username, string(oathkeeper2.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, authenticatorMockContextProvider)
	})

	t.Run("success for the request parsed as JWT flow when both normal user is present and custom authenticator are present", func(t *testing.T) {
		keys := tenantmapping.KeysExtra{
			TenantKey:         "consumerTenant",
			ExternalTenantKey: "externalTenant",
		}
		uniqueAttributeKey := "uniqueAttribute"
		uniqueAttributeValue := "uniqueAttributeValue"
		identityAttributeKey := "identity"
		identityUsername := "identityAdmin"
		scopes := "application:read"
		authenticatorName := "testAuthenticator"
		reqDataMock := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper2.UsernameKey: username,
					uniqueAttributeKey:      uniqueAttributeValue,
					identityAttributeKey:    identityUsername,
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
			KeysExtra:       keys,
			Region:          "region",
			OauthClientID:   "client_id",
			Scopes:          scopes,
			ConsumerID:      username,
			AuthFlow:        oathkeeper2.JWTAuthFlow,
			ConsumerType:    "Static User",
			ContextProvider: tenantmapping.AuthenticatorObjectContextProvider,
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

		jwtAuthDetailsWithAuthenticator := oathkeeper2.AuthDetails{AuthID: identityUsername, AuthFlow: oathkeeper2.JWTAuthFlow, Authenticator: &authn[0]}
		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper2.JWTAuthFlow) + `","identity":"` + identityUsername + `","name":"` + username + `","onBehalfOf":"","region":"region","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"client_id","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.AuthenticatorObjectContextProvider: userMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", identityUsername, string(oathkeeper2.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, userMockContextProvider)
	})

	t.Run("success for the request parsed as JWT flow when both normal user is present and custom authenticator are present but no authenticator matches", func(t *testing.T) {
		keys := tenantmapping.KeysExtra{
			TenantKey:         "consumerTenant",
			ExternalTenantKey: "externalTenant",
		}

		scopes := "application:read"
		reqDataMock := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper2.UsernameKey: username,
				},
			},
		}

		objCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			KeysExtra:    keys,
			Scopes:       scopes,
			ConsumerID:   username,
			AuthFlow:     oathkeeper2.JWTAuthFlow,
			ConsumerType: "Static User",
		}

		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper2.JWTAuthFlow) + `","name":"` + username + `","onBehalfOf":"","region":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetails, nil).Once()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetails).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.UserObjectContextProvider: userMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", username, string(oathkeeper2.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, userMockContextProvider)
	})

	t.Run("success for the request parsed as OAuth2 flow", func(t *testing.T) {
		keys := tenantmapping.KeysExtra{
			TenantKey:         "consumerTenant",
			ExternalTenantKey: "externalTenant",
		}
		scopes := "application:read"
		reqDataMock := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper2.ClientIDKey: systemAuthID.String(),
				},
			},
		}

		objCtx := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			KeysExtra:    keys,
			Scopes:       scopes,
			ConsumerID:   objID.String(),
			AuthFlow:     oathkeeper2.OAuth2Flow,
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"client_id":"` + systemAuthID.String() + `","consumerID":"` + objID.String() + `","consumerType":"Integration System","flow":"` + string(oathkeeper2.OAuth2Flow) + `","onBehalfOf":"","region":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		systemAuthMockContextProvider := getMockContextProvider()
		systemAuthMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &oAuthAuthDetails, nil).Once()
		systemAuthMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, oAuthAuthDetails).Return(objCtx, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.SystemAuthObjectContextProvider: systemAuthMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", systemAuthID.String(), string(oathkeeper2.OAuth2Flow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, systemAuthMockContextProvider)
	})

	t.Run("success for the request parsed as Certificate flow for Connector issuer", func(t *testing.T) {
		keys := tenantmapping.KeysExtra{
			TenantKey:         "consumerTenant",
			ExternalTenantKey: "externalTenant",
		}
		scopes := "application:read"
		reqDataMock := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Extra: make(map[string]interface{}),
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertKey):    []string{systemAuthID.String()},
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertIssuer): []string{oathkeeper2.ConnectorIssuer},
				},
			},
		}

		objCtx := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			KeysExtra:    keys,
			Scopes:       scopes,
			ConsumerID:   objID.String(),
			AuthFlow:     oathkeeper2.CertificateFlow,
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + objID.String() + `","consumerType":"Integration System","flow":"` + string(oathkeeper2.CertificateFlow) + `","onBehalfOf":"","region":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":{"Client-Certificate-Issuer":["` + oathkeeper2.ConnectorIssuer + `"],"Client-Id-From-Certificate":["` + systemAuthID.String() + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		systemAuthMockContextProvider := getMockContextProvider()
		systemAuthMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &certAuthDetails, nil).Once()
		systemAuthMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, certAuthDetails).Return(objCtx, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.SystemAuthObjectContextProvider: systemAuthMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", systemAuthID.String(), string(oathkeeper2.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, systemAuthMockContextProvider)
	})

	t.Run("success for the request parsed as Certificate flow for External issuer", func(t *testing.T) {
		keys := tenantmapping.KeysExtra{
			TenantKey:         "providerTenant",
			ExternalTenantKey: "providerExternalTenant",
		}
		reqDataMock := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Extra: make(map[string]interface{}),
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertKey):    []string{externalTenantID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertIssuer): []string{oathkeeper2.ExternalIssuer},
				},
			},
		}

		objCtx := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         externalTenantID,
			},
			KeysExtra:    keys,
			ConsumerID:   externalTenantID,
			AuthFlow:     oathkeeper2.CertificateFlow,
			ConsumerType: consumer.Runtime,
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper2.CertificateFlow) + `","onBehalfOf":"","region":"","scope":"` + "" + `","tenant":"{\\\"consumerTenant\\\":\\\"` + externalTenantID + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":{"Client-Certificate-Issuer":["` + oathkeeper2.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		certServiceMockContextProvider := getMockContextProvider()
		certServiceMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &externalCertAuthDetails, nil).Once()
		certServiceMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, externalCertAuthDetails).Return(objCtx, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.CertServiceObjectContextProvider: certServiceMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper2.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, certServiceMockContextProvider)
	})

	t.Run("success for the request parsed as OneTimeToken flow", func(t *testing.T) {
		keys := tenantmapping.KeysExtra{
			TenantKey:         "consumerTenant",
			ExternalTenantKey: "externalTenant",
		}
		scopes := "application:read"
		reqDataMock := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Extra: make(map[string]interface{}),
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDTokenKey): []string{systemAuthID.String()},
				},
			},
		}

		objCtx := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			KeysExtra:    keys,
			Scopes:       scopes,
			ConsumerID:   objID.String(),
			AuthFlow:     oathkeeper2.OneTimeTokenFlow,
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + objID.String() + `","consumerType":"Integration System","flow":"` + string(oathkeeper2.OneTimeTokenFlow) + `","onBehalfOf":"","region":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":{"Client-Id-From-Token":["` + systemAuthID.String() + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		systemAuthMockContextProvider := getMockContextProvider()
		systemAuthMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &oneTimeTokenAuthDetails, nil).Once()
		systemAuthMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, oneTimeTokenAuthDetails).Return(objCtx, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.SystemAuthObjectContextProvider: systemAuthMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", systemAuthID.String(), string(oathkeeper2.OneTimeTokenFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, systemAuthMockContextProvider)
	})

	t.Run("success for the request parsed as Certificate flow for External issuer and JWT flow with custom authenticator", func(t *testing.T) {
		certKeys := tenantmapping.KeysExtra{
			TenantKey:         "providerTenant",
			ExternalTenantKey: "providerExternalTenant",
		}

		JWTKeys := tenantmapping.KeysExtra{
			TenantKey:         "consumerTenant",
			ExternalTenantKey: "externalTenant",
		}

		uniqueAttributeKey := "uniqueAttribute"
		uniqueAttributeValue := "uniqueAttributeValue"
		identityAttributeKey := "identity"
		scopes := "application:read test"
		authenticatorName := "testAuthenticator"
		reqDataMock := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Extra: map[string]interface{}{
					uniqueAttributeKey:   uniqueAttributeValue,
					identityAttributeKey: username,
					authenticator.CoordinatesKey: authenticator.Coordinates{
						Name:  authenticatorName,
						Index: 0,
					},
				},
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertKey):    []string{externalTenantID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper2.ClientIDCertIssuer): []string{oathkeeper2.ExternalIssuer},
				},
			},
		}

		certObjCtx := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         externalTenantID,
			},
			Scopes:          "test",
			KeysExtra:       certKeys,
			ConsumerID:      externalTenantID,
			AuthFlow:        oathkeeper2.CertificateFlow,
			ConsumerType:    consumer.Runtime,
			ContextProvider: tenantmapping.CertServiceObjectContextProvider,
		}

		certServiceMockContextProvider := getMockContextProvider()
		certServiceMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &externalCertAuthDetails, nil).Once()
		certServiceMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, externalCertAuthDetails).Return(certObjCtx, nil).Once()

		JWTObjCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			KeysExtra:       JWTKeys,
			Scopes:          scopes,
			Region:          "region",
			OauthClientID:   "client_id",
			ConsumerID:      username,
			AuthFlow:        oathkeeper2.JWTAuthFlow,
			ConsumerType:    "Static User",
			ContextProvider: tenantmapping.AuthenticatorObjectContextProvider,
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

		jwtAuthDetailsWithAuthenticator := oathkeeper2.AuthDetails{AuthID: username, AuthFlow: oathkeeper2.JWTAuthFlow, Authenticator: &authn[0], ScopePrefix: ""}

		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper2.CertificateFlow) + `","identity":"` + username + `","onBehalfOf":"admin","region":"region","scope":"test","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"client_id","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":{"Client-Certificate-Issuer":["` + oathkeeper2.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		authenticatorMockContextProvider := getMockContextProvider()
		authenticatorMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		authenticatorMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(JWTObjCtxMock, nil).Once()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatSucceeds()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.CertServiceObjectContextProvider:   certServiceMockContextProvider,
			tenantmapping.AuthenticatorObjectContextProvider: authenticatorMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper2.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		//require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))
		require.Equal(t, gjson.Get(expectedRespPayload, "subject"), gjson.Get(strings.TrimSpace(string(body)), "subject"))
		require.Equal(t, gjson.Get(expectedRespPayload, "extra"), gjson.Get(strings.TrimSpace(string(body)), "extra"))
		require.Equal(t, gjson.Get(expectedRespPayload, "header"), gjson.Get(strings.TrimSpace(string(body)), "header"))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, certServiceMockContextProvider)
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

	t.Run("error when object context provider fails to provide object context", func(t *testing.T) {
		reqDataMock := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper2.UsernameKey: username,
				},
			},
		}

		expectedRespPayload := `{"subject":"","extra":{"name":"` + username + `"},"header":null}`
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		persist, transact := txGen.ThatDoesntExpectCommit()

		expectedError := errors.New("error")
		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetails, nil).Once()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetails).Return(tenantmapping.ObjectContext{}, expectedError).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.UserObjectContextProvider: userMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", systemAuthID.String(), string(oathkeeper2.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, userMockContextProvider)
	})

	t.Run("error when body parser returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(oathkeeper2.ReqData{}, errors.New("some error")).Once()

		handler := tenantmapping.NewHandler(reqDataParserMock, nil, nil, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := oathkeeper2.ReqData{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, oathkeeper2.ReqData{}, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock)
	})

	t.Run("error when transaction commit fails", func(t *testing.T) {
		username := "admin"
		scopes := "application:read"
		reqData := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper2.UsernameKey: username,
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
			AuthFlow:     oathkeeper2.OAuth2Flow,
			ConsumerType: "Static User",
		}

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqData, nil).Once()

		persist, transact := txGen.ThatFailsOnCommit()

		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("Match", mock.Anything, reqData).Return(true, &jwtAuthDetails, nil).Once()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqData, jwtAuthDetails).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmapping.UserObjectContextProvider: userMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", username, string(oathkeeper2.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := oathkeeper2.ReqBody{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, reqData.Body, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact, userMockContextProvider)
	})

	t.Run("error when transaction begin fails", func(t *testing.T) {
		reqData := oathkeeper2.ReqData{
			Body: oathkeeper2.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper2.UsernameKey: "test",
				},
			},
		}

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader("{}"))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqData, nil).Once()

		persist, transact := txGen.ThatFailsOnBegin()

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", "test", string(oathkeeper2.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, transact, nil, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := oathkeeper2.ReqBody{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, reqData.Body, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist, transact)
	})

	t.Run("error when GetAuthID returns error when looking for Auth ID", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(oathkeeper2.ReqData{}, nil).Once()

		persist, transact := txGen.ThatFailsOnBegin()
		handler := tenantmapping.NewHandler(reqDataParserMock, transact, nil, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := oathkeeper2.ReqBody{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, oathkeeper2.ReqBody{}, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist)
	})
}

func getMockContextProvider() *automock.ObjectContextProvider {
	provider := &automock.ObjectContextProvider{}
	return provider
}
