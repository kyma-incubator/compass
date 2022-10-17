package tenant

import (
	"database/sql"
)

// Entity represents a Compass tenant.
type Entity struct {
	ID             string         `db:"id"`
	Name           string         `db:"external_name"`
	ExternalTenant string         `db:"external_tenant"`
	Parent         sql.NullString `db:"parent"`
	Type           Type           `db:"type"`
	ProviderName   string         `db:"provider_name"`
	Initialized    *bool          `db:"initialized"` // computed value
	Status         Status         `db:"status"`
}

// Type missing godoc
type Type string

const (
	// Unknown tenant type is used when the tenant type cannot be determined when the tenant's being created.
	Unknown Type = "unknown"
	// Customer tenants can be parents of account tenants.
	Customer Type = "customer"
	// Account tenant type may have a parent with type Customer.
	Account Type = "account"
	// Subaccount tenants must have a parent of type Account.
	Subaccount Type = "subaccount"
	// Organization tenants can be parents of Folder or ResourceGroup tenants.
	Organization Type = "organization"
	// Folder tenants must have a parent of type Organization.
	Folder Type = "folder"
	// ResourceGroup tenants must have a parent of type Folder or Organization.
	ResourceGroup Type = "resource-group"
)

// Status is used to determine if a tenant is currently being used or not.
type Status string

const (
	// Active status represents tenants, which are currently active and their resources can be operated.
	Active Status = "Active"
	// Inactive status represents tenants, whose resources cannot be operated.
	Inactive Status = "Inactive"
)

// EntityCollection is a wrapper type for slice of entities.
type EntityCollection []Entity

// Len returns the current number of entities in the collection.
func (a EntityCollection) Len() int {
	return len(a)
}

// WithStatus sets the provided status to the entity.
func (e Entity) WithStatus(status Status) Entity {
	e.Status = status
	return e
}

// StrToType returns the tenant Type value of the provided string or "Unknown" if there's no type matching the string.
func StrToType(value string) Type {
	switch value {
	case string(Account):
		return Account
	case string(Customer):
		return Customer
	case string(Subaccount):
		return Subaccount
	case string(Organization):
		return Organization
	case string(Folder):
		return Folder
	case string(ResourceGroup):
		return ResourceGroup
	default:
		return Unknown
	}
}

// TypeToStr returns the string value of the provided tenant Type.
func TypeToStr(value Type) string {
	switch value {
	case Account:
		return string(Account)
	case Customer:
		return string(Customer)
	case Subaccount:
		return string(Subaccount)
	case Organization:
		return string(Organization)
	case Folder:
		return string(Folder)
	case ResourceGroup:
		return string(ResourceGroup)
	default:
		return string(Unknown)
	}
}
