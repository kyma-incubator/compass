package tenantfetcher

type EventsType int

const (
	CreatedEventsType EventsType = iota
	DeletedEventsType
	UpdatedEventsType

	//TODO: Think for better naming
	MovedSubAccountEventsType
)

type TenantEventsResponse []byte
