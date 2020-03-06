# Labels

A label is a key-value pair that you can add to every top-level entity, such as an Application or a Runtime. It allows you to search for and find out about all Applications and Runtimes label keys used in a given tenant. You can also define validation rules for labels with a given key using LabelDefinitions.

## LabelDefinitions

For every label, you can create a LabelDefinition to set validation rules for values. LabelDefinitions are optional, but they are created automatically if the user adds labels for which LabelDefinitions do not exist. You can manage LabelDefinitions using mutations and queries. See the examples:

```graphql
scalar JSONSchema # the same as Any

type LabelDefinition {
    key: String!
    schema: JSONSchema
}

input LabelDefinitionInput {
    key: String!
    schema: JSONSchema
}


type Query {
    labelDefinitions: [LabelDefinition!]!
    labelDefinition(key: String!): LabelDefinition
}

type Mutation {
    createLabelDefinition(in: LabelDefinitionInput!): LabelDefinition!
    updateLabelDefinition(in: LabelDefinitionInput!): LabelDefinition!
    deleteLabelDefinition(key: String!, deleteRelatedLabels: Boolean=false): LabelDefinition!
}
```

LabelDefinition key has to be unique for a given tenant. For the given label, you can provide only one value, however, this label can contain many elements, depending on the LabelDefinition schema. In the following example, the type of label's value is `Any`, which means that the value can be of any type, such as `JSON`, `string`, `int`:

```graphql
setApplicationLabel(applicationID: ID!, key: String!, value: Any!): Label!
```

### Label Applications without creating LabelDefinitions

You can label an Application or a Runtime without providing validation rules. Adding LabelDefinitions is optional, however, they are created automatically if you add labels for which LabelDefinitions do not exist. In such a case, the schema is empty and labels validation is not performed, which means that the user can provide any value.

### Define and set LabelDefinitions

See the example of how you can create and set a LabelDefinition:

```
createLabelDefinition(in: {
  key:"supportedLanguages",
  schema:"{
              "type": "array",
              "items": {
                  "type": "string",
                  "enum": ["Go", "Java", "C#"]
              }
          }"
}) {...}


setApplicationLabel(applicationID: "123", key: "supportedLanguages", value:["Go"]) {...}

```

### Edit LabelDefinitions

You can edit LabelDefinitions using a mutation. When editing a LabelDefinition, make sure that all labels are compatible with the new definition. Otherwise, the mutation is rejected with a clear message that there are Applications or Runtimes that have an invalid label according to the new LabelDefinition. In such a case, you must either:
- Remove incompatible labels from specific Applications and Runtimes, or
- Remove the old LabelDefinition from all labels using cascading deletion.

For example, let's assume we have the following LabelDefinition:

```graphql
 key:"supportedLanguages",
  schema:{
              "type": "array",
              "items": {
                  "type": "string",
                  "enum": ["Go", "Java", "C#"]
              }
          }
```

If you want to add a new language to the list, provide such a mutation:

```
updateLabelDefinition(in: {
                        key:"supportedLanguages",
                        schema:{
                                    "type": "array",
                                    "items": {
                                        "type": "string",
                                        "enum": ["Go", "Java", "C#","ABAP"]
                                    }
                                }
                      }) {...}
```

### Remove LabelDefinitions

Use this mutation to remove a LabelDefinition:

```graphql
deleteLabelDefinition(key: String!, deleteRelatedLabels: Boolean=false): LabelDefinition

```
This mutation allows you to remove only definitions that are not used. If you want to delete a LabelDefinition with all values, set the **deleteRelatedLabels** parameter to `true`.

### Search for objects using LabelFilters

You can define a LabelFilter to list the objects according to their labels. To search for a given Application or Runtime, use this query:

```graphql
 applications(filter: [LabelFilter!], first: Int = 100, after: PageCursor):  ApplicationPage!
```

To search for all objects with a given label despite their values, use this query:

```graphql
query {
  applications(filter:[{key:"scenarios"
    }]) {
    data {
      name
      labels
    }
    totalCount
  }
}
```

To search for all objects assigned to the `default` scenario, use this query:

```graphql
query {
  applications(filter:[{key:"scenarios",
    # This field is optional. If not provided, the query returns every object with the given label, regardless its value.
    query:"$[*] ? (@ == \"DEFAULT\")"}]) {
    data {
      name
      labels
    }
    totalCount
  }
}
```

In the **query** field, use only the limited SQL/JSON path expressions. The supported syntax is `$[*] ? (@ == "{VALUE}" )`.

## **Scenarios** label

Every Application is labeled with the special **Scenarios** label which by default has the `default` value assigned. As every Application has to be assigned to at least one scenario, if not specified explicitly, the `default` scenario is used.

When you create a new tenant, the **Scenarios** LabelDefinition is created. It defines a list of possible values that can be used for the **Scenarios** label. Every time you create or modify an Application, there is a step that ensures that **Scenarios** label exists. You can add or remove values from the **Scenarios** LabelDefinition list, but neither the `default` value, nor the **Scenarios** label can be removed.
