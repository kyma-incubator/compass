# Labeling

## Requirements

- User can label every top-level entity like Application, Runtime
- User can search Application/Runtime by specific label and its value
- User can learn about all label keys used in given tenant
- User can define validation rules for label with given key but it is optional
- There is one special label: **Scenarios**, that has additional requirements:
    - Every object is labeled with **Scenarios**
    - By default, **Scenarios** has one possible value: **default**
    
    
## Proposed solution

### 10 Commandments
1. Every Label is validated against JSON schema.
Thanks to that our API is extremely simple and we can treat every label in the same way.
Even UI developers can benefit from it, because there are libraries that render proper JS component
on the JSON schema basis. See [this library](https://github.com/json-editor/json-editor).

2. For given Label you can provide only one value. Still value can be JSON array, depending on LabelDefintions schema.
3. LabelDefinition is optional, but it will be created automatically if user uses a new label.
 
### API Changes

1. Extend GraphQL API with  mutations and queries for a new type: **LabelDefinition** 

```graphql
type LabelDefinition {
    key: String!
    schema: String
}

input LabelDefinitionInput {
    key: String!
    schema: String
}


type Query {
    labelDefinitions: [LabelDefinition!]!
    labelDefinition(key: String!): LabelDefinition
}

type Mutation {
    createLabelDefinition(in: LabelDefinitionInput!): LabelDefinition!
    updateLabelDefinition(in: LabelDefinitionInput!): LabelDefinition!
    deleteLabelDefinition(key: String!, force: Boolean=false): LabelDefinition!
}
```

LabelsDefinition key has to be unique for given tenant. Schema defines JSON schema used when user adds label
to the Application or Runtime. 

1. In JSON schema you are able to define if given label accepts one or many values. Because of that, we have to change
or API and allow to specify only one value for given label. This value can contains many elements, depending on LabelDefinition's schema.  
Change from
```graphql
addApplicationLabel(applicationID: ID!, key: String!, values: [String!]!): Label!
```
to:
```graphql
setApplicationLabel(applicationID: ID!, key: String!, value: String!): Label!
```

### Workflows

#### Label an application without creation label definition
We want to keep our API extremely simple. There can be a case that a user wants to label Application/Runtime without providing 
validation rules. Because of that, adding LabelDefinition is optional, and will be created internally. 
Schema in such case will be following: 
```
{ "type": "string" }
```

### Define a LabelDefinition and use that Label

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


setApplicationLabel(applicationID: "123", key: "supportedLanguages", value:"[Go]") {...}

```

### Editing Label Definition
Label definition can be edited. This will be used for example for label **Scenarios**.
When editing definition, we need to ensure that all labels are comatible with a new definition.
If this is not a case, such mutation has to be rejected, with clear message that currently exist Applicatiohns or Runtimes
that have this value with invalid value.
In such case user has two possibilities:
- remove offending labels from specific App/Runtimes  
- remove old LabelDefinition with cascading deletion of all Labels

### Removing Label Definition
```graphql
deleteLabelDefinition(key: String!, force: Boolean=false): LabelDefinition

```
By default, above mutation allows to remove only definitions that it is not used. If you want to 
delete definition and all valuess, set `force` argument to `true`. 

### Editing label definition
Let assume that we have following label definition:
```graphql
 key:"supportedLanguages",
  schema:"{
              "type": "array",
              "items": {
                  "type": "string",
                  "enum": ["Go", "Java", "C#"]
              }
          }"
```
and you application is already labeled with one language:
```
setApplicationLabel(applicationID: "123", key: "priority", value:"[Go]") {...}
```
If you want to add new language to this list, you have to repeat previous values:
```
setApplicationLabel(applicationID: "123", key: "supportedLanguages", value:"[Go,Java]") {...}
```

### Getting list of possible labels
Label definitions are created every time, even when user directly label Application or Runtime with a new key.
Thanks to that, to provide list of possible keys, we need to return all Label Definition for specific tenant.

### Search
There are queries for Applications/Labels where use can define labelFilter.
TODO fix schema
LabelFilter can define:
- label key
- query expression similar to PostgreSQL JSON query language.

TODO test this!

### Special case: Scenario Label
For scenario label we have additional requirements:

- there is a always `default` scenario
- every application/runtime has to be assigned to at least one scenario. If not specified explicitly, `default` scenario is used.

1. On creation of a new tenant, label `Scenario` is created
2. Label `Scenario` cannot be removed
3. `Scenario` is implemented as a list of enums.
4. For `Scenario` label definition, new enum values can be added or removed, but `default` value cannot be removed.
5. On creation/modification of Application/Runtime there is a step that ensures that `Scenario` label exist.


TODO: 
cascading delete - we need another table for that

## Database
When removing LabelDefinition or modyfing it, we need to perform cascading delete or check if all values are compliant with schema definition.
Because of that, it can be beneficial to have a separate table for storing labels.

```
tenant_id | application_id | runtime_id | labelKey | labelID | value
```

Searching by scenario `default`
```sql
select * from application a join labels lab on a.id = lab.application_id where lab.labelKey="scenario" 
    and lab.value ? "default"
```

TODO translation
