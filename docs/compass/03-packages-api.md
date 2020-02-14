# API for Packages


Package is an entity that groups multiple API Definitions, Event Definitions, and Documents. It also stores shared credentials for all APIs defined under the particular Package.

![API Packages Diagram](./assets/packages-api.svg)

In Kyma Runtime, every Application is represented as a single ServiceClass, and every Package of a given Application is represented as a single ServicePlan in the Service Catalog. It allows users to consume multiple APIs and Events with a single ServiceInstance.


A single Package can contain many different API Definitions/Event Definitions/Documents but the same API Definition/Event Definition/Document cannot belong to two different Packages. One Packages can belong only to one Application.


## GraphQL API

In order to manage Packages, Director exposes the following GraphQL API:

```graphql
type Package {
  id: ID!
  name: String!
  description: String

  # (...) Auth-related fields described in the `Credentials Request for Packages` document

  apiDefinitions(
    group: String
    first: Int = 100
    after: PageCursor
  ): APIDefinitionPage
  eventDefinitions(
    group: String
    first: Int = 100
    after: PageCursor
  ): EventDefinitionPage
  documents(first: Int = 100, after: PageCursor): DocumentPage
  apiDefinition(id: ID!): APIDefinition
  eventDefinition(id: ID!): EventDefinition
  document(id: ID!): Document
}

type PackagePage implements Pageable {
  data: [Package!]!
  pageInfo: PageInfo!
  totalCount: Int!
}

type Mutation {
  # (...)

  """
  Temporary name before doing a breaking change. Eventually, the `addAPIDefinition` mutation will be changed and there will be just one mutation: `addAPIDefinitionToPackage`.
  """
  addAPIDefinitionToPackage(
    packageID: ID!
    in: APIDefinitionInput! @validate
  ): APIDefinition!
  """
  Temporary name before doing a breaking change. Eventually, the `addEventDefinition` mutation will be changed and there will be just one mutation: `addEventDefinitionToPackage`.
  """
  addEventDefinitionToPackage(
    packageID: ID!
    in: EventDefinitionInput! @validate
  ): EventDefinition!
  """
  Temporary name before doing a breaking change. Eventually, the `addDocument` mutation will be changed and there will be just one mutation: `addDocumentToPackage`.
  """
  addDocumentToPackage(packageID: ID!, in: DocumentInput! @validate): Document!
    @hasScopes(path: "graphql.mutation.addDocumentToPackage")

  addPackage(applicationID: ID!, in: PackageCreateInput! @validate): Package!
  updatePackage(id: ID!, in: PackageUpdateInput! @validate): Package!
  deletePackage(id: ID!): Package!
}
```

## Package credentials

To learn about credentials flow for Packages and how to provide optional input parameters when provisioning a ServiceInstance, read [this](./03-packages-credential-requests.md) document.
