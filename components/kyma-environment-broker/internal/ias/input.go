package ias

import "github.com/pkg/errors"

type SPInputID int

const ( // enum SPInputID
	SPDexID     = 1
	SPGrafanaID = 2
)

const ( // enum SsoType
	OIDC = "openIdConnect"
	SAML = "saml2"
)

type ServiceProviderParam struct {
	domain        string
	ssoType       string
	redirectPath  string
	allowedGroups []string
}

var ServiceProviderInputs = map[SPInputID]ServiceProviderParam{
	SPDexID: {
		domain:        "dex",
		ssoType:       SAML,
		redirectPath:  "/callback",
		allowedGroups: []string{"runtimeOperator", "runtimeAdmin"},
	},
	SPGrafanaID: {
		domain:        "grafana",
		ssoType:       OIDC,
		redirectPath:  "/login/generic_oauth",
		allowedGroups: []string{"skr-monitoring-admin", "skr-monitoring-viewer"},
	},
}

func (id SPInputID) isValid() error {
	switch id {
	case SPGrafanaID, SPDexID:
		return nil
	}
	return errors.Errorf("Invalid Service Provider input ID: %d", id)
}
