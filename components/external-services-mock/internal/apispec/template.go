package apispec

const specJSONTemplate = `{
  "swagger": "2.0",
  "info": {
    "title": "Sample API",
    "description": "API description in Markdown.",
    "version": "1.0.0"
  },
  "host": "api.example.com",
  "basePath": "/v1",
  "schemes": [
    "https"
  ],
  "paths": {
    "/users": {
      "get": {
        "summary": "%s",
        "description": "Optional extended description in Markdown.",
        "produces": [
          "application/json"
        ],
        "responses": {
          "200": {
            "description": "OK"
          }
        }
      }
    }
  }
}`

const specYAMLTemplate = `
asyncapi: '2.0.0'
info:
  title: AnyOf example
  description: %s
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

const specXMLTemplate = `
<?xml version="1.0" encoding="UTF-8"?>
<edmx:Edmx xmlns:edmx="http://docs.oasis-open.org/odata/ns/edmx" xmlns="http://docs.oasis-open.org/odata/ns/edm" Version="4.0">
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
            <Property Name="%s" Type="Edm.String" Nullable="false" />
            <Property Name="Description" Type="Edm.String">
               <Annotation Term="Core.IsLanguageDependent" />
            </Property>
            <Property Name="ReleaseDate" Type="Edm.Date" />
            <Property Name="DiscontinuedDate" Type="Edm.Date" />
            <Property Name="Rating" Type="Edm.Int32" />
            <Property Name="Price" Type="Edm.Decimal" Scale="variable">
               <Annotation Term="Measures.ISOCurrency" Path="Currency" />
            </Property>
            <Property Name="Currency" Type="Edm.String" MaxLength="3" />
            <NavigationProperty Name="Category" Type="ODataDemo.Category" Nullable="false" Partner="Products" />
            <NavigationProperty Name="Supplier" Type="ODataDemo.Supplier" Partner="Products" />
         </EntityType>
         <EntityType Name="Category">
            <Key>
               <PropertyRef Name="ID" />
            </Key>
            <Property Name="ID" Type="Edm.Int32" Nullable="false" />
            <Property Name="Name" Type="Edm.String" Nullable="false">
               <Annotation Term="Core.IsLanguageDependent" />
            </Property>
            <NavigationProperty Name="Products" Partner="Category" Type="Collection(ODataDemo.Product)">
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
            <NavigationProperty Name="Products" Partner="Supplier" Type="Collection(ODataDemo.Product)" />
         </EntityType>
         <EntityType Name="Country">
            <Key>
               <PropertyRef Name="Code" />
            </Key>
            <Property Name="Code" Type="Edm.String" MaxLength="2" Nullable="false" />
            <Property Name="Name" Type="Edm.String" />
         </EntityType>
         <ComplexType Name="Address">
            <Property Name="Street" Type="Edm.String" />
            <Property Name="City" Type="Edm.String" />
            <Property Name="State" Type="Edm.String" />
            <Property Name="ZipCode" Type="Edm.String" />
            <Property Name="CountryName" Type="Edm.String" />
            <NavigationProperty Name="Country" Type="ODataDemo.Country">
               <ReferentialConstraint Property="CountryName" ReferencedProperty="Name" />
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
               <NavigationPropertyBinding Path="Address/Country" Target="Countries" />
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
            <FunctionImport Name="ProductsByRating" EntitySet="Products" Function="ODataDemo.ProductsByRating" />
         </EntityContainer>
      </Schema>
   </edmx:DataServices>
</edmx:Edmx>`
