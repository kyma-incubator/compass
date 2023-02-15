package model

import (
	"encoding/json"

	"github.com/kyma-incubator/compass/components/director/pkg/pagination"
)

// DefaultTemplateName will be used as default formation template name if no other options are provided
const DefaultTemplateName = "Side-by-side extensibility with Kyma"

// FormationState represents the possible states a formation can be in
type FormationState string

const (
	// InitialFormationState indicates that nothing has been done with the formation
	InitialFormationState FormationState = "INITIAL"
	// ReadyFormationState indicates that the formation is in a ready state
	ReadyFormationState FormationState = "READY"
	// CreateErrorFormationState indicates that an error occurred during the creation of the formation
	CreateErrorFormationState FormationState = "CREATE_ERROR"
	// DeleteErrorFormationState indicates that an error occurred during the deletion of the formation
	DeleteErrorFormationState FormationState = "DELETE_ERROR"
	// DeletingFormationState indicates that the formation is in deleting state
	DeletingFormationState FormationState = "DELETING"
)

// FormationOperation defines the kind of operation done on a given formation
type FormationOperation string

const (
	// AssignFormation represents the assign operation done on a given formation
	AssignFormation FormationOperation = "assign"
	// UnassignFormation represents the unassign operation done on a given formation
	UnassignFormation FormationOperation = "unassign"
	// CreateFormation represents the create operation on a given formation
	CreateFormation FormationOperation = "createFormation"
	// DeleteFormation represents the delete operation on a given formation
	DeleteFormation FormationOperation = "deleteFormation"
)

// Formation missing godoc
type Formation struct {
	ID                  string
	TenantID            string
	FormationTemplateID string
	Name                string
	State               FormationState
	Error               json.RawMessage
}

// FormationPage contains Formation data with page info
type FormationPage struct {
	Data       []*Formation
	PageInfo   *pagination.Page
	TotalCount int
}
