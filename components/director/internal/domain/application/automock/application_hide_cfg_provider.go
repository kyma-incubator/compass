// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"

// ApplicationHideCfgProvider is an autogenerated mock type for the ApplicationHideCfgProvider type
type ApplicationHideCfgProvider struct {
	mock.Mock
}

// GetApplicationHideSelectors provides a mock function with given fields:
func (_m *ApplicationHideCfgProvider) GetApplicationHideSelectors() (map[string][]string, error) {
	ret := _m.Called()

	var r0 map[string][]string
	if rf, ok := ret.Get(0).(func() map[string][]string); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(map[string][]string)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func() error); ok {
		r1 = rf()
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
