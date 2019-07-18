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

Use script `./../investigations/storage/sql-toolbox/run_postgres.sh`

Then, following query are possible:
```sql

select app_id,jsonb_path_query(value,'$[*] ? (@ == "bbb" )') from labels where label_key='scenarios';
```

Result:
```
 app_id | jsonb_path_query
--------+------------------
 app-1  | "bbb"
 app-2  | "bbb"
```


https://www.depesz.com/2019/03/19/waiting-for-postgresql-12-partial-implementation-of-sql-json-path-language/

https://www.postgresql.org/docs/12/functions-json.html

https://www.postgresql.org/developer/roadmap/
```
The next major release of PostgreSQL is planned to be the 12 release. A tentative schedule for this version has a release in the third quarter of 2019.

```

https://www.postgresql.org/about/news/1943/

```
JSON path queries per SQL/JSON specification
PostgreSQL 12 now allows execution of JSON path queries per the SQL/JSON specification in the SQL:2016 standard. Similar to XPath expressions for XML, JSON path expressions let you evaluate a variety of arithmetic expressions and functions in addition to comparing values within JSON documents.

A subset of these expressions can be accelerated with GIN indexes, allowing the execution of highly performant lookups across sets of JSON data.
```

???
security considerations


https://kubernetes.io/docs/reference/kubectl/jsonpath/
https://github.com/stedolan/jq/wiki/For-JSONPath-users

https://github.com/kubernetes/kubernetes/blob/7faeee22b109d644f9a0ed736447e8867cae728e/staging/src/k8s.io/client-go/util/jsonpath/jsonpath.go


```
$[?(@ == "aaa")]
```



---
https://news.ycombinator.com/item?id=19949240


GCP: Postgres: 9.6, 11
