package tenantfetcher

type EventsType int

const (
	CreatedEventsType EventsType = iota
	DeletedEventsType
	UpdatedEventsType

	//TODO: Think for better naming
	MovedEventType
)

type TenantEventsResponse []byte
