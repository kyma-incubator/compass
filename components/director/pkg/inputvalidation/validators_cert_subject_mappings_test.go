package inputvalidation_test

import (
	"fmt"
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	tenantEntity "github.com/kyma-incubator/compass/components/director/pkg/tenant"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCertMappingSubjectValidator_Validate(t *testing.T) {
	var nilPointer *string

	testCases := []struct {
		Name  string
		Input interface{}
		Valid bool
	}{
		{
			Name:  "Valid certificate subject",
			Input: "C=DE, L=test, O=SAP SE, OU=TestRegion, OU=SAP Cloud Platform Clients, OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10, CN=test-compass",
			Valid: true,
		},
		{
			Name:  "Invalid when the input is not a string",
			Input: 123,
			Valid: false,
		},
		{
			Name:  "Invalid when the input is `nil`",
			Input: nil,
			Valid: false,
		},
		{
			Name: "Valid when the input is empty pointer",
			Input: nilPointer,
			Valid: true,
		},
		{
			Name:  "Invalid when the certificate relative distinguished names are less than 5",
			Input: "L=test, O=SAP SE, CN=test-compass",
			Valid: false,
		},
		{
			Name:  "Invalid when Country property is missing from the subject",
			Input: "L=test, O=SAP SE, OU=TestRegion, OU=SAP Cloud Platform Clients, OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10, CN=test-compass",
			Valid: false,
		},
		{
			Name:  "Invalid when Organization property is missing from the subject",
			Input: "C=DE, L=test, OU=TestRegion, OU=SAP Cloud Platform Clients, OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10, CN=test-compass",
			Valid: false,
		},
		{
			Name:  "Invalid when Organization Unit is missing from the subject",
			Input: "C=DE, L=test, O=SAP SE, CN=test-compass, GN=given-name, DC=domain-component",
			Valid: false,
		},
		{
			Name:  "Invalid when Locality property is missing from the subject",
			Input: "C=DE, O=SAP SE, OU=TestRegion, OU=SAP Cloud Platform Clients, OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10, CN=test-compass",
			Valid: false,
		},
		{
			Name:  "Invalid when Common Name property is missing from the subject",
			Input: "C=DE, L=test, O=SAP SE, OU=TestRegion, OU=SAP Cloud Platform Clients, OU=2c0fe288-bb13-4814-ac49-ac88c4a76b10",
			Valid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			certSubjectValidator := inputvalidation.IsValidCertSubject

			// WHEN
			err := validation.Validate(testCase.Input, certSubjectValidator)

			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestCertMappingConsumerTypeValidator_Validate(t *testing.T) {
	var nilPointer *string
	invalidType := "invalidType"

	testCases := []struct {
		Name  string
		Input interface{}
		Valid bool
		ErrMsg string
	}{
		{
			Name:  "Invalid when the input is not a string",
			Input: 123,
			Valid: false,
			ErrMsg: "type has to be a string",
		},
		{
			Name:  "Invalid when the input is `nil`",
			Input: nil,
			Valid: false,
			ErrMsg: "type has to be a string",
		},
		{
			Name: "Valid when the input is empty pointer",
			Input: nilPointer,
			Valid: true,
		},
		{
			Name: "Valid when the consumer type is supported",
			Input: inputvalidation.RuntimeType,
			Valid: true,
		},
		{
			Name: "Invalid when the consumer type is unsupported",
			Input: invalidType,
			Valid: false,
			ErrMsg: fmt.Sprintf("consumer type %s is not valid", invalidType),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			certMappingSubjectValidator := inputvalidation.IsValidConsumerType

			// WHEN
			err := validation.Validate(testCase.Input, certMappingSubjectValidator)

			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ErrMsg)
			}
		})
	}
}

func TestCertMappingTenantAccessLevelValidator_Validate(t *testing.T) {
	tenantAccessLevels := []string{inputvalidation.GlobalAccessLevel, string(tenantEntity.Subaccount)}
	unsupportedTntAccessLevels := []string{"unsupported1", "unsupported2"}
	invalidTenantAccessLevelsType := "invalidTenantAccessLevelsType"

	testCases := []struct {
		Name  string
		Input interface{}
		Valid bool
		ErrMsg string
	}{
		{
			Name:  "Valid when tenant access levels are supported",
			Input: tenantAccessLevels,
			Valid: true,
		},
		{
			Name:  "Invalid when tenant access levels are unsupported",
			Input: unsupportedTntAccessLevels,
			Valid: false,
			ErrMsg: fmt.Sprintf("tenant access level %s is not valid", unsupportedTntAccessLevels[0]),
		},
		{
			Name:  "Invalid when tenant access levels type is incorrect",
			Input: invalidTenantAccessLevelsType,
			Valid: false,
			ErrMsg: fmt.Sprintf("invalid type, expected []string, got: %T", invalidTenantAccessLevelsType),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			certMappingTenantAccessLevelValidator := inputvalidation.AreTenantAccessLevelsValid

			// WHEN
			err := validation.Validate(testCase.Input, certMappingTenantAccessLevelValidator)

			// THEN
			if testCase.Valid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ErrMsg)
			}
		})
	}
}
