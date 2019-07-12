# Labeling

The following document describes the labeling feature.

## Requirements

- Creating, reading and deleting all labels within tenant
- Defining `scenarios` label values and using them for labeling Applications and Runtimes
- Reading all label keys for tenant
- Validating label values against JSON schema / type
- Labeling new Application/Runtime with `scenarios: default` label, if no scenario label is specified

## Terms
In this document there are few terms, which are used:

**Label**

Label is a tag in a form of `key:value`, which can be assigned to an Application or Runtime. Label references **LabelDefinition** which defines the shape for value.
It holds actual label key and value. A single Label can be assigned to single object: Runtime or Application. 

Label is tenant specific. You cannot change the label key, once you created the label.

```go
type Label struct {
    Key string // Key 
    Value interface{} // Value
    Tenant string
    DefinitionID string // LabelDefinition reference
    ObjectID string
    ObjectType LabelObjectType
}

type LabelObjectType string

const (
    RuntimeLabelObjectType = "Runtime"
    ApplicationLabelObjectType = "Application"
)
```

**LabelDefinition** 

Describes what's the type of value for a label key. Validation will be done against the LabelDefinition JSON Schema or type.
Label definition contains `key` property, which is unique and represents a label key.

LabelDefinition is *not* reusable between two labels with different keys. LabelDefinition is tenant-specific.

```go
type LabelDefinition struct {
    ID string
    Key string
    Tenant string
    Type LabelDefinitionType
    Schema *string
    Enum *[]string
}
```

**LabelDefinitionType**

Enum which represents a type of **LabelDefinition**.

LabelDefinitionType is not dynamic - you cannot add new values in runtime.

```go
type LabelDefinitionType string

const (
    JSONSchemaLabelDefinition LabelDefinitionType = "JSONSchema"
    StringLabelDefinition LabelDefinitionType = "String"
    EnumLabelDefinition LabelDefinitionType = "Enum"
)
```

**Scenario**

It is a Label with key `scenarios`, which references LabelDefinition of type `Enum`. Based on the `scenarios` Label, Applications and Runtimes are connected. Runtime Agent does query for all Scenarios assigned to Application.

## API

To manage Labels and LabelDefinitions, the following GraphQL API is proposed:

```graphql
type Query {
    # (...)
    labelDefinitions: [LabelDefinition!]!
    labels: [Label!]!
    label(key: String!): Label
}

type Mutation {
    # (...)
    createLabelDefinition(in: LabelDefinitionInput!): LabelDefinition!

    """It won't allow to delete LabelDefinition if some labels use it"""
    deleteLabelDefinition(key: String!): LabelDefinition

    """If the LabelDefinition for the key is not specified, it will create LabelDefinition to String"""
    setSingleLabel(objectID: ID!, objectType: ObjectType!, key: String!, value: Any!): Label!

    """If the LabelDefinition for the key is not specified, it will create LabelDefinition to String"""
    setArrayLabel(objectID: ID!, objectType: ObjectType!, key: String!, values: [Any!]!): Label!

    """Removes Label along with all its values. It doesn't remove LabelDefinition"""
    removeLabel(objectID: ID!, objectType: ObjectType!, key: String!): Label!

    """It won't allow to update the Label if the Label or LabelDefinition for the key is missing."""
    addArrayLabelValues(objectID: ID!, objectType: ObjectType!, key: String!, value: [Any!]!): Label!

    """It won't allow to remove label value if the label is not of array type"""
    removeArrayLabelValues(objectID: ID!, objectType: ObjectType!, key: String!, values: [Any!]!): Label!
}

# Label Definition

interface LabelDefinitionBase {
    key: String!
    type: LabelDefinitionType!
}

type GenericLabelDefinition implements LabelDefinitionBase {
    key: String!
    type: LabelDefinitionType!
}

type EnumLabelDefinition implements LabelDefinitionBase {
    key: String!
    type: LabelDefinitionType!
    enum: [String!]!
}

type JSONSchemaLabelDefinition implements LabelDefinitionBase {
    key: String!
    type: LabelDefinitionType!
    schema: CLOB!
}

union LabelDefinition = GenericLabelDefinition | EnumLabelDefinition | JSONSchemaLabelDefinition

enum LabelDefinitionType! {
    JSON_SCHEMA
    STRING
    ENUM
}

type LabelDefinitionInput {
    key: String!
    type: LabelDefinitionType!
    jsonSchema: CLOB
    enum: [String!]
}

# Label

interface LabelBase {
    key: String!
    definition: LabelDefinition!
    objectID: ID!
    objectType: ObjectType!
    object: Object!
}

type SingleLabel implements LabelBase {
    key: String!
    definition: LabelDefinition!
    value: Any!
    objectID: ID!
    objectType: ObjectType!
    object: Object!
}

type ArrayLabel implements LabelBase {
    key: String!
    definition: LabelDefinition!
    values: [Any!]!
    objectID: ID!
    objectType: ObjectType!
    object: Object!
}

union Label = SingleLabel | ArrayLabel

enum ObjectType {
    RUNTIME,
    APPLICATION
}

union Object = Runtime | Application

```


## Storage

// TODO:

## Workflows

### Creating Application / Runtime without label

1. User creates an Application without label.
1. Label `scenarios` is automatically created with the reference
