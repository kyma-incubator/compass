package graphql

type Application struct {
	ID             string             `json:"id"`
	Name           string             `json:"name"`
	Description    *string            `json:"description"`
	Labels         Labels             `json:"labels"`
	Status         *ApplicationStatus `json:"status"`
	HealthCheckURL *string            `json:"healthCheckURL"`
}
