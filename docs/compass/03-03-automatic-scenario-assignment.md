# Automatic Scenario Assignment

Automatic Scenario Assignment (ASA) is a feature that allows you to define an external subaccount tenant ID, in which all Runtimes are assigned to the given formation, assuming that the formation is in a parent tenant of type `account`. The following diagram outlines the process and refers to the formation as a Scenario.

![](./assets/automatic-scenario-assign.svg) 

1. An administrator creates a formation in a tenant of type `account`.
2. The administrator assigns a subaccount X to the newly created formation. 
3. A user registers a Runtime in subaccount X.
4. The Runtime is automatically assigned to the matching formation. 

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

A tenant-matching condition is defined as a label selector in the `selector` field. This behavior is intended for backwards compatibility. Yet, the assignment is based on the tenant, in which the Runtime is created and not on the labels of the Runtime. The label key is validated to be a `global_subaccount_id`. It is also validated that the subaccount tenant is a child of the tenant, in which the formation and the ASA resides. Then, if a Runtime is created in the `global_subaccount_id` tenant, that Runtime is automatically assigned to the given formation.

### Mutations

Director API cannot create or delete ASAs directly. Instead, the ASAs are created implicitly when assigning a tenant of type `subaccount` to the formation, and then, the ASAs are deleted when the tenant is unassigned from the formation:
```graphql
   assignFormation(objectID: ID!, objectType: FormationObjectType!, formation: FormationInput!): Formation!
   unassignFormation(objectID: ID!, objectType: FormationObjectType!, formation: FormationInput!): Formation! 
```

### Queries

Director API contains queries that allow you to fetch all assignments, or fetch assignments for the given formation, or fetch assignments for the given label selector:
```graphql
   automaticScenarioAssignments(first: Int = 100, after: PageCursor): AutomaticScenarioAssignmentPage 
   automaticScenarioAssignmentForScenario(scenarioName: String!): AutomaticScenarioAssignment 
   automaticScenarioAssignmentsForSelector(selector: LabelSelectorInput!): [AutomaticScenarioAssignment!]! 
```

## Assign Runtime to Formation

You can assign a Runtime to a formation as follows:
1. Create a formation.
2. Assign Runtime to the formation using `assignFromation` mutation with `objectType` - `RUNTIME`

### Create ASA and Runtime

1. Create a formation with name `WAREHOUSE`: 

```graphql
mutation {
   result: createFormation(formation: { name: "WAREHOUSE" }) {
      id
      name
      formationTemplateId
   }
}
```

2. Assign the WEAREHOUSE Administrators subaccount to the formation. This implicitly creates an ASA with a condition that the Runtimes created by a `WAREHOUSE` Administrator are assigned to the `WAREHOUSE` formation:
```graphql
mutation {
   result: assignFormation(
      objectID: "0ccd19fd-671e-4024-8b0f-887bb7e4ed4f"
      objectType: TENANT
      formation: { name: "WAREHOUSE" }
   ) {
      id
      name
      formationTemplateId
   }
}
```

3. Register a Runtime into the `0ccd19fd-671e-4024-8b0f-887bb7e4ed4f` subaccount tenant:
   Create a Runtime in the `0ccd19fd-671e-4024-8b0f-887bb7e4ed4f` tenant:
    - Run the request in the context of the wanted `subaccount` tenant.
    - Run the request in the context of the parent tenant, and label the Runtime with the wanted `subaccount` tenant. The flow is intended for backwards compatibility, and results in Runtime registration within the tenant provided as a label and not the tenant from the request context:
      ```graphql
        mutation  {
            registerRuntime(in:{name: "warehouse-runtime-1", labels:{global_subaccount_id:"0ccd19fd-671e-4024-8b0f-887bb7e4ed4f"}}) {
                name
                labels
            }
        }
      ```

ASA assigns the Runtime to the `WAREHOUSE` formation: 
```json
{
  "data": {
    "registerRuntime": {
      "name": "warehouse-runtime-1",
      "labels": {
        "global_subaccount_id": "0ccd19fd-671e-4024-8b0f-887bb7e4ed4f",
        "scenarios": [
          "WAREHOUSE"
        ]
      }
    }
  }
}
```

4. Unassign the subaccount from the formation. This results in removal of the ASA.
```graphql
mutation {
   result: unassignFormation(
      objectID: "0ccd19fd-671e-4024-8b0f-887bb7e4ed4f"
      objectType: TENANT
      formation: { name: "WAREHOUSE" }
   ) {
      id
      name
      formationTemplateId
   }
}

```

5. Fetch information about previously created Runtime, for example, by listing all Runtimes.
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

Runtime is unassigned from the `WAREHOUSE` formation.
```json
{
  "data": {
    "runtimes": {
      "data": [
        {
          "id": "b5e1bf58-e8bb-4bde-a9c0-87650b0909c0",
          "name": "warehouse-runtime-1",
          "labels": {
            "global_subaccount_id": "0ccd19fd-671e-4024-8b0f-887bb7e4ed4f"
          }
        }
      ]
    }
  }
}
```

### Create ASA when Runtime exists

You can also assign a Runtime to a given formation using ASA when the Runtime already exists. If there is a Runtime that matches a new assignment, that is, it is in the wanted `subaccount` tenant, then, it is automatically assigned to the formation.
All of the following requests are done in the context of a tenant of type `account`, which is a parent of the given `subaccount` tenant.

1. Create a Runtime in the `0ccd19fd-671e-4024-8b0f-887bb7e4ed4f` tenant:
    - Run the request in the context of the wanted `subaccount` tenant
    - Run the request in the context of the parent tenant, and label the runtime with the wanted `subaccount` tenant:
      ```graphql
        mutation  {
            registerRuntime(in:{name: "warehouse-runtime-1", labels:{global_subaccount_id:"0ccd19fd-671e-4024-8b0f-887bb7e4ed4f"}}) {
                name
                labels
            }
        }
      ```

2. Create a formation with name `MARKETING`:

```graphql
mutation {
   result: createFormation(formation: { name: "MARKETING" }) {
      id
      name
      formationTemplateId
   }
}
```

3. Assign subaccount to the formation. This action implicitly creates the ASA.
```graphql
mutation {
   result: assignFormation(
      objectID: "0ccd19fd-671e-4024-8b0f-887bb7e4ed4f"
      objectType: TENANT
      formation: { name: "MARKETING" }
   ) {
      id
      name
      formationTemplateId
   }
}
```

4. Check if the Runtime is assigned to the `MARKETING` formation.
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

In the response you can see that your Runtime is assigned to the `MARKETING` formation.

```
{
  "data": {
    "runtimes": {
      "data": [
        {
          "id": "5b55bc5a-5a4d-443c-b519-7f5dcba2e6de",
          "name": "marketing-runtime-1",
          "labels": {
            "global_subaccount_id": "0ccd19fd-671e-4024-8b0f-887bb7e4ed4f",
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
