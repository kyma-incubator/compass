package statusreport

import (
	"encoding/json"
	"unsafe"
)

// NotificationStatusReport represents the response to the notification in bot SYNC and ASYNC flows
type NotificationStatusReport struct {
	Configuration json.RawMessage
	State         string
	Error         string
}

// NewNotificationStatusReport is a constructor for NotificationStatusReport
func NewNotificationStatusReport(configuration json.RawMessage, state string, errorMessage string) *NotificationStatusReport {
	return &NotificationStatusReport{
		Configuration: configuration,
		State:         state,
		Error:         errorMessage,
	}
}

// GetState returns the State from the NotificationStatusReport
func (n *NotificationStatusReport) GetState() string {
	return n.State
}

// GetConfiguration returns the Configuration from the NotificationStatusReport
func (n *NotificationStatusReport) GetConfiguration() json.RawMessage {
	return n.Configuration
}

// GetError returns the Error from the NotificationStatusReport
func (n *NotificationStatusReport) GetError() string {
	return n.Error
}

func (n *NotificationStatusReport) GetAddress() uintptr {
	return uintptr(unsafe.Pointer(n))
}
