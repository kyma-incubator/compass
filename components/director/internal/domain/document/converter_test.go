package document_test

import (
	"database/sql"
	"testing"

	"github.com/kyma-incubator/compass/components/director/pkg/strings"

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
	sut := document.NewConverter(nil)

	modelWithRequiredFields := model.Document{
		ID:            "givenID",
		Tenant:        "givenTenant",
		ApplicationID: "givenApplicationID",
		Title:         "givenTitle",
		Description:   "givenDescription",
		DisplayName:   "givenDisplayName",
		Format:        "givenFormat",
	}

	t.Run("only required fields", func(t *testing.T) {
		givenModel := modelWithRequiredFields
		// WHEN
		actual, err := sut.ToEntity(givenModel)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, document.Entity{
			ID:          "givenID",
			TenantID:    "givenTenant",
			AppID:       "givenApplicationID",
			Title:       "givenTitle",
			Description: "givenDescription",
			DisplayName: "givenDisplayName",
			Format:      "givenFormat",
		}, actual)
	})

	t.Run("all fields", func(t *testing.T) {
		givenModel := modelWithRequiredFields
		givenModel.Data = strings.Ptr("givenData")
		givenModel.FetchRequestID = strings.Ptr("fetchRequestID")
		givenModel.Kind = strings.Ptr("givenKind")
		// WHEN
		actual, err := sut.ToEntity(givenModel)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, sql.NullString{Valid: true, String: "givenData"}, actual.Data)
		assert.Equal(t, sql.NullString{Valid: true, String: "givenKind"}, actual.Kind)
		assert.Equal(t, sql.NullString{Valid: true, String: "fetchRequestID"}, actual.FetchRequestID)
	})
}

func TestFromEntity(t *testing.T) {
	// GIVEN
	sut := document.NewConverter(nil)
	entityWithRequiredFields := document.Entity{
		ID:          "givenID",
		TenantID:    "givenTenant",
		AppID:       "givenAppID",
		Title:       "givenTitle",
		DisplayName: "givenDisplayName",
		Description: "givenDescription",
		Format:      "MARKDOWN",
	}

	t.Run("only required fields", func(t *testing.T) {
		givenEntity := entityWithRequiredFields
		// WHEN
		actualModel, err := sut.FromEntity(givenEntity)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, model.Document{
			ID:            "givenID",
			Tenant:        "givenTenant",
			ApplicationID: "givenAppID",
			Title:         "givenTitle",
			DisplayName:   "givenDisplayName",
			Description:   "givenDescription",
			Format:        model.DocumentFormatMarkdown,
		}, actualModel)

	})

	t.Run("all fields", func(t *testing.T) {
		givenEntity := entityWithRequiredFields
		givenEntity.Data = sql.NullString{
			Valid:  true,
			String: "givenData",
		}
		givenEntity.Kind = sql.NullString{
			Valid:  true,
			String: "givenKind",
		}
		givenEntity.FetchRequestID = sql.NullString{
			Valid:  true,
			String: "fetchRequestID",
		}

		// WHEN
		actualModel, err := sut.FromEntity(givenEntity)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, strings.Ptr("givenData"), actualModel.Data)
		assert.Equal(t, strings.Ptr("givenKind"), actualModel.Kind)
		assert.Equal(t, strings.Ptr("fetchRequestID"), actualModel.FetchRequestID)
	})
}
