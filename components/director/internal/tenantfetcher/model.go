package tenantfetcher

type EventsType int

const (
	CreatedEventsType EventsType = iota
	DeletedEventsType
	UpdatedEventsType
)

type TenantEventsResponse struct {
	Events       []byte
	TotalResults int
	TotalPages   int
}

type Event map[string]interface{}
