# Formation Templates

A Formation Template is the basis for the formations. 
It establishes a contract for each formation of the given type which states what types of applications and what type of runtime is allowed to be assigned to the given formation.
Based on their type we can process the formations differently depending on the Formation Template restrictions and specifics.


Each Formation Template provides a list of application types that are allowed to be assigned to a formation that is based on the template. Currently, every Formation Template allows exactly one runtime type to be included in a formation.
The template also provides additional metadata fields such as `runtimeTypeDisplayName` and `runtimeArtifactKind`. 
The runtime type display name is a short name describing the runtime type and the runtime artifact kind is an enum with the following allowed values: `SUBSCRIPTION`, `SERVICE_INSTANCE`  and `ENVIRONMENT_INSTANCE`.

## GraphQL API
Formation Templates are defined in the following way:
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
Director API exposes the following GraphQL mutations for managing Formation Templates: 
```graphql
type Mutation {
    createFormationTemplate(in: FormationTemplateInput! @validate): FormationTemplate @hasScopes(path: "graphql.mutation.createFormationTemplate")
    deleteFormationTemplate(id: ID!): FormationTemplate @hasScopes(path: "graphql.mutation.deleteFormationTemplate")
    updateFormationTemplate(id: ID!, in: FormationTemplateInput! @validate): FormationTemplate @hasScopes(path: "graphql.mutation.updateFormationTemplate")
}
```
> **TIP:** For the GraphQL mutation examples that you can use, go to the [create](https://github.com/kyma-incubator/compass/tree/main/components/director/examples/create-formation-template/create-formation-template.graphql), [update](https://github.com/kyma-incubator/compass/tree/main/components/director/examples/update-formation-template/update-formation-template.graphql) or [delete](https://github.com/kyma-incubator/compass/tree/main/components/director/examples/delete-formation-template/delete-formation-template.graphql) examples.


### Queries 
Director API exposes the following GraphQL queries for fetching all Formation Templates or a single one given an ID:
```graphql
type Query {
    formationTemplate(id: ID!): FormationTemplate @hasScopes(path: "graphql.query.formationTemplate")
    formationTemplates(first: Int = 200, after: PageCursor): FormationTemplatePage! @hasScopes(path: "graphql.query.formationTemplates")
}
```
> **TIP:** For the GraphQL query examples that you can use, go to the [query formation template](https://github.com/kyma-incubator/compass/tree/main/components/director/examples/query-formation-template/query-formation-template.graphql) or [query formation templates](https://github.com/kyma-incubator/compass/tree/main/components/director/examples/query-formation-templates/query-formation-templates.graphql) examples.
