package bundle_test

import (
	"database/sql"
	"errors"
	"testing"

	"github.com/kyma-incubator/compass/components/director/internal/repo"
	"github.com/stretchr/testify/mock"

	"github.com/kyma-incubator/compass/components/director/internal/domain/auth"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle"

	"github.com/stretchr/testify/require"

	"github.com/kyma-incubator/compass/components/director/internal/model"

	"github.com/kyma-incubator/compass/components/director/internal/domain/bundle/automock"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestEntityConverter_ToEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		// GIVEN
		name := "foo"
		desc := "bar"
		bndlModel := fixBundleModel(name, desc)
		testErrMsg := "test-err"
		bndlModel.Error = &testErrMsg
		require.NotNil(t, bndlModel)
		authConv := auth.NewConverter()
		conv := bundle.NewConverter(authConv, nil, nil, nil)
		// WHEN
		entity, err := conv.ToEntity(bndlModel)
		// THEN
		require.NoError(t, err)

		expectedBndl := fixEntityBundleWithAppID(bundleID, name, desc)
		expectedBndl.Error = sql.NullString{
			String: testErrMsg,
			Valid:  true,
		}
		assert.Equal(t, expectedBndl, entity)
	})
	t.Run("success all nullable properties empty", func(t *testing.T) {
		// GIVEN
		name := "foo"
		bndlModel := &model.Bundle{
			ApplicationID:                  &appID,
			Name:                           name,
			Description:                    nil,
			InstanceAuthRequestInputSchema: nil,
			DefaultInstanceAuth:            nil,
			BaseEntity:                     &model.BaseEntity{ID: bundleID},
		}

		expectedEntity := &bundle.Entity{
			ApplicationID:                 repo.NewValidNullableString(appID),
			Name:                          name,
			Description:                   sql.NullString{},
			InstanceAuthRequestJSONSchema: sql.NullString{},
			DefaultInstanceAuth:           sql.NullString{},
			BaseEntity:                    &repo.BaseEntity{ID: bundleID},
		}

		require.NotNil(t, bndlModel)
		authConv := auth.NewConverter()
		conv := bundle.NewConverter(authConv, nil, nil, nil)
		// WHEN
		entity, err := conv.ToEntity(bndlModel)
		// THEN
		require.NoError(t, err)
		assert.Equal(t, expectedEntity, entity)
	})
}

func TestEntityConverter_FromEntity(t *testing.T) {
	t.Run("success all nullable properties filled", func(t *testing.T) {
		// GIVEN
		name := "foo"
		desc := "bar"
		testErrMsg := "test-err"
		entity := fixEntityBundleWithAppID(bundleID, name, desc)
		entity.Error = sql.NullString{
			String: testErrMsg,
			Valid:  true,
		}
		authConv := auth.NewConverter()
		conv := bundle.NewConverter(authConv, nil, nil, nil)
		// WHEN
		bndlModel, err := conv.FromEntity(entity)
		// THEN
		require.NoError(t, err)
		expectedBdnl := fixBundleModel(name, desc)
		expectedBdnl.Error = &testErrMsg
		assert.Equal(t, expectedBdnl, bndlModel)
	})
	t.Run("success all nullable properties empty", func(t *testing.T) {
		// GIVEN
		name := "foo"
		entity := &bundle.Entity{
			ApplicationID:                 repo.NewValidNullableString(appID),
			Name:                          name,
			Description:                   sql.NullString{},
			InstanceAuthRequestJSONSchema: sql.NullString{},
			DefaultInstanceAuth:           sql.NullString{},
			BaseEntity:                    &repo.BaseEntity{ID: bundleID},
		}
		expectedModel := &model.Bundle{
			ApplicationID:                  &appID,
			Name:                           name,
			Description:                    nil,
			InstanceAuthRequestInputSchema: nil,
			DefaultInstanceAuth:            nil,
			BaseEntity:                     &model.BaseEntity{ID: bundleID},
		}
		authConv := auth.NewConverter()
		conv := bundle.NewConverter(authConv, nil, nil, nil)
		// WHEN
		bndlModel, err := conv.FromEntity(entity)
		// THEN
		require.NoError(t, err)
		require.NotNil(t, expectedModel)
		assert.Equal(t, expectedModel, bndlModel)
	})
}

