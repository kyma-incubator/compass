# Labeling

The following document describes the labeling feature.

## Terms

In this document there are few terms, which are used:

**Label**

Label is a tag in a form of `key:value`, which can be assigned to an Application or Runtime. Label references **LabelDefinition** which defines the shape for value.

Label key is unique

Label is tenant specific. You cannot change the label key, once you created the label.

```go
type Label struct {
    ID string
    Key string // Unique
    DefinitionID string // LabelDefinition reference
    Tenant string
}
```

**LabelUsage**

It holds actual label values. A single LabelUsage can be assigned to one labelable object: Runtime or Application. LabelUsage references **LabelID**.

```go
type LabelUsage struct {
    ID string
    LabelID string // or label key?
    ObjectID string // LabelableObject reference
    ObjectType LabelableObjectType // LabelableObjectType
    Values []interface{}
}

type LabelableObjectType string

const (
    RuntimeLabelObjectType LabelableObjectType = "Runtime"
    ApplicationLabelObjectType LabelableObjectType = "Application"
)
```

**LabelDefinition**

Describes what's the type of value for a label key. Validation will be done against the LabelDefinition JSON Schema or type.
Label definition contains `key` property, which is unique and represents a label key.

LabelDefinition is _not_ reusable between two labels with different keys. LabelDefinition is tenant-specific.

It can store default value. When a new Runtime/Application is created, new LabelUsage will be created along with the value described in `DefaultValue` field.

```go
type LabelDefinition struct {
    Key string // unique
    Tenant string
    Type LabelDefinitionType
    DefaultValue interface{}
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

It is a Label with key `scenarios`, which references LabelDefinition of type `ENUM`. Based on the `scenarios` Label, Applications and Runtimes are connected. Runtime Agent does query for all Scenarios assigned to Application.

While adding new tenat, LabelDefinition `scenarios` is automatically created.

## Requirements

- Creating, reading and deleting all labels within tenant
- Defining `scenarios` label values and using them for labeling Applications and Runtimes
- Reading all label keys for tenant
- Validating label values against JSON schema / type
- Labeling new Application/Runtime with `scenarios: default` label, if no scenario label is specified
- Query Applications / Runtimes by particular scenario (key: value)

## API

To manage Labels and LabelDefinitions, the following GraphQL API is proposed:

```graphql

type Query {
    # (...)
    labelDefinitions: [LabelDefinition!]!
    labelDefinition(key: String!): LabelDefinition

    applications(filter)
    runtimes(filter)
}

type Mutation {
    # (...)
    createLabelDefinition(in: LabelDefinitionInput!): LabelDefinition!

    """It won't allow to delete LabelDefinition if some labels use it"""
    deleteLabelDefinition(key: String!): LabelDefinition

    addEnumLabelDefinitionValues(key: String!, values: [String!]!): LabelDefinition!
    deleteEnumLabelDefinitionValues(key: String!, values: [String!]!): LabelDefinition!

    """If the LabelDefinition for the key is not specified, it will create LabelDefinition to String"""
    addLabelUsage(objectID: ID!, objectType: LabelableObjectType!, key: String!, values: [Any!]!): LabelUsage!

    """Removes Label along with all its values. It doesn't remove LabelDefinition"""
    removeLabelUsage(objectID: ID!, objectType: LabelableObjectType!, key: String!): LabelUsage!

    """It won't allow to update the Label if the Label or LabelDefinition for the key is missing."""
    addLabelUsageValues(objectID: ID!, objectType: LabelableObjectType!, key: String!, values: [Any!]!): LabelUsage!

    """Overwrites array label values"""
    updateLabelUsageValues(objectID: ID!, objectType: LabelableObjectType!, key: String!, values: [Any!]!): LabelUsage!

    """It won't allow to remove label value if the label is not of array type"""
    deleteLabelUsageValues(objectID: ID!, objectType: LabelableObjectType!, key: String!, values: [Any!]!): LabelUsage!
}

# Label Definition

interface LabelDefinition {
    key: String!
    # type is stored in __typename
    defaultValue: Any #TODO: Find better name
}

type StringLabelDefinition implements LabelDefinition {
    key: String!
    defaultValue: Any
}

type EnumLabelDefinition implements LabelDefinition {
    key: String!
    enum: [String!]!
    defaultValue: Any
}

type JSONSchemaLabelDefinition implements LabelDefinition {
    key: String!
    schema: CLOB!
    defaultValue: Any
}

