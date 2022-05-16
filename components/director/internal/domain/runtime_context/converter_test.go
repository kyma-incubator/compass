package runtimectx_test

import (
	"testing"

	runtimectx "github.com/kyma-incubator/compass/components/director/internal/domain/runtime_context"

	"github.com/kyma-incubator/compass/components/director/internal/model"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/assert"
)

func TestConverter_ToGraphQL(t *testing.T) {
	id := "test_id"
	key := "key"
	val := "val"
	// GIVEN
	testCases := []struct {
		Name     string
		Input    *model.RuntimeContext
		Expected *graphql.RuntimeContext
	}{
		{
			Name: "All properties given",
			Input: &model.RuntimeContext{
				ID:        id,
				RuntimeID: runtimeID,
				Key:       key,
				Value:     val,
			},
			Expected: &graphql.RuntimeContext{
				ID:    id,
				Key:   key,
				Value: val,
			},
		},
		{
			Name:     "Empty",
			Input:    &model.RuntimeContext{},
			Expected: &graphql.RuntimeContext{},
		},
		{
			Name:     "Nil",
			Input:    nil,
			Expected: nil,
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			converter := runtimectx.NewConverter()
			res := converter.ToGraphQL(testCase.Input)

			// THEN
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_MultipleToGraphQL(t *testing.T) {
	id := "test_id"
	runtimeID := "test_runtime_id"
	key := "key"
	val := "val"

	// GIVEN
	input := []*model.RuntimeContext{
		{
			ID:        id,
			RuntimeID: runtimeID,
			Key:       key,
			Value:     val,
		},
		{
			ID:        id + "2",
			RuntimeID: runtimeID + "2",
			Key:       key + "2",
			Value:     val + "2",
		},
		nil,
	}
	expected := []*graphql.RuntimeContext{
		{
			ID:    id,
			Key:   key,
			Value: val,
		},
		{
			ID:    id + "2",
			Key:   key + "2",
			Value: val + "2",
		},
	}

	// WHEN
	converter := runtimectx.NewConverter()
	res := converter.MultipleToGraphQL(input)

	// THEN
	assert.Equal(t, expected, res)
}

func TestConverter_InputFromGraphQL(t *testing.T) {
	key := "key"
	val := "val"

	// GIVEN
	testCases := []struct {
		Name     string
		Input    graphql.RuntimeContextInput
		Expected model.RuntimeContextInput
	}{
		{
			Name: "All properties given",
			Input: graphql.RuntimeContextInput{
				Key:   key,
				Value: val,
			},
			Expected: model.RuntimeContextInput{
				Key:   key,
				Value: val,
			},
		},
		{
			Name:     "Empty",
			Input:    graphql.RuntimeContextInput{},
			Expected: model.RuntimeContextInput{},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			converter := runtimectx.NewConverter()
			res := converter.InputFromGraphQL(testCase.Input)

			// THEN
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_InputFromGraphQLWithRuntimeID(t *testing.T) {
	key := "key"
	val := "val"
	runtimeID := "runtime_id"

	// GIVEN
	testCases := []struct {
		Name     string
		Input    graphql.RuntimeContextInput
		Expected model.RuntimeContextInput
	}{
		{
			Name: "All properties given",
			Input: graphql.RuntimeContextInput{
				Key:   key,
				Value: val,
			},
			Expected: model.RuntimeContextInput{
				Key:       key,
				Value:     val,
				RuntimeID: runtimeID,
			},
		},
		{
			Name:  "Empty",
			Input: graphql.RuntimeContextInput{},
			Expected: model.RuntimeContextInput{
				RuntimeID: runtimeID,
			},
		},
	}

	for _, testCase := range testCases {
		t.Run(testCase.Name, func(t *testing.T) {
			// WHEN
			converter := runtimectx.NewConverter()
			res := converter.InputFromGraphQLWithRuntimeID(testCase.Input, runtimeID)

			// THEN
			assert.Equal(t, testCase.Expected, res)
		})
	}
}

func TestConverter_EntityFromRuntimeModel(t *testing.T) {
	// GIVEN
	modelRuntimeCtx := model.RuntimeContext{
		ID:        "id",
		RuntimeID: "runtime_id",
		Key:       "key",
		Value:     "value",
	}

	conv := runtimectx.NewConverter()
	// WHEN
	entityRuntimeCtx := conv.ToEntity(&modelRuntimeCtx)

	// THEN
	assert.Equal(t, modelRuntimeCtx.ID, entityRuntimeCtx.ID)
	assert.Equal(t, modelRuntimeCtx.RuntimeID, entityRuntimeCtx.RuntimeID)
	assert.Equal(t, modelRuntimeCtx.Key, entityRuntimeCtx.Key)
	assert.Equal(t, modelRuntimeCtx.Value, entityRuntimeCtx.Value)
}

func TestConverter_RuntimeContextToModel(t *testing.T) {
	// GIVEN
	entityRuntimeCtx := &runtimectx.RuntimeContext{
		ID:        "id",
		RuntimeID: "runtime_id",
		Key:       "key",
		Value:     "value",
	}

	conv := runtimectx.NewConverter()
	// WHEN
	modelRuntimeCtx := conv.FromEntity(entityRuntimeCtx)

	// THEN
	assert.Equal(t, entityRuntimeCtx.ID, modelRuntimeCtx.ID)
	assert.Equal(t, entityRuntimeCtx.RuntimeID, modelRuntimeCtx.RuntimeID)
	assert.Equal(t, entityRuntimeCtx.Key, modelRuntimeCtx.Key)
	assert.Equal(t, entityRuntimeCtx.Value, modelRuntimeCtx.Value)
}
