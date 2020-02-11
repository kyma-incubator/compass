# Packages API

## Introduction

This document describes API for Packages. Package is an entity, which groups multiple API Definitions, Event Definitions and Documents. It also stores shared credentials for all APIs defined under the particular Package.

![API Packages Diagram](./assets/packages-api.svg)

On Kyma Runtime, every Application is represented as a single Service Class, and every Package of a given Application is represented as a single Service Plan in Service Catalog. It allows user to consume multiple APIs and Events with a single Service Instance.

## Assumptions

- A single API, Event Definition and Document can be a part of a single Package. A single Package can contain multiple API, Event Definitions and Documents.
- Package belongs to a single Application entity.

## GraphQL API

In order to manage Packages, Director exposes the following GraphQL API:

```graphql
type Package {
  id: ID!
  name: String!
  description: String

  # (...) Auth-related fields, described in Credentials Request for Packages document

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
  Temporary name before doing breaking change. Eventually the `addAPIDefinition` mutation will be changed and there will be just one mutation: `addAPIDefinitionToPackage`.
  """
  addAPIDefinitionToPackage(
    packageID: ID!
    in: APIDefinitionInput! @validate
  ): APIDefinition!
  """
  Temporary name before doing breaking change. Eventually the `addEventDefinition` mutation will be changed and there will be just one mutation: `addEventDefinitionToPackage`.
  """
  addEventDefinitionToPackage(
    packageID: ID!
    in: EventDefinitionInput! @validate
  ): EventDefinition!
  """
  Temporary name before doing breaking change. Eventually the `addDocument` mutation will be changed and there will be just one mutation: `addDocumentToPackage`.
  """
  addDocumentToPackage(packageID: ID!, in: DocumentInput! @validate): Document!
    @hasScopes(path: "graphql.mutation.addDocumentToPackage")

  addPackage(applicationID: ID!, in: PackageCreateInput! @validate): Package!
  updatePackage(id: ID!, in: PackageUpdateInput! @validate): Package!
  deletePackage(id: ID!): Package!
}
```

## Package credentials

To read about Package credentials flow, how to provide optional input parameters during Service Instance creation, see the [Credential requests for Packages](./03-packages-credential-requests.md) document.
