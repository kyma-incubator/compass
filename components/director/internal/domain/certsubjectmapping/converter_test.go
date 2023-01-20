package certsubjectmapping_test

import (
	"github.com/kyma-incubator/compass/components/director/internal/domain/certsubjectmapping"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
	"testing"
)

var converter = certsubjectmapping.NewConverter()

func TestConverter_ToGraphQL(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          *model.CertSubjectMapping
		Expected       *graphql.CertificateSubjectMapping
		ExpectedErrMsg string
	}{
		{
			Name:     "Success",
			Input:    CertSubjectMappingModel,
			Expected: CertSubjectMappingGQLModel,
		},
		{
			Name:     "Success when input is nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			r := converter.ToGraphQL(testCase.Input)

			// THEN
			require.Equal(t, r, testCase.Expected)
		})
	}
}

func TestConverter_FromGraphql(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          graphql.CertificateSubjectMappingInput
		Expected       *model.CertSubjectMapping
		ExpectedErrMsg string
	}{
		{
			Name:     "Success",
			Input:    CertSubjectMappingGQLModelInput,
			Expected: CertSubjectMappingModel,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			r := converter.FromGraphql(TestID, testCase.Input)

			// THEN
			require.Equal(t, r, testCase.Expected)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          []*model.CertSubjectMapping
		Expected       []*graphql.CertificateSubjectMapping
		ExpectedErrMsg string
	}{
		{
			Name:     "Success",
			Input:    []*model.CertSubjectMapping{CertSubjectMappingModel},
			Expected: []*graphql.CertificateSubjectMapping{CertSubjectMappingGQLModel},
		},
		{
			Name:     "Success when input is nil",
			Input:    nil,
			Expected: nil,
		},
		{
			Name:     "Success when in the input slice there is a nil element",
			Input:    []*model.CertSubjectMapping{nil},
			Expected: []*graphql.CertificateSubjectMapping{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			r := converter.MultipleToGraphQL(testCase.Input)

			// THEN
			require.Equal(t, r, testCase.Expected)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          *model.CertSubjectMapping
		Expected       *certsubjectmapping.Entity
		ExpectedErrMsg string
	}{
		{
			Name:     "Success",
			Input:    CertSubjectMappingModel,
			Expected: CertSubjectMappingEntity,
		},
		{
			Name:     "Success when input is nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			r, err := converter.ToEntity(testCase.Input)

			if testCase.ExpectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, r, testCase.Expected)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	testCases := []struct {
		Name           string
		Input          *certsubjectmapping.Entity
		Expected       *model.CertSubjectMapping
		ExpectedErrMsg string
	}{
		{
			Name:     "Success",
			Input:    CertSubjectMappingEntity,
			Expected: CertSubjectMappingModel,
		},
		{
			Name:     "Success when input is nil",
			Input:    nil,
			Expected: nil,
		},
		{
			Name: "Error when unmarhalling fails",
			Input: CertSubjectMappingEntityInvalidTntAccessLevels,
			Expected: nil,
			ExpectedErrMsg: "while unmarshalling tenant access levels",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			r, err := converter.FromEntity(testCase.Input)

			if testCase.ExpectedErrMsg != "" {
				require.Error(t, err)
				require.Contains(t, err.Error(), testCase.ExpectedErrMsg)
			} else {
				require.NoError(t, err)
			}

			// THEN
			require.Equal(t, r, testCase.Expected)
		})
	}
}
