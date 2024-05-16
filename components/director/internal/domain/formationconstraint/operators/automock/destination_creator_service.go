// Code generated by mockery. DO NOT EDIT.

package automock

import (
	context "context"
	json "encoding/json"

	destinationcreator "github.com/kyma-incubator/compass/components/director/pkg/destinationcreator"

	mock "github.com/stretchr/testify/mock"

	model "github.com/kyma-incubator/compass/components/director/internal/model"

	operators "github.com/kyma-incubator/compass/components/director/internal/domain/formationconstraint/operators"
)

// DestinationCreatorService is an autogenerated mock type for the destinationCreatorService type
type DestinationCreatorService struct {
	mock.Mock
}

// CreateCertificate provides a mock function with given fields: ctx, destinationsDetails, destinationAuthType, formationAssignment, depth, skipSubaccountValidation, useSelfSignedCert
func (_m *DestinationCreatorService) CreateCertificate(ctx context.Context, destinationsDetails []operators.Destination, destinationAuthType destinationcreator.AuthType, formationAssignment *model.FormationAssignment, depth uint8, skipSubaccountValidation bool, useSelfSignedCert bool) (*operators.CertificateData, error) {
	ret := _m.Called(ctx, destinationsDetails, destinationAuthType, formationAssignment, depth, skipSubaccountValidation, useSelfSignedCert)

	var r0 *operators.CertificateData
	var r1 error
	if rf, ok := ret.Get(0).(func(context.Context, []operators.Destination, destinationcreator.AuthType, *model.FormationAssignment, uint8, bool, bool) (*operators.CertificateData, error)); ok {
		return rf(ctx, destinationsDetails, destinationAuthType, formationAssignment, depth, skipSubaccountValidation, useSelfSignedCert)
	}
	if rf, ok := ret.Get(0).(func(context.Context, []operators.Destination, destinationcreator.AuthType, *model.FormationAssignment, uint8, bool, bool) *operators.CertificateData); ok {
		r0 = rf(ctx, destinationsDetails, destinationAuthType, formationAssignment, depth, skipSubaccountValidation, useSelfSignedCert)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(*operators.CertificateData)
		}
	}

	if rf, ok := ret.Get(1).(func(context.Context, []operators.Destination, destinationcreator.AuthType, *model.FormationAssignment, uint8, bool, bool) error); ok {
		r1 = rf(ctx, destinationsDetails, destinationAuthType, formationAssignment, depth, skipSubaccountValidation, useSelfSignedCert)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EnrichAssignmentConfigWithCertificateData provides a mock function with given fields: assignmentConfig, destinationTypePath, certData
func (_m *DestinationCreatorService) EnrichAssignmentConfigWithCertificateData(assignmentConfig json.RawMessage, destinationTypePath string, certData *operators.CertificateData) (json.RawMessage, error) {
	ret := _m.Called(assignmentConfig, destinationTypePath, certData)

	var r0 json.RawMessage
	var r1 error
	if rf, ok := ret.Get(0).(func(json.RawMessage, string, *operators.CertificateData) (json.RawMessage, error)); ok {
		return rf(assignmentConfig, destinationTypePath, certData)
	}
	if rf, ok := ret.Get(0).(func(json.RawMessage, string, *operators.CertificateData) json.RawMessage); ok {
		r0 = rf(assignmentConfig, destinationTypePath, certData)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(json.RawMessage)
		}
	}

	if rf, ok := ret.Get(1).(func(json.RawMessage, string, *operators.CertificateData) error); ok {
		r1 = rf(assignmentConfig, destinationTypePath, certData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// EnrichAssignmentConfigWithSAMLCertificateData provides a mock function with given fields: assignmentConfig, destinationTypePath, certData
func (_m *DestinationCreatorService) EnrichAssignmentConfigWithSAMLCertificateData(assignmentConfig json.RawMessage, destinationTypePath string, certData *operators.CertificateData) (json.RawMessage, error) {
	ret := _m.Called(assignmentConfig, destinationTypePath, certData)

	var r0 json.RawMessage
	var r1 error
	if rf, ok := ret.Get(0).(func(json.RawMessage, string, *operators.CertificateData) (json.RawMessage, error)); ok {
		return rf(assignmentConfig, destinationTypePath, certData)
	}
	if rf, ok := ret.Get(0).(func(json.RawMessage, string, *operators.CertificateData) json.RawMessage); ok {
		r0 = rf(assignmentConfig, destinationTypePath, certData)
	} else {
		if ret.Get(0) != nil {
			r0 = ret.Get(0).(json.RawMessage)
		}
	}

	if rf, ok := ret.Get(1).(func(json.RawMessage, string, *operators.CertificateData) error); ok {
		r1 = rf(assignmentConfig, destinationTypePath, certData)
	} else {
		r1 = ret.Error(1)
	}

	return r0, r1
}

// NewDestinationCreatorService creates a new instance of DestinationCreatorService. It also registers a testing interface on the mock and a cleanup function to assert the mocks expectations.
// The first argument is typically a *testing.T value.
func NewDestinationCreatorService(t interface {
	mock.TestingT
	Cleanup(func())
}) *DestinationCreatorService {
	mock := &DestinationCreatorService{}
	mock.Mock.Test(t)

	t.Cleanup(func() { mock.AssertExpectations(t) })

	return mock
}
