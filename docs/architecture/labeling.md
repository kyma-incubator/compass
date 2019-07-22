# Labeling

## Requirements

- User can label every top-level entity like Application, Runtime
- User can search Application/Runtime by the specific label and its value
- User can find out about all label keys used in the given tenant
- User can define validation rules for a label with a given key but it is optional
- There is one special label: **Scenarios**, that has additional requirements:
    - Every Application is labeled with **Scenarios**
    - By default, **Scenarios** has one possible value: **default**
    
    
## Proposed solution

1. Application or Runtime can be tagged with a label that consists of **key** and **value**.
1. A label can have related LabelDefinition when a user can define JSON schema for validation of values.
Thanks to that our API is extremely simple and we can treat every label in the same way.
Even UI developers can benefit from it because there are libraries that render proper JS component
on the JSON schema basis. See [this library](https://github.com/json-editor/json-editor).

2. For the given label you can provide only one value. Still, a value can be JSON array, depending on LabelDefintions schema.
3. LabelDefinition is optional, but it will be created automatically if the user adds a label for which LabelDefinition does not exist.
 
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
setApplicationLabel(applicationID: ID!, key: String!, value: String!): Label!
```

### Workflows

#### Label an application without creation label definition
We want to keep our API extremely simple. There can be a case that a user wants to label Application/Runtime without providing 
validation rules. Because of that, adding LabelDefinition is optional, and will be created internally. 
Schema in such case will empty, validation for label is not be performed and user is able to specify any value.

#### Define a LabelDefinition and use that Label

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

#### Editing Label Definition
Label definition can be edited. This will be used for example for label **Scenarios**.
Let assume that we have the following label definition:
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

If you want to add new language to this list, you have to provide a full definition:
```
updateLabelDefinition(in: {
                        key:"supportedLanguages",
                        schema:"{
                                    "type": "array",
                                    "items": {
                                        "type": "string",
                                        "enum": ["Go", "Java", "C#","ABAP"]
                                    }
                                }"
                      }) {...}
```

When editing definition, we need to ensure that all labels are compatible with the new definition.
If this is not a case, such mutation has to be rejected, with a clear message that there are Applications or Runtimes that
have invalid label according to the new LabelDefinition.
In such case a user has two possibilities:
- remove offending labels from specific App/Runtimes  
- remove old LabelDefinition with cascading deletion of all Labels

#### Removing Label Definition
```graphql
deleteLabelDefinition(key: String!, deleteRelatedLabels: Boolean=false): LabelDefinition

```
By default, the above mutation allows removing only definitions that it is not used. If you want to delete definition and all values, set the `deleteRelatedLabels` parameter to `true`. 

#### Getting list of possible labels
Label definitions are created every time, even when a user directly label Application or Runtime with a new key.
Thanks to that, to provide a list of possible label keys, we need to return all Label Definition for a specific tenant.
This functionality can be used in UI, for suggesting already existing label's key.

#### Search
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
- JSON Path (https://goessner.net/articles/JsonPath/)
- jq
- jmespath
- mongo DB has it's own query language
- kubectl does not mention which standard they are supporting: https://kubernetes.io/docs/reference/kubectl/jsonpath/
but it looks the same as Goessner's JSON Path. It looks that they implemented parsing on their own:
 https://github.com/kubernetes/kubernetes/blob/7faeee22b109d644f9a0ed736447e8867cae728e/staging/src/k8s.io/client-go/util/jsonpath/jsonpath.go

- SQL/JSON Path Expressions (part of SQL-2016 specification http://www.sai.msu.su/~megera/postgres/talks/jsonpath-pgday.it-2019.pdf, https://news.ycombinator.com/item?id=19949240). 

Options enumerated above are not compatible with each other.

The simplest solution will be to use SQL/JSON Path, then we can propagate that value directly to the PostgreSQL. 
See [this section](#database-schema) to learn how it can be implemented.
Unfortunately, this functionality is planned for PostgreSQL 12, which is going to be released in Q3 2019, see [roadmap](https://www.postgresql.org/developer/roadmap/) and [features highlights](https://www.postgresql.org/about/news/1943/).
We don't know when this version will be available on GCP or AWS, so, for now, we will be forced to use Postgres running inside the cluster.
Also, not all relational databases support JSON Path Expressions, other than Postgres is [SQL Server](https://docs.microsoft.com/en-us/sql/relational-databases/json/json-path-expressions-sql-server?view=sql-server-2017) 
Because of that, the safest approach will be to use limited SQL/JSON Path Expressions syntax, that supports currently only **Scenarios** label (`$[*] ? (@ == "default" )`) and internally translate it to PostgreSQL 11 JSON syntax.


#### Special case: Scenario Label
For scenario label we have additional requirements:

- there is always a `default` scenario
- every Application has to be assigned to at least one scenario. If not specified explicitly, the `default` scenario is used.

1. On creation of a new tenant, label `Scenario` is created. Because right now we don't have a mutation for creating tenant, we need to
perform that on every Runtime/Application creation.
2. Label `Scenario` cannot be removed. This requires additional custom validation.
3. `Scenario` is implemented as a list of enums.
4. For `Scenario` label definition, new enum values can be added or removed, but `default` value cannot be removed. This requires additional custom validation.
5. On creation/modification of Application, there is a step that ensures that `Scenario` label exists.

### Database Schema
When removing LabelDefinition or modifying it, we need to perform cascading delete or check if all values are compliant with the schema definition.
Because of that, it can be beneficial to have a separate table for storing labels.

```
tenant_id | application_id | runtime_id | labelKey | labelID | value
```

Full example:
```sql
create database compass;

\connect compass


create table labels (
  id varchar(100),
  tenant varchar(100),
  app_id varchar(100),
  runtime_id varchar(100),
  label_key varchar(100),
  label_id varchar(100),
  value JSONB
);

insert into labels(id,tenant,app_id,label_key, label_id,value) values
  ('1','adidas','app-1','scenarios','label-def-1','["aaa","bbb"]'),
  ('2','adidas','app-2','scenarios', 'label-def-1','["bbb","ccc"]'),
  ('3','adidas','app-3', 'abc','label-def-2','{"name": "John", "age": 32}'),
  ('4','adidas','app-4', 'abc','label-def-2','{"name": "Pamela", "age": 48}');

```

Use script `./../investigations/storage/sql-toolbox/run_postgres.sh` to run PostgreSQL in version 12.

Then, following query returns all applications that have **scenarios** `bbb`:
```sql

select app_id,jsonb_path_query(value,'$[*] ? (@ == "bbb" )') from labels where label_key='scenarios';
```

Label key and SQL/JSON Path query will be provided by the user, which makes implementation extremely simple.

Result:
```
 app_id | jsonb_path_query
--------+------------------
 app-1  | "bbb"
 app-2  | "bbb"
```


Full list of supported operations can be found [here](https://www.postgresql.org/docs/12/functions-json.html#FUNCTIONS-SQLJSON-PATH).
