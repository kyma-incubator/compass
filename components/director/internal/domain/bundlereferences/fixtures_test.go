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
	bundleRefID    = "qqqqqqqqq-qqqq-qqqq-qqqq-qqqqqqqqqqqq"
	apiDefID       = "ddddddddd-dddd-dddd-dddd-dddddddddddd"
	eventDefID     = "eeeeeeeee-eeee-eeee-eeee-eeeeeeeeeeee"
	bundleID       = "bbbbbbbbb-bbbb-bbbb-bbbb-bbbbbbbbbbbb"
	secondBundleID = "ppppppppp-pppp-pppp-pppp-pppppppppppp"

	apiDefTargetURL = "http://test.com"
)

func fixAPIBundleReferenceModel() model.BundleReference {
	return model.BundleReference{
		ID:                  bundleRefID,
		BundleID:            str.Ptr(bundleID),
		ObjectType:          model.BundleAPIReference,
		ObjectID:            str.Ptr(apiDefID),
		APIDefaultTargetURL: str.Ptr(apiDefTargetURL),
		IsDefaultBundle:     false,
	}
}

func fixAPIBundleReferenceEntity() bundlereferences.Entity {
	return bundlereferences.Entity{
		ID:                  bundleRefID,
		BundleID:            repo.NewValidNullableString(bundleID),
		APIDefID:            repo.NewValidNullableString(apiDefID),
		EventDefID:          sql.NullString{},
		APIDefaultTargetURL: repo.NewValidNullableString(apiDefTargetURL),
		IsDefaultBundle:     false,
	}
}

func fixAPIBundleReferenceEntityWithArgs(bndlID, apiID, targetURL string) bundlereferences.Entity {
	return bundlereferences.Entity{
		ID:                  bundleRefID,
		BundleID:            repo.NewValidNullableString(bndlID),
		APIDefID:            repo.NewValidNullableString(apiID),
		EventDefID:          sql.NullString{},
		APIDefaultTargetURL: repo.NewValidNullableString(targetURL),
		IsDefaultBundle:     false,
	}
}

func fixInvalidAPIBundleReferenceEntity() bundlereferences.Entity {
	return bundlereferences.Entity{
		ID:                  bundleRefID,
		BundleID:            repo.NewValidNullableString(bundleID),
		APIDefID:            sql.NullString{},
		EventDefID:          sql.NullString{},
		APIDefaultTargetURL: repo.NewValidNullableString(apiDefTargetURL),
		IsDefaultBundle:     false,
	}
}

func fixEventBundleReferenceModel() model.BundleReference {
	return model.BundleReference{
		ID:              bundleRefID,
		BundleID:        str.Ptr(bundleID),
		ObjectType:      model.BundleEventReference,
		ObjectID:        str.Ptr(eventDefID),
		IsDefaultBundle: false,
	}
}

func fixEventBundleReferenceEntity() bundlereferences.Entity {
	return bundlereferences.Entity{
		ID:                  bundleRefID,
		BundleID:            repo.NewValidNullableString(bundleID),
		APIDefID:            sql.NullString{},
		EventDefID:          repo.NewValidNullableString(eventDefID),
		APIDefaultTargetURL: sql.NullString{},
		IsDefaultBundle:     false,
	}
}

func fixEventBundleReferenceEntityWithArgs(bndlID, eventID string) bundlereferences.Entity {
	return bundlereferences.Entity{
		ID:              bundleRefID,
		BundleID:        repo.NewValidNullableString(bndlID),
		EventDefID:      repo.NewValidNullableString(eventID),
		IsDefaultBundle: false,
	}
}

func fixInvalidEventBundleReferenceEntity() bundlereferences.Entity {
	return bundlereferences.Entity{
		ID:                  bundleRefID,
		BundleID:            repo.NewValidNullableString(bundleID),
		APIDefID:            sql.NullString{},
		EventDefID:          sql.NullString{},
		APIDefaultTargetURL: sql.NullString{},
		IsDefaultBundle:     false,
	}
}

func fixBundleReferenceColumns() []string {
	return []string{"api_def_id", "event_def_id", "bundle_id", "api_def_url", "id", "is_default_bundle"}
}

func fixBundleReferenceRowWithoutEventID() []driver.Value {
	return []driver.Value{repo.NewValidNullableString(apiDefID), sql.NullString{}, repo.NewValidNullableString(bundleID), repo.NewValidNullableString(apiDefTargetURL), bundleRefID, false}
}

func fixBundleReferenceRowWithoutEventIDWithArgs(bndlID, apiID, targetURL string) []driver.Value {
	return []driver.Value{repo.NewValidNullableString(apiID), sql.NullString{}, repo.NewValidNullableString(bndlID), repo.NewValidNullableString(targetURL), bundleRefID, false}
}

func fixBundleReferenceRowWithoutAPIIDWithArgs(bndlID, eventID string) []driver.Value {
	return []driver.Value{sql.NullString{}, repo.NewValidNullableString(eventID), repo.NewValidNullableString(bndlID), sql.NullString{}, bundleRefID, false}
}

func fixBundleIDs(id string) []driver.Value {
	return []driver.Value{repo.NewValidNullableString(id)}
}

func fixBundleReferenceCreateArgs(bRef *model.BundleReference) []driver.Value {
	return []driver.Value{repo.NewValidNullableString(*bRef.ObjectID), sql.NullString{}, repo.NewValidNullableString(*bRef.BundleID), repo.NewValidNullableString(*bRef.APIDefaultTargetURL), bRef.ID, bRef.IsDefaultBundle}
}
