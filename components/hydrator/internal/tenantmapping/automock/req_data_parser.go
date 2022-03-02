// Code generated by mockery v2.10.0. DO NOT EDIT.

package automock

import (
	http "net/http"

	oathkeeper "github.com/kyma-incubator/compass/components/director/pkg/oathkeeper"
	mock "github.com/stretchr/testify/mock"
)

// ReqDataParser is an autogenerated mock type for the ReqDataParser type
type ReqDataParser struct {
	mock.Mock
}

// Parse provides a mock function with given fields: req
func (_m *ReqDataParser) Parse(req *http.Request) (oathkeeper.ReqData, error) {
	ret := _m.Called(req)

	var r0 oathkeeper.ReqData
	if rf, ok := ret.Get(0).(func(*http.Request) oathkeeper.ReqData); ok {
		r0 = rf(req)
	} else {
		r0 = ret.Get(0).(oathkeeper.ReqData)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Request) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
