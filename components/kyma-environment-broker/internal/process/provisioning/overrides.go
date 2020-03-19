package provisioning

import (
	"context"
	"time"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/ptr"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/storage"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	coreV1 "k8s.io/api/core/v1"
	"k8s.io/apimachinery/pkg/labels"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

const (
	namespace      = "compass-system"
	overrideLabel  = "provisioning-runtime-override=true"
	componentLabel = "component"
)

type OverridesStep struct {
	k8sClient        client.Client
	operationManager *process.OperationManager
}

func NewOverridesStep(cli client.Client, os storage.Operations) *OverridesStep {
	return &OverridesStep{
		k8sClient:        cli,
		operationManager: process.NewOperationManager(os),
	}
}

func (s *OverridesStep) Name() string {
	return "Overrides"
}

func (s *OverridesStep) Run(operation internal.ProvisioningOperation, log logrus.FieldLogger) (internal.ProvisioningOperation, time.Duration, error) {
	overrides := make(map[string][]*gqlschema.ConfigEntryInput, 0)
	ctx := context.Background()

	options, err := s.prepareSearchOptions()
	if err != nil {
		log.Errorf("cannot prepare options: %s", err)
		return s.operationManager.OperationFailed(operation, "invalid options parameters")
	}

	secretList := &coreV1.SecretList{}
	err = s.k8sClient.List(ctx, secretList, options)
	if err != nil {
		log.Errorf("cannot fetch list of secrets: %s", err)
		return operation, 10 * time.Second, nil
	}

	for _, secret := range secretList.Items {
		cName, err := componentName(secret.Labels)
		if err != nil {
			log.Errorf("cannot fetch component name from resource: %s", err)
			return s.operationManager.OperationFailed(operation, "invalid override resource")
		}
		for key, value := range secret.Data {
			overrides[cName] = append(overrides[cName], &gqlschema.ConfigEntryInput{
				Key:    key,
				Value:  string(value),
				Secret: ptr.Bool(true),
			})
		}
	}

	configMapList := &coreV1.ConfigMapList{}
	err = s.k8sClient.List(ctx, configMapList, options)
	if err != nil {
		log.Errorf("cannot fetch list of config maps: %s", err)
		return operation, 10 * time.Second, nil
	}

	for _, cm := range configMapList.Items {
		cName, err := componentName(cm.Labels)
		if err != nil {
			log.Errorf("cannot fetch component name from resource: %s", err)
			return s.operationManager.OperationFailed(operation, "invalid override resource")
		}
		for key, value := range cm.Data {
			overrides[cName] = append(overrides[cName], &gqlschema.ConfigEntryInput{
				Key:   key,
				Value: value,
			})
		}
	}

	for component, ovs := range overrides {
		operation.InputCreator.AppendOverrides(component, ovs)
	}
	return operation, 0, nil
}

func (s *OverridesStep) prepareSearchOptions() (*client.ListOptions, error) {
	options := &client.ListOptions{Namespace: namespace}
	labelSelector, err := labels.Parse(overrideLabel)
	if err != nil {
		return options, errors.Wrapf(err, "while parsing overrides labels %q", overrideLabel)
	}
	options.LabelSelector = labelSelector
	return options, nil
}

func componentName(labels map[string]string) (string, error) {
	for name, value := range labels {
		if name == componentLabel {
			return value, nil
		}
	}
	return "", errors.Errorf("resource does not have a suitable label: %s ", componentLabel)
}
