package repo

const (
	M2MTenantIDColumn   = "tenant_id"
	M2MResourceIDColumn = "id"
	M2MOwnerColumn      = "owner"
)

var m2mColumns = []string{M2MTenantIDColumn, M2MResourceIDColumn, M2MOwnerColumn}

type TenantAccess struct {
	TenantID   string `db:"tenant_id"`
	ResourceID string `db:"id"`
	Owner      bool   `db:"owner"`
}
