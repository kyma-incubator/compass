package formationconstraint

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/webhook"
)

// MatchingDetails contains information used to match the reached join point with the applicable constraints
type MatchingDetails struct {
	ResourceType    model.ResourceType
	ResourceSubtype string
}

// JoinPointDetails provides an interface for join point details
type JoinPointDetails interface {
	GetMatchingDetails() MatchingDetails
}

// CRUDFormationOperationDetails contains details applicable to createFormation and deleteFormation join points
type CRUDFormationOperationDetails struct {
	FormationType       string
	FormationTemplateID string
	FormationName       string
	TenantID            string
}

// GetMatchingDetails returns matching details for CRUDFormationOperationDetails
func (d *CRUDFormationOperationDetails) GetMatchingDetails() MatchingDetails {
	return MatchingDetails{
		ResourceType:    model.FormationResourceType,
		ResourceSubtype: d.FormationType,
	}
}

// AssignFormationOperationDetails contains details applicable to assignFormation join point
type AssignFormationOperationDetails struct {
	ResourceType        model.ResourceType
	ResourceSubtype     string
	ResourceID          string
	FormationType       string
	FormationTemplateID string
	FormationID         string
	TenantID            string
}

// GetMatchingDetails returns matching details for AssignFormationOperationDetails
func (d *AssignFormationOperationDetails) GetMatchingDetails() MatchingDetails {
	return MatchingDetails{
		ResourceType:    d.ResourceType,
		ResourceSubtype: d.ResourceSubtype,
	}
}

// UnassignFormationOperationDetails contains details applicable to unassignFormation join point
type UnassignFormationOperationDetails struct {
	ResourceType        model.ResourceType
	ResourceSubtype     string
	ResourceID          string
	FormationType       string
	FormationTemplateID string
	FormationID         string
	TenantID            string
}

// GetMatchingDetails returns matching details for UnassignFormationOperationDetails
func (d *UnassignFormationOperationDetails) GetMatchingDetails() MatchingDetails {
	return MatchingDetails{
		ResourceType:    d.ResourceType,
		ResourceSubtype: d.ResourceSubtype,
	}
}

// GenerateNotificationOperationDetails contains details applicable to generateNotifications join point
type GenerateNotificationOperationDetails struct {
	Operation       model.FormationOperation
	FormationID     string
	ResourceType    model.ResourceType
	ResourceSubtype string
	ResourceID      string

	// fields used when generating notifications from configuration changed webhooks
	ApplicationTemplate *webhook.ApplicationTemplateWithLabels
	Application         *webhook.ApplicationWithLabels
	Runtime             *webhook.RuntimeWithLabels
	RuntimeContext      *webhook.RuntimeContextWithLabels
	Assignment          *webhook.FormationAssignment
	ReverseAssignment   *webhook.FormationAssignment

	// fields used when generating notification from application tenant mapping webhooks
	SourceApplicationTemplate *webhook.ApplicationTemplateWithLabels
	// SourceApplication is the application that the notification is about
	SourceApplication         *webhook.ApplicationWithLabels
	TargetApplicationTemplate *webhook.ApplicationTemplateWithLabels
	// TargetApplication is the application that the notification is for (the one with the webhook / the one receiving the notification)
	TargetApplication *webhook.ApplicationWithLabels

	TenantID string
}

// GetMatchingDetails returns matching details for GenerateNotificationOperationDetails
func (d *GenerateNotificationOperationDetails) GetMatchingDetails() MatchingDetails {
	return MatchingDetails{
		ResourceType:    d.ResourceType,
		ResourceSubtype: d.ResourceSubtype,
	}
}
