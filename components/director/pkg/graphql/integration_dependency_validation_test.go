package graphql_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/inputvalidation/inputvalidationtest"
	"github.com/kyma-incubator/compass/components/director/pkg/str"
	"github.com/stretchr/testify/require"
)

// validate integration dependency input
func TestIntegrationDependencyInput_Validate_Name(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "name-123.com",
			ExpectedValid: true,
		},
		{
			Name:          "Valid Printable ASCII",
			Value:         "V1 +=_-)(*&^%$#@!?/>.<,|\\\"':;}{][",
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 255 chars",
			Value:         inputvalidationtest.String257Long,
			ExpectedValid: false,
		},
		{
			Name:          "String contains invalid ASCII",
			Value:         "ąćńłóęǖǘǚǜ",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidIntegrationDependencyInput()
			obj.Name = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestIntegrationDependencyInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("this is a valid description"),
			ExpectedValid: true,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         str.Ptr(inputvalidationtest.EmptyString),
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 2000 chars",
			Value:         str.Ptr(inputvalidationtest.String2001Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidIntegrationDependencyInput()
			obj.Description = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestIntegrationDependencyInput_Validate_OrdID(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("sap.foo.bar:integrationDependency:CustomerOrder:v1"),
			ExpectedValid: true,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         str.Ptr(inputvalidationtest.EmptyString),
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 255 chars",
			Value:         str.Ptr(inputvalidationtest.String257Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidIntegrationDependencyInput()
			obj.OrdID = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestIntegrationDependencyInput_Validate_PartOfPackage(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("sap.xref:package:SomePackage:v1"),
			ExpectedValid: true,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         str.Ptr(inputvalidationtest.EmptyString),
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 255 chars",
			Value:         str.Ptr(inputvalidationtest.String257Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidIntegrationDependencyInput()
			obj.PartOfPackage = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestIntegrationDependencyInput_Validate_Visibility(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("public"),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("internal"),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("private"),
			ExpectedValid: true,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         str.Ptr(inputvalidationtest.EmptyString),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidIntegrationDependencyInput()
			obj.Visibility = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestIntegrationDependencyInput_Validate_ReleaseStatus(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("beta"),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("active"),
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("deprecated"),
			ExpectedValid: true,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         str.Ptr(inputvalidationtest.EmptyString),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidIntegrationDependencyInput()
			obj.ReleaseStatus = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestIntegrationDependencyInput_Validate_Mandatory(t *testing.T) {
	mandatoryTrue := true
	mandatoryFalse := false
	testCases := []struct {
		Name          string
		Value         *bool
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &mandatoryTrue,
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid",
			Value:         &mandatoryFalse,
			ExpectedValid: true,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidIntegrationDependencyInput()
			obj.Mandatory = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestIntegrationDependencyInput_Validate_Version(t *testing.T) {
	validObj := fixValidVersionInput()
	emptyObj := graphql.VersionInput{}

	testCases := []struct {
		Name          string
		Value         *graphql.VersionInput
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid obj",
			Value:         &validObj,
			ExpectedValid: true,
		},
		{
			Name:          "Nil object",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Invalid object",
			Value:         &emptyObj,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidIntegrationDependencyInput()
			obj.Version = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// validate aspect input
func TestAspectInput_Validate_Name(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "name-123.com",
			ExpectedValid: true,
		},
		{
			Name:          "Valid Printable ASCII",
			Value:         "V1 +=_-)(*&^%$#@!?/>.<,|\\\"':;}{][",
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 255 chars",
			Value:         inputvalidationtest.String257Long,
			ExpectedValid: false,
		},
		{
			Name:          "String contains invalid ASCII",
			Value:         "ąćńłóęǖǘǚǜ",
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAspectInput()
			obj.Name = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAspectInput_Validate_Description(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("this is a valid description"),
			ExpectedValid: true,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         str.Ptr(inputvalidationtest.EmptyString),
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 2000 chars",
			Value:         str.Ptr(inputvalidationtest.String2001Long),
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAspectInput()
			obj.Description = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func TestAspectInput_Validate_Mandatory(t *testing.T) {
	mandatoryTrue := true
	mandatoryFalse := false
	testCases := []struct {
		Name          string
		Value         *bool
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         &mandatoryTrue,
			ExpectedValid: true,
		},
		{
			Name:          "ExpectedValid",
			Value:         &mandatoryFalse,
			ExpectedValid: true,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: true,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAspectInput()
			obj.Mandatory = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// validate AspectAPIDefinitionInput
func TestAspectAPIDefinitionInput_Validate_OrdID(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "sap.s4:apiResource:API_BILL_OF_MATERIAL_SRV:v1",
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 255 chars",
			Value:         inputvalidationtest.String257Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAspectAPIDefinitionInput()
			obj.OrdID = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// validate AspectEventDefinitionInput
func TestAspectEventDefinitionInput_Validate_OrdID(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         "sap.billing.sb:eventResource:BusinessEvents_SubscriptionEvents:v1",
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         inputvalidationtest.EmptyString,
			ExpectedValid: false,
		},
		{
			Name:          "String longer than 255 chars",
			Value:         inputvalidationtest.String257Long,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAspectEventDefinitionInput()
			obj.OrdID = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

// validate AspectEventDefinitionSubsetInput
func TestAspectEventDefinitionSubsetInput_Validate_EventType(t *testing.T) {
	testCases := []struct {
		Name          string
		Value         *string
		ExpectedValid bool
	}{
		{
			Name:          "ExpectedValid",
			Value:         str.Ptr("sap.cic.Order.OrderTransferred.v1"),
			ExpectedValid: true,
		},
		{
			Name:          "Empty string",
			Value:         str.Ptr(inputvalidationtest.EmptyString),
			ExpectedValid: false,
		},
		{
			Name:          "Nil pointer",
			Value:         nil,
			ExpectedValid: false,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			//GIVEN
			obj := fixValidAspectEventDefinitionSubsetInput()
			obj.EventType = testCase.Value
			// WHEN
			err := obj.Validate()
			// THEN
			if testCase.ExpectedValid {
				require.NoError(t, err)
			} else {
				require.Error(t, err)
			}
		})
	}
}

func fixValidIntegrationDependencyInput() graphql.IntegrationDependencyInput {
	return graphql.IntegrationDependencyInput{
		Name: inputvalidationtest.ValidName,
	}
}

func fixValidAspectInput() graphql.AspectInput {
	return graphql.AspectInput{
		Name: inputvalidationtest.ValidName,
	}
}

func fixValidAspectAPIDefinitionInput() graphql.AspectAPIDefinitionInput {
	return graphql.AspectAPIDefinitionInput{
		OrdID: "sap.s4:apiResource:API_BILL_OF_MATERIAL_SRV:v1",
	}
}

func fixValidAspectEventDefinitionInput() graphql.AspectEventDefinitionInput {
	return graphql.AspectEventDefinitionInput{
		OrdID:  "sap.billing.sb:eventResource:BusinessEvents_SubscriptionEvents:v1",
		Subset: []*graphql.AspectEventDefinitionSubsetInput{{EventType: str.Ptr("sap.cic.Order.OrderTransferred.v1")}},
	}
}

func fixValidAspectEventDefinitionSubsetInput() graphql.AspectEventDefinitionSubsetInput {
	return graphql.AspectEventDefinitionSubsetInput{
		EventType: str.Ptr("sap.cic.Order.OrderTransferred.v1"),
	}
}
