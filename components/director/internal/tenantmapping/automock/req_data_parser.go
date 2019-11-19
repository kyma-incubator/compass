// Code generated by mockery v1.0.0. DO NOT EDIT.

package automock

import (
	http "net/http"

	tenantmapping "github.com/kyma-incubator/compass/components/director/internal/tenantmapping"
	mock "github.com/stretchr/testify/mock"
)

// ReqDataParser is an autogenerated mock type for the ReqDataParser type
type ReqDataParser struct {
	mock.Mock
}

// Parse provides a mock function with given fields: req
func (_m *ReqDataParser) Parse(req *http.Request) (tenantmapping.ReqData, error) {
	ret := _m.Called(req)

	var r0 tenantmapping.ReqData
	if rf, ok := ret.Get(0).(func(*http.Request) tenantmapping.ReqData); ok {
		r0 = rf(req)
	} else {
		r0 = ret.Get(0).(tenantmapping.ReqData)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*http.Request) error); ok {
		r1 = rf(req)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
