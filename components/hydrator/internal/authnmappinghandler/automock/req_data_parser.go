// Code generated by mockery. DO NOT EDIT.

package automock

import (
	http "net/http"
	testing "testing"

	oathkeeper "github.com/kyma-incubator/compass/components/hydrator/pkg/oathkeeper"
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

// NewReqDataParser creates a new instance of ReqDataParser. It also registers the testing.TB interface on the mock and a cleanup function to assert the mocks expectations.
func NewReqDataParser(t testing.TB) *ReqDataParser {
	mock := &ReqDataParser{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
