package fixtures

import (
	"fmt"

	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql/graphqlizer"
	json2 "github.com/kyma-incubator/compass/tests/pkg/json"
	"github.com/kyma-incubator/compass/tests/pkg/ptr"
	"github.com/kyma-incubator/compass/tests/pkg/testctx"
	gcli "github.com/machinebox/graphql"
	"github.com/stretchr/testify/require"
)

func FixBundleCreateInput(name string) graphql.BundleCreateInput {
	return graphql.BundleCreateInput{
		Name: name,
	}
}

func FixAddAPIToBundleRequest(bundleID, APIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addAPIDefinitionToBundle(bundleID: "%s", in: %s) {
				%s
			}
		}
		`, bundleID, APIInputGQL, testctx.Tc.GQLFieldsProvider.ForAPIDefinition()))
}

func FixBundleInstanceAuthRequestInput(ctx, inputParams *graphql.JSON) graphql.BundleInstanceAuthRequestInput {
	return graphql.BundleInstanceAuthRequestInput{
		Context:     ctx,
		InputParams: inputParams,
	}
}

func FixBundleInstanceAuthCreateInput(ctx, inputParams *graphql.JSON, auth *graphql.AuthInput, rtmID, rtmCtxID *string) graphql.BundleInstanceAuthCreateInput {
	return graphql.BundleInstanceAuthCreateInput{
		Context:          ctx,
		InputParams:      inputParams,
		Auth:             auth,
		RuntimeID:        rtmID,
		RuntimeContextID: rtmCtxID,
	}
}

func FixBundleInstanceAuthUpdateInput(ctx, inputParams *graphql.JSON, auth *graphql.AuthInput) graphql.BundleInstanceAuthUpdateInput {
	return graphql.BundleInstanceAuthUpdateInput{
		Context:     ctx,
		InputParams: inputParams,
		Auth:        auth,
	}
}

func FixBundleInstanceAuthSetInputSucceeded(auth *graphql.AuthInput) graphql.BundleInstanceAuthSetInput {
	return graphql.BundleInstanceAuthSetInput{
		Auth: auth,
	}
}

func FixAddBundleRequest(appID, bundleCreateInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: addBundle(applicationID: "%s", in: %s) {
				%s
			}}`, appID, bundleCreateInput, testctx.Tc.GQLFieldsProvider.ForBundle()))
}

func FixUpdateBundleRequest(bundleID, bndlUpdateInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: updateBundle(id: "%s", in: %s) {
				%s
			}
		}`, bundleID, bndlUpdateInput, testctx.Tc.GQLFieldsProvider.ForBundle()))
}

func FixDeleteBundleRequest(bundleID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteBundle(id: "%s") {
				%s
			}
		}`, bundleID, testctx.Tc.GQLFieldsProvider.ForBundle()))
}

func FixBundleRequest(applicationID string, bundleID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, testctx.Tc.GQLFieldsProvider.ForApplication(graphqlizer.FieldCtx{
			"Application.bundle": fmt.Sprintf(`bundle(id: "%s") {%s}`, bundleID, testctx.Tc.GQLFieldsProvider.ForBundle()),
		})))
}

func FixAddDocumentToBundleRequest(bundleID, documentInputInGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addDocumentToBundle(bundleID: "%s", in: %s) {
 				%s
			}				
		}`, bundleID, documentInputInGQL, testctx.Tc.GQLFieldsProvider.ForDocument()))
}

func FixAddEventAPIToBundleRequest(bndlID, eventAPIInputGQL string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
		result: addEventDefinitionToBundle(bundleID: "%s", in: %s) {
				%s
			}
		}
		`, bndlID, eventAPIInputGQL, testctx.Tc.GQLFieldsProvider.ForEventDefinition()))
}

