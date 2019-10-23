# ApplicationClass and ApplicationTemplate

Managing [Applications](./../terminology.md#Application) can be passed on some external system, to remove the burden of integration with Compass from Applications.
Because of that [IntegrationSystems](./../terminology.md#Integration-System) were introduced.

To simplify creation of Applications, Director API is extended with Application Template. 
An ApplicationTemplate is reusable input for creating Applications that can be customized with placeholders provided 
during Application creation.

## Manage Application by Integration System
Managing Integrations System is described in [a separate document](./integration-systems.md).
Integration System can be uniquely identified by it's name. To create an Application that is managed by a Integration System,
specify `integrationSystemName` in the ApplicationCreateInput. 

```graphql
input ApplicationCreateInput {
    name: String!
    description: String
    labels: Labels
    webhooks: [WebhookInput!]
    healthCheckURL: String
    apis: [APIDefinitionInput!]
    eventAPIs: [EventAPIDefinitionInput!]
    documents: [DocumentInput!]
    integrationSystemName: String #optional
    labels: [LabelInput!]
}
```
IntegrationSystemID is an optional. 
### Example
In this example, Integration System is required and then it configures Application. 
1. Register Integration System
```graphql
mutation {
    createIntegrationSystem(in: {name: "simpleIntegrationSystem"} ) {
      id
      name
    }
}

```
2. Create Application with provided `integrationSystemName`
```graphql
mutation {
    createApplication(in:{name:"simpleApplication", integrationSystemName:"simpleIntegrationSystem"}) {
        id
        name
        integrationSystemName
        labels
    }
}
```
Compass add labels with integrationSystemName for just created Application, so output of the previous mutation is the following:
```
{
  "data": {
    "createApplication": {
      "id": "d046590f-934f-411f-91e2-d446b404a2a2",
      "name": "simpleApplication",
      "integrationSystemName": "simpleIntegrationSystem",
      "labels": {
        "scenarios":["DEFULT"],
        "integrationSystemName":"simpleIntegrationSystem",
      },
      
    }
  }
}
```

Thanks to that, no store

4. IntegrationSystem query for all related Applications by specifying label `integrationSystemName` and reconciles 
their state.

## Integration System supporting many Application types
In the previous example, Integration System was limited to support only one type of Application.
In case it supports many types of Application Types, information about which Application to provision should be
included in the Application Labels. To standaradize whole process, Application Template can be used.


Managing ApplicationTemplates:
```graphql
input ApplicationTemplateInput {
    name: String!
    description: String

    applicationInput: ApplicationCreateInput!
    placeholders:       [PlaceholderDefinitionInput!]

}

input PlaceholderDefinitionInput {
    Name            String!
    Description     String
}

type Mutation {
    createApplicationTemplate(in: ApplicationTemplateInput!): ApplicationTemplate!
    updateApplicationTemplate(id: ID!, in: ApplicationTemplateInput!): ApplicationTemplate!
    deleteApplicationTemplate(id: ID!): ApplicationTemplate!

}
type Query {
    applicationTemplates(first: Int = 100, after: PageCursor): ApplicationTemplatePage!
    applicationTemplate(id: ID!): ApplicationTemplate
} 
```

ApplicationTemplate defined ApplicationInput used to create Application. ApplicationInput can contains variable part - placeholders.
Placeholders are defined in `placeholders` fields. Thanks to that, clear definition what required input parameters is defined.


To create Application from template, `createApplicationFromTemplate` mutation is used:

```graphql

createApplicationFromTemplate(templateName: String!, values: [TemplateValueInput]): Application!

```

## Examples

### Create simple Application not managed by Integration System

In this case creating Application the same as before introducing Application Templates.
```graphql
mutation {
    createApplication(in: {name: "simpleApplication"}) {
        id 
        name
    }
}
```

### Create Application managed by Integration System
```graphql
mutation {
    createApplication(in: {name: "simpleApplication", integrationSystemName: "integrationSystem"}) {
        id
        name
        labels
    }
}
```

### Create simple Application from Application Template
1. Create an Application Template
```graphql
mutation  {

}

```
2. Create an Application From Template
```graphql
mutation  {

}

```

### Create Application managed by Integration System from Application Template
1. Create an Application Templates by Integration System

2. Create an Appliction from Application Template

    
 
## Later
- everything is labelled, including integration system
- application input can define icon
- versioning of application class


describe UI:
- 2 links: create Application and create Application from Template

- security