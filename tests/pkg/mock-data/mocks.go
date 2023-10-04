package mock_data

import (
	"github.com/kyma-incubator/compass/tests/pkg/fixtures"
	"k8s.io/utils/strings/slices"
)

type FANNotificationExpectationsBuilder struct {
	expectations map[string]map[string]fixtures.AssignmentState
}

func NewFANNotificationExpectationsBuilder() *FANNotificationExpectationsBuilder {
	return &FANNotificationExpectationsBuilder{
		expectations: map[string]map[string]fixtures.AssignmentState{},
	}
}

func (b *FANNotificationExpectationsBuilder) GetExpectations() map[string]map[string]fixtures.AssignmentState {
	return b.expectations
}

func (b *FANNotificationExpectationsBuilder) GetExpectedAssignmentsCount() int {
	count := 0
	for _, val := range b.expectations {
		count += len(val)
	}
	return count
}

func (b *FANNotificationExpectationsBuilder) getCurrentParticipants() []string {
	participants := make([]string, 0, len(b.expectations))
	for participantID, _ := range b.expectations {
		participants = append(participants, participantID)
	}

	return participants
}

func (b *FANNotificationExpectationsBuilder) WithParticipant(participantID string) *FANNotificationExpectationsBuilder {
	if _, ok := b.expectations[participantID]; ok {
		return b
	}

	for _, val := range b.expectations {
		val[participantID] = fixtures.AssignmentState{State: "READY", Config: nil, Value: nil, Error: nil}
	}

	currentParticipants := b.getCurrentParticipants()
	b.expectations[participantID] = make(map[string]fixtures.AssignmentState, len(currentParticipants))
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

func (b *FANNotificationExpectationsBuilder) WithNotifications(notifications []*NotificationData) *FANNotificationExpectationsBuilder {
	for _, notification := range notifications {
		b.expectations[notification.TargetID][notification.SourceID] = notification.getAssignmentState()
	}
	return b
}

func (b *FANNotificationExpectationsBuilder) WithoutParticipantAsync(participantID string, participantsAsyncForRemovedParticipant, asyncForParticipants []string) *FANNotificationExpectationsBuilder {
	for key, val := range b.expectations {
		if !slices.Contains(asyncForParticipants, key) {
			delete(val, participantID)
		}
	}

	if len(participantsAsyncForRemovedParticipant) == 0 {
		delete(b.expectations, participantID)
	} else {
		for key, _ := range b.expectations[participantID] {
			if !slices.Contains(participantsAsyncForRemovedParticipant, key) {
				delete(b.expectations[participantID], key)
			}
		}
	}

	return b
}
