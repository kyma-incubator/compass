package service

import (
	"testing"

	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/assert"
)

func TestGqlRequestBuilder_UnregisterApplicationRequest(t *testing.T) {
	id := "test"
	builder := NewGqlRequestBuilder()
	expectedRq := gcli.NewRequest(`mutation {
		result: unregisterApplication(id: "test") {
			id
		}	
	}`)

	rq := builder.UnregisterApplicationRequest(id)

	assert.Equal(t, expectedRq, rq)
}

func TestGqlRequestBuilder_RegisterApplicationRequest(t *testing.T) {
	builder := NewGqlRequestBuilder()
	expectedRq := gcli.NewRequest("mutation {\n\t\t\tresult: registerApplication(in: {\n\t\tname: \"test\",\n\t\tdescription: \"Lorem ipsum\",\n\t\tlabels: {\n\t\t\ttest: [\"val\",\"val2\" ],\n\t},\n\t\twebhooks: [ {\n\t\ttype: ,\n\t\turl: \"webhook1.foo.bar\",\n\n\t}, {\n\t\ttype: ,\n\t\turl: \"webhook2.foo.bar\",\n\n\t} ],\n\t\tapiDefinitions: [ {\n\t\tname: \"api1\",\n\t\ttargetURL: \"foo.bar\",\n\t}, {\n\t\tname: \"api2\",\n\t\ttargetURL: \"foo.bar2\",\n\t}],\n\t}) {\n\t\t\t\tid\n\t\t\t}\t\n\t\t}")

	input := fixGQLApplicationRegisterInput("test", "Lorem ipsum")

	rq, err := builder.RegisterApplicationRequest(input)

	assert.NoError(t, err)
	assert.Equal(t, expectedRq, rq)
}
