package repo

const (
	// M2MTenantIDColumn is the column name of the tenant_id in each tenant access table / view.
	M2MTenantIDColumn = "tenant_id"
	// M2MResourceIDColumn is the column name of the resource id in each tenant access table / view.
	M2MResourceIDColumn = "id"
	// M2MOwnerColumn is the column name of the owner in each tenant access table / view.
	M2MOwnerColumn = "owner"
)

// M2MColumns are the column names of the tenant access tables / views.
var M2MColumns = []string{M2MTenantIDColumn, M2MResourceIDColumn, M2MOwnerColumn}

// TenantAccess represents the tenant access table/views that are used for tenant isolation queries.
type TenantAccess struct {
	TenantID   string `db:"tenant_id"`
	ResourceID string `db:"id"`
	Owner      bool   `db:"owner"`
}

// TenantAccessCollection is a wrapper type for slice of entities.
type TenantAccessCollection []TenantAccess

// Len returns the current number of entities in the collection.
func (tc TenantAccessCollection) Len() int {
	return len(tc)
}
