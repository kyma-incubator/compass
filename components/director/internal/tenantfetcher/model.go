package tenantfetcher

type EventsType int

const (
	CreatedEventsType EventsType = iota
	DeletedEventsType
	UpdatedEventsType
)

type TenantEventsResponse []byte
