package systemssync_test

import (
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/domain/systemssync"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/pkg/errors"
)

var (
	syncID        = "684aa2a7-3b96-4374-936a-bb758d631b6b"
	syncTenantID  = "11111111-2222-3333-4444-555555555555"
	syncProductID = "PR"
	lastSyncTime  = time.Now()
	testError     = errors.New("test error")
)

func fixSystemsSyncModel(id, tenantID, productId string, lastSyncTimestamp time.Time) *model.SystemSynchronizationTimestamp {
	return &model.SystemSynchronizationTimestamp{
		ID:                id,
		TenantID:          tenantID,
		ProductID:         productId,
		LastSyncTimestamp: lastSyncTimestamp,
	}
}

func fixSystemsSyncEntity(id, tenantID, productId string, lastSyncTimestamp time.Time) *systemssync.Entity {
	return &systemssync.Entity{
		ID:                id,
		TenantID:          tenantID,
		ProductID:         productId,
		LastSyncTimestamp: lastSyncTimestamp,
	}
}
