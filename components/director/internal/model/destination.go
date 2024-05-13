package model

// DestinationInput missing godoc
type DestinationInput struct {
	Name              string `json:"Name"`
	Type              string `json:"Type"`
	URL               string `json:"URL"`
	Authentication    string `json:"Authentication"`
	XCorrelationID    string `json:"correlationIds"`
	XSystemTenantID   string `json:"x-system-id"`
	XSystemTenantName string `json:"x-system-name"`
	XSystemType       string `json:"x-system-type"`
	XSystemBaseURL    string `json:"x-system-base-url"`
}

// Destination is an internal model representation of the destination entity
type Destination struct {
	ID                    string  `json:"id"`
	Name                  string  `json:"name"`
	Type                  string  `json:"type"`
	URL                   string  `json:"url"`
	Authentication        string  `json:"authentication"`
	SubaccountID          string  `json:"subaccount_id"`
	InstanceID            *string `json:"instanceId"`
	FormationAssignmentID *string `json:"formationAssignmentID"`
}

// HasValidIdentifiers checks if the destination has either one of the pairs: XSystemTenantID and XSystemType, or XSystemBaseURL and XSystemTenantName
func (d *DestinationInput) HasValidIdentifiers() bool {
	hasSystemIDAndType := d.XSystemTenantID != "" && d.XSystemType != ""
	hasNameAndURL := d.XSystemBaseURL != "" && d.XSystemTenantName != ""

	return hasSystemIDAndType || hasNameAndURL
}
