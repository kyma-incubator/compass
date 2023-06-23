package document_test

import (
	"database/sql"
	"testing"
	"time"

	"github.com/kyma-incubator/compass/components/director/internal/repo"

	"github.com/kyma-incubator/compass/components/director/pkg/str"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/domain/document"
	"github.com/kyma-incubator/compass/components/director/internal/domain/document/automock"
	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *model.Document
		Expected *graphql.Document
	}{
		{
			Name:     "All properties given",
			Input:    fixModelDocumentForApp("1", "foo"),
			Expected: fixGQLDocument("1", "foo"),
		},
		{
			Name:     "Empty",
			Input:    &model.Document{BaseEntity: &model.BaseEntity{}},
			Expected: &graphql.Document{BaseEntity: &graphql.BaseEntity{}},
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

			// WHEN
			res := converter.ToGraphQL(testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			frConv.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// GIVEN
	input := []*model.Document{
		fixModelDocumentForApp("1", "foo"),
		fixModelDocumentForApp("2", "bar"),
		{BaseEntity: &model.BaseEntity{}},
		nil,
	}
	expected := []*graphql.Document{
		fixGQLDocument("1", "foo"),
		fixGQLDocument("2", "bar"),
		{BaseEntity: &graphql.BaseEntity{}},
	}
	frConv := &automock.FetchRequestConverter{}
	converter := document.NewConverter(frConv)

	// WHEN
	res := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
	frConv.AssertExpectations(t)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	// GIVEN
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
				frConv.On("InputFromGraphQL", testCase.Input.FetchRequest).Return(testCase.Expected.FetchRequest, nil)
			}
			converter := document.NewConverter(frConv)

			// WHEN
			res, err := converter.InputFromGraphQL(testCase.Input)

			// then
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
			frConv.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleInputFromGraphQL(t *testing.T) {
	// GIVEN
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
	frConv.On("InputFromGraphQL", input[0].FetchRequest).Return(expected[0].FetchRequest, nil)
	frConv.On("InputFromGraphQL", (*graphql.FetchRequestInput)(nil)).Return(nil, nil)
	converter := document.NewConverter(frConv)

	// WHEN
	res, err := converter.MultipleInputFromGraphQL(input)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
	frConv.AssertExpectations(t)
}

func TestToEntity(t *testing.T) {
	// GIVEN
	sut := document.NewConverter(nil)

	modelWithRequiredFields := model.Document{
		BundleID:    "givenBundleID",
		AppID:       str.Ptr("givenAppID"),
		Title:       "givenTitle",
		Description: "givenDescription",
		DisplayName: "givenDisplayName",
		Format:      "givenFormat",
		BaseEntity: &model.BaseEntity{
			ID:        "givenID",
			Ready:     true,
			CreatedAt: &fixedTimestamp,
			UpdatedAt: &time.Time{},
			DeletedAt: &time.Time{},
			Error:     nil,
		},
	}

	t.Run("only required fields", func(t *testing.T) {
		givenModel := modelWithRequiredFields
		// WHEN
		actual, err := sut.ToEntity(&givenModel)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, &document.Entity{
			BndlID:      givenModel.BundleID,
			AppID:       repo.NewNullableString(givenModel.AppID),
			Title:       givenModel.Title,
			Description: givenModel.Description,
			DisplayName: givenModel.DisplayName,
			Format:      string(givenModel.Format),
			BaseEntity: &repo.BaseEntity{
				ID:        givenModel.ID,
				Ready:     givenModel.Ready,
				CreatedAt: givenModel.CreatedAt,
				UpdatedAt: givenModel.UpdatedAt,
				DeletedAt: givenModel.DeletedAt,
				Error:     repo.NewNullableString(givenModel.Error),
			},
		}, actual)
	})

	t.Run("all fields", func(t *testing.T) {
		givenModel := modelWithRequiredFields
		givenModel.Data = str.Ptr("givenData")
		givenModel.Kind = str.Ptr("givenKind")
		// WHEN
		actual, err := sut.ToEntity(&givenModel)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, sql.NullString{Valid: true, String: "givenData"}, actual.Data)
		assert.Equal(t, sql.NullString{Valid: true, String: "givenKind"}, actual.Kind)
	})
}

func TestFromEntity(t *testing.T) {
	// GIVEN
	sut := document.NewConverter(nil)
	entityWithRequiredFields := document.Entity{
		BndlID:      "givenBundleID",
		AppID:       repo.NewValidNullableString("givenAppID"),
		Title:       "givenTitle",
		DisplayName: "givenDisplayName",
		Description: "givenDescription",
		Format:      "MARKDOWN",
		BaseEntity: &repo.BaseEntity{
			ID: "givenID",
		},
	}

	t.Run("only required fields", func(t *testing.T) {
		givenEntity := entityWithRequiredFields
		// WHEN
		actualModel, err := sut.FromEntity(&givenEntity)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, &model.Document{
			BundleID:    "givenBundleID",
			AppID:       str.Ptr("givenAppID"),
			Title:       "givenTitle",
			DisplayName: "givenDisplayName",
			Description: "givenDescription",
			Format:      model.DocumentFormatMarkdown,
			BaseEntity: &model.BaseEntity{
				ID: "givenID",
			},
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

		// WHEN
		actualModel, err := sut.FromEntity(&givenEntity)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, str.Ptr("givenData"), actualModel.Data)
		assert.Equal(t, str.Ptr("givenKind"), actualModel.Kind)
	})
}