type LabelDefinitionInput {
    key: String!
    type: LabelDefinitionType!
    jsonSchema: CLOB
    enum: [String!]
}

enum LabelDefinitionType {
    STRING,
    ENUM,
    JSON_SCHEMA
}

# used in Runtime and Application types
type ObjectLabel {
    key: String!
    values: [Any!]! # If the LabelDefinition is set to JSON_SCHEMA one, validate every single object against that schema.
    definition: LabelDefinition! #resolver
}

interface LabelUsage {
    values: [Any!]! # If the LabelDefinition is set to JSON_SCHEMA one, validate every single object against that schema.
    objectID: ID! #TODO:
    objectType: LabelableObjectType!
    referencedObject: LabelableObject # resolver
}

input LabelUsageFilter {
    values: [String!]
    objectType: LabelableObjectType
    operator: FilterOperator = ALL
}

enum FilterOperator {
    ALL, ANY
}

union LabelableObject = Runtime | Application

enum LabelableObjectType {
    RUNTIME,
    APPLICATION
}

#
# Changes in Application and Runtime types
#

type Runtime {
    # (...)
    labels: [ObjectLabel!]! # resolver
    label(key: String!): ObjectLabel # resolver
}

type Application {
    # (...)
    labels: [ObjectLabel!]! # resolver
    label(key: String!): ObjectLabel # resolver
}
```

## Workflows

In this proposal, there were considered a plenty of different cases, described in this section.

### Creating Application / Runtime with Scenario label

1. User creates Application / Runtime (`createRuntime` or `createApplication` ), defining scenarios values in input.
1. Scenarios values are automatically validated against LabelDefinition enum values.

### Creating Application / Runtime without Scenario label

1. User creates Application / Runtime without label (`createRuntime` or `createApplication` mutation).
1. Label `scenarios` is automatically created for the Application / Runtime, with the reference to `EnumLabelDefinition` for Scenarios. The label `scenarios` has `DEFAULT` value inserted.

### Creating Application / Runtime with custom single label

1. User creates Application / Runtime (`createRuntime` or `createApplication` mutation).
1. User adds new LabelDefinition of `string`/`enum`/`JSON Schema` type for the label key.
1. User adds new label to the Application / Runtime (`setSingleLabel` with reference to the LabelDefinition created in previous step).
1. The label value is automatically validated against LabelDefinition.

### Creating Application / Runtime with custom array label

1. User creates Application / Runtime (`createRuntime` or `createApplication` mutation).
1. User adds new LabelDefinition of `string`/`enum`/`JSON Schema` type for the label key.
1. User adds new array label to the Application / Runtime (`setArrayLabel`).
1. The label values are automatically validated LabelDefinition.

### Adding new Scenario to Runtime / Application

1. User modifies LabelDefinition for `scenarios`, adding new enum value (`addEnumLabelDefinitionValues` mutation). 
1. User adds new scenario value to Runtime / Application. It can be done in two methods:
    - User updates scenarios through `updateApplication` or `updateRuntime` mutation.
    - User adds array label value (`addArrayLabelValues` mutation) to Runtime/Application `scenarios` label.
1. The new `scenarios` label value is validated with the related LabelDefinition type.

> **NOTE:** This workflow also applies to any enum array label, except that in second step user has to use `addArrayLabelValues` mutation.

### Removing one Scenario value, when it is used in Runtime or Application

1. User removes label the value of `scenarios` label (`deleteArrayLabelValues` mutation). It is required to be able to remove value from `LabelDefinition` of type `Enum`.
1. User removes enum value from LabelDefinition (`deleteEnumLabelDefinitionValues` mutation).

> **NOTE**: This is the first iteration. In future we can introduce cascade deleting + setting up "DEFAULT" label where there are no other values for `scenarios` label.

### Adding new string array value to existing Runtime / Application label 

1. User adds array label value (`addArrayLabelValues` mutation) for the particular Runtime / Application label.
1. The values types are validated with the related LabelDefinition type.

### Removing string array values for existing Runtime / Application label 

1. User removes array label values (`deleteArrayLabelValues` mutation) for the particular Runtime / Application label.

### Removing LabelDefinition when it is used in at least one label

1. User removes label that uses the LabelDefinition which is going to be deleted (`removeLabel` mutation). It is required to be able to remove `LabelDefinition`.
1. User removes LabelDefinition (`removeLabelDefinition` mutation).
