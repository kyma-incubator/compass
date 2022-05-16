// Code generated by mockery. DO NOT EDIT.

package mocks

import (
	graphqlizer "github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	mock "github.com/stretchr/testify/mock"
)

// GqlFieldsProvider is an autogenerated mock type for the GqlFieldsProvider type
type GqlFieldsProvider struct {
	mock.Mock
}

// ForAPIDefinition provides a mock function with given fields: ctx
func (_m *GqlFieldsProvider) ForAPIDefinition(ctx ...graphqlizer.FieldCtx) string {
	_va := make([]interface{}, len(ctx))
	for _i := range ctx {
		_va[_i] = ctx[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 string
	if rf, ok := ret.Get(0).(func(...graphqlizer.FieldCtx) string); ok {
		r0 = rf(ctx...)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ForApplication provides a mock function with given fields: ctx
func (_m *GqlFieldsProvider) ForApplication(ctx ...graphqlizer.FieldCtx) string {
	_va := make([]interface{}, len(ctx))
	for _i := range ctx {
		_va[_i] = ctx[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 string
	if rf, ok := ret.Get(0).(func(...graphqlizer.FieldCtx) string); ok {
		r0 = rf(ctx...)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ForBundle provides a mock function with given fields: ctx
func (_m *GqlFieldsProvider) ForBundle(ctx ...graphqlizer.FieldCtx) string {
	_va := make([]interface{}, len(ctx))
	for _i := range ctx {
		_va[_i] = ctx[_i]
	}
	var _ca []interface{}
	_ca = append(_ca, _va...)
	ret := _m.Called(_ca...)

	var r0 string
	if rf, ok := ret.Get(0).(func(...graphqlizer.FieldCtx) string); ok {
		r0 = rf(ctx...)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ForBundleInstanceAuth provides a mock function with given fields:
func (_m *GqlFieldsProvider) ForBundleInstanceAuth() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ForBundleInstanceAuthStatus provides a mock function with given fields:
func (_m *GqlFieldsProvider) ForBundleInstanceAuthStatus() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ForDocument provides a mock function with given fields:
func (_m *GqlFieldsProvider) ForDocument() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ForEventDefinition provides a mock function with given fields:
func (_m *GqlFieldsProvider) ForEventDefinition() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// ForLabel provides a mock function with given fields:
func (_m *GqlFieldsProvider) ForLabel() string {
	ret := _m.Called()

	var r0 string
	if rf, ok := ret.Get(0).(func() string); ok {
		r0 = rf()
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// OmitForApplication provides a mock function with given fields: omit
func (_m *GqlFieldsProvider) OmitForApplication(omit []string) string {
	ret := _m.Called(omit)

	var r0 string
	if rf, ok := ret.Get(0).(func([]string) string); ok {
		r0 = rf(omit)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}

// Page provides a mock function with given fields: item
func (_m *GqlFieldsProvider) Page(item string) string {
	ret := _m.Called(item)

	var r0 string
	if rf, ok := ret.Get(0).(func(string) string); ok {
		r0 = rf(item)
	} else {
		r0 = ret.Get(0).(string)
	}

	return r0
}
