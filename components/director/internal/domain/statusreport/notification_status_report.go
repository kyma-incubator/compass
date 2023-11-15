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

func (n *NotificationStatusReport) GetAddress() uintptr {
	return uintptr(unsafe.Pointer(n))
}
