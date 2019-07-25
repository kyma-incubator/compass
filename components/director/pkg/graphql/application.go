package graphql

type Application struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	Description    *string            `json:"description"`
	Status         *ApplicationStatus `json:"status"`
	HealthCheckURL *string            `json:"healthCheckURL"`
}
