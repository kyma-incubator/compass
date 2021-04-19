package tenantfetcher

type EventsType int

const (
	CreatedEventsType EventsType = iota
	DeletedEventsType
	UpdatedEventsType
	MovedRuntimeByLabelEventsType
)

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

type TenantEventsResponse []byte
