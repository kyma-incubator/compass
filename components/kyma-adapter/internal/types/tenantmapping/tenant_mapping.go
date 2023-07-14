package tenantmapping

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
	Operation string `json:"operation"`
}

// BasicAuthentication is a structure used to JSON decode the basicAuthentication in the OutboundCommunication
type BasicAuthentication struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Oauth2ClientCredentials is a structure used to JSON decode the oauth2ClientCredentials in the OutboundCommunication
type Oauth2ClientCredentials struct {
	TokenServiceURL string `json:"tokenServiceUrl"`
	ClientID        string `json:"clientId"`
	ClientSecret    string `json:"clientSecret"`
}

// OutboundCommunication is a structure used to JSON decode the outboundCommunication in the Credentials
type OutboundCommunication struct {
	BasicAuthentication     BasicAuthentication     `json:"basicAuthentication"`
	Oauth2ClientCredentials Oauth2ClientCredentials `json:"oauth2ClientCredentials"`
}

// Credentials is a structure used to JSON decode the credentials in the Configuration
type Credentials struct {
	OutboundCommunication OutboundCommunication `json:"outboundCommunication"`
}

// Configuration is a structure used to JSON decode the configuration in the AssignedTenant
type Configuration struct {
	Credentials Credentials `json:"credentials"`
}

// ReceiverTenant is a structure used to JSON decode the receiverTenant in the Body
type ReceiverTenant struct {
	UclSystemTenantID string `json:"uclSystemTenantId"`
	OwnerTenant       string `json:"ownerTenant"`
}

// AssignedTenant is a structure used to JSON decode the assignedTenant in the Body
type AssignedTenant struct {
	UclSystemTenantID string        `json:"uclSystemTenantId"`
	Configuration     Configuration `json:"configuration"`
}

// Body is a structure used to JSON decode the request body sent to the adapter handler
type Body struct {
	Context        Context        `json:"context"`
	ReceiverTenant ReceiverTenant `json:"receiverTenant"`
	AssignedTenant AssignedTenant `json:"assignedTenant"`
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

// GetRuntimeID returns the Body runtime ID
func (b Body) GetRuntimeID() string {
	return b.ReceiverTenant.UclSystemTenantID
}

// GetApplicationID returns the Body application ID
func (b Body) GetApplicationID() string {
	return b.AssignedTenant.UclSystemTenantID
}

// GetApplicationConfiguration returns the Body application configuration
func (b Body) GetApplicationConfiguration() Configuration {
	return b.AssignedTenant.Configuration
}
