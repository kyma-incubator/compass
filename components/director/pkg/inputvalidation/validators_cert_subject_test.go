package inputvalidation_test

import (
	validation "github.com/go-ozzo/ozzo-validation/v4"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation"
	"github.com/stretchr/testify/require"
	"testing"
)

func TestCertSubjectValidator_Validate(t *testing.T) {
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
