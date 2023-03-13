package types

type Status string

const (
	StatusUp   Status = "Up"
	StatusDown Status = "Down"
)

type HealthStatus struct {
	Storage Status `json:"storageStatus"`
}
