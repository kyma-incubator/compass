package tenantmapping_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/internal/consumer"

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
	certAuthDetails := oathkeeper.AuthDetails{AuthID: systemAuthID.String(), AuthFlow: oathkeeper.CertificateFlow, CertIssuer: oathkeeper.ConnectorIssuer}
	externalCertAuthDetails := oathkeeper.AuthDetails{AuthID: externalTenantID, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: oathkeeper.ExternalIssuer}
	oneTimeTokenAuthDetails := oathkeeper.AuthDetails{AuthID: systemAuthID.String(), AuthFlow: oathkeeper.OneTimeTokenFlow}

	t.Run("success for the request parsed as JWT flow", func(t *testing.T) {
		scopes := "application:read"

		keys := tenantmapping.KeysExtra{
			TenantKey:         "consumerTenant",
			ExternalTenantKey: "externalTenant",
		}

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
			KeysExtra:    keys,
			Scopes:       scopes,
			ConsumerID:   username,
			AuthFlow:     oathkeeper.JWTAuthFlow,
			ConsumerType: "Static User",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper.JWTAuthFlow) + `","name":"` + username + `","onBehalfOf":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}"},"header":null}`
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
		clientInstrumenter.On("InstrumentClient", username, string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
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
			KeysExtra:    keys,
			Scopes:       scopes,
			ConsumerID:   username,
			AuthFlow:     oathkeeper.JWTAuthFlow,
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
		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper.JWTAuthFlow) + `","identity":"` + username + `","onBehalfOf":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":null}`

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
		clientInstrumenter.On("InstrumentClient", username, string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(authn, reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
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
			KeysExtra:    keys,
			Scopes:       scopes,
			ConsumerID:   username,
			AuthFlow:     oathkeeper.JWTAuthFlow,
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
		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper.JWTAuthFlow) + `","identity":"` + identityUsername + `","name":"` + username + `","onBehalfOf":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":null}`

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
		clientInstrumenter.On("InstrumentClient", identityUsername, string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(authn, reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
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
			KeysExtra:    keys,
			Scopes:       scopes,
			ConsumerID:   username,
			AuthFlow:     oathkeeper.JWTAuthFlow,
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

		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper.JWTAuthFlow) + `","name":"` + username + `","onBehalfOf":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}"},"header":null}`

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
		clientInstrumenter.On("InstrumentClient", username, string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(authn, reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
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
			KeysExtra:    keys,
			Scopes:       scopes,
			ConsumerID:   objID.String(),
			AuthFlow:     oathkeeper.OAuth2Flow,
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"client_id":"` + systemAuthID.String() + `","consumerID":"` + objID.String() + `","consumerType":"Integration System","flow":"` + string(oathkeeper.OAuth2Flow) + `","onBehalfOf":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}"},"header":null}`

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
		clientInstrumenter.On("InstrumentClient", systemAuthID.String(), string(oathkeeper.OAuth2Flow), mock.Anything)

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
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
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: make(map[string]interface{}),
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{systemAuthID.String()},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ConnectorIssuer},
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
			AuthFlow:     oathkeeper.CertificateFlow,
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + objID.String() + `","consumerType":"Integration System","flow":"` + string(oathkeeper.CertificateFlow) + `","onBehalfOf":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}"},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ConnectorIssuer + `"],"Client-Id-From-Certificate":["` + systemAuthID.String() + `"]}}`

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
		clientInstrumenter.On("InstrumentClient", systemAuthID.String(), string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
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
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: make(map[string]interface{}),
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{externalTenantID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
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
			AuthFlow:     oathkeeper.CertificateFlow,
			ConsumerType: consumer.Runtime,
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper.CertificateFlow) + `","onBehalfOf":"","scope":"` + "" + `","tenant":"{\\\"consumerTenant\\\":\\\"` + externalTenantID + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}"},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

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
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
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
			KeysExtra:    keys,
			Scopes:       scopes,
			ConsumerID:   objID.String(),
			AuthFlow:     oathkeeper.OneTimeTokenFlow,
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + objID.String() + `","consumerType":"Integration System","flow":"` + string(oathkeeper.OneTimeTokenFlow) + `","onBehalfOf":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}"},"header":{"Client-Id-From-Token":["` + systemAuthID.String() + `"]}}`

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
		clientInstrumenter.On("InstrumentClient", systemAuthID.String(), string(oathkeeper.OneTimeTokenFlow), mock.Anything)

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
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
				Header: http.Header{
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertKey):    []string{externalTenantID},
					textproto.CanonicalMIMEHeaderKey(oathkeeper.ClientIDCertIssuer): []string{oathkeeper.ExternalIssuer},
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
			AuthFlow:        oathkeeper.CertificateFlow,
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
			ConsumerID:      username,
			AuthFlow:        oathkeeper.JWTAuthFlow,
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

		jwtAuthDetailsWithAuthenticator := oathkeeper.AuthDetails{AuthID: username, AuthFlow: oathkeeper.JWTAuthFlow, Authenticator: &authn[0], ScopePrefix: ""}

		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper.CertificateFlow) + `","identity":"` + username + `","onBehalfOf":"admin","scope":"test","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

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
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
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

		handler := tenantmapping.NewHandler(nil, nil, nil, nil, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, "Bad request method. Got GET, expected POST", strings.TrimSpace(string(body)))
	})

	t.Run("error when object context provider fails to provide object context", func(t *testing.T) {
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.UsernameKey: username,
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
		clientInstrumenter.On("InstrumentClient", systemAuthID.String(), string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
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
		reqDataParserMock.On("Parse", mock.Anything).Return(oathkeeper.ReqData{}, errors.New("some error")).Once()

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, nil, nil, nil)
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
			AuthFlow:     oathkeeper.OAuth2Flow,
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
		clientInstrumenter.On("InstrumentClient", username, string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, objectContextProviders, clientInstrumenter)
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

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", "test", string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, nil, clientInstrumenter)
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

		persist, transact := txGen.ThatFailsOnBegin()
		handler := tenantmapping.NewHandler(nil, reqDataParserMock, transact, nil, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := oathkeeper.ReqBody{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, oathkeeper.ReqBody{}, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock, persist)
	})
}

func getMockContextProvider() *automock.ObjectContextProvider {
	provider := &automock.ObjectContextProvider{}
	return provider
}
