package service_test

import (
	"testing"

	"github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/model"
	svcautomock "github.com/kyma-incubator/compass/components/connectivity-adapter/internal/appregistry/service/automock"
	"github.com/kyma-incubator/compass/components/connectivity-adapter/pkg/gqlcli/automock"
	"github.com/kyma-incubator/compass/components/director/pkg/graphql"
	gcli "github.com/machinebox/graphql"
	log "github.com/sirupsen/logrus"
	"github.com/sirupsen/logrus/hooks/test"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func fixGQLApplicationRegisterInput(name, description string) graphql.ApplicationRegisterInput {
	labels := graphql.Labels{
		"test": []string{"val", "val2"},
	}
	return graphql.ApplicationRegisterInput{
		Name:        name,
		Description: &description,
		Labels:      &labels,
		Webhooks: []*graphql.WebhookInput{
			{URL: "webhook1.foo.bar"},
			{URL: "webhook2.foo.bar"},
		},
		APIDefinitions: []*graphql.APIDefinitionInput{
			{Name: "api1", TargetURL: "foo.bar"},
			{Name: "api2", TargetURL: "foo.bar2"},
		},
	}
}

func SuccessfulValidatorFn(input model.ServiceDetails) func() *svcautomock.Validator {
	return func() *svcautomock.Validator {
		validator := &svcautomock.Validator{}
		validator.On("Validate", input).Return(nil).Once()
		return validator
	}
}

func SuccessfulDetailsToGQLInputConverterFn(input model.ServiceDetails, output graphql.ApplicationRegisterInput) func() *svcautomock.Converter {
	return func() *svcautomock.Converter {
		converter := &svcautomock.Converter{}
		converter.On("DetailsToGraphQLInput", input).Return(output, nil).Once()
		return converter
	}
}

func EmptyConverterFn() func() *svcautomock.Converter {
	return func() *svcautomock.Converter {
		return &svcautomock.Converter{}
	}
}

func EmptyValidatorFn() func() *svcautomock.Validator {
	return func() *svcautomock.Validator {
		return &svcautomock.Validator{}
	}
}

func EmptyGraphQLClientFn() func() *automock.GraphQLClient {
	return func() *automock.GraphQLClient {
		return &automock.GraphQLClient{}
	}
}

func EmptyGraphQLRequestBuilderFn() func() *svcautomock.GraphQLRequestBuilder {
	return func() *svcautomock.GraphQLRequestBuilder {
		return &svcautomock.GraphQLRequestBuilder{}
	}
}

func SuccessfulRegisterAppGraphQLRequestBuilderFn(input graphql.ApplicationRegisterInput, output *gcli.Request) func() *svcautomock.GraphQLRequestBuilder {
	return func() *svcautomock.GraphQLRequestBuilder {
		gqlRequestBuilder := &svcautomock.GraphQLRequestBuilder{}
		gqlRequestBuilder.On("RegisterApplicationRequest", input).Return(output, nil).Once()
		return gqlRequestBuilder
	}
}

func SingleErrorLoggerAssertions(errMessage string) func(t *testing.T, hook *test.Hook) {
	return func(t *testing.T, hook *test.Hook) {
		assert.Equal(t, 1, len(hook.AllEntries()))
		entry := hook.LastEntry()
		require.NotNil(t, entry)
		assert.Equal(t, log.ErrorLevel, entry.Level)
		assert.Equal(t, errMessage, entry.Message)
	}
}

func fixAPIOpenAPIYAML() model.API {
	spec := `openapi: 3.0.0
info:
  title: Sample API
  description: Optional multiline or single-line description in [CommonMark](http://commonmark.org/help/) or HTML.
  version: 0.1.9
servers:
  - url: http://api.example.com/v1
    description: Optional server description, e.g. Main (production) server
  - url: http://staging-api.example.com
    description: Optional server description, e.g. Internal staging server for testing
paths:
  /users:
    get:
      summary: Returns a list of users.
      description: Optional extended description in CommonMark or HTML.
      responses:
        '200':    # status code
          description: A JSON array of user names
          content:
            application/json:
              schema: 
                type: array
                items: 
                  type: string`

	return model.API{
		Spec: []byte(spec),
	}
}

func fixAPIOpenAPIJSON() model.API {
	spec := `{
  "swagger": "2.0",
  "info": {
    "version": "1.0.0",
    "title": "Swagger Petstore",
    "description": "A sample API that uses a petstore as an <b>example</b> to demonstrate features in the swagger-2.0 specification",
    "termsOfService": "http://swagger.io/terms/",
    "contact": {
      "name": "Swagger API Team"
    },
    "license": {
      "name": "MIT"
    }
  },
  "host": "petstore.swagger.io",
  "basePath": "/api",
  "schemes": [
    "http"
  ],
  "consumes": [
    "application/json"
  ],
  "produces": [
    "application/json"
  ],
  "paths": {
    "/pets": {
      "get": {
        "description": "Returns all pets from the system that the user has access to",
        "produces": [
          "application/json"
        ],
        "responses": {
          "200": {
            "description": "A list of pets.",
            "schema": {
              "type": "array",
              "items": {
                "$ref": "#/definitions/Pet"
              }
            }
          }
        }
      }
    }
  },
  "definitions": {
    "Pet": {
      "type": "object",
      "required": [
        "id",
        "name"
      ],
      "properties": {
        "id": {
          "type": "integer",
          "format": "int64"
        },
        "name": {
          "type": "string"
        },
        "tag": {
          "type": "string"
        }
      }
    }
  }
}`

	return model.API{
		Spec: []byte(spec),
	}
}

func fixAPIODataXML() model.API {
	spec := `<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx"

           xmlns="http://docs.oasis-open.org/odata/ns/edm" Version="4.0">
  <edmx:Reference Uri="https://oasis-tcs.github.io/odata-vocabularies/vocabularies/Org.OData.Core.V1.xml">

    <edmx:Include Namespace="Org.OData.Core.V1" Alias="Core">

      <Annotation Term="Core.DefaultNamespace" />

    </edmx:Include>

  </edmx:Reference>

  <edmx:Reference Uri="https://oasis-tcs.github.io/odata-vocabularies/vocabularies/Org.OData.Measures.V1.xml">

    <edmx:Include Alias="Measures" Namespace="Org.OData.Measures.V1" />

  </edmx:Reference>

  <edmx:DataServices>
    <Schema Namespace="ODataDemo">
      <EntityType Name="Product" HasStream="true">
        <Key>
          <PropertyRef Name="ID" />
        </Key>
        <Property Name="ID" Type="Edm.Int32" Nullable="false" />
        <Property Name="Description" Type="Edm.String" >
          <Annotation Term="Core.IsLanguageDependent" />

        </Property>

        <Property Name="ReleaseDate" Type="Edm.Date" />
        <Property Name="DiscontinuedDate" Type="Edm.Date" />
        <Property Name="Rating" Type="Edm.Int32" />
        <Property Name="Price" Type="Edm.Decimal" Scale="variable">
          <Annotation Term="Measures.ISOCurrency" Path="Currency" />

        </Property>

        <Property Name="Currency" Type="Edm.String" MaxLength="3" />

        <NavigationProperty Name="Category" Type="ODataDemo.Category"

                            Nullable="false" Partner="Products" />

        <NavigationProperty Name="Supplier" Type="ODataDemo.Supplier"

                            Partner="Products" />
      </EntityType>
      <EntityType Name="Category">
        <Key>
         <PropertyRef Name="ID" />
        </Key>
        <Property Name="ID" Type="Edm.Int32" Nullable="false" />
        <Property Name="Name" Type="Edm.String" Nullable="false">
          <Annotation Term="Core.IsLanguageDependent" />

        </Property>

        <NavigationProperty Name="Products" Partner="Category"

                            Type="Collection(ODataDemo.Product)">
          <OnDelete Action="Cascade" />

        </NavigationProperty>
      </EntityType>
      <EntityType Name="Supplier">
        <Key>
          <PropertyRef Name="ID" />
        </Key>
        <Property Name="ID" Type="Edm.String" Nullable="false" />
        <Property Name="Name" Type="Edm.String" />
        <Property Name="Address" Type="ODataDemo.Address" Nullable="false" />
        <Property Name="Concurrency" Type="Edm.Int32" Nullable="false" />
        <NavigationProperty Name="Products" Partner="Supplier"

                            Type="Collection(ODataDemo.Product)" />

      </EntityType>
      <EntityType Name="Country">

        <Key>

          <PropertyRef Name="Code" />

        </Key>

        <Property Name="Code" Type="Edm.String" MaxLength="2"

                              Nullable="false" />

        <Property Name="Name" Type="Edm.String" />

      </EntityType>

      <ComplexType Name="Address">
        <Property Name="Street" Type="Edm.String" />
        <Property Name="City" Type="Edm.String" />
        <Property Name="State" Type="Edm.String" />
        <Property Name="ZipCode" Type="Edm.String" />
        <Property Name="CountryName" Type="Edm.String" />
        <NavigationProperty Name="Country" Type="ODataDemo.Country">

          <ReferentialConstraint Property="CountryName"  

                                 ReferencedProperty="Name" />

        </NavigationProperty>

      </ComplexType>
      <Function Name="ProductsByRating">
        <Parameter Name="Rating" Type="Edm.Int32" />

        <ReturnType Type="Collection(ODataDemo.Product)" />
      </Function>
      <EntityContainer Name="DemoService">
        <EntitySet Name="Products" EntityType="ODataDemo.Product">
          <NavigationPropertyBinding Path="Category" Target="Categories" />

        </EntitySet>
        <EntitySet Name="Categories" EntityType="ODataDemo.Category">
          <NavigationPropertyBinding Path="Products" Target="Products" />

          <Annotation Term="Core.Description" String="Product Categories" />

        </EntitySet>
        <EntitySet Name="Suppliers" EntityType="ODataDemo.Supplier">
          <NavigationPropertyBinding Path="Products" Target="Products" />

          <NavigationPropertyBinding Path="Address/Country"

                                     Target="Countries" />

          <Annotation Term="Core.OptimisticConcurrency">

            <Collection>

              <PropertyPath>Concurrency</PropertyPath>

            </Collection>

          </Annotation>

        </EntitySet>
        <Singleton Name="MainSupplier" Type="self.Supplier">
          <NavigationPropertyBinding Path="Products" Target="Products" />

          <Annotation Term="Core.Description" String="Primary Supplier" />

        </Singleton>

        <EntitySet Name="Countries" EntityType="ODataDemo.Country" />

        <FunctionImport Name="ProductsByRating" EntitySet="Products"

                        Function="ODataDemo.ProductsByRating" />
      </EntityContainer>
    </Schema>
  </edmx:DataServices>
</edmx:Edmx>`

	return model.API{
		Spec:    []byte(spec),
		ApiType: "odata",
	}
}

func fixEventsAsyncAPIYAML() model.Events {
	spec := `asyncapi: '2.0.0'
info:
  title: AnyOf example
  version: '1.0.0'

channels:
  test:
    publish:
      message:
        $ref: '#/components/messages/testMessages'

components:
  messages:
    testMessages:
      payload:
        anyOf: # anyOf in payload schema
          - $ref: "#/components/schemas/objectWithKey"
          - $ref: "#/components/schemas/objectWithKey2"

  schemas:
    objectWithKey:
      type: object
      properties:
        key:
          type: string
          additionalProperties: false
    objectWithKey2:
      type: object
      properties:
        key2:
          type: string`

	return model.Events{
		Spec: []byte(spec),
	}
}

func fixEventsAsyncAPIJSON() model.Events {
	spec := `{
	"asyncapi": "2.0.0",
	"info": {
		"title": "My API",
		"version": "1.0.0"
	},
	"channels": {
		"mychannel": {
			"publish": {
				"message": {
					"payload": {
						"type": "object",
						"properties": {
							"name": {
								"type": "string"
							}
						}
					}
				}
			}
		}
	}
}`

	return model.Events{
		Spec: []byte(spec),
	}
}
