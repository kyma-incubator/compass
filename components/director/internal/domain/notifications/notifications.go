package notifications

import (
	"github.com/kyma-incubator/compass/components/director/internal/model"
	webhookclient "github.com/kyma-incubator/compass/components/director/pkg/webhook_client"
)

// FormationAssignmentRequestMapping represents the mapping between the notification request and formation assignment
type FormationAssignmentRequestMapping struct {
	Request             *webhookclient.FormationAssignmentNotificationRequest
	FormationAssignment *model.FormationAssignment
}

// Clone returns a copy of the FormationAssignmentRequestMapping
func (f *FormationAssignmentRequestMapping) Clone() *FormationAssignmentRequestMapping {
	var request *webhookclient.FormationAssignmentNotificationRequest
	if f.Request != nil {
		request = f.Request.Clone()
	}
	var formationAssignment *model.FormationAssignment
	if f.FormationAssignment != nil {
		formationAssignment = &model.FormationAssignment{
			ID:                            f.FormationAssignment.ID,
			FormationID:                   f.FormationAssignment.FormationID,
			TenantID:                      f.FormationAssignment.TenantID,
			Source:                        f.FormationAssignment.Source,
			SourceType:                    f.FormationAssignment.SourceType,
			Target:                        f.FormationAssignment.Target,
			TargetType:                    f.FormationAssignment.TargetType,
			State:                         f.FormationAssignment.State,
			Value:                         f.FormationAssignment.Value,
			Error:                         f.FormationAssignment.Error,
			LastStateChangeTimestamp:      f.FormationAssignment.LastStateChangeTimestamp,
			LastNotificationSentTimestamp: f.FormationAssignment.LastNotificationSentTimestamp,
		}
	}

	return &FormationAssignmentRequestMapping{
		Request:             request,
		FormationAssignment: formationAssignment,
	}
}

// AssignmentMappingPair represents a pair of FormationAssignmentRequestMapping and its reverse
type AssignmentMappingPair struct {
	AssignmentReqMapping        *FormationAssignmentRequestMapping
	ReverseAssignmentReqMapping *FormationAssignmentRequestMapping
}

// AssignmentMappingPairWithOperation represents an AssignmentMappingPair and the formation operation
type AssignmentMappingPairWithOperation struct {
	*AssignmentMappingPair
	Operation model.FormationOperation
}
