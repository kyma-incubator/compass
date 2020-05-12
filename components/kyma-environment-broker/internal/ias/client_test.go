package ias

import (
	"encoding/base64"
	"encoding/json"
	"fmt"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

const (
	companyID         = "sandbox"
	serviceProviderID = "5e8c5e9e5f25d83ebec7e89e"
	clientID          = "42"
	clientSecret      = "floda"
	userForRest       = "22b13c44-a1ae-41a5-b549-0649c4a5bd25"
)

func TestClient_GetCompany(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	// when
	company, err := client.GetCompany()

	// then
	assert.NoError(t, err)
	assert.Len(t, company.ServiceProviders, 1)
	assert.Len(t, company.IdentityProviders, 1)
}

func TestClient_CreateServiceProvider(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	// when
	err := client.CreateServiceProvider("someName", companyID)

	// then
	assert.NoError(t, err)

	response, err := server.Client().Get(fmt.Sprintf("%s/getSP", server.URL))
	assert.NoError(t, err)

	body, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.Equal(t, "someName", string(body))
}

func TestClient_SetOIDCConfiguration(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	// when
	iasType := OIDCType{
		SsoType:             "openID",
		ServiceProviderName: "example.com",
		OpenIDConnectConfig: OpenIDConnectConfig{
			RedirectURIs:           []string{"https://example.com"},
			PostLogoutRedirectURIs: nil,
		},
	}
	err := client.SetOIDCConfiguration(serviceProviderID, iasType)

	// then
	assert.NoError(t, err)

	response, err := server.Client().Get(fmt.Sprintf("%s/get", server.URL))
	assert.NoError(t, err)

	var conf OIDCType
	err = json.NewDecoder(response.Body).Decode(&conf)
	assert.NoError(t, err)

	assert.Equal(t, "openID", conf.SsoType)
	assert.Equal(t, "example.com", conf.ServiceProviderName)
	assert.Equal(t, "https://example.com", conf.OpenIDConnectConfig.RedirectURIs[0])
}

func TestClient_SetSAMLConfiguration(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	// when
	iasType := SAMLType{
		ServiceProviderName: "example.com",
		ACSEndpoints: []ACSEndpoint{
			{
				Location:  "https://example.com",
				Index:     0,
				IsDefault: true,
			},
		},
	}
	err := client.SetSAMLConfiguration(serviceProviderID, iasType)

	// then
	assert.NoError(t, err)

	response, err := server.Client().Get(fmt.Sprintf("%s/get", server.URL))
	assert.NoError(t, err)

	var conf SAMLType
	err = json.NewDecoder(response.Body).Decode(&conf)
	assert.NoError(t, err)

	assert.Equal(t, "example.com", conf.ServiceProviderName)
	assert.Equal(t, "https://example.com", conf.ACSEndpoints[0].Location)
	assert.Equal(t, int32(0), conf.ACSEndpoints[0].Index)
	assert.Equal(t, true, conf.ACSEndpoints[0].IsDefault)
}

func TestClient_SetAssertionAttribute(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	// when
	attributes := PostAssertionAttributes{
		AssertionAttributes: []AssertionAttribute{
			{
				AssertionAttribute: "first_name",
				UserAttribute:      "firstName",
			},
		},
	}
	err := client.SetAssertionAttribute(serviceProviderID, attributes)

	// then
	assert.NoError(t, err)

	response, err := server.Client().Get(fmt.Sprintf("%s/get", server.URL))
	assert.NoError(t, err)

	var conf PostAssertionAttributes
	err = json.NewDecoder(response.Body).Decode(&conf)
	assert.NoError(t, err)

	assert.Equal(t, "first_name", conf.AssertionAttributes[0].AssertionAttribute)
	assert.Equal(t, "firstName", conf.AssertionAttributes[0].UserAttribute)
}

func TestClient_SetSubjectNameIdentifier(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	// when
	sni := SubjectNameIdentifier{
		NameIDAttribute: "email",
	}
	err := client.SetSubjectNameIdentifier(serviceProviderID, sni)

	// then
	assert.NoError(t, err)

	response, err := server.Client().Get(fmt.Sprintf("%s/get", server.URL))
	assert.NoError(t, err)

	var conf SubjectNameIdentifier
	err = json.NewDecoder(response.Body).Decode(&conf)
	assert.NoError(t, err)

	assert.Equal(t, "email", conf.NameIDAttribute)
}

func TestClient_SetAuthenticationAndAccess(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	// when
	auth := AuthenticationAndAccess{
		ServiceProviderAccess: ServiceProviderAccess{
			RBAConfig: RBAConfig{
				RBARules: []RBARules{
					{
						Action:    "Allow",
						Group:     "admins",
						GroupType: "cloud",
					},
				},
				DefaultAction: "Allow",
			},
		},
	}
	err := client.SetAuthenticationAndAccess(serviceProviderID, auth)

	// then
	assert.NoError(t, err)

	response, err := server.Client().Get(fmt.Sprintf("%s/get", server.URL))
	assert.NoError(t, err)

	var conf AuthenticationAndAccess
	err = json.NewDecoder(response.Body).Decode(&conf)
	assert.NoError(t, err)

	assert.Equal(t, "Allow", conf.ServiceProviderAccess.RBAConfig.DefaultAction)
	assert.Equal(t, "Allow", conf.ServiceProviderAccess.RBAConfig.RBARules[0].Action)
	assert.Equal(t, "admins", conf.ServiceProviderAccess.RBAConfig.RBARules[0].Group)
	assert.Equal(t, "cloud", conf.ServiceProviderAccess.RBAConfig.RBARules[0].GroupType)
}

func TestClient_SetDefaultAuthenticatingIDP(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	// when
	authIDP := DefaultAuthIDPConfig{
		Organization:   companyID,
		ID:             serviceProviderID,
		DefaultAuthIDP: "http://example.com",
	}
	err := client.SetDefaultAuthenticatingIDP(authIDP)

	// then
	assert.NoError(t, err)
	response, err := server.Client().Get(fmt.Sprintf("%s/get", server.URL))
	assert.NoError(t, err)

	var conf DefaultAuthIDPConfig
	err = json.NewDecoder(response.Body).Decode(&conf)
	assert.NoError(t, err)

	assert.Equal(t, companyID, conf.Organization)
	assert.Equal(t, serviceProviderID, conf.ID)
	assert.Equal(t, "http://example.com", conf.DefaultAuthIDP)
}

func TestClient_GenerateServiceProviderSecret(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	// when
	sc := SecretConfiguration{
		Organization: companyID,
		ID:           serviceProviderID,
		RestAPIClientSecret: RestAPIClientSecret{
			Description: "test",
			Scopes:      []string{"OAuth"},
		},
	}
	secret, err := client.GenerateServiceProviderSecret(sc)

	// then
	assert.NoError(t, err)
	assert.Equal(t, clientID, secret.ClientID)
	assert.Equal(t, clientSecret, secret.ClientSecret)
}

func TestClient_DeleteServiceProvider(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	err := client.CreateServiceProvider(serviceProviderID, companyID)
	assert.NoError(t, err)

	// when
	err = client.DeleteServiceProvider(serviceProviderID)

	// then
	assert.NoError(t, err)

	response, err := server.Client().Get(fmt.Sprintf("%s/getSP", server.URL))
	assert.NoError(t, err)

	body, err := ioutil.ReadAll(response.Body)
	assert.NoError(t, err)
	assert.Equal(t, "", string(body))

	// when
	err = client.DeleteServiceProvider(serviceProviderID)

	// then
	assert.NoError(t, err)
}

func TestClient_DeleteSecret(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	for i := 0; i < 3; i++ {
		sc := SecretConfiguration{
			Organization: companyID,
			ID:           serviceProviderID,
			RestAPIClientSecret: RestAPIClientSecret{
				Description: "test",
				Scopes:      []string{"OAuth"},
			},
		}
		_, err := client.GenerateServiceProviderSecret(sc)
		assert.NoError(t, err)
	}

	// when
	err := client.DeleteSecret(DeleteSecrets{
		ClientID:         userForRest,
		ClientSecretsIDs: []string{fmt.Sprintf("%s-next", clientID)},
	})

	// then
	assert.NoError(t, err)

	response, err := server.Client().Get(fmt.Sprintf("%s/getS", server.URL))
	assert.NoError(t, err)

	var secrets []ServiceProviderSecret
	err = json.NewDecoder(response.Body).Decode(&secrets)
	assert.NoError(t, err)

	assert.Len(t, secrets, 2)
}

var companies = `{
	"company_id" : "global",
	"default_sso_domain" : "https://sandbox.accounts400.ondemand.com/service/idp/5e46ar7cc92eec206b93893b",
	"service_providers" : [{
		"id" : "50c1bb7ce4b01ab0481c49a3",
		"name" : "oac.accounts.sap.com",
		"group" : "system",
		"uri" : "https://sandbox.accounts400.ondemand.com/service/sps/50c1bb7ce5b01av0481c49a3",
		"active_users" : "-1",
		"self_registration" : "allowed",
		"company_id" : "global"
	}],
	"certificates_counter_exceeded" : false,
	"identity_providers" : [{
		"id" : "5e46ad8cc92edc106b93893b",
		"display_name" : "SAP Cloud Platform Identity Authentication",
		"name" : "https://sandbox.accounts400.ondemand.com",
		"uri" : "https://sandbox.accounts400.ondemand.com/service/idp/5e46ad7cc92eec106b93893b",
		"alias" : "sandbox"
	}]
}`

type server struct {
	t *testing.T

	serviceProvider []byte
	configuration   []byte
	secrets         []ServiceProviderSecret
}

func fixHTTPServer(t *testing.T) *httptest.Server {
	s := server{t: t}

	r := mux.NewRouter()
	r.HandleFunc("/service/company/global", s.authorized(s.companies)).Methods(http.MethodGet)
	r.HandleFunc("/service/sps", s.authorized(s.createSP)).Methods(http.MethodPost)
	r.HandleFunc("/service/sps", s.authorized(s.configureSPBody)).Methods(http.MethodPut)
	r.HandleFunc("/service/sps/delete", s.authorized(s.deleteSP)).Methods(http.MethodPut)
	r.HandleFunc("/service/sps/{spID}", s.authorized(s.configureSP)).Methods(http.MethodPut)
	r.HandleFunc("/service/sps/{spID}/rba", s.authorized(s.configureSP)).Methods(http.MethodPut)
	r.HandleFunc("/service/sps/clientSecret", s.authorized(s.deleteSecrets)).Methods(http.MethodDelete)

	r.HandleFunc("/getSP", s.getServiceProvider).Methods(http.MethodGet)
	r.HandleFunc("/getS", s.getSecrets).Methods(http.MethodGet)
	r.HandleFunc("/get", s.getConfiguration).Methods(http.MethodGet)

	return httptest.NewServer(r)
}

func (s *server) authorized(pass func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
	return func(w http.ResponseWriter, r *http.Request) {
		auth := strings.SplitN(r.Header.Get("Authorization"), " ", 2)

		if len(auth) != 2 || auth[0] != "Basic" {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}

		payload, _ := base64.StdEncoding.DecodeString(auth[1])
		pair := strings.SplitN(string(payload), ":", 2)

		if len(pair) != 2 || !(pair[0] == "admin" && pair[1] == "admin123") {
			w.WriteHeader(http.StatusUnauthorized)
			return
		}
		pass(w, r)
	}
}

func (s *server) companies(w http.ResponseWriter, _ *http.Request) {
	_, err := w.Write([]byte(companies))
	if err != nil {
		s.t.Errorf("test server cannot write response body: %s", err)
		w.WriteHeader(http.StatusInternalServerError)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *server) createSP(w http.ResponseWriter, r *http.Request) {
	if err := r.ParseForm(); err != nil {
		s.t.Errorf("cannot parse form: %s", err)
		return
	}
	if r.FormValue("company_id") != companyID {
		w.WriteHeader(http.StatusForbidden)
	}

	s.serviceProvider = []byte(r.FormValue("sp_name"))
	w.WriteHeader(http.StatusCreated)
}

func (s *server) configureSP(w http.ResponseWriter, r *http.Request) {
	val, ok := mux.Vars(r)["spID"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if val != serviceProviderID {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.t.Errorf("test server cannot read request body: %s", err)
		return
	}

	s.configuration = body
	w.WriteHeader(http.StatusOK)
}

func (s *server) configureSPBody(w http.ResponseWriter, r *http.Request) {
	var sc SecretConfiguration
	var authIDP DefaultAuthIDPConfig
	body, err := ioutil.ReadAll(r.Body)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.t.Errorf("test server cannot read request body: %s", err)
		return
	}

	err = json.Unmarshal(body, &authIDP)
	err2 := json.Unmarshal(body, &sc)
	if err != nil || err2 != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.t.Errorf("test server cannot unmarshal request body: (%s;%s)", err, err2)
		return
	}

	if authIDP.ID != serviceProviderID || authIDP.Organization != companyID {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	if authIDP.DefaultAuthIDP != "" {
		s.configuration = body
		w.WriteHeader(http.StatusOK)
	} else if sc.RestAPIClientSecret.Description != "" {

		secret := ServiceProviderSecret{
			ClientID:     s.generateSecretID(),
			ClientSecret: clientSecret,
		}
		rawSecret, err := json.Marshal(secret)
		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.t.Errorf("test server cannot marshal secret struct: %s", err)
			return
		}
		s.secrets = append(s.secrets, secret)

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(rawSecret)

		if err != nil {
			w.WriteHeader(http.StatusInternalServerError)
			s.t.Errorf("test server cannot write response body: %s", err)
			return
		}
	}
}

func (s *server) generateSecretID() string {
	if len(s.secrets) == 0 {
		return clientID
	}
	return fmt.Sprintf("%s-next", s.secrets[len(s.secrets)-1].ClientID)
}

func (s *server) deleteSP(w http.ResponseWriter, r *http.Request) {
	keys, ok := r.URL.Query()["sp_id"]
	if !ok {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if keys[0] != serviceProviderID {
		w.WriteHeader(http.StatusNotFound)
		return
	}
	if string(s.serviceProvider) != serviceProviderID {
		w.WriteHeader(http.StatusNotFound)
		return
	}

	s.serviceProvider = []byte{}
	w.WriteHeader(http.StatusOK)
}

func (s *server) deleteSecrets(w http.ResponseWriter, r *http.Request) {
	var deleteSecrets DeleteSecrets
	err := json.NewDecoder(r.Body).Decode(&deleteSecrets)
	if err != nil {
		w.WriteHeader(http.StatusInternalServerError)
		s.t.Errorf("test server cannot decode request body: %s", err)
		return
	}

	if deleteSecrets.ClientID != userForRest {
		w.WriteHeader(http.StatusBadRequest)
		return
	}

	for _, sID := range deleteSecrets.ClientSecretsIDs {
		for index, secret := range s.secrets {
			if secret.ClientID == sID {
				s.secrets[index] = s.secrets[len(s.secrets)-1]
				s.secrets[len(s.secrets)-1] = ServiceProviderSecret{}
				s.secrets = s.secrets[:len(s.secrets)-1]
			}
		}
	}

	s.t.Log(deleteSecrets)
}

func (s *server) getServiceProvider(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(s.serviceProvider)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.t.Errorf("test server cannot write response body: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *server) getConfiguration(w http.ResponseWriter, r *http.Request) {
	_, err := w.Write(s.configuration)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.t.Errorf("test server cannot write response body: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}

func (s *server) getSecrets(w http.ResponseWriter, r *http.Request) {
	data, err := json.Marshal(s.secrets)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.t.Errorf("test server cannot marshal secrets: %s", err)
		return
	}

	_, err = w.Write(data)
	if err != nil {
		w.WriteHeader(http.StatusBadRequest)
		s.t.Errorf("test server cannot write response body: %s", err)
		return
	}
	w.WriteHeader(http.StatusOK)
}
