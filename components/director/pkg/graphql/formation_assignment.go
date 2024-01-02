package graphql

// FormationAssignment is a formation assignment graphQL type
type FormationAssignment struct {
	ID            string                  `json:"id"`
	Source        string                  `json:"source"`
	SourceType    FormationAssignmentType `json:"sourceType"`
	Target        string                  `json:"target"`
	TargetType    FormationAssignmentType `json:"targetType"`
	State         string                  `json:"state"`
	Value         *string                 `json:"value"`
	Configuration *string                 `json:"configuration"`
	Error         *string                 `json:"error"`
}

// FormationAssignmentPageExt is an extended types used by external API
type FormationAssignmentPageExt struct {
	FormationAssignmentPage
	Data []*FormationAssignmentExt `json:"data"`
}

// FormationAssignmentExt  is an extended types used by external API
type FormationAssignmentExt struct {
	FormationAssignment
	SourceEntity FormationParticipant
	TargetEntity FormationParticipant
}
