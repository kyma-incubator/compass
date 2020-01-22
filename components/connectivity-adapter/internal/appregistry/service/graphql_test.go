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

func TestGqlRequestBuilder_GetApplicationRequest(t *testing.T) {
	id := "test"
	builder := NewGqlRequestBuilder()
	expectedRq := gcli.NewRequest("query {\n\t\t\tresult: application(id: \"test\") {\n\t\t\t\t\tid\n\t\tname\n\t\tdescription\n\t\tlabels\n\t\teventingConfiguration { defaultURL }\n\t\tstatus {condition timestamp}\n\t\twebhooks {id\n\t\tapplicationID\n\t\ttype\n\t\turl\n\t\tauth {\n\t\t  credential {\n\t\t\t\t... on BasicCredentialData {\n\t\t\t\t\tusername\n\t\t\t\t\tpassword\n\t\t\t\t}\n\t\t\t\t...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t}\n\t\t\t}\n\t\t\tadditionalHeaders\n\t\t\tadditionalQueryParams\n\t\t\trequestAuth { \n\t\t\t  csrf {\n\t\t\t\ttokenEndpointURL\n\t\t\t\tcredential {\n\t\t\t\t  ... on BasicCredentialData {\n\t\t\t\t  \tusername\n\t\t\t\t\tpassword\n\t\t\t\t  }\n\t\t\t\t  ...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t  }\n\t\t\t    }\n\t\t\t\tadditionalHeaders\n\t\t\t\tadditionalQueryParams\n\t\t\t}\n\t\t\t}\n\t\t\n\t\t}}\n\t\thealthCheckURL\n\t\tproviderName\n\t\tapiDefinitions {data {\n\t\t\t\tid\n\t\tname\n\t\tdescription\n\t\tspec {data\n\t\tformat\n\t\ttype\n\t\tfetchRequest {url\n\t\tauth {credential {\n\t\t\t\t... on BasicCredentialData {\n\t\t\t\t\tusername\n\t\t\t\t\tpassword\n\t\t\t\t}\n\t\t\t\t...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t}\n\t\t\t}\n\t\t\tadditionalHeaders\n\t\t\tadditionalQueryParams\n\t\t\trequestAuth { \n\t\t\t  csrf {\n\t\t\t\ttokenEndpointURL\n\t\t\t\tcredential {\n\t\t\t\t  ... on BasicCredentialData {\n\t\t\t\t  \tusername\n\t\t\t\t\tpassword\n\t\t\t\t  }\n\t\t\t\t  ...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t  }\n\t\t\t    }\n\t\t\t\tadditionalHeaders\n\t\t\t\tadditionalQueryParams\n\t\t\t}\n\t\t\t}\n\t\t}\n\t\tmode\n\t\tfilter\n\t\tstatus {condition timestamp}}}\n\t\ttargetURL\n\t\tgroup\n\t\tauths {runtimeID\n\t\tauth {credential {\n\t\t\t\t... on BasicCredentialData {\n\t\t\t\t\tusername\n\t\t\t\t\tpassword\n\t\t\t\t}\n\t\t\t\t...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t}\n\t\t\t}\n\t\t\tadditionalHeaders\n\t\t\tadditionalQueryParams\n\t\t\trequestAuth { \n\t\t\t  csrf {\n\t\t\t\ttokenEndpointURL\n\t\t\t\tcredential {\n\t\t\t\t  ... on BasicCredentialData {\n\t\t\t\t  \tusername\n\t\t\t\t\tpassword\n\t\t\t\t  }\n\t\t\t\t  ...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t  }\n\t\t\t    }\n\t\t\t\tadditionalHeaders\n\t\t\t\tadditionalQueryParams\n\t\t\t}\n\t\t\t}\n\t\t}}\n\t\tdefaultAuth {credential {\n\t\t\t\t... on BasicCredentialData {\n\t\t\t\t\tusername\n\t\t\t\t\tpassword\n\t\t\t\t}\n\t\t\t\t...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t}\n\t\t\t}\n\t\t\tadditionalHeaders\n\t\t\tadditionalQueryParams\n\t\t\trequestAuth { \n\t\t\t  csrf {\n\t\t\t\ttokenEndpointURL\n\t\t\t\tcredential {\n\t\t\t\t  ... on BasicCredentialData {\n\t\t\t\t  \tusername\n\t\t\t\t\tpassword\n\t\t\t\t  }\n\t\t\t\t  ...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t  }\n\t\t\t    }\n\t\t\t\tadditionalHeaders\n\t\t\t\tadditionalQueryParams\n\t\t\t}\n\t\t\t}\n\t\t}\n\t\tversion {value\n\t\tdeprecated\n\t\tdeprecatedSince\n\t\tforRemoval}\n\t}\n\tpageInfo {startCursor\n\t\tendCursor\n\t\thasNextPage}\n\ttotalCount\n\t}\n\t\teventDefinitions {data {\n\t\t\n\t\t\tid\n\t\t\tapplicationID\n\t\t\tname\n\t\t\tdescription\n\t\t\tgroup \n\t\t\tspec {data\n\t\ttype\n\t\tformat\n\t\tfetchRequest {url\n\t\tauth {credential {\n\t\t\t\t... on BasicCredentialData {\n\t\t\t\t\tusername\n\t\t\t\t\tpassword\n\t\t\t\t}\n\t\t\t\t...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t}\n\t\t\t}\n\t\t\tadditionalHeaders\n\t\t\tadditionalQueryParams\n\t\t\trequestAuth { \n\t\t\t  csrf {\n\t\t\t\ttokenEndpointURL\n\t\t\t\tcredential {\n\t\t\t\t  ... on BasicCredentialData {\n\t\t\t\t  \tusername\n\t\t\t\t\tpassword\n\t\t\t\t  }\n\t\t\t\t  ...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t  }\n\t\t\t    }\n\t\t\t\tadditionalHeaders\n\t\t\t\tadditionalQueryParams\n\t\t\t}\n\t\t\t}\n\t\t}\n\t\tmode\n\t\tfilter\n\t\tstatus {condition timestamp}}}\n\t\t\tversion {value\n\t\tdeprecated\n\t\tdeprecatedSince\n\t\tforRemoval}\n\t\t\n\t}\n\tpageInfo {startCursor\n\t\tendCursor\n\t\thasNextPage}\n\ttotalCount\n\t}\n\t\tdocuments {data {\n\t\t\n\t\tid\n\t\tapplicationID\n\t\ttitle\n\t\tdisplayName\n\t\tdescription\n\t\tformat\n\t\tkind\n\t\tdata\n\t\tfetchRequest {url\n\t\tauth {credential {\n\t\t\t\t... on BasicCredentialData {\n\t\t\t\t\tusername\n\t\t\t\t\tpassword\n\t\t\t\t}\n\t\t\t\t...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t}\n\t\t\t}\n\t\t\tadditionalHeaders\n\t\t\tadditionalQueryParams\n\t\t\trequestAuth { \n\t\t\t  csrf {\n\t\t\t\ttokenEndpointURL\n\t\t\t\tcredential {\n\t\t\t\t  ... on BasicCredentialData {\n\t\t\t\t  \tusername\n\t\t\t\t\tpassword\n\t\t\t\t  }\n\t\t\t\t  ...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t  }\n\t\t\t    }\n\t\t\t\tadditionalHeaders\n\t\t\t\tadditionalQueryParams\n\t\t\t}\n\t\t\t}\n\t\t}\n\t\tmode\n\t\tfilter\n\t\tstatus {condition timestamp}}\n\t}\n\tpageInfo {startCursor\n\t\tendCursor\n\t\thasNextPage}\n\ttotalCount\n\t}\n\t\tauths {\n\t\tid\n\t\tauth {credential {\n\t\t\t\t... on BasicCredentialData {\n\t\t\t\t\tusername\n\t\t\t\t\tpassword\n\t\t\t\t}\n\t\t\t\t...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t}\n\t\t\t}\n\t\t\tadditionalHeaders\n\t\t\tadditionalQueryParams\n\t\t\trequestAuth { \n\t\t\t  csrf {\n\t\t\t\ttokenEndpointURL\n\t\t\t\tcredential {\n\t\t\t\t  ... on BasicCredentialData {\n\t\t\t\t  \tusername\n\t\t\t\t\tpassword\n\t\t\t\t  }\n\t\t\t\t  ...  on OAuthCredentialData {\n\t\t\t\t\tclientId\n\t\t\t\t\tclientSecret\n\t\t\t\t\turl\n\t\t\t\t\t\n\t\t\t\t  }\n\t\t\t    }\n\t\t\t\tadditionalHeaders\n\t\t\t\tadditionalQueryParams\n\t\t\t}\n\t\t\t}\n\t\t}}\n\t\n\t\t\t}\n\t\t}")

	rq := builder.GetApplicationRequest(id)

	assert.Equal(t, expectedRq, rq)
}
