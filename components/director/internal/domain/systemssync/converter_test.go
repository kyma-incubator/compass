package systemssync_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemssync"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToEntity(t *testing.T) {
	systemsSyncModel := fixSystemsSyncModel(syncID, syncTenantID, syncProductID, lastSyncTime)
	systemsSyncEntity := fixSystemsSyncEntity(syncID, syncTenantID, syncProductID, lastSyncTime)
	testCases := []struct {
		Name     string
		Input    *model.SystemSynchronizationTimestamp
		Expected *systemssync.Entity
	}{
		{
			Name:     "All properties given",
			Input:    systemsSyncModel,
			Expected: systemsSyncEntity,
		},
		{
			Name:     "Empty",
			Input:    &model.SystemSynchronizationTimestamp{},
			Expected: &systemssync.Entity{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := systemssync.NewConverter()

			// WHEN
			res := conv.ToEntity(testCase.Input)

			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	systemsSyncModel := fixSystemsSyncModel(syncID, syncTenantID, syncProductID, lastSyncTime)
	systemsSyncEntity := fixSystemsSyncEntity(syncID, syncTenantID, syncProductID, lastSyncTime)

	testCases := []struct {
		Name     string
		Input    *systemssync.Entity
		Expected *model.SystemSynchronizationTimestamp
	}{
		{
			Name:     "All properties given",
			Input:    systemsSyncEntity,
			Expected: systemsSyncModel,
		},
		{
			Name:     "Empty",
			Input:    &systemssync.Entity{},
			Expected: &model.SystemSynchronizationTimestamp{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := systemssync.NewConverter()

			// WHEN
			res := conv.FromEntity(testCase.Input)

			assert.Equal(t, testCase.Expected, res)
		})
	}
}
