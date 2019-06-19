// Code generated by mockery v1.0.0
package automock

import labelfilter "github.com/kyma-incubator/compass/components/director/internal/labelfilter"
import mock "github.com/stretchr/testify/mock"
import model "github.com/kyma-incubator/compass/components/director/internal/model"

// ApplicationRepository is an autogenerated mock type for the ApplicationRepository type
type ApplicationRepository struct {
	mock.Mock
}

// Create provides a mock function with given fields: item
func (_m *ApplicationRepository) Create(item *model.ApplicationInput) error {
	ret := _m.Called(item)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.ApplicationInput) error); ok {
		r0 = rf(item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// Delete provides a mock function with given fields: item
func (_m *ApplicationRepository) Delete(item *model.Application) error {
	ret := _m.Called(item)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.Application) error); ok {
		r0 = rf(item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetByID provides a mock function with given fields: id
func (_m *ApplicationRepository) GetByID(id string) (*model.Application, error) {
	ret := _m.Called(id)

	var r0 *model.Application
	if rf, ok := ret.Get(0).(func(string) *model.Application); ok {
		r0 = rf(id)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.Application)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(string) error); ok {
		r1 = rf(id)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// List provides a mock function with given fields: filter, pageSize, cursor
func (_m *ApplicationRepository) List(filter []*labelfilter.LabelFilter, pageSize *int, cursor *string) (*model.ApplicationPage, error) {
	ret := _m.Called(filter, pageSize, cursor)

	var r0 *model.ApplicationPage
	if rf, ok := ret.Get(0).(func([]*labelfilter.LabelFilter, *int, *string) *model.ApplicationPage); ok {
		r0 = rf(filter, pageSize, cursor)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*model.ApplicationPage)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func([]*labelfilter.LabelFilter, *int, *string) error); ok {
		r1 = rf(filter, pageSize, cursor)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// Update provides a mock function with given fields: item
func (_m *ApplicationRepository) Update(item *model.ApplicationInput) error {
	ret := _m.Called(item)

	var r0 error
	if rf, ok := ret.Get(0).(func(*model.ApplicationInput) error); ok {
		r0 = rf(item)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}
