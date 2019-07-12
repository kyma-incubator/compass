# Labeling

The following documents describes the labeling feature.

## Requirements

- Creating, Reading and deleting all labels within tenant
- Defining `scenarios` label values and using them for labeling Applications and Runtimes
- Reading all label keys for tenant
- Validating label values against JSON schema / type
- Labeling new Application/Runtime with `scenarios: default` label, if no scenario label is specified

## Terms
In this document there are few terms, which are used:

**Label**

Label is a tag in a form of `key:value`, which can be assigned to an Application or Runtime. Label references **LabelDefinition** which defines the shape for value.
It holds actual label key and value. A single Label can be assigned to Runtime or Application. 

Label is tenant specific. You cannot change the label key, once you created the label.

```go
type Label struct {
    Key string // Key 
    Value interface{} // Value
    Tenant string
    DefinitionID int // LabelDefinition reference
    ApplicationIDs []*int
    RuntimeIDs []*int
}
```

**LabelDefinition** 

Describes what's the type of value for a label key. Validation will be done against the LabelDefinition JSON Schema or type.
Label definition contains `key` property, which is unique and represents a label key.

LabelDefinition is *not* reusable between two labels with different keys. LabelDefinition is tenant-specific.

```go
type LabelDefinition struct {
    Key string
    Tenant string
    Type LabelDefinitionType
    Schema *string
    enum []string
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


#TODO: split to separate app and runtime mutations
    """It won't allow to create Label if LabelDefinition for the key is missing. Also it doesn't allow to set Label if it does already exist"""
    setSingleLabelForApplication(applicationID: ID!, key: String!, value: Any!): Label!
    """It won't allow to create Label if LabelDefinition for the key is missing. Also it doesn't allow to set Label if it does already exist"""
    setSingleLabelForRuntime(runtimeID: ID!, key: String!, value: Any!): Label!

    """It won't allow to update the Label if the LabelDefinition for the key is missing."""
    setArrayLabelForRuntime(runtimeID: ID!, key: String!, values: [Any!]!): Label!
    """It won't allow to update the Label if the LabelDefinition for the key is missing."""
    setArrayLabelForApplication(applicationID: ID!, key: String!, values: [Any!]!): Label!

    """Removes Label along with all its values. It doesn't remove LabelDefinition"""
    removeApplicationLabel(applicationID: ID!, key: String!): Label!
    """Removes Label along with all its values. It doesn't remove LabelDefinition"""
    removeRuntimeLabel(runtimeID: ID!, key: String!): Label!

    """It won't allow to update the Label if the Label or LabelDefinition for the key is missing."""
    addArrayLabelValuesForRuntime(runtimeID: ID!, key: String!, value: [Any!]!): Label!
    """It won't allow to update the Label if the Label or LabelDefinition for the key is missing."""
    addArrayLabelValuesForApplication(applicationID: ID!, key: String!, value: [Any!]!): Label!

    """It won't allow to remove label value if the label is not of array type"""
    removeArrayLabelValuesForApplication(applicationID: ID!, key: String!, values: [Any!]!): Label!
    """It won't allow to remove label value if the label is not of array type"""
    removeArrayLabelValuesForRuntime(runtimeID: ID!, key: String!, values: [Any!]!): Label!
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
    runtimes(first: Int = 100, after: PageCursor): RuntimePage! # Resolver
    applications(first: Int = 100, after: PageCursor): ApplicationPage! # Resolver
}

type SingleLabel implements LabelBase {
    key: String!
    definition: LabelDefinition!
    value: Any!
    runtimes(first: Int = 100, after: PageCursor): RuntimePage! # Resolver
    applications(first: Int = 100, after: PageCursor): ApplicationPage! # Resolver
}

type ArrayLabel implements LabelBase {
    key: String!
    definition: LabelDefinition!
    values: [Any!]!
    runtimes(first: Int = 100, after: PageCursor): RuntimePage! # Resolver
    applications(first: Int = 100, after: PageCursor): ApplicationPage! # Resolver
}

union Label = SingleLabel | ArrayLabel

```


## Storage

// TODO:

## Workflows

### Creating Application / Runtime without label

1. User creates an Application without label.
1. Label `scenarios` is automatically created with the reference
