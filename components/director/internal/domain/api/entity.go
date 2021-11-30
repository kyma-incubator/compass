package api

import (
	"database/sql"

	"github.com/kyma-incubator/compass/components/director/internal/domain/version"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
)

// Entity is a representation of a API in the database.
type Entity struct {
	ApplicationID                           string         `db:"app_id"`
	PackageID                               sql.NullString `db:"package_id"`
	Name                                    string         `db:"name"`
	Description                             sql.NullString `db:"description"`
	Group                                   sql.NullString `db:"group_name"`
	TargetURLs                              sql.NullString `db:"target_urls"`
	OrdID                                   sql.NullString `db:"ord_id"`
	ShortDescription                        sql.NullString `db:"short_description"`
	SystemInstanceAware                     sql.NullBool   `db:"system_instance_aware"`
	APIProtocol                             sql.NullString `db:"api_protocol"`
	Tags                                    sql.NullString `db:"tags"`
	Countries                               sql.NullString `db:"countries"`
	Links                                   sql.NullString `db:"links"`
	APIResourceLinks                        sql.NullString `db:"api_resource_links"`
	ReleaseStatus                           sql.NullString `db:"release_status"`
	SunsetDate                              sql.NullString `db:"sunset_date"`
	Successors                              sql.NullString `db:"successors"`
	ChangeLogEntries                        sql.NullString `db:"changelog_entries"`
	Labels                                  sql.NullString `db:"labels"`
	Visibility                              sql.NullString `db:"visibility"`
	Disabled                                sql.NullBool   `db:"disabled"`
	PartOfProducts                          sql.NullString `db:"part_of_products"`
	LineOfBusiness                          sql.NullString `db:"line_of_business"`
	Industry                                sql.NullString `db:"industry"`
	ImplementationStandard                  sql.NullString `db:"implementation_standard"`
	CustomImplementationStandard            sql.NullString `db:"custom_implementation_standard"`
	CustomImplementationStandardDescription sql.NullString `db:"custom_implementation_standard_description"`
	Extensible                              sql.NullString `db:"extensible"`
	ResourceHash                            sql.NullString `db:"resource_hash"`

	*repo.BaseEntity
	version.Version
}

// GetParentID returns the parent ID of the entity.
func (e *Entity) GetParentID() string {
	return e.ApplicationID
}

// DecorateWithTenantID decorates the entity with the given tenant ID.
func (e *Entity) DecorateWithTenantID(tenant string) interface{} {
	return struct {
		*Entity
		TenantID string `db:"tenant_id"`
	}{
		Entity:   e,
		TenantID: tenant,
	}
}
