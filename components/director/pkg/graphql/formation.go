package graphql

// Formation missing godoc
type Formation struct {
	ID                  string `json:"id"`
	Name                string `json:"name"`
	FormationTemplateID string `json:"formationTemplateId"`
}

// FormationPageExt is an extended types used by external API
type FormationPageExt struct {
	FormationPage
	Data []*FormationExt `json:"data"`
}

// FormationExt  is an extended types used by external API
type FormationExt struct {
	Formation
	FormationAssignment  FormationAssignment     `json:"formationAssignment"`
	FormationAssignments FormationAssignmentPage `json:"formationAssignments"`
}
