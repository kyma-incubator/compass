package expectations_builders

import (
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
)

type FAExpectationsBuilder struct {
	// target: source: state
	expectations map[string]map[string]fixtures.Assignment
}

func NewFAExpectationsBuilder() *FAExpectationsBuilder {
	return &FAExpectationsBuilder{
		expectations: map[string]map[string]fixtures.Assignment{},
	}
}

func (b *FAExpectationsBuilder) GetExpectations() map[string]map[string]fixtures.Assignment {
	return b.expectations
}

func (b *FAExpectationsBuilder) GetExpectedAssignmentsCount() int {
	count := 0
	for _, val := range b.expectations {
		count += len(val)
	}
	return count
}

func (b *FAExpectationsBuilder) getCurrentParticipantIDs() []string {
	participantIDs := make([]string, 0, len(b.expectations))
	for participantID, _ := range b.expectations {
		participantIDs = append(participantIDs, participantID)
	}

	return participantIDs
}

func (b *FAExpectationsBuilder) WithParticipant(newParticipantID string) *FAExpectationsBuilder {
	return b.WithParticipantAndStates(newParticipantID, "READY", "READY", "READY")
}

func (b *FAExpectationsBuilder) WithParticipantAndStates(newParticipantID, targetState, sourceState, selfState string) *FAExpectationsBuilder {
	if _, ok := b.expectations[newParticipantID]; ok {
		return b
	}

	// add records for assignments where the new participant is target and a participant that was already added to the expectations structure is source
	for previousParticipant, expectationsForPreviouslyAddedParticipantAsSource := range b.expectations {
		expectationsForPreviouslyAddedParticipantAsSource[newParticipantID] = fixtures.Assignment{
			AssignmentStatus: fixtures.AssignmentState{State: targetState, Config: nil, Value: nil, Error: nil},
			Operations: []*fixtures.Operation{{
				SourceID:    previousParticipant,
				TargetID:    newParticipantID,
				Type:        "ASSIGN",
				TriggeredBy: "ASSIGN_OBJECT",
				IsFinished:  true,
			}},
		}
	}

	currentParticipantIDs := b.getCurrentParticipantIDs()
	// add expectations where the newly added participant is source
	b.expectations[newParticipantID] = make(map[string]fixtures.Assignment, len(currentParticipantIDs)+1)
	// add record for the loop assignment
	b.expectations[newParticipantID][newParticipantID] = fixtures.Assignment{
		AssignmentStatus: fixtures.AssignmentState{State: selfState, Config: nil, Value: nil, Error: nil},
		Operations: []*fixtures.Operation{{
			SourceID:    newParticipantID,
			TargetID:    newParticipantID,
			Type:        "ASSIGN",
			TriggeredBy: "ASSIGN_OBJECT",
			IsFinished:  true,
		}},
	}
	// add records for assignments where the new participant is source and the target is a participant that was already added to the expectations structure
	for _, previouslyAddedParticipantID := range currentParticipantIDs {
		b.expectations[newParticipantID][previouslyAddedParticipantID] = fixtures.Assignment{
			AssignmentStatus: fixtures.AssignmentState{State: sourceState, Config: nil, Value: nil, Error: nil},
			Operations: []*fixtures.Operation{{
				SourceID:    newParticipantID,
				TargetID:    previouslyAddedParticipantID,
				Type:        "ASSIGN",
				TriggeredBy: "ASSIGN_OBJECT",
				IsFinished:  true,
			}},
		}
	}

	return b
}

func (b *FAExpectationsBuilder) WithCustomParticipants(newParticipantIDs []string) *FAExpectationsBuilder {
	for _, participant := range newParticipantIDs {
		b.expectations[participant] = make(map[string]fixtures.Assignment)
	}
	return b
}

func (b *FAExpectationsBuilder) WithNotifications(notifications []*NotificationData) *FAExpectationsBuilder {
	for _, notification := range notifications {
		assignmentExpectation := b.expectations[notification.SourceID][notification.TargetID]
		assignmentExpectation.AssignmentStatus = notification.getAssignmentState()
		b.expectations[notification.SourceID][notification.TargetID] = assignmentExpectation
	}
	return b
}

func (b *FAExpectationsBuilder) WithOperations(operations []*fixtures.Operation) *FAExpectationsBuilder {
	// Remove operations generated from adding the participant to the expectations only for the assignments where operations are provided externally.
	// We can not directly replace as we can have two or more operations provided externally for one and the same assignment(in case of successful assign and then reset)
	for _, operation := range operations {
		assignmentExpectation := b.expectations[operation.SourceID][operation.TargetID]
		assignmentExpectation.Operations = nil
		b.expectations[operation.SourceID][operation.TargetID] = assignmentExpectation
	}

	// add the new operations
	for _, operation := range operations {
		assignmentExpectation := b.expectations[operation.SourceID][operation.TargetID]
		assignmentExpectation.Operations = append(assignmentExpectation.Operations, operation)
		b.expectations[operation.SourceID][operation.TargetID] = assignmentExpectation
	}
	return b
}

type NotificationData struct {
	SourceID string
	TargetID string
	State    string
	Config   *string
	Error    *string
}

func NewNotificationData(sourceID string, targetID string, state string, config *string, error *string) *NotificationData {
	return &NotificationData{SourceID: sourceID, TargetID: targetID, State: state, Config: config, Error: error}
}

func (n *NotificationData) getAssignmentState() fixtures.AssignmentState {
	as := fixtures.AssignmentState{
		Config: n.Config,
		Value:  n.Config,
		Error:  n.Error,
		State:  n.State,
	}
	if n.Error != nil {
		as.Value = n.Error
	}

	return as
}
