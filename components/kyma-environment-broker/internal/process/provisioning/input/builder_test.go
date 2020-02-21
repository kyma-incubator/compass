package input

import (
	"testing"

	"github.com/kyma-incubator/compass/components/kyma-environment-broker/internal/broker"

	"github.com/kyma-project/kyma/components/kyma-operator/pkg/apis/installer/v1alpha1"
	"github.com/stretchr/testify/assert"
)

// Currently on production only azure is supported

func TestInputBuilderFactory_IsPlanSupport(t *testing.T) {
	// given
	ibf := NewInputBuilderFactory(nil, []v1alpha1.KymaComponent{}, Config{}, "1.10")

	// when/then
	assert.True(t, ibf.IsPlanSupport(broker.GcpPlanID))
	assert.True(t, ibf.IsPlanSupport(broker.AzurePlanID))
	assert.False(t, ibf.IsPlanSupport(broker.AwsPlanID))
}

func TestInputBuilderFactory_ForPlan(t *testing.T) {
	// given
	ibf := NewInputBuilderFactory(nil, []v1alpha1.KymaComponent{}, Config{}, "1.10")

	// when
	input, found := ibf.ForPlan(broker.GcpPlanID)

	// Then
	assert.True(t, found)
	assert.IsType(t, &RuntimeInput{}, input)
}
