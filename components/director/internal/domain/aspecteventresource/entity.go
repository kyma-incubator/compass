package aspecteventresource

import (
	"database/sql"
	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/kyma-incubator/compass/components/director/pkg/resource"
)

// Entity is a representation of an Aspect in the database.
type Entity struct {
	ApplicationID                sql.NullString `db:"app_id"`
	ApplicationTemplateVersionID sql.NullString `db:"app_template_version_id"`
	AspectID                     string         `db:"integration_dependency_id"`
	OrdID                        string         `db:"ord_id"`
	MinVersion                   sql.NullString `db:"min_version"`
	Subset                       sql.NullString `db:"subset"`

	*repo.BaseEntity
}

// GetParent returns the parent type and the parent ID of the entity.
func (e *Entity) GetParent(_ resource.Type) (resource.Type, string) {
	if e.AspectID != "" {
		return resource.AspectEventResource, e.AspectID
	}

	return "", ""
}
