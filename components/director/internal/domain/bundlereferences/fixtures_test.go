package bundlereferences_test

import (
	"database/sql"
	"database/sql/driver"
	"fmt"
	"regexp"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundlereferences"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
)

const (
	apiDefID         = "ddddddddd-dddd-dddd-dddd-dddddddddddd"
	eventDefID       = "eeeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	tenantID         = "ttttttttt-tttt-tttt-tttt-tttttttttttt"
	externalTenantID = "xxxxxxxxx-xxxx-xxxx-xxxx-xxxxxxxxxxxx"
	bundleID         = "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	secondBundleID   = "ppppppppp-pppp-pppp-pppp-pppppppppppp"

	apiDefTargetURL = "http://test.com"
)

func fixAPIBundleReferenceModel() model.BundleReference {
	return model.BundleReference{
		Tenant:              tenantID,
		BundleID:            str.Ptr(bundleID),
		ObjectType:          model.BundleAPIReference,
		ObjectID:            str.Ptr(apiDefID),
		APIDefaultTargetURL: str.Ptr(apiDefTargetURL),
	}
}

func fixAPIBundleReferenceEntity() bundlereferences.Entity {
	return bundlereferences.Entity{
		TenantID:            tenantID,
		BundleID:            repo.NewValidNullableString(bundleID),
		APIDefID:            repo.NewValidNullableString(apiDefID),
		EventDefID:          sql.NullString{},
		APIDefaultTargetURL: repo.NewValidNullableString(apiDefTargetURL),
	}
}

func fixInvalidAPIBundleReferenceEntity() bundlereferences.Entity {
	return bundlereferences.Entity{
		TenantID:            tenantID,
		BundleID:            repo.NewValidNullableString(bundleID),
		APIDefID:            sql.NullString{},
		EventDefID:          sql.NullString{},
		APIDefaultTargetURL: repo.NewValidNullableString(apiDefTargetURL),
	}
}

func fixEventBundleReferenceModel() model.BundleReference {
	return model.BundleReference{
		Tenant:     tenantID,
		BundleID:   str.Ptr(bundleID),
		ObjectType: model.BundleEventReference,
		ObjectID:   str.Ptr(eventDefID),
	}
}

func fixEventBundleReferenceEntity() bundlereferences.Entity {
	return bundlereferences.Entity{
		TenantID:            tenantID,
		BundleID:            repo.NewValidNullableString(bundleID),
		APIDefID:            sql.NullString{},
		EventDefID:          repo.NewValidNullableString(eventDefID),
		APIDefaultTargetURL: sql.NullString{},
	}
}

func fixInvalidEventBundleReferenceEntity() bundlereferences.Entity {
	return bundlereferences.Entity{
		TenantID:            tenantID,
		BundleID:            repo.NewValidNullableString(bundleID),
		APIDefID:            sql.NullString{},
		EventDefID:          sql.NullString{},
		APIDefaultTargetURL: sql.NullString{},
	}
}

func fixBundleReferenceColumns() []string {
	return []string{"tenant_id", "api_def_id", "event_def_id", "bundle_id", "api_def_url"}
}

func fixBundleReferenceRowWithoutEventID() []driver.Value {
	return []driver.Value{tenantID, repo.NewValidNullableString(apiDefID), sql.NullString{}, repo.NewValidNullableString(bundleID), repo.NewValidNullableString(apiDefTargetURL)}
}

func fixBundleIDs(id string) []driver.Value {
	return []driver.Value{repo.NewValidNullableString(id)}
}

func fixBundleReferenceCreateArgs(bRef *model.BundleReference) []driver.Value {
	return []driver.Value{bRef.Tenant, repo.NewValidNullableString(*bRef.ObjectID), sql.NullString{}, repo.NewValidNullableString(*bRef.BundleID), repo.NewValidNullableString(*bRef.APIDefaultTargetURL)}
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
