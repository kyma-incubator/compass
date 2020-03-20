package provisioning

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/sirupsen/logrus"
	coreV1 "k8s.io/api/core/v1"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespace      = "compass-system"
	overrideLabel  = "provisioning-runtime-override"
	componentLabel = "component"
)

type OverridesFromSecretsAndConfigStep struct {
	ctx              context.Context
	k8sClient        client.Client
	operationManager *process.OperationManager
}

func NewOverridesFromSecretsAndConfigStep(c context.Context, cli client.Client, os storage.Operations) *OverridesFromSecretsAndConfigStep {
	return &OverridesFromSecretsAndConfigStep{
		ctx:              c,
		k8sClient:        cli,
		operationManager: process.NewOperationManager(os),
	}
}

func (s *OverridesFromSecretsAndConfigStep) Name() string {
	return "Overrides_From_Secrets_And_Config_Step"
}

func (s *OverridesFromSecretsAndConfigStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	overrides := make(map[string][]*gqlschema.ConfigEntryInput, 0)
	globalOverrides := make([]*gqlschema.ConfigEntryInput, 0)

	secretList := &coreV1.SecretList{}
	if err := s.k8sClient.List(s.ctx, secretList, s.listOptions()...); err != nil {
		log.Errorf("cannot fetch list of secrets: %s", err)
		return operation, 10 * time.Second, nil
	}

	for _, secret := range secretList.Items {
		cName, global := componentName(secret.Labels)
		for key, value := range secret.Data {
			if global {
				globalOverrides = append(globalOverrides, &gqlschema.ConfigEntryInput{
					Key:    key,
					Value:  string(value),
					Secret: ptr.Bool(true),
				})
			} else {
				overrides[cName] = append(overrides[cName], &gqlschema.ConfigEntryInput{
					Key:    key,
					Value:  string(value),
					Secret: ptr.Bool(true),
				})
			}
		}
	}

	configMapList := &coreV1.ConfigMapList{}
	if err := s.k8sClient.List(s.ctx, configMapList, s.listOptions()...); err != nil {
		log.Errorf("cannot fetch list of config maps: %s", err)
		return operation, 10 * time.Second, nil
	}

	for _, cm := range configMapList.Items {
		cName, global := componentName(cm.Labels)
		for key, value := range cm.Data {
			if global {
				globalOverrides = append(globalOverrides, &gqlschema.ConfigEntryInput{
					Key:   key,
					Value: value,
				})
			} else {
				overrides[cName] = append(overrides[cName], &gqlschema.ConfigEntryInput{
					Key:   key,
					Value: value,
				})
			}
		}
	}

	for component, ovs := range overrides {
		operation.InputCreator.AppendOverrides(component, ovs)
	}
	if len(globalOverrides) > 0 {
		operation.InputCreator.AppendGlobalOverrides(globalOverrides)
	}

	return operation, 0, nil
}

func (s *OverridesFromSecretsAndConfigStep) listOptions() []client.ListOption {
	label := map[string]string{
		overrideLabel: "true",
	}

	return []client.ListOption{
		client.InNamespace(namespace),
		client.MatchingLabels(label),
	}
}

// componentName returns component name from label and determines whether the override is global or not
func componentName(labels map[string]string) (string, bool) {
	for name, value := range labels {
		if name == componentLabel {
			return value, false
		}
	}
	return "", true
}
