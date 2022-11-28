package oathkeeper

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/kyma-incubator/compass/components/director/pkg/cert"
	"github.com/kyma-incubator/compass/components/director/pkg/log"
	"github.com/kyma-incubator/compass/components/hydrator/pkg/authenticator"

	"github.com/form3tech-oss/jwt-go"
	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/pkg/errors"
)

const (
	// CertificateFlow missing godoc
	CertificateFlow AuthFlow = "Certificate"
	// OneTimeTokenFlow missing godoc
	OneTimeTokenFlow AuthFlow = "OneTimeToken"
	// OAuth2Flow missing godoc
	OAuth2Flow AuthFlow = "OAuth2"
	// JWTAuthFlow missing godoc
	JWTAuthFlow AuthFlow = "JWT"
	// ConsumerProviderFlow is using when we have consumer-provider and subscription relationship between them
	ConsumerProviderFlow AuthFlow = "Consumer-Provider"

	// ClientIDKey missing godoc
	ClientIDKey = "client_id"
	// EmailKey missing godoc
	EmailKey = "email"
	// UsernameKey missing godoc
	UsernameKey = "name"
	// GroupsKey missing godoc
	GroupsKey = "groups"
	// ClientIDCertKey missing godoc
	ClientIDCertKey = "client-id-from-certificate"
	// ClientIDCertIssuer missing godoc
	ClientIDCertIssuer = "client-certificate-issuer"
	// ClientIDTokenKey missing godoc
	ClientIDTokenKey = "client-id-from-token"
	// ExternalTenantKey missing godoc
	ExternalTenantKey = "tenant"
	// UserContextKey is a header key containing consumer data
	UserContextKey = "User_context"
	// ScopesKey missing godoc
	ScopesKey = "scope"

	// ConnectorIssuer missing godoc
	ConnectorIssuer = "connector"
	// ExternalIssuer missing godoc
	ExternalIssuer = "certificate-service"
)

// AuthDetails contains information about the currently authenticated client - AuthID, AuthFlow and Authenticator to use for further processing
type AuthDetails struct {
	AuthID        string
	AuthFlow      AuthFlow
	CertIssuer    string
	Authenticator *authenticator.Config
	ScopePrefix   string
	Region        string
}

// AuthFlow wraps possible flows of auth like OAuth2, JWT and certificate
type AuthFlow string

// IsCertFlow missing godoc
func (f AuthFlow) IsCertFlow() bool {
	return f == CertificateFlow
}

// IsOneTimeTokenFlow missing godoc
func (f AuthFlow) IsOneTimeTokenFlow() bool {
	return f == OneTimeTokenFlow
}

// IsOAuth2Flow missing godoc
func (f AuthFlow) IsOAuth2Flow() bool {
	return f == OAuth2Flow
}

// IsJWTFlow missing godoc
func (f AuthFlow) IsJWTFlow() bool {
	return f == JWTAuthFlow
}

// ReqBody represents parsed request input to the handler
type ReqBody struct {
	Subject string                 `json:"subject"`
	Extra   map[string]interface{} `json:"extra"`
	Header  http.Header            `json:"header"`
}

// ReqData represents incoming request with parsed body and its header
type ReqData struct {
	Body   ReqBody
	Header http.Header
	ctx    context.Context
}

// ExtraData represents the extra fields that might be provided in the incoming request
type ExtraData struct {
	InternalConsumerID string
	ConsumerType       model.SystemAuthReferenceObjectType
	AccessLevels       []string
}

// NewReqData missing godoc
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

	if tenantVal, ok := d.Body.Extra[ExternalTenantKey]; ok && tenantVal != "" {
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
	userGroups := make([]string, 0)
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
	} else if groupsStr, err := str.Cast(groupsVal); err == nil {
		userGroups = append(userGroups, groupsStr)
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

// ExtractCoordinates extracts the authenticator coordinates from ReqData. The coordinates are stored in Body.Extra and the key for them is "authenticator_coordinates".
func (d *ReqData) ExtractCoordinates() (authenticator.Coordinates, bool, error) {
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

// IsIntegrationSystemFlow returns true if a tenant header is missing or is provided, but it differs from
// the client ID found in the certificate
func (d *ReqData) IsIntegrationSystemFlow() bool {
	clientIDFromCert := d.Body.Header.Get(ClientIDCertKey)
	tenant, err := d.GetExternalTenantID()
	if err != nil {
		return false
	}
	return clientIDFromCert != tenant && d.ConsumerType() == model.IntegrationSystemReference
}

// TenantAccessLevels gets the granted tenant access levels from body extra if they exist.
func (d *ReqData) TenantAccessLevels() []string {
	if d.Body.Extra == nil {
		return nil
	}
	if _, found := d.Body.Extra[cert.AccessLevelsExtraField]; !found {
		return nil
	}
	accessLevelsRaw, ok := d.Body.Extra[cert.AccessLevelsExtraField].([]interface{})
	if !ok {
		return nil
	}
	accessLevels := make([]string, 0)
	for _, al := range accessLevelsRaw {
		accessLevels = append(accessLevels, fmt.Sprintf("%s", al))
	}
	return accessLevels
}

// ConsumerType gets consumer type from body extra if it exists.
func (d *ReqData) ConsumerType() model.SystemAuthReferenceObjectType {
	defaultConsumerType := model.ExternalCertificateReference
	if d.Body.Extra == nil {
		return defaultConsumerType
	}
	consumerType, found := d.Body.Extra[cert.ConsumerTypeExtraField]
	if !found {
		return defaultConsumerType
	}
	return model.SystemAuthReferenceObjectType(fmt.Sprint(consumerType))
}

// InternalConsumerID gets internal consumer id from body extra if it exists.
func (d *ReqData) InternalConsumerID() string {
	if d.Body.Extra == nil {
		return ""
	}
	if _, found := d.Body.Extra[cert.InternalConsumerIDField]; !found {
		return ""
	}
	return fmt.Sprint(d.Body.Extra[cert.InternalConsumerIDField])
}

// GetExtraDataWithDefaults gets body extra.
func (d *ReqData) GetExtraDataWithDefaults() ExtraData {
	return ExtraData{
		InternalConsumerID: d.InternalConsumerID(),
		ConsumerType:       d.ConsumerType(),
		AccessLevels:       d.TenantAccessLevels(),
	}
}
