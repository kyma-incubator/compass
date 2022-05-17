// Code generated by mockery. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	service "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service"
)

// AppLabeler is an autogenerated mock type for the AppLabeler type
type AppLabeler struct {
	mock.Mock
}

// DeleteServiceReference provides a mock function with given fields: appLabels, serviceID
func (_m *AppLabeler) DeleteServiceReference(appLabels graphql.Labels, serviceID string) (graphql.LabelInput, error) {
	ret := _m.Called(appLabels, serviceID)

	var r0 graphql.LabelInput
	if rf, ok := ret.Get(0).(func(graphql.Labels, string) graphql.LabelInput); ok {
		r0 = rf(appLabels, serviceID)
	} else {
		r0 = ret.Get(0).(graphql.LabelInput)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(graphql.Labels, string) error); ok {
		r1 = rf(appLabels, serviceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ListServiceReferences provides a mock function with given fields: appLabels
func (_m *AppLabeler) ListServiceReferences(appLabels graphql.Labels) ([]service.LegacyServiceReference, error) {
	ret := _m.Called(appLabels)

	var r0 []service.LegacyServiceReference
	if rf, ok := ret.Get(0).(func(graphql.Labels) []service.LegacyServiceReference); ok {
		r0 = rf(appLabels)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]service.LegacyServiceReference)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(graphql.Labels) error); ok {
		r1 = rf(appLabels)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ReadServiceReference provides a mock function with given fields: appLabels, serviceID
func (_m *AppLabeler) ReadServiceReference(appLabels graphql.Labels, serviceID string) (service.LegacyServiceReference, error) {
	ret := _m.Called(appLabels, serviceID)

	var r0 service.LegacyServiceReference
	if rf, ok := ret.Get(0).(func(graphql.Labels, string) service.LegacyServiceReference); ok {
		r0 = rf(appLabels, serviceID)
	} else {
		r0 = ret.Get(0).(service.LegacyServiceReference)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(graphql.Labels, string) error); ok {
		r1 = rf(appLabels, serviceID)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// WriteServiceReference provides a mock function with given fields: appLabels, serviceReference
func (_m *AppLabeler) WriteServiceReference(appLabels graphql.Labels, serviceReference service.LegacyServiceReference) (graphql.LabelInput, error) {
	ret := _m.Called(appLabels, serviceReference)

	var r0 graphql.LabelInput
	if rf, ok := ret.Get(0).(func(graphql.Labels, service.LegacyServiceReference) graphql.LabelInput); ok {
		r0 = rf(appLabels, serviceReference)
	} else {
		r0 = ret.Get(0).(graphql.LabelInput)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(graphql.Labels, service.LegacyServiceReference) error); ok {
		r1 = rf(appLabels, serviceReference)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
