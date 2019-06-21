package graphql

type Application struct {
	ID             string             `json:"id"`
	Tenant         Tenant             `json:"tenant"`
	Name           string             `json:"name"`
	Description    *string            `json:"description"`
	Labels         Labels             `json:"labels"`
	Annotations    Annotations        `json:"annotations"`
	Status         *ApplicationStatus `json:"status"`
	HealthCheckURL *string            `json:"healthCheckURL"`
}
