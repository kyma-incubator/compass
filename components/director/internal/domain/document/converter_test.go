package document_test

import (
	"database/sql"
	"testing"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
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
			Expected: fixGQLDocument("1", "foo"),
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

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			frConv := &automock.FetchRequestConverter{}
			if testCase.Input != nil {
				frConv.On("ToGraphQL", testCase.Input.FetchRequest).Return(testCase.Expected.FetchRequest)
			}
			converter := document.NewConverter(frConv)

			// when
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			frConv.AssertExpectations(t)
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
		fixGQLDocument("1", "foo"),
		fixGQLDocument("2", "bar"),
		{},
	}
	frConv := &automock.FetchRequestConverter{}
	frConv.On("ToGraphQL", input[0].FetchRequest).Return(expected[0].FetchRequest)
	frConv.On("ToGraphQL", (*model.FetchRequest)(nil)).Return(nil)
	converter := document.NewConverter(frConv)

	// when
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
	frConv.AssertExpectations(t)
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

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			frConv := &automock.FetchRequestConverter{}
			if testCase.Input != nil {
				frConv.On("InputFromGraphQL", testCase.Input.FetchRequest).Return(testCase.Expected.FetchRequest)
			}
			converter := document.NewConverter(frConv)

			// when
			res := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			frConv.AssertExpectations(t)
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
	frConv := &automock.FetchRequestConverter{}
	frConv.On("InputFromGraphQL", input[0].FetchRequest).Return(expected[0].FetchRequest)
	frConv.On("InputFromGraphQL", (*graphql.FetchRequestInput)(nil)).Return(nil)
	converter := document.NewConverter(frConv)

	// when
	res := converter.MultipleInputFromGraphQL(input)

	// then
	assert.Equal(t, expected, res)
	frConv.AssertExpectations(t)
}

func TestToEntity(t *testing.T) {
	// GIVEN
	frConv := &automock.FetchRequestConverter{}
	sut := document.NewConverter(frConv)
	// WHEN
	modelDocument := fixModelDocument("id", "appID")
	actual, err := sut.ToEntity(*modelDocument)
	// THEN
	require.NoError(t, err)
	assert.Equal(t, "tenant", actual.TenantID)
	assert.Equal(t, "id", actual.ID)
	assert.Equal(t, "appID", actual.AppID)
}

func TestFromEntity(t *testing.T) {
	// GIVEN
	frConv := &automock.FetchRequestConverter{}
	sut := document.NewConverter(frConv)
	// WHEN
	id := "id"
	tenantID := "tenantID"

	actual, err := sut.FromEntity(document.Entity{
		ID:          id,
		TenantID:    tenantID,
		AppID:       "appID",
		Title:       "title",
		DisplayName: "dispName",
		Description: "desc",
		Format:      "MARKDOWN",
		Kind: sql.NullString{
			String: "kind",
			Valid:  true,
		},
		Data: sql.NullString{
			String: "data",
			Valid:  true,
		},
		FetchRequest: sql.NullString{},
	})
	// THEN
	require.NoError(t, err)
	assert.Equal(t, id, actual.ID)
	assert.Equal(t, tenantID, actual.Tenant)
	assert.Equal(t, "kind", *actual.Kind)
}
