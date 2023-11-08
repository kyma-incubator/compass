package expectations_builders

import (
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
)

type FAExpectationsBuilder struct {
	// target: source: state
	expectations map[string]map[string]fixtures.AssignmentState
}

func NewFAExpectationsBuilder() *FAExpectationsBuilder {
	return &FAExpectationsBuilder{
		expectations: map[string]map[string]fixtures.AssignmentState{},
	}
}

func (b *FAExpectationsBuilder) GetExpectations() map[string]map[string]fixtures.AssignmentState {
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
	if _, ok := b.expectations[newParticipantID]; ok {
		return b
	}

	// add records for assignments where the new participant is target and a participant that was already added to the expectations structure is source
	for _, expectationsForPreviouslyAddedParticipantAsSource := range b.expectations {
		expectationsForPreviouslyAddedParticipantAsSource[newParticipantID] = fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}
	}

	currentParticipantIDs := b.getCurrentParticipantIDs()
	// add expectations where the newly added participant is source
	b.expectations[newParticipantID] = make(map[string]fixtures.AssignmentState, len(currentParticipantIDs)+1)
	// add record for the loop assignment
	b.expectations[newParticipantID][newParticipantID] = fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}
	// add records for assignments where the new participant is source and the target is a participant that was already added to the expectations structure
	for _, previouslyAddedParticipantID := range currentParticipantIDs {
		b.expectations[newParticipantID][previouslyAddedParticipantID] = fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}
	}

	return b
}

func (b *FAExpectationsBuilder) WithNotifications(notifications []*NotificationData) *FAExpectationsBuilder {
	for _, notification := range notifications {
		b.expectations[notification.SourceID][notification.TargetID] = notification.getAssignmentState()
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
