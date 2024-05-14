package graphql

// Formation missing godoc
type Formation struct {
	ID                            string         `json:"id"`
	Name                          string         `json:"name"`
	TenantID                      string         `json:"tenantID"`
	FormationTemplateID           string         `json:"formationTemplateId"`
	State                         string         `json:"state"`
	Error                         FormationError `json:"error"`
	LastStateChangeTimestamp      *Timestamp     `json:"lastStateChangeTimestamp"`
	LastNotificationSentTimestamp *Timestamp     `json:"lastNotificationSentTimestamp"`
}

// FormationPageExt is an extended types used by external API
type FormationPageExt struct {
	FormationPage
	Data []*FormationExt `json:"data"`
}

// FormationExt is an extended types used by external API
type FormationExt struct {
	Formation
	FormationAssignment  FormationAssignmentExt     `json:"formationAssignment"`
	FormationAssignments FormationAssignmentPageExt `json:"formationAssignments"`
	Status               FormationStatus            `json:"status"`
}
