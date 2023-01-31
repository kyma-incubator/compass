// Code generated by mockery. DO NOT EDIT.

package automock

import (
	certsubjectmapping "github.com/kyma-incubator/compass/components/hydrator/internal/certsubjectmapping"
	mock "github.com/stretchr/testify/mock"

	testing "testing"
)

// Cache is an autogenerated mock type for the Cache type
type Cache struct {
	mock.Mock
}

// Get provides a mock function with given fields:
func (_m *Cache) Get() []certsubjectmapping.SubjectConsumerTypeMapping {
	ret := _m.Called()

	var r0 []certsubjectmapping.SubjectConsumerTypeMapping
	if rf, ok := ret.Get(0).(func() []certsubjectmapping.SubjectConsumerTypeMapping); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]certsubjectmapping.SubjectConsumerTypeMapping)
		}
	}

	return r0
}

// Put provides a mock function with given fields: certSubjectMappings
func (_m *Cache) Put(certSubjectMappings []certsubjectmapping.SubjectConsumerTypeMapping) {
	_m.Called(certSubjectMappings)
}

// NewCache creates a new instance of Cache. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewCache(t testing.TB) *Cache {
	mock := &Cache{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
