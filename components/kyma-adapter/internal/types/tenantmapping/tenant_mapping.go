package tenantmapping

import (
	"encoding/json"
	"fmt"

	"github.com/pkg/errors"

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
	UclSystemTenantID string          `json:"uclSystemTenantId"`
	Configuration     json.RawMessage `json:"configuration"`
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

	applicationConfiguration, err := b.GetApplicationConfiguration()
	if err != nil {
		return apperrors.NewInvalidDataError("failed to get application configuration %s", err)
	}

	if applicationConfiguration != (Configuration{}) {
		oauthCredentials := applicationConfiguration.GetOauthCredentials()
		basicCredentials := applicationConfiguration.GetBasicCredentials()

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

// GetOauthCredentials returns the Configuration oauth credentials
func (c Configuration) GetOauthCredentials() Oauth2ClientCredentials {
	return c.Credentials.OutboundCommunication.Oauth2ClientCredentials
}

// GetBasicCredentials returns the Configuration basic credentials
func (c Configuration) GetBasicCredentials() BasicAuthentication {
	return c.Credentials.OutboundCommunication.BasicAuthentication
}

// GetRuntimeID returns the Body runtime ID
func (b Body) GetRuntimeID() string {
	return b.ReceiverTenant.UclSystemTenantID
}

// GetApplicationID returns the Body application ID
func (b Body) GetApplicationID() string {
	return b.AssignedTenant.UclSystemTenantID
}

// GetApplicationConfiguration returns the Body application configuration as Configuration struct
func (b Body) GetApplicationConfiguration() (Configuration, error) {
	if b.AssignedTenant.Configuration != nil && (string(b.AssignedTenant.Configuration) == "" || string(b.AssignedTenant.Configuration) == "{}" || string(b.AssignedTenant.Configuration) == "\"\"" || string(b.AssignedTenant.Configuration) == "null") {
		return Configuration{}, nil
	}

	var applicationConfiguration Configuration
	if err := json.Unmarshal(b.AssignedTenant.Configuration, &applicationConfiguration); err != nil {
		return Configuration{}, errors.Wrapf(err, "while unmarshalling application configuration for application with ID: %q", b.AssignedTenant.Configuration)
	}

	return applicationConfiguration, nil
}
