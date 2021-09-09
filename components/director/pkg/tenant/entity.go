package tenant

import (
	"database/sql"
)

// Entity missing godoc
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
	// Unknown missing godoc
	Unknown Type = "unknown"
	// Account missing godoc
	Account Type = "account"
	// Customer missing godoc
	Customer Type = "customer"
)

// Status missing godoc
type Status string

const (
	// Active missing godoc
	Active Status = "Active"
	// Inactive missing godoc
	Inactive Status = "Inactive"
)

// EntityCollection missing godoc
type EntityCollection []Entity

// Len missing godoc
func (a EntityCollection) Len() int {
	return len(a)
}

// WithStatus missing godoc
func (e Entity) WithStatus(status Status) Entity {
	e.Status = status
	return e
}

// StrToType missing godoc
func StrToType(value string) Type {
	switch value {
	case string(Account):
		return Account
	case string(Customer):
		return Customer
	default:
		return Unknown
	}
}

// TypeToStr missing godoc
func TypeToStr(value Type) string {
	switch value {
	case Account:
		return string(Account)
	case Customer:
		return string(Customer)
	default:
		return string(Unknown)
	}
}