func FixBundleWithOnlyOdataAPIs() *graphql.BundleCreateInput {
	return &graphql.BundleCreateInput{
		Name:        "test-bndl",
		Description: ptr.String("foo-descr"),
		APIDefinitions: []*graphql.APIDefinitionInput{
			{
				Name:        "reviews-v1",
				Description: ptr.String("api for adding reviews"),
				TargetURL:   "http://mywordpress.com/reviews",
				Version:     FixActiveVersion(),
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOdata,
					Format: graphql.SpecFormatJSON,
					Data:   ptr.CLOB(`{"openapi":"3.0.1"}`),
				},
			},
			{
				Name:        "xml",
				Description: ptr.String("xml api"),
				Version:     FixDecommissionedVersion(),
				TargetURL:   "http://mywordpress.com/xml",
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOdata,
					Format: graphql.SpecFormatXML,
					Data:   ptr.CLOB("odata"),
				},
			},
		},
	}
}

func FixBundleCreateInputWithRelatedObjects(t require.TestingT, name string) graphql.BundleCreateInput {
	desc := "Foo bar"
	return graphql.BundleCreateInput{
		Name:        name,
		Description: &desc,
		APIDefinitions: []*graphql.APIDefinitionInput{
			{
				Name:        "comments-v1",
				Description: ptr.String("api for adding comments"),
				TargetURL:   "http://mywordpress.com/comments",
				Group:       ptr.String("comments"),
				Version:     FixDeprecatedVersion(),
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOpenAPI,
					Format: graphql.SpecFormatYaml,
					Data:   ptr.CLOB(`{"openapi":"3.0.2"}`),
				},
			},
			{
				Name:      "reviews-v1",
				TargetURL: "http://mywordpress.com/reviews",
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOdata,
					Format: graphql.SpecFormatJSON,
					FetchRequest: &graphql.FetchRequestInput{
						URL:    "http://mywordpress.com/apis",
						Mode:   ptr.FetchMode(graphql.FetchModeBundle),
						Filter: ptr.String("odata.json"),
						Auth:   FixBasicAuth(t),
					},
				},
			},
			{
				Name:      "xml",
				TargetURL: "http://mywordpress.com/xml",
				Spec: &graphql.APISpecInput{
					Type:   graphql.APISpecTypeOdata,
					Format: graphql.SpecFormatXML,
					Data:   ptr.CLOB("odata"),
				},
			},
		},
		EventDefinitions: []*graphql.EventDefinitionInput{
			{
				Name:        "comments-v1",
				Description: ptr.String("comments events"),
				Version:     FixDeprecatedVersion(),
				Group:       ptr.String("comments"),
				Spec: &graphql.EventSpecInput{
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: graphql.SpecFormatYaml,
					Data:   ptr.CLOB(`{"asyncapi":"1.2.0"}`),
				},
			},
			{
				Name:        "reviews-v1",
				Description: ptr.String("review events"),
				Spec: &graphql.EventSpecInput{
					Type:   graphql.EventSpecTypeAsyncAPI,
					Format: graphql.SpecFormatYaml,
					FetchRequest: &graphql.FetchRequestInput{
						URL:    "http://mywordpress.com/events",
						Mode:   ptr.FetchMode(graphql.FetchModeBundle),
						Filter: ptr.String("async.json"),
						Auth:   FixOauthAuth(t),
					},
				},
			},
		},
		Documents: []*graphql.DocumentInput{
			{
				Title:       "Readme",
				Description: "Detailed description of project",
				Format:      graphql.DocumentFormatMarkdown,
				DisplayName: "display-name",
				FetchRequest: &graphql.FetchRequestInput{
					URL:    "kyma-project.io",
					Mode:   ptr.FetchMode(graphql.FetchModeBundle),
					Filter: ptr.String("/docs/README.md"),
					Auth:   FixBasicAuth(t),
				},
			},
			{
				Title:       "Troubleshooting",
				Description: "Troubleshooting description",
				Format:      graphql.DocumentFormatMarkdown,
				DisplayName: "display-name",
				Data:        ptr.CLOB("No problems, everything works on my machine"),
			},
		},
	}
}

func FixBundleCreateInputWithDefaultAuth(name string, authInput *graphql.AuthInput) graphql.BundleCreateInput {
	return graphql.BundleCreateInput{
		Name:                name,
		DefaultInstanceAuth: authInput,
	}
}

func FixBundleUpdateInput(name string) graphql.BundleUpdateInput {
	return graphql.BundleUpdateInput{
		Name: name,
	}
}

