// Code generated by mockery v2.9.4. DO NOT EDIT.

package automock

import (
	context "context"

	mock "github.com/stretchr/testify/mock"
)

// ApplicationService is an autogenerated mock type for the ApplicationService type
type ApplicationService struct {
	mock.Mock
}

// TryUpdateBaseUrl provides a mock function with given fields: ctx, appID, targetURL
func (_m *ApplicationService) TryUpdateBaseUrl(ctx context.Context, appID string, targetURL string) error {
	ret := _m.Called(ctx, appID, targetURL)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string) error); ok {
		r0 = rf(ctx, appID, targetURL)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
