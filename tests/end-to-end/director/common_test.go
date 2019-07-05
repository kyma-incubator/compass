package director

import (
	"context"
	"os"
	"reflect"
	"testing"

	gcli "github.com/machinebox/graphql"
)

var tc = testContext{graphqlizer: graphqlizer{}, gqlFieldsProvider: gqlFieldsProvider{}, cli: gcli.NewClient(getDirectorURL())}

func (tc *testContext) RunQuery(ctx context.Context, req *gcli.Request, resp interface{}) error {
	if req.Header["Tenant"] == nil {
		req.Header["Tenant"] = []string{"test-end-to-end"}
	}
	m := resultMapperFor(&resp)
	return tc.cli.Run(ctx, req, &m)
}

// testContext contains dependencies that help executing tests
type testContext struct {
	graphqlizer       graphqlizer
	gqlFieldsProvider gqlFieldsProvider
	cli               *gcli.Client
}

func getDirectorURL() string {
	url := os.Getenv("DIRECTOR_GRAPHQL_API")
	if url == "" {
		url = "http://127.0.0.1:3000/graphql"
	}
	return url
}

// resultMapperFor returns generic object that can be passed to Run method for storing response.
// In GraphQL, set `result` alias for your query
func resultMapperFor(target interface{}) genericGQLResponse {
	if reflect.ValueOf(target).Kind() != reflect.Ptr {
		panic("target has to be a pointer")
	}
	return genericGQLResponse{
		Result: target,
	}
}

type genericGQLResponse struct {
	Result interface{} `json:"result"`
}

func saveQueryInExamples(t *testing.T, query string, exampleName string) {
	// TODO uncomment this in https://github.com/kyma-incubator/compass/issues/77
	//t.Helper()
	//sanitizedName := strings.Replace(exampleName, " ", "-", -1)
	//// replace uuids with constant value
	//r, err := regexp.Compile("[a-f0-9]{8}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{4}-[a-f0-9]{12}")
	//require.NoError(t, err)
	//query = r.ReplaceAllString(query, "aaaaaaaa-aaaa-aaaa-aaaa-aaaaaaaaaaaa")
	//content := fmt.Sprintf("# This example is generated from test compass_e2e_test.go\n%s", query)
	//err = ioutil.WriteFile(fmt.Sprintf("%s/src/github.com/kyma-incubator/compass/components/director/examples/%s.graphql", os.Getenv("GOPATH"), sanitizedName), []byte(content), 0660)
	//require.NoError(t, err)
}
