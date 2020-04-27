package api

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
)

const OpenAPISpec = "https://raw.githubusercontent.com/kyma-incubator/github-slack-connectors/master/github-connector/internal/registration/configs/githubopenAPI.json"

func Test_FetchRequestAddAPIToPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkg := createPackage(t, ctx, application.ID, pkgName)
	defer deletePackage(t, ctx, pkg.ID)

	apiInput := graphql.APIDefinitionInput{
		Name:      "test",
		TargetURL: "https://target.url",
		Spec: &graphql.APISpecInput{
			Format: graphql.SpecFormatJSON,
			Type:   graphql.APISpecTypeOpenAPI,
			FetchRequest: &graphql.FetchRequestInput{
				URL: OpenAPISpec,
			},
		},
	}
	api := addAPIToPackageWithInput(t, ctx, pkg.ID, apiInput)
	assert.NotNil(t, api.Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, api.Spec.FetchRequest.Status.Condition)
}

func TestFetchRequestAddPackage(t *testing.T) {
	ctx := context.Background()

	appName := "app-test-package"
	application := registerApplication(t, ctx, appName)
	defer unregisterApplication(t, application.ID)

	pkgName := "test-package"
	pkgInput := graphql.PackageCreateInput{
		Name: pkgName,
		APIDefinitions: []*graphql.APIDefinitionInput{{
			Name:      "test",
			TargetURL: "https://target.url",
			Spec: &graphql.APISpecInput{
				Format: graphql.SpecFormatJSON,
				Type:   graphql.APISpecTypeOpenAPI,
				FetchRequest: &graphql.FetchRequestInput{
					URL: OpenAPISpec,
				},
			},
		},
		},
	}

	pkg := createPackageWithInput(t, ctx, application.ID, pkgInput)
	defer deletePackage(t, ctx, pkg.ID)

	assert.NotNil(t, pkg.APIDefinitions.Data[0].Spec.Data)
	assert.Equal(t, graphql.FetchRequestStatusConditionSucceeded, pkg.APIDefinitions.Data[0].Spec.FetchRequest.Status.Condition)
}
