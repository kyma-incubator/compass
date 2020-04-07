package ias

import (
	"encoding/base64"
	"encoding/json"
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
}

func TestClient_SetType(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	// when
	iasType := Type{
		SsoType:             "openID",
		ServiceProviderName: "example.com",
		OpenIDConnectConfig: OpenIDConnectConfig{
			RedirectURIs:           []string{"https://example.com"},
			PostLogoutRedirectURIs: nil,
		},
	}
	err := client.SetType(serviceProviderID, iasType)

	// then
	assert.NoError(t, err)
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
}

func TestClient_GenerateServiceProviderSecret(t *testing.T) {
	// given
	server := fixHTTPServer(t)
	defer server.Close()

	client := NewClient(server.Client(), ClientConfig{URL: server.URL, ID: "admin", Secret: "admin123"})

	// when
	sc := SecretConfiguration{
		Organization:   companyID,
		ID:             serviceProviderID,
		DefaultAuthIDp: "http://example.com",
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

	// when
	err := client.DeleteServiceProvider(serviceProviderID)

	// then
	assert.NoError(t, err)
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

func fixHTTPServer(t *testing.T) *httptest.Server {
	var authorized = func(pass func(w http.ResponseWriter, r *http.Request)) func(w http.ResponseWriter, r *http.Request) {
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
	var handleConfiguration = func(w http.ResponseWriter, r *http.Request) {
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
			t.Errorf("test server cannot read request body: %s", err)
		}
		if !isJSON(string(body)) {
			w.WriteHeader(http.StatusBadRequest)
			return
		}
		w.WriteHeader(http.StatusOK)
	}

	r := mux.NewRouter()
	r.HandleFunc("/service/company/global", authorized(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		_, err := w.Write([]byte(companies))
		if err != nil {
			t.Errorf("test server cannot write response body: %s", err)
		}
	})).Methods(http.MethodGet)
	r.HandleFunc("/service/sps", authorized(func(w http.ResponseWriter, r *http.Request) {
		if err := r.ParseForm(); err != nil {
			t.Errorf("cannot parse form: %s", err)
			return
		}
		if r.FormValue("company_id") != companyID {
			w.WriteHeader(http.StatusForbidden)
		}
		w.WriteHeader(http.StatusCreated)
	})).Methods(http.MethodPost)
	r.HandleFunc("/service/sps", authorized(func(w http.ResponseWriter, r *http.Request) {
		var sc SecretConfiguration
		err := json.NewDecoder(r.Body).Decode(&sc)
		if err != nil {
			t.Errorf("test server cannot read request body: %s", err)
			return
		}

		if sc.ID != serviceProviderID || sc.Organization != companyID {
			w.WriteHeader(http.StatusNotFound)
			return
		}

		secret := ServiceProviderSecret{
			ClientID:     clientID,
			ClientSecret: clientSecret,
		}
		rawSecret, err := json.Marshal(secret)
		if err != nil {
			t.Errorf("test server cannot marshal secret struct: %s", err)
			return
		}

		w.WriteHeader(http.StatusOK)
		_, err = w.Write(rawSecret)
		if err != nil {
			t.Errorf("test server cannot write response body: %s", err)
			return
		}
	})).Methods(http.MethodPut)
	r.HandleFunc("/service/sps/delete", authorized(func(w http.ResponseWriter, r *http.Request) {
		keys, ok := r.URL.Query()["sp_id"]
		if !ok {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		if keys[0] != serviceProviderID {
			w.WriteHeader(http.StatusNotFound)
			return
		}
		w.WriteHeader(http.StatusOK)
	})).Methods(http.MethodPut)
	r.HandleFunc("/service/sps/{spID}", authorized(handleConfiguration)).Methods(http.MethodPut)
	r.HandleFunc("/service/sps/{spID}/rba", authorized(handleConfiguration)).Methods(http.MethodPut)

	return httptest.NewServer(r)
}

func isJSON(s string) bool {
	var js map[string]interface{}
	return json.Unmarshal([]byte(s), &js) == nil
}
