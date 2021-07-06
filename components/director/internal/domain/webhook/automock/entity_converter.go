// Code generated by mockery 2.9.0. DO NOT EDIT.

package automock

import (
	webhook "github.com/kyma-incubator/compass/components/director/internal/domain/webhook"
	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"
)

// EntityConverter is an autogenerated mock type for the EntityConverter type
type EntityConverter struct {
	mock.Mock
}

// FromEntity provides a mock function with given fields: in
func (_m *EntityConverter) FromEntity(in webhook.Entity) (model.Webhook, error) {
	ret := _m.Called(in)

	var r0 model.Webhook
	if rf, ok := ret.Get(0).(func(webhook.Entity) model.Webhook); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(model.Webhook)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(webhook.Entity) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ToEntity provides a mock function with given fields: in
func (_m *EntityConverter) ToEntity(in model.Webhook) (webhook.Entity, error) {
	ret := _m.Called(in)

	var r0 webhook.Entity
	if rf, ok := ret.Get(0).(func(model.Webhook) webhook.Entity); ok {
		r0 = rf(in)
	} else {
		r0 = ret.Get(0).(webhook.Entity)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(model.Webhook) error); ok {
		r1 = rf(in)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}
