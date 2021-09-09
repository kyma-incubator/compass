package tenantfetcher

// EventsType missing godoc
type EventsType int

const (
	// CreatedEventsType missing godoc
	CreatedEventsType EventsType = iota
	// DeletedEventsType missing godoc
	DeletedEventsType
	// UpdatedEventsType missing godoc
	UpdatedEventsType
	// MovedRuntimeByLabelEventsType missing godoc
	MovedRuntimeByLabelEventsType
)

// String missing godoc
func (e EventsType) String() string {
	switch e {
	case CreatedEventsType:
		return "CreatedEventsType"
	case DeletedEventsType:
		return "DeletedEventsType"
	case UpdatedEventsType:
		return "UpdatedEventsType"
	case MovedRuntimeByLabelEventsType:
		return "MovedRuntimeByLabelEventsType"
	default:
		return ""
	}
}

// TenantEventsResponse missing godoc
type TenantEventsResponse []byte
