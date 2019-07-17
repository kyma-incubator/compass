// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import mock "github.com/stretchr/testify/mock"
import persistence "github.com/kyma-incubator/compass/components/director/internal/persistence"

// Transactioner is an autogenerated mock type for the Transactioner type
type Transactioner struct {
	mock.Mock
}

// Begin provides a mock function with given fields:
func (_m *Transactioner) Begin() (persistence.PersistenceTx, error) {
	ret := _m.Called()

	var r0 persistence.PersistenceTx
	if rf, ok := ret.Get(0).(func() persistence.PersistenceTx); ok {
		r0 = rf()
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(persistence.PersistenceTx)
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
