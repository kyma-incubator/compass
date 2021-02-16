package tombstone_test

import (
	"database/sql/driver"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	tenantID         = "tenantID"
	appID            = "appID"
	ordID            = "com.compass.v1"
	externalTenantID = "externalTenantID"
)

func fixEntityTombstone() *tombstone.Entity {
	return &tombstone.Entity{
		OrdID:         ordID,
		TenantID:      tenantID,
		ApplicationID: appID,
		RemovalDate:   "removalDate",
	}
}

func fixTombstoneModel() *model.Tombstone {
	return &model.Tombstone{
		OrdID:         ordID,
		TenantID:      tenantID,
		ApplicationID: appID,
		RemovalDate:   "removalDate",
	}
}

func fixTombstoneModelInput() *model.TombstoneInput {
	return &model.TombstoneInput{
		OrdID:         ordID,
		TenantID:      tenantID,
		ApplicationID: appID,
		RemovalDate:   "removalDate",
	}
}

func fixTombstoneColumns() []string {
	return []string{"ord_id", "tenant_id", "app_id", "removal_date"}
}

func fixTombstoneRow() []driver.Value {
	return []driver.Value{ordID, tenantID, appID, "removalDate"}
}

func fixTombstoneUpdateArgs() []driver.Value {
	return []driver.Value{"removalDate"}
}
