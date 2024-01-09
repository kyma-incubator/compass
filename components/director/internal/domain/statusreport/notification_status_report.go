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

// GetAddress returns the memory address of the NotificationStatusReport in the form of an uninterpreted type(integer number)
// Currently, it's used in some formation constraints input templates, so we could propagate the memory address to the formation constraints operators and later on to modify/update it.
func (n *NotificationStatusReport) GetAddress() uintptr {
	return uintptr(unsafe.Pointer(n))
}
