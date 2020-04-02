package input

import (
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	cloudProvider "github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/provider"
	"github.com/kyma-incubator/compass/components/provisioner/pkg/gqlschema"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/pkg/errors"
	"github.com/sirupsen/logrus"
	"github.com/vburenin/nsync"
)

//go:generate mockery -name=ComponentListProvider -output=automock -outpkg=automock -case=underscore

type (
	OptionalComponentService interface {
		ExecuteDisablers(components internal.ComponentConfigurationInputList, names ...string) (internal.ComponentConfigurationInputList, error)
		ComputeComponentsToDisable(optComponentsToKeep []string) []string
	}

	HyperscalerInputProvider interface {
		Defaults() *gqlschema.ClusterConfigInput
		ApplyParameters(input *gqlschema.ClusterConfigInput, params internal.ProvisioningParametersDTO)
	}

	CreatorForPlan interface {
		IsPlanSupport(planID string) bool
		ForPlan(planID, kymaVersion string) (internal.ProvisionInputCreator, error)
	}

	ComponentListProvider interface {
		AllComponents(kymaVersion string) ([]v1alpha1.KymaComponent, error)
	}
)

type InputBuilderFactory struct {
	kymaVersion        string
	config             Config
	optComponentsSvc   OptionalComponentService
	fullComponentsList internal.ComponentConfigurationInputList
	componentsProvider ComponentListProvider
}

func NewInputBuilderFactory(optComponentsSvc OptionalComponentService, componentsListProvider ComponentListProvider, config Config, defaultKymaVersion string) (CreatorForPlan, error) {
	components, err := componentsListProvider.AllComponents(defaultKymaVersion)
	if err != nil {
		return &InputBuilderFactory{}, errors.Wrap(err, "while creating components list for default Kyma version")
	}

	return &InputBuilderFactory{
		config:             config,
		kymaVersion:        defaultKymaVersion,
		optComponentsSvc:   optComponentsSvc,
		fullComponentsList: mapToGQLComponentConfigurationInput(components),
		componentsProvider: componentsListProvider,
	}, nil
}

func (f *InputBuilderFactory) IsPlanSupport(planID string) bool {
	switch planID {
	case broker.GcpPlanID, broker.AzurePlanID:
		return true
	default:
		return false
	}
}

func (f *InputBuilderFactory) ForPlan(planID, kymaVersion string) (internal.ProvisionInputCreator, error) {
	if !f.IsPlanSupport(planID) {
		return nil, errors.Errorf("plan %s in not supported", planID)
	}

	var provider HyperscalerInputProvider
	switch planID {
	case broker.GcpPlanID:
		provider = &cloudProvider.GcpInput{}
	case broker.AzurePlanID:
		provider = &cloudProvider.AzureInput{}
	// insert cases for other providers like AWS or GCP
	default:
		return nil, errors.Errorf("case with plan %s is not supported", planID)
	}

	initInput, err := f.initInput(provider, kymaVersion)
	if err != nil {
		return &RuntimeInput{}, errors.Wrap(err, "while initialization input")
	}

	return &RuntimeInput{
		input:                     initInput,
		overrides:                 make(map[string][]*gqlschema.ConfigEntryInput, 0),
		globalOverrides:           make([]*gqlschema.ConfigEntryInput, 0),
		mutex:                     nsync.NewNamedMutex(),
		hyperscalerInputProvider:  provider,
		optionalComponentsService: f.optComponentsSvc,
	}, nil
}

func (f *InputBuilderFactory) initInput(provider HyperscalerInputProvider, kymaVersion string) (gqlschema.ProvisionRuntimeInput, error) {
	var (
		version    string
		components internal.ComponentConfigurationInputList
	)

	if kymaVersion != "" {
		allComponents, err := f.componentsProvider.AllComponents(kymaVersion)
		if err != nil {
			return gqlschema.ProvisionRuntimeInput{}, errors.Wrapf(err, "while fetching components for %s Kyma version", kymaVersion)
		}
		version = kymaVersion
		components = mapToGQLComponentConfigurationInput(allComponents)
	} else {
		version = f.kymaVersion
		components = f.fullComponentsList
	}

	return gqlschema.ProvisionRuntimeInput{
		RuntimeInput:  &gqlschema.RuntimeInput{},
		ClusterConfig: provider.Defaults(),
		KymaConfig: &gqlschema.KymaConfigInput{
			Version:    version,
			Components: components.DeepCopy(),
		},
	}, nil
}

func mapToGQLComponentConfigurationInput(kymaComponents []v1alpha1.KymaComponent) internal.ComponentConfigurationInputList {
	var input internal.ComponentConfigurationInputList
	for _, component := range kymaComponents {
		var sourceURL *string
		if component.Source != nil {
			sourceURL = &component.Source.URL
			logrus.Infof("Source URL: %s", sourceURL)
		}

		input = append(input, &gqlschema.ComponentConfigurationInput{
			Component: component.Name,
			Namespace: component.Namespace,
			SourceURL: sourceURL,
		})
	}

	return input
}
