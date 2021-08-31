package tenant

import (
	"database/sql"
)

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

type Type string

const (
	Unknown    Type = "unknown"
	Customer   Type = "customer"
	Account    Type = "account"
	Subaccount Type = "subaccount"
)

type Status string

const (
	Active   Status = "Active"
	Inactive Status = "Inactive"
)

type EntityCollection []Entity

func (a EntityCollection) Len() int {
	return len(a)
}

func (e Entity) WithStatus(status Status) Entity {
	e.Status = status
	return e
}

func StrToType(value string) Type {
	switch value {
	case string(Account):
		return Account
	case string(Customer):
		return Customer
	case string(Subaccount):
		return Subaccount
	default:
		return Unknown
	}
}

func TypeToStr(value Type) string {
	switch value {
	case Account:
		return string(Account)
	case Customer:
		return string(Customer)
	case Subaccount:
		return string(Subaccount)
	default:
		return string(Unknown)
	}
}
