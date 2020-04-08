package stages

import (
	"github.com/kyma-incubator/compass/components/provisioner/internal/model"
	"github.com/kyma-incubator/compass/components/provisioner/internal/persistence/dberrors"
	"github.com/kyma-incubator/compass/components/provisioner/internal/provisioning/persistence/dbsession/mocks"
	"github.com/sirupsen/logrus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"testing"
	"time"
)

func TestNewUpdateUpgradeStateStep(t *testing.T) {
	operationID := "82f6c076-be77-49bb-acec-50abccc37c72"

	t.Run("should return next step when finished", func(t *testing.T) {
		//given
		session := &mocks.WriteSession{}
		session.On("UpdateUpgradeState", operationID, model.UpgradeSucceeded).Return(nil)

		step := NewUpdateUpgradeStateStep(session, nextStageName, time.Minute)

		//when
		result, err := step.Run(model.Cluster{}, model.Operation{ID: operationID}, logrus.New())

		//then
		require.NoError(t, err)
		assert.Equal(t, nextStageName, result.Stage)
		assert.Equal(t, time.Duration(0), result.Delay)
	})

	t.Run("should return error when upgrade fails", func(t *testing.T) {
		//given
		session := &mocks.WriteSession{}
		session.On("UpdateUpgradeState", operationID, model.UpgradeSucceeded).Return(dberrors.NotFound("oh noes"))

		step := NewUpdateUpgradeStateStep(session, nextStageName, time.Minute)

		//when
		_, err := step.Run(model.Cluster{}, model.Operation{ID: operationID}, logrus.New())

		//then
		require.Error(t, err)
	})
}
