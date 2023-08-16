// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"

	destinationcreator "github.com/kyma-incubator/compass/components/director/internal/destinationcreator"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	operators "github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
)

// DestinationCreatorService is an autogenerated mock type for the destinationCreatorService type
type DestinationCreatorService struct {
	mock.Mock
}

// CreateBasicCredentialDestinations provides a mock function with given fields: ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs, depth
func (_m *DestinationCreatorService) CreateBasicCredentialDestinations(ctx context.Context, destinationDetails operators.Destination, basicAuthenticationCredentials operators.BasicAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, depth uint8) error {
	ret := _m.Called(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs, depth)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, operators.Destination, operators.BasicAuthentication, *model.FormationAssignment, []string, uint8) error); ok {
		r0 = rf(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs, depth)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateClientCertificateDestination provides a mock function with given fields: ctx, destinationDetails, clientCertAuthCreds, formationAssignment, correlationIDs, depth
func (_m *DestinationCreatorService) CreateClientCertificateDestination(ctx context.Context, destinationDetails operators.Destination, clientCertAuthCreds *operators.ClientCertAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, depth uint8) error {
	ret := _m.Called(ctx, destinationDetails, clientCertAuthCreds, formationAssignment, correlationIDs, depth)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, operators.Destination, *operators.ClientCertAuthentication, *model.FormationAssignment, []string, uint8) error); ok {
		r0 = rf(ctx, destinationDetails, clientCertAuthCreds, formationAssignment, correlationIDs, depth)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateDesignTimeDestinations provides a mock function with given fields: ctx, destinationDetails, formationAssignment, depth
func (_m *DestinationCreatorService) CreateDesignTimeDestinations(ctx context.Context, destinationDetails operators.Destination, formationAssignment *model.FormationAssignment, depth uint8) error {
	ret := _m.Called(ctx, destinationDetails, formationAssignment, depth)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, operators.Destination, *model.FormationAssignment, uint8) error); ok {
		r0 = rf(ctx, destinationDetails, formationAssignment, depth)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// CreateSAMLAssertionDestination provides a mock function with given fields: ctx, destinationDetails, samlAssertionAuthCreds, formationAssignment, correlationIDs, depth
func (_m *DestinationCreatorService) CreateSAMLAssertionDestination(ctx context.Context, destinationDetails operators.Destination, samlAssertionAuthCreds *operators.SAMLAssertionAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string, depth uint8) error {
	ret := _m.Called(ctx, destinationDetails, samlAssertionAuthCreds, formationAssignment, correlationIDs, depth)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, operators.Destination, *operators.SAMLAssertionAuthentication, *model.FormationAssignment, []string, uint8) error); ok {
		r0 = rf(ctx, destinationDetails, samlAssertionAuthCreds, formationAssignment, correlationIDs, depth)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteCertificate provides a mock function with given fields: ctx, certificateName, externalDestSubaccountID, formationAssignment
func (_m *DestinationCreatorService) DeleteCertificate(ctx context.Context, certificateName string, externalDestSubaccountID string, formationAssignment *model.FormationAssignment) error {
	ret := _m.Called(ctx, certificateName, externalDestSubaccountID, formationAssignment)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *model.FormationAssignment) error); ok {
		r0 = rf(ctx, certificateName, externalDestSubaccountID, formationAssignment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// DeleteDestination provides a mock function with given fields: ctx, destinationName, externalDestSubaccountID, formationAssignment
func (_m *DestinationCreatorService) DeleteDestination(ctx context.Context, destinationName string, externalDestSubaccountID string, formationAssignment *model.FormationAssignment) error {
	ret := _m.Called(ctx, destinationName, externalDestSubaccountID, formationAssignment)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, string, string, *model.FormationAssignment) error); ok {
		r0 = rf(ctx, destinationName, externalDestSubaccountID, formationAssignment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// EnsureDestinationSubaccountIDsCorrectness provides a mock function with given fields: ctx, destinationsDetails, formationAssignment
func (_m *DestinationCreatorService) EnsureDestinationSubaccountIDsCorrectness(ctx context.Context, destinationsDetails []operators.Destination, formationAssignment *model.FormationAssignment) error {
	ret := _m.Called(ctx, destinationsDetails, formationAssignment)

	var r0 error
	if rf, ok := ret.Get(0).(func(context.Context, []operators.Destination, *model.FormationAssignment) error); ok {
		r0 = rf(ctx, destinationsDetails, formationAssignment)
	} else {
		r0 = ret.Error(0)
	}

	return r0
}

// GetConsumerTenant provides a mock function with given fields: ctx, formationAssignment
func (_m *DestinationCreatorService) GetConsumerTenant(ctx context.Context, formationAssignment *model.FormationAssignment) (string, error) {
	ret := _m.Called(ctx, formationAssignment)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, *model.FormationAssignment) string); ok {
		r0 = rf(ctx, formationAssignment)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, *model.FormationAssignment) error); ok {
		r1 = rf(ctx, formationAssignment)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// PrepareBasicRequestBody provides a mock function with given fields: ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs
func (_m *DestinationCreatorService) PrepareBasicRequestBody(ctx context.Context, destinationDetails operators.Destination, basicAuthenticationCredentials operators.BasicAuthentication, formationAssignment *model.FormationAssignment, correlationIDs []string) (*destinationcreator.BasicAuthDestinationRequestBody, error) {
	ret := _m.Called(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs)

	var r0 *destinationcreator.BasicAuthDestinationRequestBody
	if rf, ok := ret.Get(0).(func(context.Context, operators.Destination, operators.BasicAuthentication, *model.FormationAssignment, []string) *destinationcreator.BasicAuthDestinationRequestBody); ok {
		r0 = rf(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*destinationcreator.BasicAuthDestinationRequestBody)
		}
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, operators.Destination, operators.BasicAuthentication, *model.FormationAssignment, []string) error); ok {
		r1 = rf(ctx, destinationDetails, basicAuthenticationCredentials, formationAssignment, correlationIDs)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// ValidateDestinationSubaccount provides a mock function with given fields: ctx, externalDestSubaccountID, formationAssignment
func (_m *DestinationCreatorService) ValidateDestinationSubaccount(ctx context.Context, externalDestSubaccountID string, formationAssignment *model.FormationAssignment) (string, error) {
	ret := _m.Called(ctx, externalDestSubaccountID, formationAssignment)

	var r0 string
	if rf, ok := ret.Get(0).(func(context.Context, string, *model.FormationAssignment) string); ok {
		r0 = rf(ctx, externalDestSubaccountID, formationAssignment)
	} else {
		r0 = ret.Get(0).(string)
	}

	var r1 error
	if rf, ok := ret.Get(1).(func(context.Context, string, *model.FormationAssignment) error); ok {
		r1 = rf(ctx, externalDestSubaccountID, formationAssignment)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

type mockConstructorTestingTNewDestinationCreatorService interface {
	mock.TestingT
	Cleanup(func())
}

// NewDestinationCreatorService creates a new instance of DestinationCreatorService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
func NewDestinationCreatorService(t mockConstructorTestingTNewDestinationCreatorService) *DestinationCreatorService {
	mock := &DestinationCreatorService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
