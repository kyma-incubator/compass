// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	model "github.com/kyma-incubator/compass/components/director/internal/model"
	mock "github.com/stretchr/testify/mock"

	operators "github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
)

// DestinationService is an autogenerated mock type for the destinationService type
type DestinationService struct {
	mock.Mock
}

// CreateBasicCredentialDestinations provides a mock function with given fields: ctx, destinationsDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs, skipSubaccountValidation
func (_m *DestinationService) CreateBasicCredentialDestinations(ctx context.Context, destinationsDetails []operators.Destination, basicAuthenticationCredentials operators.BasicAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, skipSubaccountValidation bool) error {
	ret := _m.Called(ctx, destinationsDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs, skipSubaccountValidation)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []operators.Destination, operators.BasicAuthentication, *model.FormationAssignment, []string, bool) error); ok {
		r0 = rf(ctx, destinationsDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs, skipSubaccountValidation)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateClientCertificateAuthenticationDestination provides a mock function with given fields: ctx, destinationsDetails, clientCertAuthCredentials, formationAssignment, correlationIDs, skipSubaccountValidation
func (_m *DestinationService) CreateClientCertificateAuthenticationDestination(ctx context.Context, destinationsDetails []operators.Destination, clientCertAuthCredentials *operators.ClientCertAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, skipSubaccountValidation bool) error {
	ret := _m.Called(ctx, destinationsDetails, clientCertAuthCredentials, formationAssignment, correlationIDs, skipSubaccountValidation)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []operators.Destination, *operators.ClientCertAuthentication, *model.FormationAssignment, []string, bool) error); ok {
		r0 = rf(ctx, destinationsDetails, clientCertAuthCredentials, formationAssignment, correlationIDs, skipSubaccountValidation)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateDesignTimeDestinations provides a mock function with given fields: ctx, destinationsDetails, formationAssignment, skipSubaccountValidation
func (_m *DestinationService) CreateDesignTimeDestinations(ctx context.Context, destinationsDetails []operators.Destination, formationAssignment *model.FormationAssignment, skipSubaccountValidation bool) error {
	ret := _m.Called(ctx, destinationsDetails, formationAssignment, skipSubaccountValidation)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []operators.Destination, *model.FormationAssignment, bool) error); ok {
		r0 = rf(ctx, destinationsDetails, formationAssignment, skipSubaccountValidation)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateSAMLAssertionDestination provides a mock function with given fields: ctx, destinationsDetails, samlAssertionAuthCredentials, formationAssignment, correlationIDs, skipSubaccountValidation
func (_m *DestinationService) CreateSAMLAssertionDestination(ctx context.Context, destinationsDetails []operators.Destination, samlAssertionAuthCredentials *operators.SAMLAssertionAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, skipSubaccountValidation bool) error {
	ret := _m.Called(ctx, destinationsDetails, samlAssertionAuthCredentials, formationAssignment, correlationIDs, skipSubaccountValidation)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []operators.Destination, *operators.SAMLAssertionAuthentication, *model.FormationAssignment, []string, bool) error); ok {
		r0 = rf(ctx, destinationsDetails, samlAssertionAuthCredentials, formationAssignment, correlationIDs, skipSubaccountValidation)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteDestinations provides a mock function with given fields: ctx, formationAssignment, skipSubaccountValidation
func (_m *DestinationService) DeleteDestinations(ctx context.Context, formationAssignment *model.FormationAssignment, skipSubaccountValidation bool) error {
	ret := _m.Called(ctx, formationAssignment, skipSubaccountValidation)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, *model.FormationAssignment, bool) error); ok {
		r0 = rf(ctx, formationAssignment, skipSubaccountValidation)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

type mockConstructorTestingTNewDestinationService interface {
	mock.TestingT
	Cleanup(func())
}

// NewDestinationService creates a new instance of DestinationService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDestinationService(t mockConstructorTestingTNewDestinationService) *DestinationService {
	mock := &DestinationService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
