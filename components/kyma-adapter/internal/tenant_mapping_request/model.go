package tenant_mapping_request

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

const (
	assignOperation   = "assign"
	unassignOperation = "unassign"
)

type Context struct {
	Operation string `json:"operation,omitempty"`
}

type BasicAuthentication struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

type Oauth2ClientCredentials struct {
	TokenServiceUrl string `json:"tokenServiceUrl,omitempty"`
	ClientId        string `json:"clientId,omitempty"`
	ClientSecret    string `json:"clientSecret,omitempty"`
}

type OutboundCommunication struct {
	BasicAuthentication     BasicAuthentication     `json:"basicAuthentication,omitempty"`
	Oauth2ClientCredentials Oauth2ClientCredentials `json:"oauth2ClientCredentials,omitempty"`
}

type Credentials struct {
	OutboundCommunication OutboundCommunication `json:"outboundCommunication,omitempty"`
}

type Configuration struct {
	Credentials Credentials `json:"credentials,omitempty"`
}

type ReceiverTenant struct {
	UclSystemTenantId string `json:"uclSystemTenantId,omitempty"`
}

type AssignedTenant struct {
	UclSystemTenantId string        `json:"uclSystemTenantId,omitempty"`
	Configuration     Configuration `json:"configuration,omitempty"`
}

type Body struct {
	Context        Context        `json:"context,omitempty"`
	ReceiverTenant ReceiverTenant `json:"receiverTenant,omitempty"`
	AssignedTenant AssignedTenant `json:"assignedTenant,omitempty"`
}

func (b Body) Validate() error {
	if b.Context.Operation != assignOperation && b.Context.Operation != unassignOperation {
		return apperrors.NewInvalidDataError(fmt.Sprintf("Context operation must be either %q or %q", assignOperation, unassignOperation))
	}

	if b.GetApplicationConfiguration() != (Configuration{}) {
		oauthCredentials := b.GetOauthCredentials()
		basicCredentials := b.GetBasicCredentials()

		if oauthCredentials != (Oauth2ClientCredentials{}) &&
			(oauthCredentials.ClientId == "" || oauthCredentials.ClientSecret == "" || oauthCredentials.TokenServiceUrl == "") {
			return apperrors.NewInvalidDataError("All of OauthCredentials properties should be provided")
		}

		if basicCredentials != (BasicAuthentication{}) &&
			(basicCredentials.Username == "" || basicCredentials.Password == "") {
			return apperrors.NewInvalidDataError("All of BasicCredentials properties should be provided")
		}
	}

	return nil
}

func (b Body) GetOauthCredentials() Oauth2ClientCredentials {
	return b.AssignedTenant.Configuration.Credentials.OutboundCommunication.Oauth2ClientCredentials
}

func (b Body) GetBasicCredentials() BasicAuthentication {
	return b.AssignedTenant.Configuration.Credentials.OutboundCommunication.BasicAuthentication
}

func (b Body) GetRuntimeId() string {
	return b.ReceiverTenant.UclSystemTenantId
}

func (b Body) GetApplicationId() string {
	return b.AssignedTenant.UclSystemTenantId
}

func (b Body) GetApplicationConfiguration() Configuration {
	return b.AssignedTenant.Configuration
}
