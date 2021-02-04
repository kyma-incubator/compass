package spec_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/spec"
	"github.com/kyma-incubator/compass/components/director/internal/domain/spec/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/pkg/errors"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConverter_ToGraphQLAPISpec(t *testing.T) {
	// given
	testCases := []struct {
		Name      string
		Input     *model.Spec
		Expected  *graphql.APISpec
		ExpectErr bool
	}{
		{
			Name:     "All properties given",
			Input:    fixModelAPISpec(),
			Expected: fixGQLAPISpec(),
		},
		{
			Name:      "Referenced ObjectType is not API should return error",
			Input:     fixModelEventSpec(),
			ExpectErr: true,
		},
		{
			Name: "APIType is nil should return error",
			Input: &model.Spec{
				ObjectType: model.APISpecReference,
			},
			ExpectErr: true,
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := spec.NewConverter(&automock.FetchRequestConverter{})

			// when
			res, err := converter.ToGraphQLAPISpec(testCase.Input)

			// then
			if testCase.ExpectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_ToGraphQLEventSpec(t *testing.T) {
	// given
	testCases := []struct {
		Name      string
		Input     *model.Spec
		Expected  *graphql.EventSpec
		ExpectErr bool
	}{
		{
			Name:     "All properties given",
			Input:    fixModelEventSpec(),
			Expected: fixGQLEventSpec(),
		},
		{
			Name:      "Referenced ObjectType is not Event should return error",
			Input:     fixModelAPISpec(),
			ExpectErr: true,
		},
		{
			Name: "EventType is nil should return error",
			Input: &model.Spec{
				ObjectType: model.EventSpecReference,
			},
			ExpectErr: true,
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			converter := spec.NewConverter(&automock.FetchRequestConverter{})

			// when
			res, err := converter.ToGraphQLEventSpec(testCase.Input)

			// then
			if testCase.ExpectErr {
				assert.Error(t, err)
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_InputFromGraphQLAPISpec(t *testing.T) {
	testErr := errors.New("test")

	// given
	testCases := []struct {
		Name               string
		FetchRequestConvFn func() *automock.FetchRequestConverter
		Input              *graphql.APISpecInput
		Expected           *model.SpecInput
		ExpectedErr        error
	}{
		{
			Name:  "All properties given",
			Input: fixGQLAPISpecInputWithFetchRequest(),
			FetchRequestConvFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("InputFromGraphQL", fixGQLAPISpecInputWithFetchRequest().FetchRequest).Return(fixModelAPISpecInputWithFetchRequest().FetchRequest, nil).Once()
				return conv
			},
			Expected: fixModelAPISpecInputWithFetchRequest(),
		},
		{
			Name:  "Return error when FetchRequest convertion fails",
			Input: fixGQLAPISpecInputWithFetchRequest(),
			FetchRequestConvFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("InputFromGraphQL", fixGQLAPISpecInputWithFetchRequest().FetchRequest).Return(nil, testErr).Once()
				return conv
			},
			ExpectedErr: testErr,
		},
		{
			Name:  "Nil",
			Input: nil,
			FetchRequestConvFn: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			frConv := testCase.FetchRequestConvFn()
			converter := spec.NewConverter(frConv)

			// when
			res, err := converter.InputFromGraphQLAPISpec(testCase.Input)

			// then
			// then
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.Expected, res)
			frConv.AssertExpectations(t)
		})
	}
}

func TestConverter_InputFromGraphQLEventSpec(t *testing.T) {
	testErr := errors.New("test")

	// given
	testCases := []struct {
		Name               string
		FetchRequestConvFn func() *automock.FetchRequestConverter
		Input              *graphql.EventSpecInput
		Expected           *model.SpecInput
		ExpectedErr        error
	}{
		{
			Name:  "All properties given",
			Input: fixGQLEventSpecInputWithFetchRequest(),
			FetchRequestConvFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("InputFromGraphQL", fixGQLEventSpecInputWithFetchRequest().FetchRequest).Return(fixModelEventSpecInputWithFetchRequest().FetchRequest, nil).Once()
				return conv
			},
			Expected: fixModelEventSpecInputWithFetchRequest(),
		},
		{
			Name:  "Return error when FetchRequest convertion fails",
			Input: fixGQLEventSpecInputWithFetchRequest(),
			FetchRequestConvFn: func() *automock.FetchRequestConverter {
				conv := &automock.FetchRequestConverter{}
				conv.On("InputFromGraphQL", fixGQLEventSpecInputWithFetchRequest().FetchRequest).Return(nil, testErr).Once()
				return conv
			},
			ExpectedErr: testErr,
		},
		{
			Name:  "Nil",
			Input: nil,
			FetchRequestConvFn: func() *automock.FetchRequestConverter {
				return &automock.FetchRequestConverter{}
			},
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			frConv := testCase.FetchRequestConvFn()
			converter := spec.NewConverter(frConv)

			// when
			res, err := converter.InputFromGraphQLEventSpec(testCase.Input)

			// then
			// then
			if testCase.ExpectedErr != nil {
				assert.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				assert.NoError(t, err)
			}
			assert.Equal(t, testCase.Expected, res)
			frConv.AssertExpectations(t)
		})
	}
}

func TestConverter_FromEntity(t *testing.T) {
	// given
	testCases := []struct {
		Name               string
		Input              spec.Entity
		Expected           model.Spec
		ExpectedErrMessage string
	}{
		{
			Name:               "All properties given for API",
			Input:              fixAPISpecEntity(),
			Expected:           *fixModelAPISpec(),
			ExpectedErrMessage: "",
		},
		{
			Name:               "All properties given for Event",
			Input:              fixEventSpecEntity(),
			Expected:           *fixModelEventSpec(),
			ExpectedErrMessage: "",
		},
		{
			Name: "Error when no reference entity",
			Input: spec.Entity{
				ID:       "2",
				TenantID: "tenant",
			},
			ExpectedErrMessage: "while determining object reference: incorrect Object Reference ID and its type for Entity with ID '2'",
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := spec.NewConverter(&automock.FetchRequestConverter{})

			// when
			res, err := conv.FromEntity(testCase.Input)

			if testCase.ExpectedErrMessage != "" {
				require.Error(t, err)
				assert.Equal(t, testCase.ExpectedErrMessage, err.Error())
				return
			}

			// then
			require.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_ToEntity(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    model.Spec
		Expected spec.Entity
	}{
		{
			Name:     "All properties given for API",
			Input:    *fixModelAPISpec(),
			Expected: fixAPISpecEntity(),
		},
		{
			Name:     "All properties given for Event",
			Input:    *fixModelEventSpec(),
			Expected: fixEventSpecEntity(),
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			conv := spec.NewConverter(&automock.FetchRequestConverter{})

			// when
			res := conv.ToEntity(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}
