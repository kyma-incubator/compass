package tombstone_test

import (
	"database/sql/driver"
	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/internal/domain/tombstone"
	"github.com/kyma-incubator/compass/components/director/internal/model"
)

const (
	tombstoneID      = "tombstoneID"
	tenantID         = "b91b59f7-2563-40b2-aba9-fef726037aa3"
	ordID            = "com.compass.v1"
	externalTenantID = "externalTenantID"
	description      = "desc"
)

var (
	appID                = "appID"
	appTemplateVersionID = "appTemplateVersionID"
)

func fixEntityTombstoneForApp() *tombstone.Entity {
	return fixEntityTombstoneWithIDForApp(tombstoneID)
}

func fixEntityTombstoneForAppTemplateVersion() *tombstone.Entity {
	return fixEntityTombstoneWithIDForAppTemplateVersion(tombstoneID)
}

func fixEntityTombstoneWithID(id string) *tombstone.Entity {
	return &tombstone.Entity{
		ID:          id,
		OrdID:       ordID,
		RemovalDate: "removalDate",
		Description: repo.NewValidNullableString(description),
	}
}

func fixEntityTombstoneWithIDForApp(id string) *tombstone.Entity {
	tombstone := fixEntityTombstoneWithID(id)
	tombstone.ApplicationID = repo.NewValidNullableString(appID)
	return tombstone
}

func fixEntityTombstoneWithIDForAppTemplateVersion(id string) *tombstone.Entity {
	tombstone := fixEntityTombstoneWithID(id)
	tombstone.ApplicationTemplateVersionID = repo.NewValidNullableString(appTemplateVersionID)
	return tombstone
}

func fixTombstoneModelForApp() *model.Tombstone {
	return fixTombstoneModelWithIDForApp(tombstoneID)
}

func fixTombstoneModelForAppTemplateVersion() *model.Tombstone {
	return fixTombstoneModelWithIDForAppTemplateVersion(tombstoneID)
}

func fixTombstoneModelWithID(id string) *model.Tombstone {
	return &model.Tombstone{
		ID:          id,
		OrdID:       ordID,
		RemovalDate: "removalDate",
		Description: str.Ptr(description),
	}
}

func fixTombstoneModelWithIDForApp(id string) *model.Tombstone {
	tombstone := fixTombstoneModelWithID(id)
	tombstone.ApplicationID = &appID
	return tombstone
}

func fixTombstoneModelWithIDForAppTemplateVersion(id string) *model.Tombstone {
	tombstone := fixTombstoneModelWithID(id)
	tombstone.ApplicationTemplateVersionID = &appTemplateVersionID
	return tombstone
}

func fixTombstoneModelInput() *model.TombstoneInput {
	return &model.TombstoneInput{
		OrdID:       ordID,
		RemovalDate: "removalDate",
		Description: str.Ptr(description),
	}
}

func fixTombstoneColumns() []string {
	return []string{"ord_id", "app_id", "app_template_version_id", "removal_date", "id", "description"}
}

func fixTombstoneRowForApp() []driver.Value {
	return fixTombstoneRowWithIDForApp(tombstoneID)
}

func fixTombstoneRowForAppTemplateVersion() []driver.Value {
	return fixTombstoneRowWithIDForAppTemplateVersion(tombstoneID)
}

func fixTombstoneRowWithIDForApp(id string) []driver.Value {
	return []driver.Value{ordID, appID, repo.NewValidNullableString(""), "removalDate", id, description}
}

func fixTombstoneRowWithIDForAppTemplateVersion(id string) []driver.Value {
	return []driver.Value{ordID, repo.NewValidNullableString(""), appTemplateVersionID, "removalDate", id, description}
}

func fixTombstoneUpdateArgs() []driver.Value {
	return []driver.Value{"removalDate", description}
}
