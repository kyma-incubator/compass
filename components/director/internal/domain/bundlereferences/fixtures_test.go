package bundlereferences_test

import (
	"database/sql"
	"database/sql/driver"

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

func fixAPIBundleReferenceEntityWithArgs(bndlID, apiID, targetURL string) bundlereferences.Entity {
	return bundlereferences.Entity{
		TenantID:            tenantID,
		BundleID:            repo.NewValidNullableString(bndlID),
		APIDefID:            repo.NewValidNullableString(apiID),
		EventDefID:          sql.NullString{},
		APIDefaultTargetURL: repo.NewValidNullableString(targetURL),
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

func fixEventBundleReferenceEntityWithArgs(bndlID, eventID string) bundlereferences.Entity {
	return bundlereferences.Entity{
		TenantID:   tenantID,
		BundleID:   repo.NewValidNullableString(bndlID),
		EventDefID: repo.NewValidNullableString(eventID),
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

func fixBundleReferenceRowWithoutEventIDWithArgs(bndlID, apiID, targetURL string) []driver.Value {
	return []driver.Value{tenantID, repo.NewValidNullableString(apiID), sql.NullString{}, repo.NewValidNullableString(bndlID), repo.NewValidNullableString(targetURL)}
}

func fixBundleReferenceRowWithoutAPIIDWithArgs(bndlID, eventID string) []driver.Value {
	return []driver.Value{tenantID, sql.NullString{}, repo.NewValidNullableString(eventID), repo.NewValidNullableString(bndlID), sql.NullString{}}
}

func fixBundleIDs(id string) []driver.Value {
	return []driver.Value{repo.NewValidNullableString(id)}
}

func fixBundleReferenceCreateArgs(bRef *model.BundleReference) []driver.Value {
	return []driver.Value{bRef.Tenant, repo.NewValidNullableString(*bRef.ObjectID), sql.NullString{}, repo.NewValidNullableString(*bRef.BundleID), repo.NewValidNullableString(*bRef.APIDefaultTargetURL)}
}
