// Code generated by mockery v2.10.0. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

// ClientInstrumenter is an autogenerated mock type for the ClientInstrumenter type
type ClientInstrumenter struct {
	mock.Mock
}

// InstrumentClient provides a mock function with given fields: clientID, authFlow, details
func (_m *ClientInstrumenter) InstrumentClient(clientID string, authFlow string, details string) {
	_m.Called(clientID, authFlow, details)
}
