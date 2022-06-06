package tenantfetcher

// EventsType missing godoc
type EventsType int

const (
	// CreatedAccountType missing godoc
	CreatedAccountType EventsType = iota
	// DeletedAccountType missing godoc
	DeletedAccountType
	// UpdatedAccountType missing godoc
	UpdatedAccountType
	// CreatedSubaccountType missing godoc
	CreatedSubaccountType
	// DeletedSubaccountType missing godoc
	DeletedSubaccountType
	// UpdatedSubaccountType missing godoc
	UpdatedSubaccountType
	// MovedSubaccountType missing godoc
	MovedSubaccountType
)

// String missing godoc
func (e EventsType) String() string {
	switch e {
	case CreatedAccountType:
		return "CreatedEventsType"
	case DeletedAccountType:
		return "DeletedEventsType"
	case UpdatedAccountType:
		return "UpdatedEventsType"
	case CreatedSubaccountType:
		return "CreatedSubaccountType"
	case DeletedSubaccountType:
		return "DeletedSubaccountType"
	case UpdatedSubaccountType:
		return "UpdatedSubaccountType"
	case MovedSubaccountType:
		return "MovedSubaccountType"
	default:
		return ""
	}
}

// TenantEventsResponse missing godoc
type TenantEventsResponse []byte
