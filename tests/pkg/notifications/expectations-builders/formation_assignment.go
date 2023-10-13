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

func (b *FAExpectationsBuilder) getCurrentParticipants() []string {
	participants := make([]string, 0, len(b.expectations))
	for participantID, _ := range b.expectations {
		participants = append(participants, participantID)
	}

	return participants
}

func (b *FAExpectationsBuilder) WithParticipant(participantID string) *FAExpectationsBuilder {
	if _, ok := b.expectations[participantID]; ok {
		return b
	}

	for _, val := range b.expectations {
		val[participantID] = fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}
	}

	currentParticipants := b.getCurrentParticipants()
	b.expectations[participantID] = make(map[string]fixtures.AssignmentState, len(currentParticipants)+1)
	b.expectations[participantID][participantID] = fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}
	for _, participant := range currentParticipants {
		b.expectations[participantID][participant] = fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}
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

func (b *FAExpectationsBuilder) WithNotifications(notifications []*NotificationData) *FAExpectationsBuilder {
	for _, notification := range notifications {
		b.expectations[notification.TargetID][notification.SourceID] = notification.getAssignmentState()
	}
	return b
}