func TestConverter_ToGraphQL(t *testing.T) {
	// GIVEN
	id := bundleID
	name := "foo"
	desc := "bar"
	modelBundle := fixBundleModel(name, desc)
	gqlBundle := fixGQLBundle(id, name, desc)
	emptyModelBundle := &model.Bundle{BaseEntity: &model.BaseEntity{}}
	emptyGraphQLBundle := &graphql.Bundle{BaseEntity: &graphql.BaseEntity{}}

	testCases := []struct {
		Name            string
		Input           *model.Bundle
		Expected        *graphql.Bundle
		AuthConverterFn func() *automock.AuthConverter
		ExpectedErr     error
	}{
		{
			Name:     "All properties given",
			Input:    modelBundle,
			Expected: gqlBundle,
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", modelBundle.DefaultInstanceAuth).Return(gqlBundle.DefaultInstanceAuth, nil).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    emptyModelBundle,
			Expected: emptyGraphQLBundle,
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("ToGraphQL", emptyModelBundle.DefaultInstanceAuth).Return(nil, nil).Once()
				return conv
			},
		},
		{
			Name:        "Nil",
			Input:       nil,
			Expected:    nil,
			ExpectedErr: errors.New("the model Bundle is nil"),
			AuthConverterFn: func() *automock.AuthConverter {
				return &automock.AuthConverter{}
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			authConverter := testCase.AuthConverterFn()

			// WHEN
			converter := bundle.NewConverter(authConverter, nil, nil, nil)
			res, err := converter.ToGraphQL(testCase.Input)

			// then
			assert.EqualValues(t, testCase.Expected, res)
			if testCase.ExpectedErr != nil {
				require.Error(t, err)
				assert.Contains(t, err.Error(), testCase.ExpectedErr.Error())
			} else {
				require.Nil(t, err)
			}

			authConverter.AssertExpectations(t)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	// GIVEN
	name1 := "foo"
	name2 := "bar"
	desc := "1"
	input := []*model.Bundle{
		fixBundleModel(name1, desc),
		fixBundleModel(name2, desc),
		{BaseEntity: &model.BaseEntity{}},
		nil,
	}

	expected := []*graphql.Bundle{
		fixGQLBundle(bundleID, name1, desc),
		fixGQLBundle(bundleID, name2, desc),
		{BaseEntity: &graphql.BaseEntity{}},
	}

	authConverter := &automock.AuthConverter{}

	for i, api := range input {
		if api == nil {
			continue
		}
		authConverter.On("ToGraphQL", api.DefaultInstanceAuth).Return(expected[i].DefaultInstanceAuth, nil).Once()
	}

	// WHEN
	converter := bundle.NewConverter(authConverter, nil, nil, nil)
	res, err := converter.MultipleToGraphQL(input)

	// then
	assert.Equal(t, expected, res)
	assert.NoError(t, err)

	authConverter.AssertExpectations(t)
}

func TestConverter_CreateInputFromGraphQL(t *testing.T) {
	// GIVEN
	name := "foo"
	desc := "Lorem ipsum"
	gqlBundleCreateInput := fixGQLBundleCreateInput(name, desc)
	modelBundleCreateInput := fixModelBundleCreateInput(name, desc)
	emptyGQLBundleCreateInput := &graphql.BundleCreateInput{}
	emptyModelBundleCreateInput := &model.BundleCreateInput{}
	testCases := []struct {
		Name                string
		Input               graphql.BundleCreateInput
		Expected            model.BundleCreateInput
		APIConverterFn      func() *automock.APIConverter
		EventAPIConverterFn func() *automock.EventConverter
		DocumentConverterFn func() *automock.DocumentConverter
		AuthConverterFn     func() *automock.AuthConverter
	}{
		{
			Name:     "All properties given",
			Input:    gqlBundleCreateInput,
			Expected: modelBundleCreateInput,
			APIConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleInputFromGraphQL", gqlBundleCreateInput.APIDefinitions).Return(modelBundleCreateInput.APIDefinitions, modelBundleCreateInput.APISpecs, nil)
				return conv
			},
			EventAPIConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				conv.On("MultipleInputFromGraphQL", gqlBundleCreateInput.EventDefinitions).Return(modelBundleCreateInput.EventDefinitions, modelBundleCreateInput.EventSpecs, nil)
				return conv
			},
			DocumentConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("MultipleInputFromGraphQL", gqlBundleCreateInput.Documents).Return(modelBundleCreateInput.Documents, nil)
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlBundleCreateInput.DefaultInstanceAuth).Return(modelBundleCreateInput.DefaultInstanceAuth, nil).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    graphql.BundleCreateInput{},
			Expected: model.BundleCreateInput{},
			APIConverterFn: func() *automock.APIConverter {
				conv := &automock.APIConverter{}
				conv.On("MultipleInputFromGraphQL", emptyGQLBundleCreateInput.APIDefinitions).Return(emptyModelBundleCreateInput.APIDefinitions, emptyModelBundleCreateInput.APISpecs, nil)
				return conv
			},
			EventAPIConverterFn: func() *automock.EventConverter {
				conv := &automock.EventConverter{}
				conv.On("MultipleInputFromGraphQL", emptyGQLBundleCreateInput.EventDefinitions).Return(emptyModelBundleCreateInput.EventDefinitions, emptyModelBundleCreateInput.EventSpecs, nil)
				return conv
			},
			DocumentConverterFn: func() *automock.DocumentConverter {
				conv := &automock.DocumentConverter{}
				conv.On("MultipleInputFromGraphQL", emptyGQLBundleCreateInput.Documents).Return(emptyModelBundleCreateInput.Documents, nil)
				return conv
			},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", emptyGQLBundleCreateInput.DefaultInstanceAuth).Return(emptyModelBundleCreateInput.DefaultInstanceAuth, nil).Once()
				return conv
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			apiConverter := testCase.APIConverterFn()
			eventConverter := testCase.EventAPIConverterFn()
			documentConverter := testCase.DocumentConverterFn()
			authConverter := testCase.AuthConverterFn()

			// WHEN
			converter := bundle.NewConverter(authConverter, apiConverter, eventConverter, documentConverter)
			res, err := converter.CreateInputFromGraphQL(testCase.Input)

			// then
			assert.NoError(t, err)
			assert.Equal(t, testCase.Expected, res)
			mock.AssertExpectationsForObjects(t, apiConverter, eventConverter, documentConverter, authConverter)
		})
	}
}

