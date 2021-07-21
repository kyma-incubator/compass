package tombstone_test

import (
	"database/sql/driver"
	"fmt"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	tombstoneID      = "tombstoneID"
	tenantID         = "tenantID"
	appID            = "appID"
	ordID            = "com.compass.v1"
	externalTenantID = "externalTenantID"
)

func fixEntityTombstone() *tombstone.Entity {
	return &tombstone.Entity{
		ID:            tombstoneID,
		OrdID:         ordID,
		TenantID:      tenantID,
		ApplicationID: appID,
		RemovalDate:   "removalDate",
	}
}

func fixTombstoneModel() *model.Tombstone {
	return &model.Tombstone{
		ID:            tombstoneID,
		OrdID:         ordID,
		TenantID:      tenantID,
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
	return []string{"ord_id", "tenant_id", "app_id", "removal_date", "id"}
}

func fixTombstoneRow() []driver.Value {
	return []driver.Value{ordID, tenantID, appID, "removalDate", tombstoneID}
}

func fixTombstoneUpdateArgs() []driver.Value {
	return []driver.Value{"removalDate"}
}

func fixUpdateTenantIsolationSubquery() string {
	return `tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = ? UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`
}

func fixTenantIsolationSubquery() string {
	return fixTenantIsolationSubqueryWithArg(1)
}

func fixUnescapedTenantIsolationSubquery() string {
	return fixUnescapedTenantIsolationSubqueryWithArg(1)
}

func fixTenantIsolationSubqueryWithArg(i int) string {
	return regexp.QuoteMeta(fmt.Sprintf(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $%d UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`, i))
}

func fixUnescapedTenantIsolationSubqueryWithArg(i int) string {
	return fmt.Sprintf(`tenant_id IN ( with recursive children AS (SELECT t1.id, t1.parent FROM business_tenant_mappings t1 WHERE id = $%d UNION ALL SELECT t2.id, t2.parent FROM business_tenant_mappings t2 INNER JOIN children t on t.id = t2.parent) SELECT id from children )`, i)
}
