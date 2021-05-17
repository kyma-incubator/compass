package json

import (
	"encoding/json"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/stretchr/testify/require"
	"strconv"
)

func MarshalJSON(t require.TestingT, data interface{}) *graphql.JSON {
	out, err := json.Marshal(data)
	require.NoError(t, err)
	output := strconv.Quote(string(out))
	j := graphql.JSON(output)
	return &j
}

func UnmarshalJSON(t require.TestingT, j *graphql.JSON) interface{} {
	require.NotNil(t, j)
	var output interface{}
	err := json.Unmarshal([]byte(*j), &output)
	require.NoError(t, err)

	return output
}

func MarshalJSONSchema(t require.TestingT, schema interface{}) *graphql.JSONSchema {
	out, err := json.Marshal(schema)
	require.NoError(t, err)
	output := strconv.Quote(string(out))
	jsonSchema := graphql.JSONSchema(output)
	return &jsonSchema
}

func UnmarshalJSONSchema(t require.TestingT, schema *graphql.JSONSchema) interface{} {
	require.NotNil(t, schema)
	var output interface{}
	err := json.Unmarshal([]byte(*schema), &output)
	require.NoError(t, err)

	return output
}
