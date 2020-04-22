# Automatic Scenario Assignment

Automatic Scenario Assignment (ASA) feature allows you to define a condition that specifies when a Scenario is automatically assigned to a Runtime. For example, using this feature, you can specify a label that adds a given Scenario to each Runtime created by the given user, company or any other entity specified in the label. See the diagram:

![](./assets/automatic-scenario-assign.svg) 

1. Administrator defines Scenarios.
2. Administrator defines conditions to label Runtimes using Automatic Scenario Assignment. 
3. User registers a Runtime that matches the conditions specified in the ASA.
4. Runtime is automatically assigned to the matching Scenario. 

## API

Automatic Scenario Assignment is defined in the following way:
```graphql
type AutomaticScenarioAssignment {
   scenarioName: String!
   selector: Label!
}

type Label {
   key: String!
   value: Any!
}

```

A condition is defined as a label selector in the **selector** field. If a Runtime is labeled with a label that matches the value of the **selector** parameter, the Runtime is assigned to the given Scenario.

### Mutations

Director API contains the following mutations for managing Automatic Scenario Assignments:
```graphql
   createAutomaticScenarioAssignment(in: AutomaticScenarioAssignmentSetInput!): AutomaticScenarioAssignment 
   deleteAutomaticScenarioAssignmentForScenario(scenarioName: String!): AutomaticScenarioAssignment 
   deleteAutomaticScenarioAssignmentsForSelector(selector: LabelSelectorInput!): [AutomaticScenarioAssignment!]! 
```
When creating an assignment, you must fulfill the following conditions:
- For a given Scenario, at most one Assignment exists
- A given Scenario exists
- Selector value type is a string

### Queries

Director API contains queries that allow you to fetch all assignments, fetch assignments for the given Scenario, and fetch assignments for the given label selector:
```graphql
   automaticScenarioAssignments(first: Int = 100, after: PageCursor): AutomaticScenarioAssignmentPage 
   automaticScenarioAssignmentForScenario(scenarioName: String!): AutomaticScenarioAssignment 
   automaticScenarioAssignmentsForSelector(selector: LabelSelectorInput!): [AutomaticScenarioAssignment!]! 
```

## Assign Runtime to Scenario

You can assign a Runtime to a Scenario either by:
- Creating ASA and then creating a Runtime that matches
- Creating ASA when the Runtime already exists


### Create ASA and Runtime

1. Create or update the `scenarios` LabelDefinition and specify the following scenarios: 
* DEFAULT
* MARKETING
* WAREHOUSE

```graphql
mutation  {
  createLabelDefinition(in: {key: "scenarios", schema: "{\"items\":{\"enum\":[\"DEFAULT\",\"MARKETING\",\"WAREHOUSE\"],\"maxLength\":128,\"pattern\":\"^[A-Za-z0-9]([-_A-Za-z0-9\\\\s]*[A-Za-z0-9])$\",\"type\":\"string\"},\"minItems\":1,\"type\":\"array\",\"uniqueItems\":true}"}) {
    key
    schema
  }
}
```

2. Create an assignment with a condition that Runtimes created by a `WAREHOUSE` Administrator are assigned to the `WAREHOUSE` Scenario:
```graphql
mutation  {
  createAutomaticScenarioAssignment(in: {scenarioName: "WAREHOUSE", selector: {key: "owner", value: "warehouse-admin@mycompany.com"}}) {
    scenarioName
    selector {
      key
      value
    }
  }
}
```

3. Register a Runtime with a label that indicates that it was created by the `WAREHOUSE` Administrator.
```graphql
mutation  {
  registerRuntime(in:{name: "warehouse-runtime-1", labels:{owner:"warehouse-admin@mycompany.com"}}) {
    name
    labels
  }
}
```

Automatic Scenario Assignment assigns the Runtime to the `WAREHOUSE` Scenario: 
```json
{
  "data": {
    "registerRuntime": {
      "name": "warehouse-runtime-1",
      "labels": {
        "owner": "warehouse-admin@mycompany.com",
        "scenarios": [
          "WAREHOUSE"
        ]
      }
    }
  }
}
```

4. Remove Automatic Scenario Assignment:
```graphql
mutation  {
  deleteAutomaticScenarioAssignmentForScenario(scenarioName: "WAREHOUSE") {
    scenarioName
  }
}
```

5. Fetch information about previously created Runtime, for example by listing all Runtimes
```graphql
query  {
  runtimes {
    data {
      id
      name
      labels
    }
  }
}
```

Runtime is unassigned from the `WAREHOUSE` Scenario:
```json
{
  "data": {
    "runtimes": {
      "data": [
        {
          "id": "b5e1bf58-e8bb-4bde-a9c0-87650b0909c0",
          "name": "warehouse-runtime-1",
          "labels": {
            "owner": "warehouse-admin@mycompany.com"
          }
        }
      ]
    }
  }
}
```

>**NOTE:** The same situation occurs if you modify or remove the `createdBy` label for the Runtime.


### Create ASA when Runtime exists

You can also assign a Runtimes to a given Scenario using ASA when the Runtime already exists. If there is a Runtime that matches a new assignment, it is automatically assigned to the Scenario.

1. Create a Runtime with the `owner:marketing-admin@mycompany.com` label:

```graphql
mutation  {
  registerRuntime(in:{name: "marketing-runtime-1", labels:{owner:"marketing-admin@mycompany.com"}}) {
    name
    labels
  }
}

```

2. Create an assignment:
```graphql
mutation  {
  createAutomaticScenarioAssignment(in: {scenarioName: "MARKETING", selector: {key: "owner", value: "marketing-admin@mycompany.com"}}) {
    scenarioName
    selector {
      key
      value
    }
  }
}
```

3. Check if the Runtime is assigned to the `MARKETING` Scenario:
```graphql
query  {
  runtimes {
    data {
      id
      name
      labels
    }
  }
}
```

In the response you can see that your Runtime is assigned to the `MARKETING` Scenario:

```
{
  "data": {
    "runtimes": {
      "data": [
        {
          "id": "5b55bc5a-5a4d-443c-b519-7f5dcba2e6de",
          "name": "marketing-runtime-1",
          "labels": {
            "owner": "marketing-admin@mycompany.com",
            "scenarios": [
              "MARKETING"
            ]
          }
        }
        ...
      ]
    }
  }
}
```
