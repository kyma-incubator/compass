// Code generated by mockery v2.10.5. DO NOT EDIT.

package automock

import (
	graphql "github.com/kyma-incubator/compass/components/director/pkg/graphql"
	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
)

// WebhookConverter is an autogenerated mock type for the WebhookConverter type
type WebhookConverter struct {
	mock.Mock
}

// InputFromGraphQL provides a mock function with given fields: in
func (_m *WebhookConverter) InputFromGraphQL(in *graphql.WebhookInput) (*model.WebhookInput, error) {
	ret := _m.Called(in)

	var r0 *model.WebhookInput
	if rf, ok := ret.Get(0).(func(*graphql.WebhookInput) *model.WebhookInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.WebhookInput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*graphql.WebhookInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MultipleInputFromGraphQL provides a mock function with given fields: in
func (_m *WebhookConverter) MultipleInputFromGraphQL(in []*graphql.WebhookInput) ([]*model.WebhookInput, error) {
	ret := _m.Called(in)

	var r0 []*model.WebhookInput
	if rf, ok := ret.Get(0).(func([]*graphql.WebhookInput) []*model.WebhookInput); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*model.WebhookInput)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*graphql.WebhookInput) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// MultipleToGraphQL provides a mock function with given fields: in
func (_m *WebhookConverter) MultipleToGraphQL(in []*model.Webhook) ([]*graphql.Webhook, error) {
	ret := _m.Called(in)

	var r0 []*graphql.Webhook
	if rf, ok := ret.Get(0).(func([]*model.Webhook) []*graphql.Webhook); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).([]*graphql.Webhook)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*model.Webhook) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToGraphQL provides a mock function with given fields: in
func (_m *WebhookConverter) ToGraphQL(in *model.Webhook) (*graphql.Webhook, error) {
	ret := _m.Called(in)

	var r0 *graphql.Webhook
	if rf, ok := ret.Get(0).(func(*model.Webhook) *graphql.Webhook); ok {
		r0 = rf(in)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*graphql.Webhook)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(*model.Webhook) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
