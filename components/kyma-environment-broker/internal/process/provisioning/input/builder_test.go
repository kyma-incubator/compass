package input

import (
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"
	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/process/provisioning/input/automock"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/stretchr/testify/assert"
)

// Currently on production only azure is supported

func TestInputBuilderFactory_IsPlanSupport(t *testing.T) {
	// given
	componentsProvider := &automock.ComponentListProvider{}
	componentsProvider.On("AllComponents", "1.10").Return([]v1alpha1.KymaComponent{}, nil)
	defer componentsProvider.AssertExpectations(t)

	ibf, err := NewInputBuilderFactory(nil, componentsProvider, Config{}, "1.10")
	assert.NoError(t, err)

	// when/then
	assert.True(t, ibf.IsPlanSupport(broker.GcpPlanID))
	assert.True(t, ibf.IsPlanSupport(broker.AzurePlanID))
}

func TestInputBuilderFactory_ForPlan(t *testing.T) {
	// given
	componentsProvider := &automock.ComponentListProvider{}
	componentsProvider.On("AllComponents", "1.10").Return([]v1alpha1.KymaComponent{}, nil)
	defer componentsProvider.AssertExpectations(t)

	ibf, err := NewInputBuilderFactory(nil, componentsProvider, Config{}, "1.10")
	assert.NoError(t, err)

	// when
	input, err := ibf.ForPlan(broker.GcpPlanID, "")

	// Then
	assert.NoError(t, err)
	assert.IsType(t, &RuntimeInput{}, input)
}
