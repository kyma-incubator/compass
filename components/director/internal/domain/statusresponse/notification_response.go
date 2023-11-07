package statusresponse

import (
	"encoding/json"
)

// NotificationResponse represents the response to the notification in bot SYNC and ASYNC flows
type NotificationResponse struct {
	configuration json.RawMessage
	state         string
	error         string
}

// NewNotificationResponse is a constructor for NotificationResponse
func NewNotificationResponse(configuration json.RawMessage, state string, errorMessage string) *NotificationResponse {
	return &NotificationResponse{
		configuration: configuration,
		state:         state,
		error:         errorMessage,
	}
}

// GetState returns the state from the NotificationResponse
func (n *NotificationResponse) GetState() string {
	return n.state
}

// GetConfiguration returns the configuration from the NotificationResponse
func (n *NotificationResponse) GetConfiguration() json.RawMessage {
	return n.configuration
}

// GetError returns the error from the NotificationResponse
func (n *NotificationResponse) GetError() string {
	return n.error
}
