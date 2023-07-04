package types

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/apperrors"
)

const (
	assignOperation   = "assign"
	unassignOperation = "unassign"
)

// Context is a structure used to JSON decode the context in the Body
type Context struct {
	Operation string `json:"operation,omitempty"`
}

// BasicAuthentication is a structure used to JSON decode the basicAuthentication in the OutboundCommunication
type BasicAuthentication struct {
	Username string `json:"username,omitempty"`
	Password string `json:"password,omitempty"`
}

// Oauth2ClientCredentials is a structure used to JSON decode the oauth2ClientCredentials in the OutboundCommunication
type Oauth2ClientCredentials struct {
	TokenServiceURL string `json:"tokenServiceUrl,omitempty"`
	ClientID        string `json:"clientId,omitempty"`
	ClientSecret    string `json:"clientSecret,omitempty"`
}

// OutboundCommunication is a structure used to JSON decode the outboundCommunication in the Credentials
type OutboundCommunication struct {
	BasicAuthentication     BasicAuthentication     `json:"basicAuthentication,omitempty"`
	Oauth2ClientCredentials Oauth2ClientCredentials `json:"oauth2ClientCredentials,omitempty"`
}

// Credentials is a structure used to JSON decode the credentials in the Configuration
type Credentials struct {
	OutboundCommunication OutboundCommunication `json:"outboundCommunication,omitempty"`
}

// Configuration is a structure used to JSON decode the configuration in the AssignedTenant
type Configuration struct {
	Credentials Credentials `json:"credentials,omitempty"`
}

// ReceiverTenant is a structure used to JSON decode the receiverTenant in the Body
type ReceiverTenant struct {
	UclSystemTenantID string `json:"uclSystemTenantId,omitempty"`
	OwnerTenant       string `json:"ownerTenant,omitempty"`
}

// AssignedTenant is a structure used to JSON decode the assignedTenant in the Body
type AssignedTenant struct {
	UclSystemTenantID string        `json:"uclSystemTenantId,omitempty"`
	Configuration     Configuration `json:"configuration,omitempty"`
}

// Body is a structure used to JSON decode the request body sent to the adapter handler
type Body struct {
	Context        Context        `json:"context,omitempty"`
	ReceiverTenant ReceiverTenant `json:"receiverTenant,omitempty"`
	AssignedTenant AssignedTenant `json:"assignedTenant,omitempty"`
}

// Validate validates the request Body
func (b Body) Validate() error {
	if b.Context.Operation != assignOperation && b.Context.Operation != unassignOperation {
		return apperrors.NewInvalidDataError(fmt.Sprintf("Context operation must be either %q or %q", assignOperation, unassignOperation))
	}

	if len(b.ReceiverTenant.OwnerTenant) == 0 {
		return apperrors.NewInvalidDataError("Receiver tenant owner tenant must be provided.")
	}

	if b.GetApplicationConfiguration() != (Configuration{}) {
		oauthCredentials := b.GetOauthCredentials()
		basicCredentials := b.GetBasicCredentials()

		if oauthCredentials != (Oauth2ClientCredentials{}) &&
			(oauthCredentials.ClientID == "" || oauthCredentials.ClientSecret == "" || oauthCredentials.TokenServiceURL == "") {
			return apperrors.NewInvalidDataError("All of OauthCredentials properties should be provided")
		}

		if basicCredentials != (BasicAuthentication{}) &&
			(basicCredentials.Username == "" || basicCredentials.Password == "") {
			return apperrors.NewInvalidDataError("All of BasicCredentials properties should be provided")
		}
	}

	return nil
}

// GetOauthCredentials returns the Body oauth credentials
func (b Body) GetOauthCredentials() Oauth2ClientCredentials {
	return b.AssignedTenant.Configuration.Credentials.OutboundCommunication.Oauth2ClientCredentials
}

// GetBasicCredentials returns the Body basic credentials
func (b Body) GetBasicCredentials() BasicAuthentication {
	return b.AssignedTenant.Configuration.Credentials.OutboundCommunication.BasicAuthentication
}

// GetRuntimeId returns the Body runtime ID
func (b Body) GetRuntimeId() string {
	return b.ReceiverTenant.UclSystemTenantID
}

// GetApplicationId returns the Body application ID
func (b Body) GetApplicationId() string {
	return b.AssignedTenant.UclSystemTenantID
}

// GetApplicationConfiguration returns the Body application configuration
func (b Body) GetApplicationConfiguration() Configuration {
	return b.AssignedTenant.Configuration
}
