package tombstone_test

import (
	"database/sql/driver"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	tombstoneID      = "tombstoneID"
	tenantID         = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	appID            = "appID"
	ordID            = "com.compass.v1"
	externalTenantID = "externalTenantID"
)

func fixEntityTombstone() *tombstone.Entity {
	return fixEntityTombstoneWithID(tombstoneID)
}
func fixEntityTombstoneWithID(id string) *tombstone.Entity {
	return &tombstone.Entity{
		ID:            id,
		OrdID:         ordID,
		ApplicationID: appID,
		RemovalDate:   "removalDate",
	}
}

func fixTombstoneModel() *model.Tombstone {
	return fixTombstoneModelWithID(tombstoneID)
}

func fixTombstoneModelWithID(id string) *model.Tombstone {
	return &model.Tombstone{
		ID:            id,
		OrdID:         ordID,
		ApplicationID: appID,
		RemovalDate:   "removalDate",
	}
}

func fixTombstoneModelInput() *model.TombstoneInput {
	return &model.TombstoneInput{
		OrdID:       ordID,
		RemovalDate: "removalDate",
	}
}

func fixTombstoneColumns() []string {
	return []string{"ord_id", "app_id", "removal_date", "id"}
}

func fixTombstoneRow() []driver.Value {
	return fixTombstoneRowWithID(tombstoneID)
}

func fixTombstoneRowWithID(id string) []driver.Value {
	return []driver.Value{ordID, appID, "removalDate", id}
}

func fixTombstoneUpdateArgs() []driver.Value {
	return []driver.Value{"removalDate"}
}
