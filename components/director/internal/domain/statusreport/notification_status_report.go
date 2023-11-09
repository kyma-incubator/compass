package statusreport

import (
	"encoding/json"
)

// NotificationStatusReport represents the response to the notification in bot SYNC and ASYNC flows
type NotificationStatusReport struct {
	configuration json.RawMessage
	state         string
	error         string
}

// NewNotificationStatusReport is a constructor for NotificationStatusReport
func NewNotificationStatusReport(configuration json.RawMessage, state string, errorMessage string) *NotificationStatusReport {
	return &NotificationStatusReport{
		configuration: configuration,
		state:         state,
		error:         errorMessage,
	}
}

// GetState returns the state from the NotificationStatusReport
func (n *NotificationStatusReport) GetState() string {
	return n.state
}

// GetConfiguration returns the configuration from the NotificationStatusReport
func (n *NotificationStatusReport) GetConfiguration() json.RawMessage {
	return n.configuration
}

// GetError returns the error from the NotificationStatusReport
func (n *NotificationStatusReport) GetError() string {
	return n.error
}
