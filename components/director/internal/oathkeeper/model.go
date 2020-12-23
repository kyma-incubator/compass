package oathkeeper

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/authenticator"

	"github.com/kyma-incubator/compass/components/director/pkg/log"

	"github.com/form3tech-oss/jwt-go"
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

// AuthDetails contains information about the currently authenticated client - AuthID, AuthFlow and Authenticator to use for further processing
type AuthDetails struct {
	AuthID        string
	AuthFlow      AuthFlow
	Authenticator *authenticator.Config
	ScopePrefix   string
}

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
	ctx    context.Context
}

func NewReqData(ctx context.Context, reqBody ReqBody, reqHeader http.Header) ReqData {
	if reqBody.Extra == nil {
		reqBody.Extra = make(map[string]interface{})
	}

	if reqBody.Header == nil {
		reqBody.Header = make(http.Header)
	}

	return ReqData{
		Body:   reqBody,
		Header: reqHeader,
		ctx:    ctx,
	}
}

// GetAuthID looks for auth ID and identifies auth flow in the parsed request input represented by the ReqData struct
func (d *ReqData) GetAuthID(ctx context.Context) (*AuthDetails, error) {
	return d.GetAuthIDWithAuthenticators(ctx, []authenticator.Config{})
}

// GetAuthIDWithAuthenticators looks for auth ID and identifies auth flow in the parsed request input represented by the ReqData struct while taking into account existing preconfigured authenticators
func (d *ReqData) GetAuthIDWithAuthenticators(ctx context.Context, authenticators []authenticator.Config) (*AuthDetails, error) {
	coords, exist, err := d.extractCoordinates()
	if err != nil {
		return nil, errors.Wrap(err, "while extracting coordinates")
	}
	if exist {
		for _, authn := range authenticators {
			if authn.Name != coords.Name {
				continue
			}

			log.C(ctx).Infof("Request token matches %q authenticator", authn.Name)
			identity, ok := d.Body.Extra[authn.Attributes.IdentityAttribute.Key]
			if !ok {
				return nil, apperrors.NewInvalidDataError("missing identity attribute from %q authenticator token", authn.Name)
			}

			authID, err := str.Cast(identity)
			if err != nil {
				return nil, errors.Wrapf(err, "while parsing the value for %s", identity)
			}
			index := coords.Index
			return &AuthDetails{AuthID: authID, AuthFlow: JWTAuthFlow, Authenticator: &authn, ScopePrefix: authn.TrustedIssuers[index].ScopePrefix}, nil
		}
	}

	if idVal, ok := d.Body.Extra[ClientIDKey]; ok {
		authID, err := str.Cast(idVal)
		if err != nil {
			return nil, errors.Wrapf(err, "while parsing the value for %s", ClientIDKey)
		}

		return &AuthDetails{AuthID: authID, AuthFlow: OAuth2Flow}, nil
	}

	if idVal := d.Body.Header.Get(ClientIDCertKey); idVal != "" {
		return &AuthDetails{AuthID: idVal, AuthFlow: CertificateFlow}, nil
	}

	if idVal := d.Body.Header.Get(ClientIDTokenKey); idVal != "" {
		return &AuthDetails{AuthID: idVal, AuthFlow: OneTimeTokenFlow}, nil
	}

	if usernameVal, ok := d.Body.Extra[UsernameKey]; ok {
		username, err := str.Cast(usernameVal)
		if err != nil {
			return nil, errors.Wrapf(err, "while parsing the value for %s", UsernameKey)
		}
		return &AuthDetails{AuthID: username, AuthFlow: JWTAuthFlow}, nil
	}

	return nil, apperrors.NewInternalError("unable to find valid auth ID")
}

// MarshalExtra marshals the request data extra content
func (d *ReqData) MarshalExtra() (string, error) {
	extra, err := json.Marshal(d.Body.Extra)
	if err != nil {
		return "", err
	}

	return string(extra), nil
}

// GetExternalTenantID returns external tenant ID from the parsed request input if it is defined
func (d *ReqData) GetExternalTenantID() (string, error) {
	if tenantVal := d.Body.Header.Get(ExternalTenantKey); tenantVal != "" {
		return tenantVal, nil
	}

	if tenantVal, ok := d.Body.Extra[ExternalTenantKey]; ok {
		tenant, err := str.Cast(tenantVal)
		if err != nil {
			return "", errors.Wrapf(err, "while parsing the value for key=%s", ExternalTenantKey)
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

// GetUserScopes returns scopes as string array from the parsed request input if defined;
// also it strips the scopes from any potential authenticator prefixes
func (d *ReqData) GetUserScopes(scopePrefix string) ([]string, error) {
	userScopes := make([]string, 0)
	scopesVal, ok := d.Body.Extra[ScopesKey]
	if !ok {
		return userScopes, nil
	}

	if scopesArray, ok := scopesVal.([]interface{}); ok {
		for _, scope := range scopesArray {
			scopeString, err := str.Cast(scope)
			if err != nil {
				return []string{}, errors.Wrapf(err, "while parsing the value for %s", ScopesKey)
			}
			actualScope := strings.TrimPrefix(scopeString, scopePrefix)
			userScopes = append(userScopes, actualScope)
		}
	}

	return userScopes, nil
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
				log.C(d.ctx).Infof("%+v skipped because string conversion failed", group)
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

func (d *ReqData) extractCoordinates() (authenticator.Coordinates, bool, error) {
	var coords authenticator.Coordinates
	coordsInterface, exists := d.Body.Extra[authenticator.CoordinatesKey]
	if !exists {
		return coords, false, nil
	}

	coordsBytes, err := json.Marshal(coordsInterface)
	if err != nil {
		return coords, true, errors.Wrap(err, "while marshaling authenticator coordinates")
	}
	if err := json.Unmarshal(coordsBytes, &coords); err != nil {
		return coords, true, errors.Wrap(err, "while unmarshaling authenticator coordinates")
	}

	return coords, true, nil
}