func TestConverter_MultipleCreateInputFromGraphQL(t *testing.T) {
	// GIVEN
	gqlBndl1 := fixGQLBundleCreateInput("foo", "bar")
	gqlBndl2 := fixGQLBundleCreateInput("bar", "baz")
	input := []*graphql.BundleCreateInput{
		&gqlBndl1,
		&gqlBndl2,
	}

	modBndl1 := fixModelBundleCreateInput("foo", "bar")
	modBndl2 := fixModelBundleCreateInput("bar", "baz")
	expected := []*model.BundleCreateInput{
		&modBndl1, &modBndl2,
	}

	apiConv := &automock.APIConverter{}
	apiConv.On("MultipleInputFromGraphQL", gqlBndl1.APIDefinitions).Return(modBndl1.APIDefinitions, modBndl1.APISpecs, nil).Once()
	apiConv.On("MultipleInputFromGraphQL", gqlBndl2.APIDefinitions).Return(modBndl2.APIDefinitions, modBndl2.APISpecs, nil).Once()

	eventConv := &automock.EventConverter{}
	eventConv.On("MultipleInputFromGraphQL", gqlBndl1.EventDefinitions).Return(modBndl1.EventDefinitions, modBndl1.EventSpecs, nil).Once()
	eventConv.On("MultipleInputFromGraphQL", gqlBndl2.EventDefinitions).Return(modBndl2.EventDefinitions, modBndl2.EventSpecs, nil).Once()

	docConv := &automock.DocumentConverter{}
	docConv.On("MultipleInputFromGraphQL", gqlBndl1.Documents).Return(modBndl1.Documents, nil).Once()
	docConv.On("MultipleInputFromGraphQL", gqlBndl2.Documents).Return(modBndl2.Documents, nil).Once()

	authConv := &automock.AuthConverter{}
	authConv.On("InputFromGraphQL", gqlBndl1.DefaultInstanceAuth).Return(modBndl1.DefaultInstanceAuth, nil).Once()
	authConv.On("InputFromGraphQL", gqlBndl2.DefaultInstanceAuth).Return(modBndl2.DefaultInstanceAuth, nil).Once()

	converter := bundle.NewConverter(authConv, apiConv, eventConv, docConv)

	// WHEN
	res, err := converter.MultipleCreateInputFromGraphQL(input)

	// then
	assert.NoError(t, err)
	assert.Equal(t, expected, res)
	mock.AssertExpectationsForObjects(t, apiConv, eventConv, docConv, authConv)
}

func TestConverter_UpdateInputFromGraphQL(t *testing.T) {
	// GIVEN
	name := "foo"
	desc := "Lorem ipsum"
	gqlBundleCreateInput := fixGQLBundleUpdateInput(name, desc)
	modelBundleCreateInput := fixModelBundleUpdateInput(name, desc)
	emptyGQLBundleCreateInput := &graphql.BundleCreateInput{}
	emptyModelBundleCreateInput := &model.BundleCreateInput{}
	testCases := []struct {
		Name            string
		Input           *graphql.BundleUpdateInput
		Expected        *model.BundleUpdateInput
		AuthConverterFn func() *automock.AuthConverter
	}{
		{
			Name:     "All properties given",
			Input:    &gqlBundleCreateInput,
			Expected: &modelBundleCreateInput,
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", gqlBundleCreateInput.DefaultInstanceAuth).Return(modelBundleCreateInput.DefaultInstanceAuth, nil).Once()
				return conv
			},
		},
		{
			Name:     "Empty",
			Input:    &graphql.BundleUpdateInput{},
			Expected: &model.BundleUpdateInput{},
			AuthConverterFn: func() *automock.AuthConverter {
				conv := &automock.AuthConverter{}
				conv.On("InputFromGraphQL", emptyGQLBundleCreateInput.DefaultInstanceAuth).Return(emptyModelBundleCreateInput.DefaultInstanceAuth, nil).Once()
				return conv
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// GIVEN
			authConverter := testCase.AuthConverterFn()

			// WHEN
			converter := bundle.NewConverter(authConverter, nil, nil, nil)
			res, err := converter.UpdateInputFromGraphQL(*testCase.Input)

			// then
			assert.Equal(t, testCase.Expected, res)
			assert.NoError(t, err)
			authConverter.AssertExpectations(t)
		})
	}
}