func FixDocumentInBundleRequest(appID, bndlID, docID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							document(id: "%s"){
						%s
						}					
					}
				}
			}`, appID, bndlID, docID, testctx.Tc.GQLFieldsProvider.ForDocument()))
}

func FixAPIDefinitionsInBundleRequest(appID, bndlID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							apiDefinitions{
						%s
						}					
					}
				}
			}`, appID, bndlID, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForAPIDefinition())))

}

func FixEventDefinitionsInBundleRequest(appID, bndlID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							eventDefinitions{
						%s
						}					
					}
				}
			}`, appID, bndlID, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForEventDefinition())))
}

func FixDocumentsInBundleRequest(appID, bndlID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
						bundle(id: "%s"){
							documents{
						%s
						}					
					}
				}
			}`, appID, bndlID, testctx.Tc.GQLFieldsProvider.Page(testctx.Tc.GQLFieldsProvider.ForDocument())))
}

func FixSetBundleInstanceAuthRequest(authID, apiAuthInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: setBundleInstanceAuth(authID: "%s", in: %s) {
				%s
			}
		}`, authID, apiAuthInput, testctx.Tc.GQLFieldsProvider.ForBundleInstanceAuth()))
}

func FixDeleteBundleInstanceAuthRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: deleteBundleInstanceAuth(authID: "%s") {
				%s
			}
		}`, authID, testctx.Tc.GQLFieldsProvider.ForBundleInstanceAuth()))
}

func FixRequestBundleInstanceAuthCreationRequest(bundleID, bndlInstanceAuthRequestInput string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestBundleInstanceAuthCreation(bundleID: "%s", in: %s) {
				%s
			}
		}`, bundleID, bndlInstanceAuthRequestInput, testctx.Tc.GQLFieldsProvider.ForBundleInstanceAuth()))
}

func FixRequestBundleInstanceAuthDeletionRequest(authID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`mutation {
			result: requestBundleInstanceAuthDeletion(authID: "%s") {
				%s
			}
		}`, authID, testctx.Tc.GQLFieldsProvider.ForBundleInstanceAuth()))
}

func FixBundleByInstanceAuthIDRequest(packageInstanceAuthID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: packageByInstanceAuth(authID: "%s") {
				%s
				}
			}`, packageInstanceAuthID, testctx.Tc.GQLFieldsProvider.ForBundle()))
}

func FixGetBundleWithInstanceAuthRequest(applicationID string, bundleID string, instanceAuthID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID,
			testctx.Tc.GQLFieldsProvider.ForApplication(
				graphqlizer.FieldCtx{"Application.bundle": fmt.Sprintf(`bundle(id: "%s") {%s}`,
					bundleID,
					testctx.Tc.GQLFieldsProvider.ForBundle(graphqlizer.FieldCtx{
						"Bundle.instanceAuth": fmt.Sprintf(`instanceAuth(id: "%s") {%s}`,
							instanceAuthID,
							testctx.Tc.GQLFieldsProvider.ForBundleInstanceAuth()),
					})),
				})))
}

func FixGetBundlesRequest(applicationID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: application(id: "%s") {
				%s
				}
			}`, applicationID, testctx.Tc.GQLFieldsProvider.ForApplication()))
}

func FixBundleInstanceAuthRequest(packageInstanceAuthID string) *gcli.Request {
	return gcli.NewRequest(
		fmt.Sprintf(`query {
			result: bundleInstanceAuth(id: "%s") {
					%s
				}
			}`, packageInstanceAuthID, testctx.Tc.GQLFieldsProvider.ForBundleInstanceAuth()))
}

func FixBundleInstanceAuthContextAndInputParams(t require.TestingT) (*graphql.JSON, *graphql.JSON) {
	authCtxPayload := map[string]interface{}{
		"ContextData": "ContextValue",
	}
	var authCtxData interface{} = authCtxPayload

	inputParamsPayload := map[string]interface{}{
		"InKey": "InValue",
	}
	var inputParamsData interface{} = inputParamsPayload

	authCtx := json2.MarshalJSON(t, authCtxData)
	inputParams := json2.MarshalJSON(t, inputParamsData)

	return authCtx, inputParams
}
