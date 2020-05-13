package service_test

import (
	"fmt"
	"testing"

	"github.com/pkg/errors"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service"
)

const legacyServicesLabelKey = "legacy_servicesMetadata"

func TestLabeler_WriteServiceReference(t *testing.T) {
	// GIVEN
	labeler := service.NewAppLabeler()

	testCases := []struct {
		Name              string
		InputLabels       graphql.Labels
		InputSvcReference service.LegacyServiceReference
		ExpectedOutput    graphql.LabelInput
		ExpectedError     error
	}{
		{
			Name: "Success",
			InputLabels: graphql.Labels{
				legacyServicesLabelKey: "{\"foo\":{\"id\":\"foo\",\"identifier\":\"bar\"}}",
			},
			InputSvcReference: service.LegacyServiceReference{
				ID:         "biz",
				Identifier: "baz",
			},
			ExpectedOutput: graphql.LabelInput{
				Key:   legacyServicesLabelKey,
				Value: "\"{\\\"biz\\\":{\\\"id\\\":\\\"biz\\\",\\\"identifier\\\":\\\"baz\\\"},\\\"foo\\\":{\\\"id\\\":\\\"foo\\\",\\\"identifier\\\":\\\"bar\\\"}}\"",
			},
			ExpectedError: nil,
		},
		{
			Name:        "Success when map value is nil",
			InputLabels: graphql.Labels{},
			InputSvcReference: service.LegacyServiceReference{
				ID:         "foo",
				Identifier: "bar",
			},
			ExpectedOutput: graphql.LabelInput{
				Key:   legacyServicesLabelKey,
				Value: "\"{\\\"foo\\\":{\\\"id\\\":\\\"foo\\\",\\\"identifier\\\":\\\"bar\\\"}}\"",
			},
			ExpectedError: nil,
		},
		{
			Name: "Error when value is not a string",
			InputLabels: graphql.Labels{
				legacyServicesLabelKey: 10,
			},
			InputSvcReference: service.LegacyServiceReference{
				ID:         "foo",
				Identifier: "bar",
			},
			ExpectedOutput: graphql.LabelInput{},
			ExpectedError:  errors.New("invalid type: expected: string; actual: int"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			actual, err := labeler.WriteServiceReference(testCase.InputLabels, testCase.InputSvcReference)
			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
			assert.Equal(t, testCase.ExpectedOutput, actual)
		})
	}
}

func TestLabeler_ReadServiceReference(t *testing.T) {
	// GIVEN
	labeler := service.NewAppLabeler()
	svcID := "foo"
	svcIdentifier := "bar"

	testCases := []struct {
		Name           string
		InputLabels    graphql.Labels
		ExpectedOutput service.LegacyServiceReference
		ExpectedError  error
	}{
		{
			Name: "Success",
			InputLabels: graphql.Labels{
				legacyServicesLabelKey: fmt.Sprintf("{\"%[1]s\":{\"id\":\"%[1]s\",\"identifier\":\"%s\"}}", svcID, svcIdentifier),
			},
			ExpectedOutput: service.LegacyServiceReference{
				ID:         svcID,
				Identifier: svcIdentifier,
			},
			ExpectedError: nil,
		},
		{
			Name:           "Success when map value is nil",
			InputLabels:    graphql.Labels{},
			ExpectedOutput: service.LegacyServiceReference{},
			ExpectedError:  nil,
		},
		{
			Name: "Error when value is not a string",
			InputLabels: graphql.Labels{
				legacyServicesLabelKey: 10,
			},
			ExpectedOutput: service.LegacyServiceReference{},
			ExpectedError:  errors.New("invalid type: expected: string; actual: int"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			actual, err := labeler.ReadServiceReference(testCase.InputLabels, svcID)
			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
			assert.Equal(t, testCase.ExpectedOutput, actual)
		})
	}
}

func TestLabeler_ListServiceReferences(t *testing.T) {
	// GIVEN
	labeler := service.NewAppLabeler()

	testCases := []struct {
		Name           string
		InputLabels    graphql.Labels
		ExpectedOutput []service.LegacyServiceReference
		ExpectedError  error
	}{
		{
			Name: "Success",
			InputLabels: graphql.Labels{
				legacyServicesLabelKey: "{\"foo\":{\"id\":\"foo\",\"identifier\":\"foo\"}, \"bar\":{\"id\":\"bar\",\"identifier\":\"bar\"}}",
			},
			ExpectedOutput: []service.LegacyServiceReference{
				{
					ID:         "foo",
					Identifier: "foo",
				},
				{
					ID:         "bar",
					Identifier: "bar",
				},
			},
			ExpectedError: nil,
		},
		{
			Name:           "Success when map value is nil",
			InputLabels:    graphql.Labels{},
			ExpectedOutput: nil,
			ExpectedError:  nil,
		},
		{
			Name: "Error when value is not a string",
			InputLabels: graphql.Labels{
				legacyServicesLabelKey: 10,
			},
			ExpectedOutput: []service.LegacyServiceReference{},
			ExpectedError:  errors.New("invalid type: expected: string; actual: int"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			actual, err := labeler.ListServiceReferences(testCase.InputLabels)
			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
			assert.ElementsMatch(t, testCase.ExpectedOutput, actual)
		})
	}
}

func TestLabeler_DeleteServiceReference(t *testing.T) {
	// GIVEN
	labeler := service.NewAppLabeler()
	svcID := "foo"

	testCases := []struct {
		Name           string
		InputLabels    graphql.Labels
		ExpectedOutput graphql.LabelInput
		ExpectedError  error
	}{
		{
			Name: "Success",
			InputLabels: graphql.Labels{
				legacyServicesLabelKey: "{\"foo\":{\"id\":\"foo\",\"identifier\":\"foo\"}, \"bar\":{\"id\":\"bar\",\"identifier\":\"bar\"}}",
			},
			ExpectedOutput: graphql.LabelInput{
				Key:   "legacy_servicesMetadata",
				Value: "\"{\\\"bar\\\":{\\\"id\\\":\\\"bar\\\",\\\"identifier\\\":\\\"bar\\\"}}\"",
			},
			ExpectedError: nil,
		},
		{
			Name:        "Success when map value is nil",
			InputLabels: graphql.Labels{},
			ExpectedOutput: graphql.LabelInput{
				Key:   "legacy_servicesMetadata",
				Value: "\"{}\"",
			},
			ExpectedError: nil,
		},
		{
			Name: "Error when value is not a string",
			InputLabels: graphql.Labels{
				legacyServicesLabelKey: 10,
			},
			ExpectedOutput: graphql.LabelInput{},
			ExpectedError:  errors.New("invalid type: expected: string; actual: int"),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			actual, err := labeler.DeleteServiceReference(testCase.InputLabels, svcID)
			// THEN
			if testCase.ExpectedError == nil {
				require.NoError(t, err)
			} else {
				require.EqualError(t, err, testCase.ExpectedError.Error())
			}
			assert.Equal(t, testCase.ExpectedOutput, actual)
		})
	}
}
