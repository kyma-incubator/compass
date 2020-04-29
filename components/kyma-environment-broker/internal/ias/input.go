package ias

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
)

type OverridesGetter func(*ServiceProviderBundle) (string, []*gqlschema.ConfigEntryInput)
type ServiceProviderParam struct {
	Domain        string
	ssoType       string
	redirectPath  string
	allowedGroups []string
	overrides     OverridesGetter
}

var ServiceProviderInputs = map[string]ServiceProviderParam{
	"dex": {
		Domain:        "dex",
		ssoType:       "saml2",
		redirectPath:  "/callback",
		allowedGroups: []string{"runtimeOperator", "runtimeAdmin"},
		overrides:     nil,
	},
	"grafana": {
		Domain:        "grafana",
		ssoType:       "openIdConnect",
		redirectPath:  "/login/generic_oauth",
		allowedGroups: []string{"skr-monitoring-admin", "skr-monitoring-viewer"},
		overrides: func(b *ServiceProviderBundle) (string, []*gqlschema.ConfigEntryInput) {
			return "monitoring", []*gqlschema.ConfigEntryInput{
				{
					Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_CLIENT_ID",
					Value:  b.serviceProviderSecret.ClientID,
					Secret: ptr.Bool(true),
				},
				{
					Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_CLIENT_SECRET",
					Value:  b.serviceProviderSecret.ClientSecret,
					Secret: ptr.Bool(true),
				},
			}
		},
	},
}
