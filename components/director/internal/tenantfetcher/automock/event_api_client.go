// Code generated by mockery v2.10.5. DO NOT EDIT.

package automock

import (
	tenantfetcher "github.com/kyma-incubator/compass/components/director/internal/tenantfetcher"
	mock "github.com/stretchr/testify/mock"
)

// EventAPIClient is an autogenerated mock type for the EventAPIClient type
type EventAPIClient struct {
	mock.Mock
}

// FetchTenantEventsPage provides a mock function with given fields: eventsType, additionalQueryParams
func (_m *EventAPIClient) FetchTenantEventsPage(eventsType tenantfetcher.EventsType, additionalQueryParams tenantfetcher.QueryParams) (tenantfetcher.TenantEventsResponse, error) {
	ret := _m.Called(eventsType, additionalQueryParams)

	var r0 tenantfetcher.TenantEventsResponse
	if rf, ok := ret.Get(0).(func(tenantfetcher.EventsType, tenantfetcher.QueryParams) tenantfetcher.TenantEventsResponse); ok {
		r0 = rf(eventsType, additionalQueryParams)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(tenantfetcher.TenantEventsResponse)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(tenantfetcher.EventsType, tenantfetcher.QueryParams) error); ok {
		r1 = rf(eventsType, additionalQueryParams)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
