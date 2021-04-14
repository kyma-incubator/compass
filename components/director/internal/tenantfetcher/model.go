package tenantfetcher

type EventsType int

const (
	CreatedEventsType EventsType = iota
	DeletedEventsType
	UpdatedEventsType
	MovedRuntimeByLabelEventsType
)

type TenantEventsResponse []byte
