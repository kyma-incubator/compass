package tenantmapping_test

import (
	"encoding/json"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"net/textproto"
	"strings"
	"testing"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"

	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping"
	"github.com/kyma-incubator/compass/components/hydrator/internal/tenantmapping/automock"
	tenantmappingconsts "github.com/kyma-incubator/compass/components/hydrator/pkg/tenantmapping"

	"github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"

	"github.com/tidwall/gjson"

	"github.com/kyma-incubator/compass/components/director/pkg/consumer"

	"github.com/stretchr/testify/assert"

	"github.com/google/uuid"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

const (
	overrideAllScopes        = "overrideAllScopes"
	mergeWithOtherScopes     = "mergeWithOtherScopes"
	intersectWithOtherScopes = "intersectWithOtherScopes"
)

func TestHandler(t *testing.T) {
	username := "admin"
	target := "http://example.com/foo"
	externalTenantID := "external-" + uuid.New().String()
	region := "eu-1"
	tenantID := uuid.New()
	systemAuthID := uuid.New()
	objID := uuid.New()

	jwtAuthDetails := oathkeeper.AuthDetails{AuthID: username, Region: region, AuthFlow: oathkeeper.JWTAuthFlow}
	oAuthAuthDetails := oathkeeper.AuthDetails{AuthID: systemAuthID.String(), Region: region, AuthFlow: oathkeeper.OAuth2Flow}
	certAuthDetails := oathkeeper.AuthDetails{AuthID: systemAuthID.String(), Region: region, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: oathkeeper.ConnectorIssuer}
	externalCertAuthDetails := oathkeeper.AuthDetails{AuthID: externalTenantID, Region: region, AuthFlow: oathkeeper.CertificateFlow, CertIssuer: oathkeeper.ExternalIssuer}
	oneTimeTokenAuthDetails := oathkeeper.AuthDetails{AuthID: systemAuthID.String(), Region: region, AuthFlow: oathkeeper.OneTimeTokenFlow}

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
			Region:       region,
			ConsumerType: "Static User",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper.JWTAuthFlow) + `","name":"` + username + `","onBehalfOf":"","region":"` + region + `","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":null}`
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetails, nil).Once()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetails).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.UserObjectContextProvider: userMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", username, string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, userMockContextProvider)
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
		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
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
			AuthFlow:        oathkeeper.JWTAuthFlow,
			ConsumerType:    "Static User",
			ContextProvider: tenantmappingconsts.AuthenticatorObjectContextProvider,
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

		jwtAuthDetailsWithAuthenticator := oathkeeper.AuthDetails{AuthID: username, AuthFlow: oathkeeper.JWTAuthFlow, Authenticator: &authn[0], ScopePrefix: "", Region: "region"}
		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"` + clientIDAttributeKey + `":"client_id","consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper.JWTAuthFlow) + `","identity":"` + username + `","onBehalfOf":"","region":"region","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"client_id","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		authenticatorMockContextProvider := getMockContextProvider()
		authenticatorMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		authenticatorMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.AuthenticatorObjectContextProvider: authenticatorMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", username, string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, authenticatorMockContextProvider)
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
			KeysExtra:       keys,
			Region:          "region",
			OauthClientID:   "client_id",
			Scopes:          scopes,
			ConsumerID:      username,
			AuthFlow:        oathkeeper.JWTAuthFlow,
			ConsumerType:    "Static User",
			ContextProvider: tenantmappingconsts.AuthenticatorObjectContextProvider,
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
		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper.JWTAuthFlow) + `","identity":"` + identityUsername + `","name":"` + username + `","onBehalfOf":"","region":"region","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"client_id","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.AuthenticatorObjectContextProvider: userMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", identityUsername, string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, userMockContextProvider)
	})

	t.Run("success for the request parsed as JWT flow when both normal user is present and custom authenticator are present but no authenticator matches", func(t *testing.T) {
		keys := tenantmapping.KeysExtra{
			TenantKey:         "consumerTenant",
			ExternalTenantKey: "externalTenant",
		}

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
			Region:       region,
			ConsumerType: "Static User",
		}

		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper.JWTAuthFlow) + `","name":"` + username + `","onBehalfOf":"","region":"` + region + `","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetails, nil).Once()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetails).Return(objCtxMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.UserObjectContextProvider: userMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", username, string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, userMockContextProvider)
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
			Region:       region,
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"client_id":"` + systemAuthID.String() + `","consumerID":"` + objID.String() + `","consumerType":"Integration System","flow":"` + string(oathkeeper.OAuth2Flow) + `","onBehalfOf":"","region":"` + region + `","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		systemAuthMockContextProvider := getMockContextProvider()
		systemAuthMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &oAuthAuthDetails, nil).Once()
		systemAuthMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, oAuthAuthDetails).Return(objCtx, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.SystemAuthObjectContextProvider: systemAuthMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", systemAuthID.String(), string(oathkeeper.OAuth2Flow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, systemAuthMockContextProvider)
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
			Region:       region,
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + objID.String() + `","consumerType":"Integration System","flow":"` + string(oathkeeper.CertificateFlow) + `","onBehalfOf":"","region":"` + region + `","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ConnectorIssuer + `"],"Client-Id-From-Certificate":["` + systemAuthID.String() + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		systemAuthMockContextProvider := getMockContextProvider()
		systemAuthMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &certAuthDetails, nil).Once()
		systemAuthMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, certAuthDetails).Return(objCtx, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.SystemAuthObjectContextProvider: systemAuthMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", systemAuthID.String(), string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, systemAuthMockContextProvider)
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
			Region:       region,
			ConsumerType: consumer.Runtime,
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper.CertificateFlow) + `","onBehalfOf":"","region":"` + region + `","scope":"` + "" + `","tenant":"{\\\"consumerTenant\\\":\\\"` + externalTenantID + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		certServiceMockContextProvider := getMockContextProvider()
		certServiceMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &externalCertAuthDetails, nil).Once()
		certServiceMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, externalCertAuthDetails).Return(objCtx, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.CertServiceObjectContextProvider: certServiceMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, certServiceMockContextProvider)
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
			Region:       region,
			ConsumerType: "Integration System",
		}
		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + objID.String() + `","consumerType":"Integration System","flow":"` + string(oathkeeper.OneTimeTokenFlow) + `","onBehalfOf":"","region":"` + region + `","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":{"Client-Id-From-Token":["` + systemAuthID.String() + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		systemAuthMockContextProvider := getMockContextProvider()
		systemAuthMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &oneTimeTokenAuthDetails, nil).Once()
		systemAuthMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, oneTimeTokenAuthDetails).Return(objCtx, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.SystemAuthObjectContextProvider: systemAuthMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", systemAuthID.String(), string(oathkeeper.OneTimeTokenFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, systemAuthMockContextProvider)
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
			Region:          region,
			ConsumerType:    consumer.Runtime,
			ContextProvider: tenantmappingconsts.CertServiceObjectContextProvider,
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
			Region:          region,
			OauthClientID:   "client_id",
			ConsumerID:      username,
			AuthFlow:        oathkeeper.JWTAuthFlow,
			ConsumerType:    "Static User",
			ContextProvider: tenantmappingconsts.AuthenticatorObjectContextProvider,
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

		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper.CertificateFlow) + `","identity":"` + username + `","onBehalfOf":"admin","region":"` + region + `","scope":"test","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"client_id","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		authenticatorMockContextProvider := getMockContextProvider()
		authenticatorMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		authenticatorMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(JWTObjCtxMock, nil).Once()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.CertServiceObjectContextProvider:   certServiceMockContextProvider,
			tenantmappingconsts.AuthenticatorObjectContextProvider: authenticatorMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, gjson.Get(expectedRespPayload, "subject"), gjson.Get(strings.TrimSpace(string(body)), "subject"))
		require.Equal(t, gjson.Get(expectedRespPayload, "extra"), gjson.Get(strings.TrimSpace(string(body)), "extra"))
		require.Equal(t, gjson.Get(expectedRespPayload, "header"), gjson.Get(strings.TrimSpace(string(body)), "header"))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, certServiceMockContextProvider)
	})

	t.Run("success for the request parsed as Certificate flow for External issuer and JWT flow with custom authenticator: scopes with mergeStrategy merge: merge scopes", func(t *testing.T) {
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
			Scopes:              "test1 test2 test3",
			ScopesMergeStrategy: mergeWithOtherScopes,
			KeysExtra:           certKeys,
			ConsumerID:          externalTenantID,
			AuthFlow:            oathkeeper.CertificateFlow,
			Region:              region,
			ConsumerType:        consumer.Runtime,
			ContextProvider:     tenantmappingconsts.CertServiceObjectContextProvider,
		}

		certServiceMockContextProvider := getMockContextProvider()
		certServiceMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &externalCertAuthDetails, nil).Once()
		certServiceMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, externalCertAuthDetails).Return(certObjCtx, nil).Once()

		JWTObjCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			KeysExtra:           JWTKeys,
			Scopes:              scopes,
			ScopesMergeStrategy: mergeWithOtherScopes,
			Region:              region,
			OauthClientID:       "client_id",
			ConsumerID:          username,
			AuthFlow:            oathkeeper.JWTAuthFlow,
			ConsumerType:        "Static User",
			ContextProvider:     tenantmappingconsts.AuthenticatorObjectContextProvider,
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

		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper.CertificateFlow) + `","identity":"` + username + `","onBehalfOf":"admin","region":"` + region + `","scope":"application:read test test1 test2 test3","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"client_id","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		authenticatorMockContextProvider := getMockContextProvider()
		authenticatorMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		authenticatorMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(JWTObjCtxMock, nil).Once()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.CertServiceObjectContextProvider:   certServiceMockContextProvider,
			tenantmappingconsts.AuthenticatorObjectContextProvider: authenticatorMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assertExtra(t, expectedRespPayload, string(body))
		require.Equal(t, gjson.Get(expectedRespPayload, "subject"), gjson.Get(strings.TrimSpace(string(body)), "subject"))
		require.Equal(t, gjson.Get(expectedRespPayload, "header"), gjson.Get(strings.TrimSpace(string(body)), "header"))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, certServiceMockContextProvider)
	})

	t.Run("success for the request parsed as Certificate flow for External issuer and JWT flow with custom authenticator: scopes with mergeStrategy merge: deduplicate scopes", func(t *testing.T) {
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
			Scopes:              "test test1 test2 test3",
			ScopesMergeStrategy: mergeWithOtherScopes,
			KeysExtra:           certKeys,
			ConsumerID:          externalTenantID,
			AuthFlow:            oathkeeper.CertificateFlow,
			Region:              region,
			ConsumerType:        consumer.Runtime,
			ContextProvider:     tenantmappingconsts.CertServiceObjectContextProvider,
		}

		certServiceMockContextProvider := getMockContextProvider()
		certServiceMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &externalCertAuthDetails, nil).Once()
		certServiceMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, externalCertAuthDetails).Return(certObjCtx, nil).Once()

		JWTObjCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			KeysExtra:           JWTKeys,
			Scopes:              scopes,
			ScopesMergeStrategy: mergeWithOtherScopes,
			Region:              region,
			OauthClientID:       "client_id",
			ConsumerID:          username,
			AuthFlow:            oathkeeper.JWTAuthFlow,
			ConsumerType:        "Static User",
			ContextProvider:     tenantmappingconsts.AuthenticatorObjectContextProvider,
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

		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper.CertificateFlow) + `","identity":"` + username + `","onBehalfOf":"admin","region":"` + region + `","scope":"application:read test test1 test2 test3","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"client_id","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		authenticatorMockContextProvider := getMockContextProvider()
		authenticatorMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		authenticatorMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(JWTObjCtxMock, nil).Once()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.CertServiceObjectContextProvider:   certServiceMockContextProvider,
			tenantmappingconsts.AuthenticatorObjectContextProvider: authenticatorMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assertExtra(t, expectedRespPayload, string(body))
		require.Equal(t, gjson.Get(expectedRespPayload, "subject"), gjson.Get(strings.TrimSpace(string(body)), "subject"))
		require.Equal(t, gjson.Get(expectedRespPayload, "header"), gjson.Get(strings.TrimSpace(string(body)), "header"))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, certServiceMockContextProvider)
	})

	t.Run("success for the request parsed as Certificate flow for External issuer and JWT flow with custom authenticator: scopes with mergeStrategy merge and intersect", func(t *testing.T) {
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
			Scopes:              "test test1 test2 test3",
			ScopesMergeStrategy: intersectWithOtherScopes,
			KeysExtra:           certKeys,
			ConsumerID:          externalTenantID,
			AuthFlow:            oathkeeper.CertificateFlow,
			Region:              region,
			ConsumerType:        consumer.Runtime,
			ContextProvider:     tenantmappingconsts.CertServiceObjectContextProvider,
		}

		certServiceMockContextProvider := getMockContextProvider()
		certServiceMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &externalCertAuthDetails, nil).Once()
		certServiceMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, externalCertAuthDetails).Return(certObjCtx, nil).Once()

		JWTObjCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			KeysExtra:           JWTKeys,
			Scopes:              scopes,
			ScopesMergeStrategy: mergeWithOtherScopes,
			Region:              region,
			OauthClientID:       "client_id",
			ConsumerID:          username,
			AuthFlow:            oathkeeper.JWTAuthFlow,
			ConsumerType:        "Static User",
			ContextProvider:     tenantmappingconsts.AuthenticatorObjectContextProvider,
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

		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper.CertificateFlow) + `","identity":"` + username + `","onBehalfOf":"admin","region":"` + region + `","scope":"test","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"client_id","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		authenticatorMockContextProvider := getMockContextProvider()
		authenticatorMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		authenticatorMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(JWTObjCtxMock, nil).Once()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.CertServiceObjectContextProvider:   certServiceMockContextProvider,
			tenantmappingconsts.AuthenticatorObjectContextProvider: authenticatorMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assertExtra(t, expectedRespPayload, string(body))
		require.Equal(t, gjson.Get(expectedRespPayload, "subject"), gjson.Get(strings.TrimSpace(string(body)), "subject"))
		require.Equal(t, gjson.Get(expectedRespPayload, "header"), gjson.Get(strings.TrimSpace(string(body)), "header"))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, certServiceMockContextProvider)
	})

	t.Run("success for the request parsed as Certificate flow for External issuer and JWT flow with custom authenticator: scopes with mergeStrategy merge and override", func(t *testing.T) {
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
			Scopes:              "test test1 test2 test3",
			ScopesMergeStrategy: overrideAllScopes,
			KeysExtra:           certKeys,
			ConsumerID:          externalTenantID,
			AuthFlow:            oathkeeper.CertificateFlow,
			Region:              region,
			ConsumerType:        consumer.Runtime,
			ContextProvider:     tenantmappingconsts.CertServiceObjectContextProvider,
		}

		certServiceMockContextProvider := getMockContextProvider()
		certServiceMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &externalCertAuthDetails, nil).Once()
		certServiceMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, externalCertAuthDetails).Return(certObjCtx, nil).Once()

		JWTObjCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			KeysExtra:           JWTKeys,
			Scopes:              scopes,
			ScopesMergeStrategy: mergeWithOtherScopes,
			Region:              region,
			OauthClientID:       "client_id",
			ConsumerID:          username,
			AuthFlow:            oathkeeper.JWTAuthFlow,
			ConsumerType:        "Static User",
			ContextProvider:     tenantmappingconsts.AuthenticatorObjectContextProvider,
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

		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper.CertificateFlow) + `","identity":"` + username + `","onBehalfOf":"admin","region":"` + region + `","scope":"test test1 test2 test3","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"client_id","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		authenticatorMockContextProvider := getMockContextProvider()
		authenticatorMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		authenticatorMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(JWTObjCtxMock, nil).Once()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.CertServiceObjectContextProvider:   certServiceMockContextProvider,
			tenantmappingconsts.AuthenticatorObjectContextProvider: authenticatorMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assertExtra(t, expectedRespPayload, string(body))
		require.Equal(t, gjson.Get(expectedRespPayload, "subject"), gjson.Get(strings.TrimSpace(string(body)), "subject"))
		require.Equal(t, gjson.Get(expectedRespPayload, "header"), gjson.Get(strings.TrimSpace(string(body)), "header"))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, certServiceMockContextProvider)
	})

	t.Run("success for the request parsed as Certificate flow for External issuer and JWT flow with custom authenticator: scopes with mergeStrategy intersect and override", func(t *testing.T) {
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
			Scopes:              "test test1 test2 test3",
			ScopesMergeStrategy: overrideAllScopes,
			KeysExtra:           certKeys,
			ConsumerID:          externalTenantID,
			AuthFlow:            oathkeeper.CertificateFlow,
			Region:              region,
			ConsumerType:        consumer.Runtime,
			ContextProvider:     tenantmappingconsts.CertServiceObjectContextProvider,
		}

		certServiceMockContextProvider := getMockContextProvider()
		certServiceMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &externalCertAuthDetails, nil).Once()
		certServiceMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, externalCertAuthDetails).Return(certObjCtx, nil).Once()

		JWTObjCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         tenantID.String(),
			},
			KeysExtra:           JWTKeys,
			Scopes:              scopes,
			ScopesMergeStrategy: intersectWithOtherScopes,
			Region:              region,
			OauthClientID:       "client_id",
			ConsumerID:          username,
			AuthFlow:            oathkeeper.JWTAuthFlow,
			ConsumerType:        "Static User",
			ContextProvider:     tenantmappingconsts.AuthenticatorObjectContextProvider,
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

		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper.CertificateFlow) + `","identity":"` + username + `","onBehalfOf":"admin","region":"` + region + `","scope":"test test1 test2 test3","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"client_id","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		authenticatorMockContextProvider := getMockContextProvider()
		authenticatorMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		authenticatorMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(JWTObjCtxMock, nil).Once()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.CertServiceObjectContextProvider:   certServiceMockContextProvider,
			tenantmappingconsts.AuthenticatorObjectContextProvider: authenticatorMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		assertExtra(t, expectedRespPayload, string(body))
		require.Equal(t, gjson.Get(expectedRespPayload, "subject"), gjson.Get(strings.TrimSpace(string(body)), "subject"))
		require.Equal(t, gjson.Get(expectedRespPayload, "header"), gjson.Get(strings.TrimSpace(string(body)), "header"))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, certServiceMockContextProvider)
	})

	t.Run("success for the request parsed as two Certificate flow for External issuer object context matched (one is without region)", func(t *testing.T) {
		certKeys := tenantmapping.KeysExtra{
			TenantKey:         "providerTenant",
			ExternalTenantKey: "providerExternalTenant",
		}

		uniqueAttributeKey := "uniqueAttribute"
		uniqueAttributeValue := "uniqueAttributeValue"
		identityAttributeKey := "identity"
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
			Region:          region,
			ConsumerType:    consumer.Runtime,
			ContextProvider: tenantmappingconsts.CertServiceObjectContextProvider,
		}

		certObjCtxWithoutRegion := certObjCtx
		certObjCtxWithoutRegion.Region = ""

		certServiceMockContextProvider := getMockContextProvider()
		certServiceMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &externalCertAuthDetails, nil).Twice()
		certServiceMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, externalCertAuthDetails).Return(certObjCtx, nil).Once()
		certServiceMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, externalCertAuthDetails).Return(certObjCtxWithoutRegion, nil).Once()

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

		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper.CertificateFlow) + `","identity":"` + username + `","onBehalfOf":"","region":"` + region + `","scope":"test","tenant":"{\\\"consumerTenant\\\":\\\"` + externalTenantID + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.CertServiceObjectContextProvider:  certServiceMockContextProvider,
			tenantmappingconsts.TenantHeaderObjectContextProvider: certServiceMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, gjson.Get(expectedRespPayload, "subject"), gjson.Get(strings.TrimSpace(string(body)), "subject"))
		require.Equal(t, gjson.Get(expectedRespPayload, "extra"), gjson.Get(strings.TrimSpace(string(body)), "extra"))
		require.Equal(t, gjson.Get(expectedRespPayload, "header"), gjson.Get(strings.TrimSpace(string(body)), "header"))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, certServiceMockContextProvider)
	})

	t.Run("Certificate flow for External issuer and JWT flow with custom authenticator on behalf of non-existing tenant and region doesn't result in nullified hydrator request data extra field", func(t *testing.T) {
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
			Region:          region,
			ConsumerType:    consumer.Runtime,
			ContextProvider: tenantmappingconsts.CertServiceObjectContextProvider,
		}

		certServiceMockContextProvider := getMockContextProvider()
		certServiceMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &externalCertAuthDetails, nil).Once()
		certServiceMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, externalCertAuthDetails).Return(certObjCtx, nil).Once()

		JWTObjCtxMock := tenantmapping.ObjectContext{
			TenantContext: tenantmapping.TenantContext{
				ExternalTenantID: externalTenantID,
				TenantID:         "",
			},
			KeysExtra:       JWTKeys,
			Scopes:          scopes,
			Region:          "",
			OauthClientID:   "client_id",
			ConsumerID:      username,
			AuthFlow:        oathkeeper.JWTAuthFlow,
			ConsumerType:    "Static User",
			ContextProvider: tenantmappingconsts.AuthenticatorObjectContextProvider,
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

		expectedRespPayload := `{"subject":"","extra":{"authenticator_coordinates":{"name":"` + authn[0].Name + `","index":0},"consumerID":"` + externalTenantID + `","consumerType":"Runtime","flow":"` + string(oathkeeper.CertificateFlow) + `","identity":"` + username + `","onBehalfOf":"admin","region":"` + region + `","scope":"test","tenant":"{\\\"consumerTenant\\\":\\\"\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerExternalTenant\\\":\\\"` + externalTenantID + `\\\",\\\"providerTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":"client_id","` + uniqueAttributeKey + `":"` + uniqueAttributeValue + `"},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		authenticatorMockContextProvider := getMockContextProvider()
		authenticatorMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		authenticatorMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(JWTObjCtxMock, nil).Once()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.CertServiceObjectContextProvider:   certServiceMockContextProvider,
			tenantmappingconsts.AuthenticatorObjectContextProvider: authenticatorMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, gjson.Get(expectedRespPayload, "subject"), gjson.Get(strings.TrimSpace(string(body)), "subject"))
		require.Equal(t, gjson.Get(expectedRespPayload, "extra"), gjson.Get(strings.TrimSpace(string(body)), "extra"))
		require.Equal(t, gjson.Get(expectedRespPayload, "header"), gjson.Get(strings.TrimSpace(string(body)), "header"))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, certServiceMockContextProvider)
	})

	t.Run("error when sending different HTTP verb than POST", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodGet, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		handler := tenantmapping.NewHandler(nil, nil, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusBadRequest, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, "Bad request method. Got GET, expected POST", strings.TrimSpace(string(body)))
	})

	t.Run("empty response payload when object context provider fails to provide object context", func(t *testing.T) {
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

		expectedError := errors.New("error")
		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetails, nil).Once()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetails).Return(tenantmapping.ObjectContext{}, expectedError).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.UserObjectContextProvider: userMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", systemAuthID.String(), string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, userMockContextProvider)
	})

	t.Run("empty payload without region when matched object context provider doesn't have region value'", func(t *testing.T) {
		scopes := "application:read"

		keys := tenantmapping.KeysExtra{
			TenantKey:         "consumerTenant",
			ExternalTenantKey: "externalTenant",
		}

		objCtxMockWithoutRegion := tenantmapping.ObjectContext{
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

		reqDataMock := oathkeeper.ReqData{
			Body: oathkeeper.ReqBody{
				Extra: map[string]interface{}{
					oathkeeper.UsernameKey: username,
				},
			},
		}

		expectedRespPayload := `{"subject":"","extra":{"consumerID":"` + username + `","consumerType":"Static User","flow":"` + string(oathkeeper.JWTAuthFlow) + `","name":"` + username + `","onBehalfOf":"","region":"","scope":"` + scopes + `","tenant":"{\\\"consumerTenant\\\":\\\"` + tenantID.String() + `\\\",\\\"externalTenant\\\":\\\"` + externalTenantID + `\\\"}","tokenClientID":""},"header":null}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		userMockContextProvider := getMockContextProvider()
		userMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetails, nil).Once()
		userMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetails).Return(objCtxMockWithoutRegion, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.UserObjectContextProvider: userMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", username, string(oathkeeper.JWTAuthFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, userMockContextProvider)
	})

	t.Run("Certificate flow for External issuer and JWT flow with custom authenticator results in nullified hydrator request data extra field when matched object context providers don't have same region values", func(t *testing.T) {
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
		region2 := "eu-2"
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
			Region:          region,
			ConsumerType:    consumer.Runtime,
			ContextProvider: tenantmappingconsts.CertServiceObjectContextProvider,
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
			Region:          region2,
			OauthClientID:   "client_id",
			ConsumerID:      username,
			AuthFlow:        oathkeeper.JWTAuthFlow,
			ConsumerType:    "Static User",
			ContextProvider: tenantmappingconsts.AuthenticatorObjectContextProvider,
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

		expectedRespPayload := `{"subject":"","extra":{},"header":{"Client-Certificate-Issuer":["` + oathkeeper.ExternalIssuer + `"],"Client-Id-From-Certificate":["` + externalTenantID + `"]}}`

		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		authenticatorMockContextProvider := getMockContextProvider()
		authenticatorMockContextProvider.On("Match", mock.Anything, reqDataMock).Return(true, &jwtAuthDetailsWithAuthenticator, nil).Once()
		authenticatorMockContextProvider.On("GetObjectContext", mock.Anything, reqDataMock, jwtAuthDetailsWithAuthenticator).Return(JWTObjCtxMock, nil).Once()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(reqDataMock, nil).Once()

		objectContextProviders := map[string]tenantmapping.ObjectContextProvider{
			tenantmappingconsts.CertServiceObjectContextProvider:   certServiceMockContextProvider,
			tenantmappingconsts.AuthenticatorObjectContextProvider: authenticatorMockContextProvider,
		}

		clientInstrumenter := &automock.ClientInstrumenter{}
		clientInstrumenter.On("InstrumentClient", externalTenantID, string(oathkeeper.CertificateFlow), mock.Anything)

		handler := tenantmapping.NewHandler(reqDataParserMock, objectContextProviders, clientInstrumenter)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)
		body, err := ioutil.ReadAll(resp.Body)
		require.NoError(t, err)

		require.Equal(t, expectedRespPayload, strings.TrimSpace(string(body)))

		mock.AssertExpectationsForObjects(t, reqDataParserMock, certServiceMockContextProvider)
	})

	t.Run("error when body parser returns error", func(t *testing.T) {
		req := httptest.NewRequest(http.MethodPost, target, strings.NewReader(""))
		w := httptest.NewRecorder()

		reqDataParserMock := &automock.ReqDataParser{}
		reqDataParserMock.On("Parse", mock.Anything).Return(oathkeeper.ReqData{}, errors.New("some error")).Once()

		handler := tenantmapping.NewHandler(reqDataParserMock, nil, nil)
		handler.ServeHTTP(w, req)

		resp := w.Result()
		require.Equal(t, http.StatusOK, resp.StatusCode)

		out := oathkeeper.ReqData{}
		err := json.NewDecoder(resp.Body).Decode(&out)
		require.NoError(t, err)

		assert.Equal(t, oathkeeper.ReqData{}, out)

		mock.AssertExpectationsForObjects(t, reqDataParserMock)
	})
}

func getMockContextProvider() *automock.ObjectContextProvider {
	provider := &automock.ObjectContextProvider{}
	return provider
}

func assertExtra(t *testing.T, expectedBody, actualBody string) {
	expectedExtra := gjson.Get(expectedBody, "extra").Map()
	actualExtra := gjson.Get(strings.TrimSpace(actualBody), "extra").Map()
	for k, v := range expectedExtra {
		if k != "scope" {
			require.Equal(t, v.Str, actualExtra[k].Str)
		}
	}

	expectedScopes := strings.Split(expectedExtra["scope"].String(), " ")
	actualScopes := strings.Split(actualExtra["scope"].String(), " ")
	require.ElementsMatch(t, expectedScopes, actualScopes)
}
