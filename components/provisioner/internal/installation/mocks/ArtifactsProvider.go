// Code generated by mockery v1.0.0. DO NOT EDIT.

package mocks

import (
	artifacts "github.com/kyma-incubator/compass/components/provisioner/internal/installation/release"

	mock "github.com/stretchr/testify/mock"
)

// ArtifactsProvider is an autogenerated mock type for the ArtifactsProvider type
type ArtifactsProvider struct {
	mock.Mock
}

// GetRelease provides a mock function with given fields: version
func (_m *ArtifactsProvider) GetArtifacts(version string) (artifacts.Release, error) {
	ret := _m.Called(version)

	var r0 artifacts.Release
	if rf, ok := ret.Get(0).(func(string) artifacts.Release); ok {
		r0 = rf(version)
	} else {
		r0 = ret.Get(0).(artifacts.Release)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(version)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
