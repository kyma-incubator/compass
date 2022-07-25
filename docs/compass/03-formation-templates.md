# Formation Templates

A formation template is a model that is used during the creation of a specific formation type.  The formation template comprises preliminary information about what system types and runtimes are allowed to be included in a specific formation type. 
That is, the actual formation object is modelled, created, and processed, based entirely on the information and limitations that are set in the formation template.
 
The formation templates provide a list of many system types and only one runtime. This way, they control the combination of various systems and a runtime for the productive formation object when created. The runtime, specified in the formation template is a central entity that controls the nature of the formation template. Then, the formation template, in turn, determines the type of the actual formation object.

Additionally, the formation template provides the following metadata fields:
* `runtimeTypeDisplayName` - Represents the official name of the runtime. Unlike `runtimeType`, the value is suitable for UI visualizations and external documents where the official name of the runtime must be used.
* `runtimeArtifactKind` - An enum with the following allowed values: `SUBSCRIPTION`, `SERVICE_INSTANCE`, and `ENVIRONMENT_INSTANCE`.

## GraphQL API
Formation templates are defined as follows:
```graphql
type FormationTemplate {
    id: ID!
    name: String!
    applicationTypes: [String!]!
    runtimeType: String!
    runtimeTypeDisplayName: String!
    runtimeArtifactKind: ArtifactType!
}

type FormationTemplatePage implements Pageable {
    data: [FormationTemplate!]!
    pageInfo: PageInfo!
    totalCount: Int!
}
```

### Mutations
Director API exposes the following GraphQL mutations for managing formation templates: 
```graphql
type Mutation {
    createFormationTemplate(in: FormationTemplateInput! @validate): FormationTemplate @hasScopes(path: "graphql.mutation.createFormationTemplate")
    deleteFormationTemplate(id: ID!): FormationTemplate @hasScopes(path: "graphql.mutation.deleteFormationTemplate")
    updateFormationTemplate(id: ID!, in: FormationTemplateInput! @validate): FormationTemplate @hasScopes(path: "graphql.mutation.updateFormationTemplate")
}
```
> **Note:** For example GraphQL mutations, see: [create](https://github.com/kyma-incubator/compass/tree/main/components/director/examples/create-formation-template/create-formation-template.graphql), [update](https://github.com/kyma-incubator/compass/tree/main/components/director/examples/update-formation-template/update-formation-template.graphql), or [delete](https://github.com/kyma-incubator/compass/tree/main/components/director/examples/delete-formation-template/delete-formation-template.graphql).


### Queries 
Director API exposes the following GraphQL queries for fetching a single formation template by its ID or all formation templates:
```graphql
type Query {
    formationTemplate(id: ID!): FormationTemplate @hasScopes(path: "graphql.query.formationTemplate")
    formationTemplates(first: Int = 200, after: PageCursor): FormationTemplatePage! @hasScopes(path: "graphql.query.formationTemplates")
}
```
> **Note:** For example GraphQL queries, see: [Query formation template](https://github.com/kyma-incubator/compass/tree/main/components/director/examples/query-formation-template/query-formation-template.graphql), or [Query formation templates](https://github.com/kyma-incubator/compass/tree/main/components/director/examples/query-formation-templates/query-formation-templates.graphql).
