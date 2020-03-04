# Labels

You can add a label to every top-level entity, such as Application or Runtime. A label consists of a key-value pair that allows you to search for an Application/Runtime. You can find out about all label keys used in the given tenant
You can also define validation rules for a label with a given key but it is optional.

## LabelDefinitions

For every label, you can create a LabelDefinition to set validation rules for values. For the given label, you can provide only one value, however, it can also be a JSON array, depending on the LabelDefintions schema.

>**NOTE:** LabelDefinitions are optional, but they are created automatically if the user adds labels for which LabelDefinitions do not exist.

1. A label can have related LabelDefinition when a user can define JSON schema for validation of values.


### API

1. Extend the Director's GraphQL API to manage LabelDefinitions using mutations and queries for a new type: **LabelDefinition**

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

The LabelsDefinition key has to be unique for a given tenant. A schema defines JSON schema used when a user adds a label
to the Application or Runtime.

1. In JSON schema user can define if given label accepts one or many values. Because of that, we have to change
or API and allow to specify only one value for a given label. This value can contain many elements, depending on LabelDefinition's schema.  
Change from
```graphql
addApplicationLabel(applicationID: ID!, key: String!, values: [String!]!): Label!
```
to:
```graphql
setApplicationLabel(applicationID: ID!, key: String!, value: Any!): Label!
```
As you can see, type of label's value is `Any`, it can by `JSON`, `string`, `int`, etc.


### Label an Application without creating a LabelDefinition

You can label an Application or Runtime without providing validation rules. Adding LabelDefinitions is optional, and will be created internally.
Schema in such case will empty, validation for label is not be performed and user is able to specify any value.

### Define a LabelDefinition and use that label

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

### Edit LabelDefinition

Label definition can be edited. This will be used for example for label **Scenarios**.
Let assume that we have the following label definition:

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

If you want to add new language to this list, you have to provide a full definition:
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

When editing definition, we need to ensure that all labels are compatible with the new definition.
If this is not a case, such mutation has to be rejected, with a clear message that there are Applications or Runtimes that
have invalid label according to the new LabelDefinition.
In such case a user has two possibilities:
- remove offending labels from specific App/Runtimes  
- remove old LabelDefinition with cascading deletion of all Labels

### Remove LabelDefinition
```graphql
deleteLabelDefinition(key: String!, deleteRelatedLabels: Boolean=false): LabelDefinition

```
By default, the above mutation allows removing only definitions that it is not used. If you want to delete definition and all values, set the `deleteRelatedLabels` parameter to `true`.

### Get the list of possible labels
Label definitions are created every time, even when a user directly label Application or Runtime with a new key.
Thanks to that, to provide a list of possible label keys, we need to return all Label Definition for a specific tenant.
This functionality can be used in UI, for suggesting already existing label's key.

### Search
There are queries for Applications/Runtimes where user can define LabelFilter:
```graphql
 applications(filter: [LabelFilter!], first: Int = 100, after: PageCursor):  ApplicationPage!
```

Because now every label is defined by JSON Schema, LabelFilter needs to be changed, from:
```graphql
input LabelFilter {
    label: String!
    values: [String!]!
    operator: FilterOperator = ALL
}
```
to:
```graphql
input LabelFilter {
    label: String!
    query: String # optional, if not provided returns every object with given label regardless of its value.
}
```

Challenging part is how the user will provide **query** field.
There is no standard query language for JSON, see [discussion](https://stackoverflow.com/questions/777455/is-there-a-query-language-for-json).
We have many alternatives:

- SQL/JSON Path Expressions (part of SQL-2016 specification http://www.sai.msu.su/~megera/postgres/talks/jsonpath-pgday.it-2019.pdf, https://news.ycombinator.com/item?id=19949240).

The simplest solution will be to use SQL/JSON Path, then we can propagate that value directly to the PostgreSQL.
See [this section](#database-schema) to learn how it can be implemented.
Unfortunately, this functionality is planned for PostgreSQL 12, which is going to be released in Q3 2019, see [roadmap](https://www.postgresql.org/developer/roadmap/) and [features highlights](https://www.postgresql.org/about/news/1943/).
We don't know when this version will be available on GCP or AWS, so, for now, we will be forced to use Postgres running inside the cluster.
Also, not all relational databases support JSON Path Expressions, other than Postgres is [SQL Server](https://docs.microsoft.com/en-us/sql/relational-databases/json/json-path-expressions-sql-server?view=sql-server-2017)
Because of that, the safest approach will be to use limited SQL/JSON Path Expressions syntax, that supports currently only **Scenarios** label (`$[*] ? (@ == "default" )`) and internally translate it to PostgreSQL 11 JSON syntax.


## `Scenario` label

- There is one special label: **Scenarios**, that has additional requirements:
    - Every Application is labeled with **Scenarios**
    - By default, **Scenarios** has one possible value: **default**


For scenario label we have additional requirements:

- there is always a `default` scenario
- every Application has to be assigned to at least one scenario. If not specified explicitly, the `default` scenario is used.

1. On creation of a new tenant, a Label Definition `Scenario` is created. Because right now we don't have a mutation for creating tenant, we need to
perform that on every Runtime/Application creation.
2. Label `Scenario` cannot be removed. This requires additional custom validation.
3. `Scenario` is implemented as a list of enums.
4. For `Scenario` label definition, new enum values can be added or removed, but `default` value cannot be removed. This requires additional custom validation.
5. On creation/modification of Application, there is a step that ensures that `Scenario` label exists.
