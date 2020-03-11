package oathkeeper

import (
	"net/http"

	log "github.com/sirupsen/logrus"

	"github.com/dgrijalva/jwt-go"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

const (
	CertificateFlow  AuthFlow = "Certificate"
	OneTimeTokenFlow AuthFlow = "OneTimeToken"
	OAuth2Flow       AuthFlow = "OAuth2"
	JWTAuthFlow      AuthFlow = "JWT"

	ClientIDKey       = "client_id"
	EmailKey          = "email"
	UsernameKey       = "name"
	GroupsKey         = "groups"
	ClientIDCertKey   = "client-id-from-certificate"
	ClientIDTokenKey  = "client-id-from-token"
	ExternalTenantKey = "tenant"
	ScopesKey         = "scope"
)

// AuthFlow wraps possible flows of auth like OAuth2, JWT and certificate
type AuthFlow string

func (f AuthFlow) IsCertFlow() bool {
	return f == CertificateFlow
}

func (f AuthFlow) IsOneTimeTokenFlow() bool {
	return f == OneTimeTokenFlow
}

func (f AuthFlow) IsOAuth2Flow() bool {
	return f == OAuth2Flow
}

func (f AuthFlow) IsJWTFlow() bool {
	return f == JWTAuthFlow
}

// ReqBody represents parsed request input to the handler
type ReqBody struct {
	Subject string                 `json:"subject"`
	Extra   map[string]interface{} `json:"extra"`
	Header  http.Header            `json:"header"`
}

// ReqData represents incomming request with parsed body and its header
type ReqData struct {
	Body   ReqBody
	Header http.Header
}

func NewReqData(reqBody ReqBody, reqHeader http.Header) ReqData {
	if reqBody.Extra == nil {
		reqBody.Extra = make(map[string]interface{})
	}

	if reqBody.Header == nil {
		reqBody.Header = make(http.Header)
	}

	return ReqData{
		Body:   reqBody,
		Header: reqHeader,
	}
}

// GetAuthID looks for auth ID and identifies auth flow in the parsed request input represented by the ReqData struct
func (d *ReqData) GetAuthID() (string, AuthFlow, error) {
	if idVal, ok := d.Body.Extra[ClientIDKey]; ok {
		authID, err := str.Cast(idVal)
		if err != nil {
			return "", "", errors.Wrapf(err, "while parsing the value for %s", ClientIDKey)
		}

		return authID, OAuth2Flow, nil
	}

	if idVal := d.Body.Header.Get(ClientIDCertKey); idVal != "" {
		return idVal, CertificateFlow, nil
	}

	if idVal := d.Body.Header.Get(ClientIDTokenKey); idVal != "" {
		return idVal, OneTimeTokenFlow, nil
	}

	if usernameVal, ok := d.Body.Extra[UsernameKey]; ok {
		username, err := str.Cast(usernameVal)
		if err != nil {
			return "", "", errors.Wrapf(err, "while parsing the value for %s", UsernameKey)
		}
		return username, JWTAuthFlow, nil
	}

	return "", "", errors.New("unable to find valid auth ID")
}

// GetExternalTenantID returns external tenant ID from the parsed request input if it is defined
func (d *ReqData) GetExternalTenantID() (string, error) {
	if tenantVal := d.Body.Header.Get(ExternalTenantKey); tenantVal != "" {
		return tenantVal, nil
	}

	if tenantVal, ok := d.Body.Extra[ExternalTenantKey]; ok {
		tenant, err := str.Cast(tenantVal)
		if err != nil {
			return "", errors.Wrapf(err, "while parsing the value for %s", ExternalTenantKey)
		}

		return tenant, nil
	}

	if tenantVal := d.Header.Get(ExternalTenantKey); tenantVal != "" {
		return tenantVal, nil
	}

	return "", apperrors.NewKeyDoesNotExistError(ExternalTenantKey)
}

// GetScopes returns scopes from the parsed request input if defined
func (d *ReqData) GetScopes() (string, error) {
	if scopesVal, ok := d.Body.Extra[ScopesKey]; ok {
		scopes, err := str.Cast(scopesVal)
		if err != nil {
			return "", errors.Wrapf(err, "while parsing the value for %s", ScopesKey)
		}

		return scopes, nil
	}

	return "", apperrors.NewKeyDoesNotExistError(ScopesKey)
}

// GetUserGroups returns group name or empty string if there's no group
func (d *ReqData) GetUserGroups() []string {
	userGroups := []string{}
	groupsVal, ok := d.Body.Extra[GroupsKey]
	if !ok {
		return userGroups
	}

	if groupsArray, ok := groupsVal.([]interface{}); ok {
		for _, group := range groupsArray {
			groupString, err := str.Cast(group)
			if err != nil {
				log.Infof("%+v skipped because string conversion failed", group)
				continue
			}

			userGroups = append(userGroups, groupString)
		}
	}

	return userGroups
}

// SetExternalTenantID sets the external tenant ID in the Header collection
func (d *ReqData) SetExternalTenantID(id string) {
	d.Body.Header.Add(ExternalTenantKey, id)
}

// SetExtraFromClaims sets the data based on the JWT claims
func (d *ReqData) SetExtraFromClaims(claims jwt.MapClaims) {
	d.Body.Extra[EmailKey] = claims[EmailKey]
	d.Body.Extra[UsernameKey] = claims[UsernameKey]
	d.Body.Extra[GroupsKey] = claims[GroupsKey]
}
