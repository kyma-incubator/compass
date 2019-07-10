# Labeling

The following documents describes the labeling feature.


## Use cases

- Validating label values against JSON schema / type
- Reading all label keys for tenant


## Terms
In this document there are few terms, which are used:

**Label**

Label is a tag in a form of `key:value`, which can be assigned to an Application or Runtime. Label references **LabelDefinition** which defines the shape for value.
It holds actual label key and value. A single Label can be assigned to Runtime or Application. 

Label is tenant specific. 

```go
type Label struct {
    Name string // Key 
    Value interface{} // Value
    Tenant string
    DefinitionID int // LabelDefinition reference
    ApplicationID *int
    Runtime ID *int
}
```

**LabelDefinition** 

Describes what's the type of value for a label key. Validation will be done against the LabelDefinition JSON Schema or type.
Label definition contains `name`, which is unique and represents a label `key`.

LabelDefinition is not reusable between two labels with different keys. LabelDefinition is tenant-specific.

```go
type LabelDefinition struct {
    Name string // Key 
    Tenant string
    Data map[string]interface{}
    Type LabelDefinitionType
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
    NumberLabelDefinition LabelDefinitionType = "Number"
    StringArrayLabelDefinition LabelDefinitionType = "StringArray"
    EnumLabelDefinition LabelDefinitionType = "Enum"
)
```

**Scenario**

It is a Label with key `scenarios`, which references LabelDefinition of type `Enum`. Based on the `scenarios` Label, Applications and Runtimes are connected. Runtime Agent does query for all Scenarios assigned to Application.

## Workflows

## Storage

