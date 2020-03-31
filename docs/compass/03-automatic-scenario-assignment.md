# Automatic Scenario Assignment

In order to connect and group your Applications and Runtimes, assign them to the same scenario.
Automatic Scenario Assignment feature, allows to define condition, when a Scenario is automatically assigned to the Runtime.


## API

AutomaticScenarioAssignment is defined in the following way:
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

If the Runtime is labelled with a label that matches label's selector, the Runtime is assigned to the given Scenario.


Director API contains the following mutations for managing Automatic Scenario Assignments.
```graphql
	setAutomaticScenarioAssignment(in: AutomaticScenarioAssignmentSetInput!): AutomaticScenarioAssignment 
	deleteAutomaticScenarioAssignmentForScenario(scenarioName: String!): AutomaticScenarioAssignment 
	deleteAutomaticScenarioAssignmentForSelector(selector: LabelSelectorInput!): [AutomaticScenarioAssignment!]! 
}
```
When creating Assignment, the following conditions are checked:
* for given scenario, at most one Assignment exist
* scenario already exists; it is enumarated in the scenarios Label Definition
* type of selector's value is string

There are queries that allow to fetch all asssignments, assignments for given Scenario and for given label selector.
```graphql
	automaticScenarioAssignments(first: Int = 100, after: PageCursor): AutomaticScenarioAssignmentPage @hasScopes(path: "graphql.query.automaticScenarioAssignments")
	automaticScenarioAssignmentForScenario(scenarioName: String!): AutomaticScenarioAssignment @hasScopes(path: "graphql.query.automaticScenarioAssignmentForScenario")
	automaticScenarioAssignmentForSelector(selector: LabelSelectorInput!): [AutomaticScenarioAssignment!]! @hasScopes(path: "graphql.query.automaticScenarioAssignmentForSelector")
```

## Example
Let assume a situation, that Integration System that is responsible for registering Runtimes, label every runtime with 
an information about a user who triggered provisioning a Runtime. 
Then, you can define an Assignment, that every Runtime provisioned by a given person belongs to specific Scenario.

1. Create Scenarios Label Definition and specify the following scenarios: 
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

2. Create an Assignment, that runtimes created by a Warehouse administrator are assigned to `WAREHOUSE` scenario.
```graphql
mutation  {
  setAutomaticScenarioAssignment(in: {scenarioName: "WAREHOUSE", selector: {key: "owner", value: "warehouse-admin@mycompany.com"}}) {
    scenarioName
    selector {
      key
      value
    }
  }
}
```
4. Register a Runtime, with a label that indicates that it was created by the Warehouse administrator.
```graphql
mutation  {
  registerRuntime(in:{name: "warehouse-runtime-1", labels:{owner:"warehouse-admin@mycompany.com"}}) {
    name
    labels
  }
}
```
TODO: As you can see from the output, Runtime is assigned to the `WAREHOUSE` scenario, thanks to the previously
defined Automatic Scenario Assignment. 
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

6. Remove AutomaticScenarioAssignment.
```graphql
mutation  {
  deleteAutomaticScenarioAssignmentForScenario(scenarioName: "WAREHOUSE") {
    scenarioName
  }
}
```

7. Fetch information about previously created Runtime, for example by listing all Runtimes
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

As you can see, Runtime was unassigned from `WAREHOUSE` scenario.
The same will happen if you modify label `createdBy` for the Runtime.

Assignment Runtime to scenarios occurs not only when we label Runtime, but also when we define Assignment.
If there is a runtime, that matches newly created Assignment, it will be also automatically assigned to the Scenario.

1. Create runtime with label `owner`:`marketing-admin@mycompany.com`

```graphql
mutation  {
  registerRuntime(in:{name: "marketing-runtime-1", labels:{owner:"marketing-admin@mycompany.com"}}) {
    name
    labels
  }
}

```
2. Create assignment:
```graphql
mutation  {
  setAutomaticScenarioAssignment(in: {scenarioName: "MARKETIGN", selector: {key: "owner", value: "marketing-admin@mycompany.com"}}) {
    scenarioName
    selector {
      key
      value
    }
  }
}
```

3. Check that the Runitme is assigned to the `MARKETING` scenario.
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
