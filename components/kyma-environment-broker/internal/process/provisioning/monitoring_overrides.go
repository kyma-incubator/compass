package provisioning

import (
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"
	"github.com/sirupsen/logrus"
)

type MonitoringOverrideStep struct {
	operationManager *process.OperationManager
}

func (s *MonitoringOverrideStep) Name() string {
	return "Monitoring_Override"
}

func NewMonitoringOverrideStep(os storage.Operations) *MonitoringOverrideStep {
	return &MonitoringOverrideStep{
		operationManager: process.NewOperationManager(os),
	}
}

func (s *MonitoringOverrideStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	log.Info("Setting up monitoring overrides")
	//monitoringOverrides := s.setupMonitoringOverride()
	monitoringOverrides := []*gqlschema.ConfigEntryInput{
		{
			Key:    "resourceSelector.namespaces",
			Value:  "kyma-system,istio-system,knative-eventing,knative-serving,kyma-integration,kube-system",
			Secret: ptr.Bool(true),
		},
		{
			Key:    "grafana.kyma.console.enabled",
			Value:  "false",
			Secret: ptr.Bool(true),
		},
		{
			Key:    "grafana.env.GF_USERS_AUTO_ASSIGN_ORG_ROLE",
			Value:  "Admin",
			Secret: ptr.Bool(true),
		},
		{
			Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_SCOPES",
			Value:  "openid email",
			Secret: ptr.Bool(true),
		},
		{
			Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_TOKEN_URL",
			Value:  "https://kyma.blah.com/oauth2/token",
			Secret: ptr.Bool(true),
		},
		{
			Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_AUTH_URL",
			Value:  "https://kyma.foo.com/oauth2/token",
			Secret: ptr.Bool(true),
		},{
			Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_API_URL",
			Value:  "https://kyma.bar.com/oauth2/token",
			Secret: ptr.Bool(true),
		},
	}
	operation.InputCreator.AppendOverrides("monitoring", monitoringOverrides)
	return operation, 0, nil
}

//func (s *MonitoringOverrideStep) setupMonitoringOverride() []*gqlschema.ConfigEntryInput {
//	monitoringOverrides := []*gqlschema.ConfigEntryInput{
//		{
//			Key:    "resourceSelector.namespaces",
//			Value:  "kyma-system,istio-system,knative-eventing,knative-serving,kyma-integration,kube-system",
//			Secret: ptr.Bool(true),
//		},
//		{
//			Key:    "grafana.kyma.console.enabled",
//			Value:  "false",
//			Secret: ptr.Bool(true),
//		},
//		{
//			Key:    "grafana.env.GF_USERS_AUTO_ASSIGN_ORG_ROLE",
//			Value:  "Admin",
//			Secret: ptr.Bool(true),
//		},
//		{
//			Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_SCOPES",
//			Value:  "openid email",
//			Secret: ptr.Bool(true),
//		},
//		{
//			Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_TOKEN_URL",
//			Value:  "https://kyma.blah.com/oauth2/token",
//			Secret: ptr.Bool(true),
//		},
//		{
//			Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_AUTH_URL",
//			Value:  "https://kyma.foo.com/oauth2/token",
//			Secret: ptr.Bool(true),
//		},{
//			Key:    "grafana.env.GF_AUTH_GENERIC_OAUTH_API_URL",
//			Value:  "https://kyma.bar.com/oauth2/token",
//			Secret: ptr.Bool(true),
//		},
//	}
//	return monitoringOverrides
//
//}
