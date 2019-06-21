package document_test

import (
	"fmt"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document/automock"
	"github.com/kyma-incubator/compass/components/director/internal/graphql"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *model.Document
		Expected *graphql.Document
	}{
		{
			Name:     "All properties given",
			Input:    fixModelDocument("1", "foo"),
			Expected: fixGQLDocument("foo"),
		},
		{
			Name:     "Empty",
			Input:    &model.Document{},
			Expected: &graphql.Document{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			frConv := &automock.FetchRequestConverter{}
			if testCase.Input != nil {
				frConv.On("ToGraphQL", testCase.Input.FetchRequest).Return(testCase.Expected.FetchRequest)
			}
			converter := document.NewConverter(frConv)
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// given
	input := []*model.Document{
		fixModelDocument("1", "foo"),
		fixModelDocument("2", "bar"),
		{},
		nil,
	}
	expected := []*graphql.Document{
		fixGQLDocument("foo"),
		fixGQLDocument("bar"),
		{},
	}
	var nilPointer *model.FetchRequest

	// when
	frConv := &automock.FetchRequestConverter{}
	frConv.On("ToGraphQL", input[0].FetchRequest).Return(expected[0].FetchRequest)
	frConv.On("ToGraphQL", nilPointer).Return(nil)
	converter := document.NewConverter(frConv)
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// given
	testCases := []struct {
		Name     string
		Input    *graphql.DocumentInput
		Expected *model.DocumentInput
	}{
		{
			Name:     "All properties given",
			Input:    fixGQLDocumentInput("foo"),
			Expected: fixModelDocumentInput("foo"),
		},
		{
			Name:     "Empty",
			Input:    &graphql.DocumentInput{},
			Expected: &model.DocumentInput{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for i, testCase := range testCases {
		t.Run(fmt.Sprintf("%d: %s", i, testCase.Name), func(t *testing.T) {
			// when
			frConv := &automock.FetchRequestConverter{}
			if testCase.Input != nil {
				frConv.On("InputFromGraphQL", testCase.Input.FetchRequest).Return(testCase.Expected.FetchRequest)
			}
			converter := document.NewConverter(frConv)
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	// given
	input := []*graphql.DocumentInput{
		fixGQLDocumentInput("foo"),
		fixGQLDocumentInput("bar"),
		{},
		nil,
	}
	expected := []*model.DocumentInput{
		fixModelDocumentInput("foo"),
		fixModelDocumentInput("bar"),
		{},
	}
	var nilPointer *graphql.FetchRequestInput

	// when
	frConv := &automock.FetchRequestConverter{}
	frConv.On("InputFromGraphQL", input[0].FetchRequest).Return(expected[0].FetchRequest)
	frConv.On("InputFromGraphQL", nilPointer).Return(nil)
	converter := document.NewConverter(frConv)
	res := converter.MultipleInputFromGraphQL(input)

	// then
	assert.Equal(t, expected, res)
}
