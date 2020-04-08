package stages

import (
	"fmt"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/runtime/mocks"
	"github.com/kyma-incubator/compass/components/provisioner/internal/util"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

const (
	nextStageName model.OperationStage = "NextStage"
)

func TestConnectAgentStep_Run(t *testing.T) {

	cluster := model.Cluster{Kubeconfig: util.StringPtr("kubeconfig")}

	t.Run("should return next step when finished", func(t *testing.T) {
		// given
		configurator := &mocks.Configurator{}
		configurator.On("ConfigureRuntime", cluster, "kubeconfig").Return(nil)

		stage := NewConnectAgentStage(configurator, nextStageName, time.Minute)

		// when
		result, err := stage.Run(cluster, model.Operation{}, &logrus.Entry{})

		// then
		require.NoError(t, err)
		assert.Equal(t, nextStageName, result.Stage)
		assert.Equal(t, time.Duration(0), result.Delay)
	})

	t.Run("should return error when failed to configure cluster", func(t *testing.T) {
		// given
		configurator := &mocks.Configurator{}
		configurator.On("ConfigureRuntime", cluster, "kubeconfig").Return(fmt.Errorf("error"))

		stage := NewConnectAgentStage(configurator, nextStageName, time.Minute)

		// when
		_, err := stage.Run(cluster, model.Operation{}, &logrus.Entry{})

		// then
		require.Error(t, err)
	})

}
