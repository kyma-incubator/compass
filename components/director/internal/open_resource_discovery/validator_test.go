package ord_test

import (
	"context"
	"encoding/json"
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/model"

	"github.com/tidwall/sjson"

	ord "github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery"
	"github.com/kyma-incubator/compass/components/director/internal/open_resource_discovery/automock"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/mock"
	"github.com/stretchr/testify/require"
)

func TestDocumentValidator_Validate(t *testing.T) {
	testNamespace := "test.ns"
	ignoredValidationRule := "type1"
	ignorelistMapping := map[string][]string{
		testNamespace: {ignoredValidationRule},
	}

	validatorClientCallWithoutValidationErrors := func() *automock.ValidatorClient {
		clientValidator := &automock.ValidatorClient{}
		clientValidator.On("Validate", mock.Anything, "", mock.Anything).Return([]model.ValidationResult{}, nil)
		return clientValidator
	}
	validatorClientCallWithValidationErrors := func() *automock.ValidatorClient {
		clientValidator := &automock.ValidatorClient{}
		clientValidator.On("Validate", mock.Anything, "", mock.Anything).Return([]model.ValidationResult{
			{
				Code:     ignoredValidationRule,
				Severity: "error",
				Message:  "errors are found",
				Path:     []string{"packages", "0", "shortDescription"},
			},
		}, nil)
		return clientValidator
	}

	testCases := []struct {
		Name                     string
		ClientValidatorFn        func() *automock.ValidatorClient
		InputDocument            string
		InputBaseURL             string
		ExpectedRuntimeError     error
		ExpectedValidationErrors []*ord.ValidationError
		ExpectedValidResources   func() ([]*ord.Document, error)
		Namespace                string
	}{
		{
			Name:              "Success without errors",
			ClientValidatorFn: validatorClientCallWithoutValidationErrors,
			InputDocument:     fmt.Sprintf(ordDocument, baseURL),
			InputBaseURL:      baseURL,
		},
		{
			Name: "Runtime error from API Metadata Validator",
			ClientValidatorFn: func() *automock.ValidatorClient {
				clientValidator := &automock.ValidatorClient{}
				clientValidator.On("Validate", mock.Anything, "", mock.Anything).Return([]model.ValidationResult{}, errors.New("Test runtime error"))
				return clientValidator
			},
			InputDocument:        fmt.Sprintf(ordDocument, baseURL),
			InputBaseURL:         baseURL,
			ExpectedRuntimeError: errors.New("error while validating document with API Metadata validator"),
		},
		{
			Name: "Validation errors only with severity level Error",
			ClientValidatorFn: func() *automock.ValidatorClient {
				clientValidator := &automock.ValidatorClient{}
				clientValidator.On("Validate", mock.Anything, "", mock.Anything).Return(validationResultsErrorSeverity, nil)
				return clientValidator
			},
			InputDocument:            fmt.Sprintf(ordDocument, baseURL),
			InputBaseURL:             baseURL,
			ExpectedValidationErrors: validationErrorsErrorSeverity,
		},
		{
			Name: "Validation errors with severity level Warning",
			ClientValidatorFn: func() *automock.ValidatorClient {
				clientValidator := &automock.ValidatorClient{}
				clientValidator.On("Validate", mock.Anything, "", mock.Anything).Return(validationResultsWarningSeverity, nil)
				return clientValidator
			},
			InputDocument:            fmt.Sprintf(ordDocument, baseURL),
			InputBaseURL:             baseURL,
			ExpectedValidationErrors: validationErrorsWarningSeverity,
		},
		{
			Name:                     "Validation error when there is a duplicate API resource",
			ClientValidatorFn:        validatorClientCallWithoutValidationErrors,
			InputDocument:            fmt.Sprintf(ordDocumentWithDuplicates, baseURL),
			InputBaseURL:             baseURL,
			ExpectedValidationErrors: validationErrorDuplicateResources,
		},
		{
			Name:                     "Validation error when there is a resource with unknown reference",
			ClientValidatorFn:        validatorClientCallWithoutValidationErrors,
			InputDocument:            fmt.Sprintf(ordDocumentAPIHasUnknownReference, baseURL),
			InputBaseURL:             baseURL,
			ExpectedValidationErrors: validationErrorUnknownReference,
		},
		{
			Name:                     "Validation error when baseUrl is missing",
			ClientValidatorFn:        validatorClientCallWithoutValidationErrors,
			InputDocument:            fmt.Sprintf(ordDocumentWithWrongBaseURL, ""),
			InputBaseURL:             "",
			ExpectedValidationErrors: validationErrorMissingBaseURL,
		},
		{
			Name:                     "Validation error when there is a mismatch between the given baseUrls",
			ClientValidatorFn:        validatorClientCallWithoutValidationErrors,
			InputDocument:            fmt.Sprintf(ordDocumentWithWrongBaseURL, "https://differentbase.com"),
			InputBaseURL:             baseURL,
			ExpectedValidationErrors: validationErrorMismatchedBaseURL,
		},
		{
			Name:                     "Should not delete failed resources when matching the ignored list",
			ClientValidatorFn:        validatorClientCallWithValidationErrors,
			InputDocument:            fmt.Sprintf(ordDocumentWithInvalidPackage, baseURL),
			InputBaseURL:             baseURL,
			Namespace:                testNamespace,
			ExpectedValidationErrors: validationErrorForInvalidPackage,
			ExpectedValidResources: func() ([]*ord.Document, error) {
				doc := &ord.Document{}
				if err := json.Unmarshal([]byte(fmt.Sprintf(ordDocumentWithInvalidPackage, baseURL)), &doc); err != nil {
					return nil, err
				}
				return []*ord.Document{doc}, nil
			},
		},
		{
			Name:                     "Should delete failed resources when not matching the ignored list",
			ClientValidatorFn:        validatorClientCallWithValidationErrors,
			InputDocument:            fmt.Sprintf(ordDocumentWithInvalidPackage, baseURL),
			InputBaseURL:             baseURL,
			Namespace:                "non.ignorelisted",
			ExpectedValidationErrors: validationErrorForInvalidPackage,
			ExpectedValidResources: func() ([]*ord.Document, error) {
				doc := &ord.Document{}

				// remove the invalid resource (package) from the doc
				resources, err := sjson.SetRaw(fmt.Sprintf(ordDocumentWithInvalidPackage, baseURL), "packages", `[]`)
				if err != nil {
					return nil, err
				}

				if err = json.Unmarshal([]byte(resources), &doc); err != nil {
					return nil, err
				}
				return []*ord.Document{doc}, nil
			},
		},
	}

	for _, test := range testCases {
		t.Run(test.Name, func(t *testing.T) {
			validator := ord.NewDocumentValidator(test.ClientValidatorFn(), ignorelistMapping)

			result := &ord.Document{}
			err := json.Unmarshal([]byte(test.InputDocument), &result)
			if err != nil {
				return
			}
			ordDoc := []*ord.Document{result}

			validationErrors, err := validator.Validate(context.TODO(), ordDoc, test.InputBaseURL, map[string]bool{}, []string{test.InputDocument}, "", test.Namespace)

			if test.ExpectedRuntimeError != nil {
				require.Error(t, err)
				require.Contains(t, err.Error(), test.ExpectedRuntimeError.Error())
			} else {
				require.NoError(t, err)
			}

			if len(test.ExpectedValidationErrors) > 0 {
				require.Equal(t, len(test.ExpectedValidationErrors), len(validationErrors))

				for i, currentError := range validationErrors {
					require.Equal(t, test.ExpectedValidationErrors[i].OrdID, currentError.OrdID)
					require.Equal(t, test.ExpectedValidationErrors[i].Severity, currentError.Severity)
					require.Equal(t, test.ExpectedValidationErrors[i].Type, currentError.Type)
					require.Equal(t, test.ExpectedValidationErrors[i].Description, currentError.Description)
				}
			} else {
				require.Empty(t, validationErrors)
			}

			if test.ExpectedValidResources != nil {
				expectedResources, err := test.ExpectedValidResources()
				require.NoError(t, err)
				require.Equal(t, ordDoc, expectedResources)
			}
		})
	}
}
