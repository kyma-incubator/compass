package notificationresponse

import (
	"encoding/json"
)

type NotificationResponse struct {
	configuration json.RawMessage
	state         string
	error         string
}

func NewNotificationResponse(configuration json.RawMessage, state string, error string) *NotificationResponse {
	return &NotificationResponse{
		configuration: configuration,
		state:         state,
		error:         error,
	}
}

func (n *NotificationResponse) GetState() string {
	return n.state
}

func (n *NotificationResponse) GetConfiguration() json.RawMessage {
	return n.configuration
}

func (n *NotificationResponse) GetError() string {
	return n.error
}
